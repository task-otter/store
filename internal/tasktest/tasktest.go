package tasktest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

type Taskfile struct {
	Version string          `yaml:"version"`
	Vars    map[string]any  `yaml:"vars"`
	Tasks   map[string]Task `yaml:"tasks"`
}

type Task struct {
	Desc          string   `yaml:"desc"`
	Summary       string   `yaml:"summary"`
	Internal      bool     `yaml:"internal"`
	Run           string   `yaml:"run"`
	Set           []string `yaml:"set"`
	Preconditions any      `yaml:"preconditions"`
	Cmds          any      `yaml:"cmds"`
	Deps          any      `yaml:"deps"`
}

type testT interface {
	Helper()
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	TempDir() string
}

var (
	getWorkingDir = os.Getwd
	taskBinary    = "task"
	taskTimeout   = time.Minute
)

// TestMain builds the shared read-only dry-run environment once for the whole
// test binary, runs the package's tests, and tears the environment down. Test
// packages that exercise task modules should delegate to it:
//
//	func TestMain(m *testing.M) { tasktest.TestMain(m) }
//
// Delegating is optional — the environment is built lazily on first use — but
// without it the temporary tree is left behind for the OS to reclaim.
func TestMain(m *testing.M) {
	code := m.Run()
	cleanupDryRunEnv()
	os.Exit(code)
}

// AssertModule checks a module's README and Taskfile against its declared
// public surface. It does not fork the task CLI: the Taskfile is parsed here,
// and CLI loadability is covered once per family by AssertTaskCliCanLoad.
func AssertModule(t testT, module string, expectedTasks, expectedVars []string) {
	t.Helper()

	assertReadme(t, module, expectedTasks)
	assertTaskfile(t, module, expectedTasks, expectedVars)
}

// AssertTaskCliCanLoad verifies the task CLI can parse a module's Taskfile.
// This forks the CLI, so it is deduplicated per tool family within a test
// binary — every variant of a family shares the same generated structure.
func AssertTaskCliCanLoad(t testT, module string) {
	t.Helper()

	family, _, _ := strings.Cut(module, "/")
	once, _ := cliLoadOnce.LoadOrStore(family, &sync.Once{})
	once.(*sync.Once).Do(func() {
		assertTaskCliCanLoad(t, module)
	})
}

var cliLoadOnce sync.Map // family -> *sync.Once

func AssertDryRunContains(t testT, module string, args []string, tokens ...string) {
	t.Helper()

	output := DryRun(t, module, args...)
	for _, token := range tokens {
		if !strings.Contains(output, token) {
			t.Fatalf("dry-run output for %s missing %q\nargs: %v\noutput:\n%s", module, token, args, output)
		}
	}
}

// AssertInstallDryRun verifies install dry-run output. When the tool is already
// on PATH the install task is skipped ("up to date"); otherwise downloadTokens
// must appear in the output.
func AssertInstallDryRun(t testT, module, toolName string, downloadTokens ...string) {
	t.Helper()

	output := DryRun(t, module, "install")
	if strings.Contains(output, "up to date") {
		if !strings.Contains(output, toolName) {
			t.Fatalf("dry-run output for %s install skipped but missing %q\noutput:\n%s", module, toolName, output)
		}
		return
	}

	for _, token := range downloadTokens {
		if !strings.Contains(output, token) {
			t.Fatalf("dry-run output for %s install missing %q\noutput:\n%s", module, token, output)
		}
	}
}

func RootDryRun(t testT, args ...string) string {
	t.Helper()

	allArgs := append([]string{"--dry", "--yes", "--verbose"}, args...)
	return runTask(t, allArgs...)
}

func DryRun(t testT, module string, args ...string) string {
	t.Helper()

	projectDir, env := sharedDryRunEnv(t)
	allArgs := append([]string{"--taskfile", taskfilePath(t, module), "--dry", "--yes", "--verbose"}, args...)

	ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, taskBinary, allArgs...)
	cmd.Dir = projectDir
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("task command timed out: task %s", strings.Join(allArgs, " "))
	}
	if err != nil {
		t.Fatalf("task command failed: task %s\nerror: %v\noutput:\n%s", strings.Join(allArgs, " "), err, string(output))
	}

	return string(output)
}

var (
	dryRunEnvOnce sync.Once
	dryRunRoot    string
	dryRunProject string
	dryRunEnv     []string
	dryRunEnvErr  error
)

