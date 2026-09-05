// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package tasktest provides shared helpers for validating Taskfile modules.
package tasktest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	yaml "go.yaml.in/yaml/v3"
)

// Taskfile contains the top-level fields read from a Taskfile.yml.
type (
	Taskfile = struct {
		Includes map[string]any   `yaml:"includes"`
		Vars     map[string]any   `yaml:"vars"`
		Tasks    map[string]*Task `yaml:"tasks"`
		Version  string           `yaml:"version"`
	}

	// Task contains the task fields validated by the shared test helpers.
	Task = struct {
		Preconditions any      `yaml:"preconditions"`
		Cmds          any      `yaml:"cmds"`
		Deps          any      `yaml:"deps"`
		Vars          any      `yaml:"vars"`
		Status        any      `yaml:"status"`
		Desc          string   `yaml:"desc"`
		Summary       string   `yaml:"summary"`
		Run           string   `yaml:"run"`
		Set           []string `yaml:"set"`
		Internal      bool     `yaml:"internal"`
	}

	// TestingT is the subset of [testing.TB] used by the shared helpers.
	TestingT interface {
		Helper()
		Fatal(args ...any)
		Fatalf(format string, args ...any)
		TempDir() string
	}

	// ModuleExpectations describes the public surface a module should expose.
	ModuleExpectations = struct {
		Tasks []string
		Vars  []string
	}

	readmeCheck = struct {
		text          string
		expectedTasks []string
	}

	publicTaskCheck = struct {
		taskfile      *Taskfile
		expectedTasks []string
		actualTasks   []string
	}

	taskCheck = struct {
		task *Task
		name string
	}

	varCheck = struct {
		taskfile     *Taskfile
		expectedVars []string
	}

	taskCliLoadFailure = struct {
		err    error
		module string
		output string
	}

	taskCommandSettings = struct {
		timeout time.Duration
	}

	workingDirectoryFunc = func() (string, error)
)

const (
	minTaskDescriptionLength = 12
	emptyLength              = 0
	taskJSONFlag             = "--json"
	taskListAllFlag          = "--list-all"
	taskBinary               = "task"
	taskfileFlag             = "--taskfile"
	taskfileName             = "Taskfile.yml"
	windowsLineEnding        = "\r\n"
)

// AssertModule checks a module's README and Taskfile against its declared
// public surface. It does not fork the task CLI: the Taskfile is parsed here,
// and CLI loadability is covered once per family by AssertTaskCliCanLoad.
func AssertModule(tester TestingT, module string, expected *ModuleExpectations) {
	tester.Helper()

	assertReadme(tester, module, expected.Tasks)
	assertTaskfile(tester, module, expected)
}

// AssertTaskCliCanLoad verifies the task CLI can parse a module's Taskfile.
func AssertTaskCliCanLoad(tester TestingT, module string) {
	tester.Helper()

	assertTaskCLIParse(tester, module)
}

// LoadTaskfile reads and parses the Taskfile for module.
func LoadTaskfile(tester TestingT, module string) *Taskfile {
	tester.Helper()

	content := mustReadTaskfile(tester, module)
	validateTaskfileFormatting(tester, module, content)

	return mustParseTaskfile(tester, module, content)
}

// RepoRoot walks upward from the working directory to find the repository root.
func RepoRoot(tester TestingT) string {
	tester.Helper()

	return findRepoRootFrom(tester, os.Getwd)
}

func currentTaskCommandSettings() taskCommandSettings {
	return taskCommandSettings{timeout: time.Minute}
}

func mustReadTaskfile(tester TestingT, module string) []byte {
	content, err := readFile(taskfilePath(tester, module))
	if err != nil {
		tester.Fatalf("read %s Taskfile: %v", module, err)
	}

	return content
}

