// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package protolint_test

import (
	"slices"
	"testing"

	"github.com/task-otter/store/internal/taskintegration"
	"github.com/task-otter/store/internal/tasktest"
)

type (
	dependencyCheck struct {
		taskfile *tasktest.Taskfile
		taskName string
		expected []string
	}
)

const (
	ciTask          = "ci"
	ciFixTask       = "ci:fix"
	installTask     = "install"
	versionTask     = "version"
	goInstallTask   = "go:install"
	protolintModule = "protolint"
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
		protolintModule,
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

// TestOperationalTaskDependencies validates public task dependencies.
func TestOperationalTaskDependencies(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, protolintModule)

	assertDependencyMap(t, taskfile, map[string][]string{
		ciTask:      {goInstallTask, installTask},
		ciFixTask:   {goInstallTask, installTask},
		versionTask: {goInstallTask, installTask},
	})
}

func publicTasks() []string {
	return []string{
		ciTask,
		ciFixTask,
		installTask,
		versionTask,
	}
}

func publicVars() []string {
	return []string{
		"PROTOLINT_EXTRA_ARGS",
		"PROTOLINT_GO_PKG",
		"PROTOLINT_NIX_INSTALLABLE",
		"PROTOLINT_TARGETS",
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
