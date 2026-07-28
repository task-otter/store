package tasktest

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
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
	Vars          any      `yaml:"vars"`
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
