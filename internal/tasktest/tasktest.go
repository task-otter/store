// Copyright 2026 task-otter
// SPDX-License-Identifier: Apache-2.0

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
	"sort"
	"strings"
	"time"

	yaml "go.yaml.in/yaml/v3"
)

// Taskfile contains the top-level fields read from a Taskfile.yml.
type (
	Taskfile struct {
		Vars    map[string]any   `yaml:"vars"`
		Tasks   map[string]*Task `yaml:"tasks"`
		Version string           `yaml:"version"`
	}

	// Task contains the task fields validated by the shared test helpers.
	Task struct {
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

	TestingT interface {
		Helper()
		Fatal(args ...any)
		Fatalf(format string, args ...any)
		TempDir() string
	}

	taskCommandSettings struct {
		binary  string
		timeout time.Duration
	}
)

const (
	minTaskDescriptionLength = 12
)

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
func AssertModule(tester TestingT, module string, expectedTasks, expectedVars []string) {
	tester.Helper()

	assertReadme(tester, module, expectedTasks)
	assertTaskfile(tester, module, expectedTasks, expectedVars)
}

// AssertTaskCliCanLoad verifies the task CLI can parse a module's Taskfile.
func AssertTaskCliCanLoad(tester TestingT, module string) {
	tester.Helper()

	assertTaskCliCanLoad(tester, module)
}

// LoadTaskfile reads and parses the Taskfile for module.
func LoadTaskfile(tester TestingT, module string) *Taskfile {
	tester.Helper()

	content := mustReadTaskfile(tester, module)
	validateTaskfileFormatting(tester, module, content)

	return mustParseTaskfile(tester, module, content)
}

func mustReadTaskfile(tester TestingT, module string) []byte {
	content, err := readFile(taskfilePath(tester, module))
	if err != nil {
		tester.Fatalf("read %s Taskfile: %v", module, err)
	}

	return content
}

func validateTaskfileFormatting(tester TestingT, module string, content []byte) {
	if bytes.Contains(content, []byte("\r\n")) {
		tester.Fatalf("%s Taskfile must use LF line endings", module)
	}

	trimmedWhitespace := bytes.TrimRight(content, " \t\r\n")
	trimmedLineEndings := bytes.TrimRight(content, "\r\n")

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

// RepoRoot walks upward from the working directory to find the repository root.
func RepoRoot(tester TestingT) string {
	tester.Helper()

	workingDirectory, err := workingDir()
	if err != nil {
		tester.Fatalf("get working directory: %v", err)
	}

	return findRepoRoot(tester, workingDirectory)
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

func pathExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

func assertReadme(tester TestingT, module string, expectedTasks []string) {
	tester.Helper()

	text := string(mustReadReadme(tester, module))
	validateReadme(tester, module, text)
	assertReadmeTasks(tester, module, text, expectedTasks)
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

func assertReadmeTasks(tester TestingT, module, text string, expectedTasks []string) {
	for i := range expectedTasks {
		if !strings.Contains(text, "`"+expectedTasks[i]+"`") {
			tester.Fatalf("%s README.md does not mention public task %q", module, expectedTasks[i])
		}
	}
}

func assertTaskfile(tester TestingT, module string, expectedTasks, expectedVars []string) {
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
	tester TestingT,
	module string,
	taskfile *Taskfile,
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

	for i := range actualTasks {
		assertPublicTask(tester, module, actualTasks[i], taskfile.Tasks[actualTasks[i]])
	}
}

func assertPublicTask(tester TestingT, module, name string, task *Task) {
	tester.Helper()

	if len(strings.TrimSpace(task.Desc)) < minTaskDescriptionLength {
		tester.Fatalf("%s task %q desc is missing or too short: %q", module, name, task.Desc)
	}

	if task.Cmds == nil && task.Deps == nil {
		tester.Fatalf("%s task %q must define cmds or deps", module, name)
	}
}

func assertExpectedVars(tester TestingT, module string, taskfile *Taskfile, expectedVars []string) {
	tester.Helper()

	for i := range expectedVars {
		if _, ok := taskfile.Vars[expectedVars[i]]; !ok {
			tester.Fatalf("%s Taskfile vars missing %q", module, expectedVars[i])
		}
	}
}

func assertTaskCliCanLoad(tester TestingT, module string) {
	tester.Helper()

	output, err := taskListJSONOutput(tester, module)
	if err != nil {
		failTaskCliLoad(tester, module, output, err)
	}

	assertValidJSON(tester, module, output)
}

func taskListJSONOutput(tester TestingT, module string) (string, error) {
	return runTaskOutput(
		tester,
		"--taskfile",
		taskfilePath(tester, module),
		"--list-all",
		"--json",
	)
}

func failTaskCliLoad(tester TestingT, module, output string, err error) {
	tester.Fatalf(
		"%s task --list-all --json failed:\n%s\nerror: %v",
		module,
		output,
		err,
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
	names := make([]string, 0, len(taskfile.Tasks))

	for name := range taskfile.Tasks {
		if name == "default" || taskfile.Tasks[name].Internal {
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

func taskfilePath(tester TestingT, module string) string {
	tester.Helper()

	return filepath.Join(moduleDir(tester, module), "Taskfile.yml")
}

func moduleDir(tester TestingT, module string) string {
	tester.Helper()

	return filepath.Join(RepoRoot(tester), "taskfiles", module)
}

func runTaskOutput(tester TestingT, args ...string) (string, error) {
	tester.Helper()

	settings := currentTaskCommandSettings()
	ctx, cancel := context.WithTimeout(context.Background(), settings.timeout)

	defer cancel()

	cmd := newTaskCommand(tester, ctx, settings.binary, args)
	output, err := cmd.CombinedOutput()

	assertTaskCommandDidNotTimeout(tester, ctx, args)

	return string(output), err
}

func newTaskCommand(
	tester TestingT,
	ctx context.Context,
	binary string,
	args []string,
) *exec.Cmd {
	if binary != "task" {
		tester.Fatalf("unsupported task command binary %q", binary)

		return nil
	}

	cmd := exec.CommandContext(ctx, "task", args...)

	cmd.Dir = RepoRoot(tester)
	cmd.Env = os.Environ()

	return cmd
}

func assertTaskCommandDidNotTimeout(tester TestingT, ctx context.Context, args []string) {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		tester.Fatalf("task command timed out: task %s", strings.Join(args, " "))
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