func validateTaskfileFormatting(tester TestingT, module string, content []byte) {
	if bytes.Contains(content, []byte(windowsLineEnding)) {
		tester.Fatalf("%s Taskfile must use LF line endings", module)
	}

	trimmedWhitespace := bytes.TrimRight(content, " \t\r\n")
	trimmedLineEndings := bytes.TrimRight(content, windowsLineEnding)

	if !bytes.Equal(trimmedWhitespace, trimmedLineEndings) {
		tester.Fatalf("%s Taskfile has trailing whitespace", module)
	}
}

func mustParseTaskfile(tester TestingT, module string, content []byte) *Taskfile {
	taskfile := new(Taskfile)

	err := yaml.Unmarshal(content, taskfile)
	if err != nil {
		tester.Fatalf("parse %s Taskfile: %v", module, err)
	}

	return taskfile
}

func findRepoRoot(tester TestingT, directory string) string {
	for {
		if pathExists(filepath.Join(directory, "go.mod")) {
			return directory
		}

		parent := filepath.Dir(directory)

		if parent == directory {
			tester.Fatal("could not find repository root with go.mod")
		}

		directory = parent
	}
}

func findRepoRootFrom(tester TestingT, workingDir workingDirectoryFunc) string {
	tester.Helper()

	workingDirectory, err := workingDir()
	if err != nil {
		tester.Fatalf("get working directory: %v", err)
	}

	return findRepoRoot(tester, workingDirectory)
}

func pathExists(path string) bool {
	info, err := os.Stat(path)

	if info == nil {
		return false
	}

	return err == nil
}

func assertReadme(tester TestingT, module string, expectedTasks []string) {
	tester.Helper()

	text := string(mustReadReadme(tester, module))
	validateReadme(tester, module, text)
	assertReadmeTasks(tester, module, &readmeCheck{text: text, expectedTasks: expectedTasks})
}

func mustReadReadme(tester TestingT, module string) []byte {
	path := filepath.Join(moduleDir(tester, readmeModule(module)), "README.md")

	content, err := readFile(path)
	if err != nil {
		tester.Fatalf("%s must have README.md: %v", module, err)
	}

	return content
}

func readmeModule(module string) string {
	if before, _, ok := strings.Cut(module, "/"); ok {
		return before
	}

	return module
}

func validateReadme(tester TestingT, module, text string) {
	if strings.TrimSpace(text) == "" {
		tester.Fatalf("%s README.md is empty", module)
	}

	if !strings.Contains(text, "## Public Tasks") {
		tester.Fatalf("%s README.md must document public tasks", module)
	}
}

func assertReadmeTasks(tester TestingT, module string, check *readmeCheck) {
	for i := range check.expectedTasks {
		if !strings.Contains(check.text, "`"+check.expectedTasks[i]+"`") {
			tester.Fatalf(
				"%s README.md does not mention public task %q",
				module,
				check.expectedTasks[i],
			)
		}
	}
}

func assertTaskfile(tester TestingT, module string, expected *ModuleExpectations) {
	tester.Helper()

	taskfile := LoadTaskfile(tester, module)

	if taskfile.Version != "3" && !strings.HasPrefix(taskfile.Version, "3.") {
		tester.Fatalf("%s Taskfile version must be 3 or 3.x, got %q", module, taskfile.Version)
	}

	if len(taskfile.Tasks) == emptyLength {
		tester.Fatalf("%s Taskfile must define tasks", module)
	}

	actualTasks := publicTaskNames(taskfile)
	assertPublicTasks(tester, module, &publicTaskCheck{
		taskfile:      taskfile,
		expectedTasks: sortedCopy(expected.Tasks),
		actualTasks:   actualTasks,
	})
	assertExpectedVars(tester, module, &varCheck{taskfile: taskfile, expectedVars: expected.Vars})
}

func assertPublicTasks(tester TestingT, module string, check *publicTaskCheck) {
	tester.Helper()

	if !slices.Equal(check.expectedTasks, check.actualTasks) {
		tester.Fatalf(
			"%s public task drift\nexpected: %v\nactual:   %v",
			module,
			check.expectedTasks,
			check.actualTasks,
		)
	}

	for i := range check.actualTasks {
		name := check.actualTasks[i]
		assertPublicTask(tester, module, &taskCheck{name: name, task: check.taskfile.Tasks[name]})
	}
}

