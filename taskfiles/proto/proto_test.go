// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package proto_test

import (
	"testing"

	"github.com/task-otter/store/internal/taskintegration"
	"github.com/task-otter/store/internal/tasktest"
)

type (
	// taskReference is one assertion that a task field calls another task.
	taskReference struct {
		entries  any
		task     string
		expected string
	}
)

const (
	constProtoModule       = "proto"
	constProtoTaskGen      = "gen"
	constProtoTaskInstall  = "install"
	constNixInstallProfile = "nix:install:profile"
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
		constProtoModule,
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

// TestGenDependsOnInstall proves gen auto-installs through the module's public
// install task, which is the single caller of nix:install:profile.
func TestGenDependsOnInstall(t *testing.T) {
	t.Parallel()

	assertTaskDependsOn(t, constProtoTaskGen, constProtoTaskInstall)
}

// TestInstallUsesNixProfile proves the install task installs through the shared
// nix:install:profile task rather than owning an installer of its own.
func TestInstallUsesNixProfile(t *testing.T) {
	t.Parallel()

	assertTaskRuns(t, constProtoTaskInstall, constNixInstallProfile)
}

func assertTaskDependsOn(t *testing.T, task, expected string) {
	t.Helper()

	assertTaskReferences(t, &taskReference{
		task:     task,
		expected: expected,
		entries:  loadTask(t, task).Deps,
	})
}

func assertTaskRuns(t *testing.T, task, expected string) {
	t.Helper()

	assertTaskReferences(t, &taskReference{
		task:     task,
		expected: expected,
		entries:  loadTask(t, task).Cmds,
	})
}

func loadTask(t *testing.T, task string) *tasktest.Task {
	t.Helper()

	return tasktest.LoadTaskfile(t, constProtoModule).Tasks[task]
}

// assertTaskReferences proves the entries of one task field list a call to expected.
func assertTaskReferences(t *testing.T, reference *taskReference) {
	t.Helper()

	entries, ok := reference.entries.([]any)

	if !ok {
		t.Fatalf("%s entries have type %T, want []any", reference.task, reference.entries)
	}

	if !containsTaskDependency(entries, reference.expected) {
		t.Errorf("%s must reference %s; entries: %v", reference.task, reference.expected, entries)
	}
}

func publicTasks() []string {
	return []string{
		constProtoTaskGen,
		constProtoTaskInstall,
		"ungen",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"GO_MODULE",
		"PROTO_NIX_INSTALLABLE",
		"PROTO_PATH",
		"PROTO_PATTERN",
	}
}

func containsTaskDependency(deps []any, expected string) bool {
	for i := range deps {
		rawDep := deps[i]

		if taskDependencyMatches(rawDep, expected) {
			return true
		}
	}

	return false
}

func taskDependencyMatches(rawDep any, expected string) bool {
	switch dep := rawDep.(type) {
	case string:
		return dep == expected
	case map[string]any:
		return dep["task"] == expected
	default:
		return false
	}
}
