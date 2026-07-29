// Package tasktest provides shared assertions for Taskfile module tests.
package tasktest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const minTaskDescriptionLength = 12

// Taskfile contains the top-level fields read from a Taskfile.yml.
type Taskfile struct {
	Version string          `yaml:"version"`
	Vars    map[string]any  `yaml:"vars"`
	Tasks   map[string]Task `yaml:"tasks"`
}

// Task contains the task fields validated by the shared test helpers.
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
	Status        any      `yaml:"status"`
}

type testT interface {
	Helper()
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	TempDir() string
}

type taskCommandSettings struct {
	binary  string
	timeout time.Duration
}

func workingDir() (string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	return workingDirectory, nil
}

func currentTaskCommandSettings() taskCommandSettings {
	return taskCommandSettings{binary: "task", timeout: time.Minute}
}

// AssertModule checks a module's README and Taskfile against its declared
// public surface. It does not fork the task CLI: the Taskfile is parsed here,
// and CLI loadability is covered once per family by AssertTaskCliCanLoad.
func AssertModule(tester testT, module string, expectedTasks, expectedVars []string) {
	tester.Helper()

	assertReadme(tester, module, expectedTasks)
	assertTaskfile(tester, module, expectedTasks, expectedVars)
}

// AssertTaskCliCanLoad verifies the task CLI can parse a module's Taskfile.
func AssertTaskCliCanLoad(tester testT, module string) {
	tester.Helper()

	assertTaskCliCanLoad(tester, module)
}

// LoadTaskfile reads and parses the Taskfile for module.
func LoadTaskfile(tester testT, module string) Taskfile {
	tester.Helper()

	content, err := readFile(taskfilePath(tester, module))
	if err != nil {
		tester.Fatalf("read %s Taskfile: %v", module, err)
	}

	if strings.Contains(string(content), "\r\n") {
		tester.Fatalf("%s Taskfile must use LF line endings", module)
	}

	if strings.TrimRight(string(content), " \t\r\n") != strings.TrimRight(string(content), "\r\n") {
		tester.Fatalf("%s Taskfile has trailing whitespace", module)
	}

	var taskfile Taskfile

	err = yaml.Unmarshal(content, &taskfile)
	if err != nil {
		tester.Fatalf("parse %s Taskfile: %v", module, err)
	}

	return taskfile
}

// RepoRoot walks upward from the working directory to find the repository root.
func RepoRoot(tester testT) string {
	tester.Helper()

	workingDirectory, err := workingDir()
	if err != nil {
		tester.Fatalf("get working directory: %v", err)
	}

	for {
		_, err = os.Stat(filepath.Join(workingDirectory, "go.mod"))
		if err == nil {
			return workingDirectory
		}

		parent := filepath.Dir(workingDirectory)
		if parent == workingDirectory {
			tester.Fatal("could not find repository root with go.mod")
		}

		workingDirectory = parent
	}
}

func assertReadme(tester testT, module string, expectedTasks []string) {
	tester.Helper()

	// Nested tool families (e.g. "biome/node/fnm/npm") share a single README at
	// the family root; flat modules keep their own. Resolve accordingly.
	readmeModule := module
	if index := strings.IndexByte(readmeModule, '/'); index >= 0 {
		readmeModule = readmeModule[:index]
	}

	path := filepath.Join(moduleDir(tester, readmeModule), "README.md")

	content, err := readFile(path)
	if err != nil {
		tester.Fatalf("%s must have README.md: %v", module, err)
	}

	text := string(content)
	if strings.TrimSpace(text) == "" {
		tester.Fatalf("%s README.md is empty", module)
	}

	if !strings.Contains(text, "## Public Tasks") {
		tester.Fatalf("%s README.md must document public tasks", module)
	}

	for _, task := range expectedTasks {
		if !strings.Contains(text, "`"+task+"`") {
			tester.Fatalf("%s README.md does not mention public task %q", module, task)
		}
	}
}

func assertTaskfile(tester testT, module string, expectedTasks, expectedVars []string) {
	tester.Helper()

	taskfile := LoadTaskfile(tester, module)
	if taskfile.Version != "3" && !strings.HasPrefix(taskfile.Version, "3.") {
		tester.Fatalf("%s Taskfile version must be 3 or 3.x, got %q", module, taskfile.Version)
	}

	if len(taskfile.Tasks) == 0 {
		tester.Fatalf("%s Taskfile must define tasks", module)
	}

	actualTasks := publicTaskNames(taskfile)
	assertPublicTasks(tester, module, taskfile, sortedCopy(expectedTasks), actualTasks)
	assertExpectedVars(tester, module, taskfile, expectedVars)
}

func assertPublicTasks(
	tester testT,
	module string,
	taskfile Taskfile,
	expectedTasks, actualTasks []string,
) {
	tester.Helper()

	if !slices.Equal(expectedTasks, actualTasks) {
		tester.Fatalf(
			"%s public task drift\nexpected: %v\nactual:   %v",
			module,
			expectedTasks,
			actualTasks,
		)
	}

	for _, name := range actualTasks {
		assertPublicTask(tester, module, name, taskfile.Tasks[name])
	}
}

func assertPublicTask(tester testT, module, name string, task Task) {
	tester.Helper()

	if len(strings.TrimSpace(task.Desc)) < minTaskDescriptionLength {
		tester.Fatalf("%s task %q desc is missing or too short: %q", module, name, task.Desc)
	}

	if task.Cmds == nil && task.Deps == nil {
		tester.Fatalf("%s task %q must define cmds or deps", module, name)
	}
}

func assertExpectedVars(tester testT, module string, taskfile Taskfile, expectedVars []string) {
	tester.Helper()

	for _, name := range expectedVars {
		if _, ok := taskfile.Vars[name]; !ok {
			tester.Fatalf("%s Taskfile vars missing %q", module, name)
		}
	}
}

func assertTaskCliCanLoad(tester testT, module string) {
	tester.Helper()

	output, _ := runTaskOutput(
		tester,
		"--taskfile",
		taskfilePath(tester, module),
		"--list-all",
		"--json",
	)

	var payload any

	err := json.Unmarshal([]byte(output), &payload)
	if err != nil {
		tester.Fatalf(
			"%s task --list-all --json produced invalid JSON:\n%s\nerror: %v",
			module,
			output,
			err,
		)
	}
}

func publicTaskNames(taskfile Taskfile) []string {
	var names []string

	for name, task := range taskfile.Tasks {
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

func taskfilePath(tester testT, module string) string {
	tester.Helper()

	return filepath.Join(moduleDir(tester, module), "Taskfile.yml")
}

func moduleDir(tester testT, module string) string {
	tester.Helper()

	return filepath.Join(RepoRoot(tester), "taskfiles", module)
}

func runTaskOutput(tester testT, args ...string) (string, error) {
	tester.Helper()

	settings := currentTaskCommandSettings()

	ctx, cancel := context.WithTimeout(context.Background(), settings.timeout)
	defer cancel()

	commandContext := exec.CommandContext
	cmd := commandContext(ctx, settings.binary, args...)
	cmd.Dir = RepoRoot(tester)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		tester.Fatalf("task command timed out: task %s", strings.Join(args, " "))
	}

	return string(output), err
}

func readFile(path string) ([]byte, error) {
	clean := filepath.Clean(path)

	content, err := fs.ReadFile(os.DirFS(filepath.Dir(clean)), filepath.Base(clean))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", clean, err)
	}

	return content, nil
}