func assertPublicTask(tester TestingT, module string, check *taskCheck) {
	tester.Helper()

	if len(strings.TrimSpace(check.task.Desc)) < minTaskDescriptionLength {
		tester.Fatalf(
			"%s task %q desc is missing or too short: %q",
			module,
			check.name,
			check.task.Desc,
		)
	}

	if check.task.Cmds == nil && check.task.Deps == nil {
		tester.Fatalf("%s task %q must define cmds or deps", module, check.name)
	}
}

func assertExpectedVars(tester TestingT, module string, check *varCheck) {
	tester.Helper()

	for i := range check.expectedVars {
		if _, ok := check.taskfile.Vars[check.expectedVars[i]]; !ok {
			tester.Fatalf("%s Taskfile vars missing %q", module, check.expectedVars[i])
		}
	}
}

func assertTaskCLIParse(tester TestingT, module string) {
	tester.Helper()

	output, err := taskListJSONOutput(tester, module)
	if err != nil {
		failTaskCliLoad(tester, &taskCliLoadFailure{module: module, output: output, err: err})
	}

	assertValidJSON(tester, module, output)
}

func taskListJSONOutput(tester TestingT, module string) (string, error) {
	output, err := runTaskListJSONOutput(tester, module)
	if err != nil {
		return output, fmt.Errorf("list %s tasks as JSON: %w", module, err)
	}

	return output, nil
}

func failTaskCliLoad(tester TestingT, failure *taskCliLoadFailure) {
	tester.Fatalf(
		"%s task --list-all --json failed:\n%s\nerror: %v",
		failure.module,
		failure.output,
		failure.err,
	)
}

func assertValidJSON(tester TestingT, module, output string) {
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

func publicTaskNames(taskfile *Taskfile) []string {
	names := make([]string, emptyLength, len(taskfile.Tasks))

	for name := range taskfile.Tasks {
		if name == "default" || taskfile.Tasks[name].Internal {
			continue
		}

		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

func sortedCopy(values []string) []string {
	clone := slices.Clone(values)
	slices.Sort(clone)

	return clone
}

func taskfilePath(tester TestingT, module string) string {
	tester.Helper()

	return filepath.Join(moduleDir(tester, module), taskfileName)
}

func moduleDir(tester TestingT, module string) string {
	tester.Helper()

	return filepath.Join(RepoRoot(tester), "taskfiles", module)
}

func runTaskListJSONOutput(tester TestingT, module string) (string, error) {
	tester.Helper()

	settings := currentTaskCommandSettings()
	ctx, cancel := context.WithTimeout(context.Background(), settings.timeout)

	defer cancel()

	cmd := newTaskListJSONCommand(ctx, tester, module)
	output, err := cmd.CombinedOutput()

	assertTaskCommandDidNotTimeout(ctx, tester, []string{
		taskfileFlag,
		taskfileName,
		taskListAllFlag,
		taskJSONFlag,
	})

	return string(output), err
}

func newTaskListJSONCommand(ctx context.Context, tester TestingT, module string) *exec.Cmd {
	cmd := exec.CommandContext(
		ctx,
		taskBinary,
		taskfileFlag,
		taskfileName,
		taskListAllFlag,
		taskJSONFlag,
	)

	cmd.Dir = moduleDir(tester, module)
	cmd.Env = os.Environ()

	return cmd
}

func assertTaskCommandDidNotTimeout(ctx context.Context, tester TestingT, args []string) {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		tester.Fatalf("task command timed out: %s %s", taskBinary, strings.Join(args, " "))
	}
}

func readFile(path string) ([]byte, error) {
	clean := filepath.Clean(path)

	content, err := fs.ReadFile(os.DirFS(filepath.Dir(clean)), filepath.Base(clean))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", clean, err)
	}

	return content, nil
}
