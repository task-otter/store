// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package go_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/taskintegration"
	"github.com/task-otter/store/internal/tasktest"
)

type (
	taskCmdsTokenCheck struct {
		taskfile *tasktest.Taskfile
		taskName string
		tokens   []string
	}

	dependencyCheck struct {
		taskfile *tasktest.Taskfile
		taskName string
		expected []string
	}

	taskCommandCase struct {
		task   string
		tokens []string
	}
)

const (
	constGoTestTest = "test"
	benchTask       = "bench"
	fuzzTask        = "fuzz"
	verifyTask      = "verify"
	whichTask       = "which"
	installTask     = "install"
	versionTask     = "version"
	goModuleName    = "go"
	goTestCmd       = "go test"
	fmtPercentV     = "%v"
	zeroLen         = 0
)

// TestModuleIntegration runs the shared task CLI integration suite for this module.
func TestModuleIntegration(t *testing.T) {
	t.Parallel()

	taskintegration.RunHere(t)
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		goModuleName,
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

// TestTestingTaskCommands
func TestTestingTaskCommands(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, goModuleName)
	tests := taskCommandCases()

	for i := range tests {
		runTaskCommandCase(t, taskfile, &tests[i])
	}
}

// TestOperationalTaskDependencies
func TestOperationalTaskDependencies(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, goModuleName)

	assertDependencyMap(t, taskfile, map[string][]string{
		constGoTestTest: {installTask},
		benchTask:       {installTask},
		fuzzTask:        {installTask},
		whichTask:       {installTask},
		verifyTask:      {installTask},
		versionTask:     {installTask},
	})
}

func publicTasks() []string {
	return []string{
		benchTask,
		fuzzTask,
		installTask,
		constGoTestTest,
		verifyTask,
		versionTask,
		whichTask,
	}
}

func publicVars() []string {
	return []string{
		"GO_FUZZTIME",
		"GO_NIX_INSTALLABLE",
	}
}

func goTestTaskTokens() []string {
	return []string{
		"go test -v",
		"./...",
	}
}

func taskCommandCases() []taskCommandCase {
	return []taskCommandCase{
		{task: constGoTestTest, tokens: goTestTaskTokens()},
		{task: benchTask, tokens: []string{goTestCmd, "-bench", "-benchmem"}},
		{task: fuzzTask, tokens: []string{goTestCmd, "-fuzz", "-fuzztime"}},
	}
}

func runTaskCommandCase(t *testing.T, taskfile *tasktest.Taskfile, testCase *taskCommandCase) {
	t.Helper()
	t.Run(testCase.task, func(t *testing.T) {
		t.Parallel()

		assertTaskCmdsContainTokens(
			t,
			&taskCmdsTokenCheck{
				taskfile: taskfile,
				taskName: testCase.task,
				tokens:   testCase.tokens,
			},
		)
	})
}

func assertTaskCmdsContainTokens(t *testing.T, check *taskCmdsTokenCheck) {
	t.Helper()

	task, ok := check.taskfile.Tasks[check.taskName]

	if !ok {
		t.Fatalf("go Taskfile missing task %q", check.taskName)
	}

	cmds := fmt.Sprintf(fmtPercentV, task.Cmds)

	for i := range check.tokens {
		token := check.tokens[i]

		if !strings.Contains(cmds, token) {
			t.Fatalf("go task %q cmds missing %q: %s", check.taskName, token, cmds)
		}
	}
}

func assertDependencyMap(t *testing.T, taskfile *tasktest.Taskfile, deps map[string][]string) {
	t.Helper()

	for taskName := range deps {
		expected := deps[taskName]
		assertTaskDependencies(
			t,
			&dependencyCheck{taskfile: taskfile, taskName: taskName, expected: expected},
		)
	}
}

func assertTaskDependencies(t *testing.T, check *dependencyCheck) {
	t.Helper()

	rawDeps := rawTaskDeps(t, check)
	actual := taskDependencyNames(t, check.taskName, rawDeps)

	if !slices.Equal(actual, check.expected) {
		t.Fatalf(
			"%s deps mismatch\nexpected: %v\nactual:   %v",
			check.taskName,
			check.expected,
			actual,
		)
	}
}

func rawTaskDeps(t *testing.T, check *dependencyCheck) []any {
	t.Helper()

	rawDeps, ok := check.taskfile.Tasks[check.taskName].Deps.([]any)

	if !ok {
		t.Fatalf(
			"%s deps have type %T, want []any",
			check.taskName,
			check.taskfile.Tasks[check.taskName].Deps,
		)
	}

	return rawDeps
}

func taskDependencyNames(t *testing.T, taskName string, rawDeps []any) []string {
	t.Helper()

	actual := make([]string, zeroLen, len(rawDeps))

	for index := range rawDeps {
		rawDep := rawDeps[index]
		dep, ok := taskDependencyName(rawDep)

		if !ok {
			t.Fatalf("%s dependency %d has unsupported value %v", taskName, index, rawDep)
		}

		actual = append(actual, dep)
	}

	return actual
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
