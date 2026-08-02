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
	constGoTestTest      = "test"
	benchTask            = "bench"
	fuzzTask             = "fuzz"
	installTask          = "install"
	installGoJunitReport = "install:go-junit-report"
	goCoverProfileVar    = "GO_COVER_PROFILE"
	goJunitReportVar     = "GO_JUNIT_REPORT"
	goModuleName         = "go"
	goTestCmd            = "go test"
	fmtPercentV          = "%v"
	zeroLen              = 0
)

func publicTasks() []string {
	return []string{
		benchTask,
		fuzzTask,
		installTask,
		installGoJunitReport,
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
		goCoverProfileVar,
		"GO_DOWNLOAD_BASE_URL",
		"GO_FUZZTIME",
		goJunitReportVar,
		"GO_VERSION",
		"GO_ROOT_UNIX",
		"GO_VERSION_URL",
		"GLOBAL_GO_BIN",
		"INSTALL_DIR_UNIX",
	}
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

func goTestTaskTokens() []string {
	return []string{
		"go test -v",
		"-covermode atomic",
		"-coverprofile",
		"./...",
		"go-junit-report",
		"-set-exit-code",
		"-iocopy",
		"-out",
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

// TestTestingTaskCommands
func TestTestingTaskCommands(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, goModuleName)
	tests := taskCommandCases()

	for i := range tests {
		runTaskCommandCase(t, taskfile, &tests[i])
	}

	assertGoTestVarsContainTokens(t, taskfile)
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

func assertGoTestVarsContainTokens(t *testing.T, taskfile *tasktest.Taskfile) {
	t.Helper()

	testVars := fmt.Sprintf(fmtPercentV, taskfile.Tasks[constGoTestTest].Vars)

	tokens := []string{
		"GO_JUNIT_REPORT_OUT",
		goJunitReportVar,
		"junit.xml",
		"GO_COVER_OUT",
		goCoverProfileVar,
		"coverage.out",
	}

	for i := range tokens {
		if !strings.Contains(testVars, tokens[i]) {
			t.Fatalf("go test vars missing %q: %s", tokens[i], testVars)
		}
	}
}

// TestVersionVariableIsOptional
func TestVersionVariableIsOptional(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, goModuleName)

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

// TestDevelopmentToolDependencies
func TestDevelopmentToolDependencies(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, goModuleName)

	installTasks := map[string][]string{
		installGoJunitReport: {installTask},
	}
	testTasks := map[string][]string{
		constGoTestTest: {installGoJunitReport},
	}

	assertDependencyMap(t, taskfile, installTasks)
	assertDependencyMap(t, taskfile, testTasks)
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
