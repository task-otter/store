// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package govulncheck_test

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
	constGovulncheckLint    = "ci"
	constGovulncheckModule  = "govulncheck"
	constGovulncheckInstall = "install"
	constGovulncheckVersion = "version"

	envVarGovulncheckNixInstallable = "GOVULNCHECK_NIX_INSTALLABLE"

	emptyString = ""
	zeroLen     = 0
)

// TestModuleIntegration runs the shared task CLI integration suite for this module.
func TestModuleIntegration(t *testing.T) {
	t.Parallel()

	taskintegration.RunHere(t)
}

// TestTaskfileModuleContract validates the behavior covered by this test case.
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		constGovulncheckModule,
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

// TestDevelopmentToolDependencies validates the behavior covered by this test case.
func TestDevelopmentToolDependencies(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, constGovulncheckModule)

	dependencies := map[string][]string{
		constGovulncheckLint:    {constGovulncheckInstall},
		constGovulncheckVersion: {constGovulncheckInstall},
	}

	for taskName := range dependencies {
		expected := dependencies[taskName]
		assertTaskDependencies(
			t,
			&dependencyCheck{taskfile: taskfile, taskName: taskName, expected: expected},
		)
	}
}

func publicTasks() []string {
	return []string{
		constGovulncheckLint,
		constGovulncheckInstall,
		constGovulncheckVersion,
	}
}

func publicVars() []string {
	return []string{
		envVarGovulncheckNixInstallable,
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
		return emptyString, false
	}
}
