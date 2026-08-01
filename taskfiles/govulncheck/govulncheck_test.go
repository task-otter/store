// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package govulncheck_test

import (
	"slices"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

const (
	constGovulncheckInstall = "install"
	constGovulncheckLint    = "lint"
	constGoInstall          = "go:install"
)

func publicTasks() []string {
	return []string{
		constGovulncheckInstall,
		constGovulncheckLint,
	}
}

func publicVars() []string {
	return []string{
		"GOVULNCHECK_VERSION",
		"GOVULNCHECK_LINT_SKIP_PATTERN",
		"GLOBAL_GO_BIN",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"govulncheck",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func TestLintSkipPatternDefaultsEmpty(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "govulncheck")

	value, exists := taskfile.Vars["GOVULNCHECK_LINT_SKIP_PATTERN"]

	if !exists {
		t.Fatal("GOVULNCHECK_LINT_SKIP_PATTERN must be defined")
	}

	if value != "" {
		t.Fatalf("GOVULNCHECK_LINT_SKIP_PATTERN default = %#v, want empty", value)
	}
}

func TestVersionVariableIsOptional(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "govulncheck")

	value, exists := taskfile.Vars["GOVULNCHECK_VERSION"]

	if !exists {
		t.Fatal("GOVULNCHECK_VERSION must be defined")
	}

	if value != "" {
		t.Fatalf("GOVULNCHECK_VERSION default = %#v, want empty", value)
	}
}

func TestDevelopmentToolDependencies(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "govulncheck")

	dependencies := map[string][]string{
		constGovulncheckInstall: {constGoInstall},
		constGovulncheckLint:    {constGovulncheckInstall},
	}

	for taskName, expected := range dependencies {
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