// sharedDryRunEnv returns the process-wide dry-run project directory and
// environment, building the stub tree on first use. Dry runs never write to it,
// so every test — including parallel ones — shares the same read-only tree.
func sharedDryRunEnv(t testT) (projectDir string, env []string) {
	t.Helper()

	dryRunEnvOnce.Do(func() {
		dryRunRoot, dryRunEnvErr = os.MkdirTemp("", "tasktest-")
		if dryRunEnvErr != nil {
			dryRunEnvErr = fmt.Errorf("create stub root: %w", dryRunEnvErr)
			return
		}
		home := filepath.Join(dryRunRoot, "home")
		dryRunProject = filepath.Join(dryRunRoot, "project")
		dryRunEnv, dryRunEnvErr = buildDryRunEnv(home, dryRunProject)
	})
	if dryRunEnvErr != nil {
		t.Fatalf("build dry-run environment: %v", dryRunEnvErr)
	}

	return dryRunProject, dryRunEnv
}

// cleanupDryRunEnv removes the shared stub tree. Safe to call when nothing was
// ever built.
func cleanupDryRunEnv() {
	if dryRunRoot != "" {
		os.RemoveAll(dryRunRoot)
	}
}

// buildDryRunEnv populates a project directory and isolated environment for
// dry-run tests. It stubs common JS package managers, node version managers,
// and linting/formatting tools so that _install-if-missing skips installation
// and package-manager preconditions pass without real tools being present.
func buildDryRunEnv(home, projectDir string) (env []string, err error) {
	binDir := filepath.Join(projectDir, ".stub-bin")

	for _, dir := range []string{
		binDir,
		filepath.Join(home, ".bun", "bin"),
		filepath.Join(home, ".local", "share", "fnm"),
		filepath.Join(home, ".nvm"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create stub dir %s: %w", dir, err)
		}
	}

	const stub = "#!/usr/bin/env bash\nexit 0\n"

	for _, name := range []string{
		"fnm", "node", "npm", "npx", "pnpm", "yarn", "bun", "corepack", "nvm",
		"prettier", "eslint", "biome", "stylelint", "knip", "depcheck", "bru",
	} {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte(stub), 0755); err != nil {
			return nil, fmt.Errorf("write stub %s: %w", name, err)
		}
	}

	// bun:_bun:unix checks: test -f "$HOME/.bun/bin/bun"
	if err := os.WriteFile(filepath.Join(home, ".bun", "bin", "bun"), []byte(stub), 0755); err != nil {
		return nil, fmt.Errorf("write bun file stub: %w", err)
	}
	// npm/pnpm/yarn:_*:unix checks FNM_INSTALL_DIR ($HOME/.local/share/fnm/fnm)
	if err := os.WriteFile(filepath.Join(home, ".local", "share", "fnm", "fnm"), []byte(stub), 0755); err != nil {
		return nil, fmt.Errorf("write fnm file stub: %w", err)
	}
	// npm/pnpm/yarn nvm-stack checks $HOME/.nvm/nvm.sh
	if err := os.WriteFile(filepath.Join(home, ".nvm", "nvm.sh"), []byte("# nvm stub\n"), 0644); err != nil {
		return nil, fmt.Errorf("write nvm.sh stub: %w", err)
	}
	// npm/pnpm/yarn:_*:unix checks for package.json in USER_WORKING_DIR
	if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte("{}\n"), 0644); err != nil {
		return nil, fmt.Errorf("write package.json: %w", err)
	}

	env = os.Environ()
	env = dryRunSetEnv(env, "HOME", home)
	env = dryRunSetEnv(env, "PATH", binDir+":"+dryRunGetEnv(env, "PATH"))
	env = dryRunSetEnv(env, "CI", "true")
	env = dryRunSetEnv(env, "NO_COLOR", "1")
	env = dryRunSetEnv(env, "TASK_ASSUME_YES", "true")

	return env, nil
}

func dryRunSetEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, item := range env {
		if strings.HasPrefix(item, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

func dryRunGetEnv(env []string, key string) string {
	prefix := key + "="
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			return strings.TrimPrefix(item, prefix)
		}
	}
	return ""
}

