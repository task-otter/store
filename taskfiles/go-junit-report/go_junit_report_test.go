// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package gojunitreport_test

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
)

const (
	constGoJunitReportReport = "report"
	verifyTask               = "verify"
	whichTask                = "which"
	installTask              = "install"
	versionTask              = "version"
	goInstallTask            = "go:install"
	goJunitReportModule      = "go-junit-report"
	goJunitReportNixVar      = "GO_JUNIT_REPORT_NIX_INSTALLABLE"
	goJunitReportInVar       = "GO_JUNIT_REPORT_IN"
	goJunitReportOutVar      = "GO_JUNIT_REPORT_OUT"
	fmtPercentV              = "%v"
	zeroLen                  = 0
)

// TestModuleIntegration runs the shared task CLI integration suite for this module.
func TestModuleIntegration(t *testing.T) {
	t.Parallel()

	taskintegration.RunHere(t)
}

// TestTaskfileModuleContract validates the module contract.
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		goJunitReportModule,
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

// TestReportTaskCommands validates report task command tokens.
func TestReportTaskCommands(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, goJunitReportModule)

	assertTaskCmdsContainTokens(
		t,
		&taskCmdsTokenCheck{
			taskfile: taskfile,
			taskName: constGoJunitReportReport,
			tokens:   goJunitReportReportTokens(),
		},
	)
}

// TestOperationalTaskDependencies validates public task dependencies.
func TestOperationalTaskDependencies(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, goJunitReportModule)

	assertDependencyMap(t, taskfile, map[string][]string{
		constGoJunitReportReport: {goInstallTask, installTask},
		whichTask:                {goInstallTask, installTask},
		verifyTask:               {goInstallTask, installTask},
		versionTask:              {goInstallTask, installTask},
	})
}

func publicTasks() []string {
	return []string{
		installTask,
		constGoJunitReportReport,
		verifyTask,
		versionTask,
		whichTask,
	}
}

func publicVars() []string {
	return []string{
		goJunitReportInVar,
		goJunitReportOutVar,
		goJunitReportNixVar,
	}
}

func goJunitReportReportTokens() []string {
	return []string{
		"go-junit-report",
		"-in",
		"-out",
		"-set-exit-code",
		goJunitReportInVar,
		goJunitReportOutVar,
	}
}

func assertTaskCmdsContainTokens(t *testing.T, check *taskCmdsTokenCheck) {
	t.Helper()

	task, ok := check.taskfile.Tasks[check.taskName]

	if !ok {
		t.Fatalf("go-junit-report Taskfile missing task %q", check.taskName)
	}

	cmds := fmt.Sprintf(fmtPercentV, task.Cmds)

	for i := range check.tokens {
		token := check.tokens[i]

		if !strings.Contains(cmds, token) {
			t.Fatalf("go-junit-report task %q cmds missing %q: %s", check.taskName, token, cmds)
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
