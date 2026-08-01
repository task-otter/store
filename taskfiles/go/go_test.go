// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package go_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

const constGoTestTest = "test"

func publicTasks() []string {
	return []string{
		"bench",
		"fuzz",
		"install",
		"install:go-junit-report",
		"install:undo",
		constGoTestTest,
		"upgrade",
		"verify",
		"version",
		"which",
	}
}

func publicVars() []string {
	return []string{
		"GO_BIN_UNIX",
		"GO_CMD_UNIX",
		"GO_COVER_PROFILE",
		"GO_DOWNLOAD_BASE_URL",
		"GO_FUZZTIME",
		"GO_JUNIT_REPORT",
		"GO_VERSION",
		"GO_ROOT_UNIX",
		"GO_VERSION_URL",
		"GLOBAL_GO_BIN",
		"INSTALL_DIR_UNIX",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"go",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func TestTestingTaskCommands(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "go")

	tests := []struct {
		task   string
		tokens []string
	}{
		{
			task: constGoTestTest,
			tokens: []string{
				"go test -v",
				"-covermode atomic",
				"-coverprofile",
				"./...",
				"go-junit-report",
				"-set-exit-code",
				"-iocopy",
				"-out",
			},
		},
		{task: "bench", tokens: []string{"go test", "-bench", "-benchmem"}},
		{task: "fuzz", tokens: []string{"go test", "-fuzz", "-fuzztime"}},
	}

	for _, testCase := range tests {
		t.Run(testCase.task, func(t *testing.T) {
			t.Parallel()

			task, ok := taskfile.Tasks[testCase.task]

			if !ok {
				t.Fatalf("go Taskfile missing task %q", testCase.task)
			}

			cmds := fmt.Sprintf("%v", task.Cmds)

			for _, token := range testCase.tokens {
				if !strings.Contains(cmds, token) {
					t.Fatalf("go task %q cmds missing %q: %s", testCase.task, token, cmds)
				}
			}
		})
	}

	testVars := fmt.Sprintf("%v", taskfile.Tasks[constGoTestTest].Vars)

	for _, token := range []string{
		"GO_JUNIT_REPORT_OUT",
		"GO_JUNIT_REPORT",
		"junit.xml",
		"GO_COVER_OUT",
		"GO_COVER_PROFILE",
		"coverage.out",
	} {
		if !strings.Contains(testVars, token) {
			t.Fatalf("go test vars missing %q: %s", token, testVars)
		}
	}
}

func TestVersionVariableIsOptional(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "go")

	if _, exists := taskfile.Vars["VERSION"]; exists {
		t.Fatal("shared VERSION variable must not be defined")
	}

	value, exists := taskfile.Vars["GO_VERSION"]

	if !exists {
		t.Fatal("GO_VERSION must be defined")
	}

	if value != "" {
		t.Fatalf("GO_VERSION default = %#v, want empty", value)
	}
}

func TestDevelopmentToolDependencies(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "go")

	installTasks := map[string][]string{
		"install:go-junit-report": {"install"},
	}
	testTasks := map[string][]string{
		constGoTestTest: {"install:go-junit-report"},
	}

	for taskName, expected := range installTasks {
		assertTaskDependencies(
			t,
			dependencyCheck{taskfile: taskfile, taskName: taskName, expected: expected},
		)
	}

	for taskName, expected := range testTasks {
		assertTaskDependencies(
			t,
			dependencyCheck{taskfile: taskfile, taskName: taskName, expected: expected},
		)
	}
}

type dependencyCheck struct {
	taskfile *tasktest.Taskfile
	taskName string
	expected []string
}

func assertTaskDependencies(t *testing.T, check dependencyCheck) {
	t.Helper()

	rawDeps, ok := check.taskfile.Tasks[check.taskName].Deps.([]any)

	if !ok {
		t.Fatalf(
			"%s deps have type %T, want []any",
			check.taskName,
			check.taskfile.Tasks[check.taskName].Deps,
		)
	}

	actual := make([]string, len(rawDeps))

	for index, rawDep := range rawDeps {
		dep, ok := taskDependencyName(rawDep)

		if !ok {
			t.Fatalf("%s dependency %d has unsupported value %v", check.taskName, index, rawDep)
		}

		actual[index] = dep
	}

	if !slices.Equal(actual, check.expected) {
		t.Fatalf(
			"%s deps mismatch\nexpected: %v\nactual:   %v",
			check.taskName,
			check.expected,
			actual,
		)
	}
}

func taskDependencyName(rawDep any) (string, bool) {
	switch dep := rawDep.(type) {
	case string:
		return dep, true
	case map[string]any:
		name, ok := dep["task"].(string)

		return name, ok
	default:
		return "", false
	}
}