func LoadTaskfile(t testT, module string) Taskfile {
	t.Helper()

	content, err := os.ReadFile(taskfilePath(t, module))
	if err != nil {
		t.Fatalf("read %s Taskfile: %v", module, err)
	}

	if strings.Contains(string(content), "\r\n") {
		t.Fatalf("%s Taskfile must use LF line endings", module)
	}
	if strings.TrimRight(string(content), " \t\r\n") != strings.TrimRight(string(content), "\r\n") {
		t.Fatalf("%s Taskfile has trailing whitespace", module)
	}

	var tf Taskfile
	if err := yaml.Unmarshal(content, &tf); err != nil {
		t.Fatalf("parse %s Taskfile: %v", module, err)
	}

	return tf
}

func RepoRoot(t testT) string {
	t.Helper()

	wd, err := getWorkingDir()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("could not find repository root with go.mod")
		}
		wd = parent
	}
}

func assertReadme(t testT, module string, expectedTasks []string) {
	t.Helper()

	// Nested tool families (e.g. "biome/node/fnm/npm") share a single README at
	// the family root; flat modules keep their own. Resolve accordingly.
	readmeModule := module
	if index := strings.IndexByte(readmeModule, '/'); index >= 0 {
		readmeModule = readmeModule[:index]
	}
	path := filepath.Join(moduleDir(t, readmeModule), "README.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s must have README.md: %v", module, err)
	}

	text := string(content)
	if strings.TrimSpace(text) == "" {
		t.Fatalf("%s README.md is empty", module)
	}
	if !strings.Contains(text, "## Public Tasks") {
		t.Fatalf("%s README.md must document public tasks", module)
	}
	for _, task := range expectedTasks {
		if !strings.Contains(text, "`"+task+"`") {
			t.Fatalf("%s README.md does not mention public task %q", module, task)
		}
	}
}

func assertTaskfile(t testT, module string, expectedTasks, expectedVars []string) {
	t.Helper()

	tf := LoadTaskfile(t, module)
	if tf.Version != "3" && !strings.HasPrefix(tf.Version, "3.") {
		t.Fatalf("%s Taskfile version must be 3 or 3.x, got %q", module, tf.Version)
	}
	if len(tf.Tasks) == 0 {
		t.Fatalf("%s Taskfile must define tasks", module)
	}

	actualTasks := publicTaskNames(tf)
	expectedTasks = sortedCopy(expectedTasks)
	if !slices.Equal(expectedTasks, actualTasks) {
		t.Fatalf("%s public task drift\nexpected: %v\nactual:   %v", module, expectedTasks, actualTasks)
	}

	for _, name := range actualTasks {
		task := tf.Tasks[name]
		if len(strings.TrimSpace(task.Desc)) < 12 {
			t.Fatalf("%s task %q desc is missing or too short: %q", module, name, task.Desc)
		}
		if task.Cmds == nil && task.Deps == nil {
			t.Fatalf("%s task %q must define cmds or deps", module, name)
		}
	}

	for _, name := range expectedVars {
		if _, ok := tf.Vars[name]; !ok {
			t.Fatalf("%s Taskfile vars missing %q", module, name)
		}
	}
}

func assertTaskCliCanLoad(t testT, module string) {
	t.Helper()

	output, _ := runTaskOutput(t, "--taskfile", taskfilePath(t, module), "--list-all", "--json")
	var payload any
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("%s task --list-all --json produced invalid JSON:\n%s\nerror: %v", module, output, err)
	}
}

func publicTaskNames(tf Taskfile) []string {
	var names []string
	for name, task := range tf.Tasks {
		if name == "default" || task.Internal {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedCopy(values []string) []string {
	clone := slices.Clone(values)
	sort.Strings(clone)
	return clone
}

func taskfilePath(t testT, module string) string {
	t.Helper()
	return filepath.Join(moduleDir(t, module), "Taskfile.yml")
}

func moduleDir(t testT, module string) string {
	t.Helper()
	return filepath.Join(RepoRoot(t), "taskfiles", module)
}

func runTask(t testT, args ...string) string {
	t.Helper()

	output, err := runTaskOutput(t, args...)
	if err != nil {
		t.Fatalf("task command failed: task %s\nerror: %v\noutput:\n%s", strings.Join(args, " "), err, output)
	}

	return output
}

func runTaskOutput(t testT, args ...string) (string, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, taskBinary, args...)
	cmd.Dir = RepoRoot(t)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("task command timed out: task %s", strings.Join(args, " "))
	}

	return string(output), err
}
