// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package proto_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

const (
	constProtoModule       = "proto"
	constProtoTaskGen      = "gen"
	constNixInstallProfile = "nix:install:profile"
)

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		constProtoModule,
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

// TestGenInstallsViaNixProfile
func TestGenInstallsViaNixProfile(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, constProtoModule)

	deps, ok := taskfile.Tasks[constProtoTaskGen].Deps.([]any)

	if !ok {
		t.Fatalf(
			"%s deps have type %T, want []any",
			constProtoTaskGen,
			taskfile.Tasks[constProtoTaskGen].Deps,
		)
	}

	if !containsTaskDependency(deps, constNixInstallProfile) {
		t.Errorf("%s must depend on %s; deps: %v", constProtoTaskGen, constNixInstallProfile, deps)
	}
}

func publicTasks() []string {
	return []string{
		constProtoTaskGen,
		"ungen",
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
