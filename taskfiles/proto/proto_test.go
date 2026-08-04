// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package proto_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

const (
	constProtoModule      = "proto"
	constProtoTaskInstall = "install"
	constProtoTaskUpgrade = "upgrade"
)

func publicTasks() []string {
	return []string{
		"gen",
		constProtoTaskInstall,
		"install:undo",
		constProtoTaskUpgrade,
		"ungen",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"GO_CMD",
		"GO_GLOBAL_BIN",
		"PROTO_PATH",
		"PROTO_PATTERN",
		"PROTOC_GEN_GO_GRPC_VERSION",
		"PROTOC_GEN_GO_VERSION",
		"PROTOC_VERSION",
	}
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

// TestPluginWorkflowsInstallGoFirst
func TestPluginWorkflowsInstallGoFirst(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, constProtoModule)

	for i := range []string{constProtoTaskInstall, constProtoTaskUpgrade} {
		taskName := []string{constProtoTaskInstall, constProtoTaskUpgrade}[i]
		deps, ok := taskfile.Tasks[taskName].Deps.([]any)

		if !ok {
			t.Fatalf("%s deps have type %T, want []any", taskName, taskfile.Tasks[taskName].Deps)
		}

		if !containsTaskDependency(deps, "go:install") {
			t.Errorf("%s must depend on go:install; deps: %v", taskName, deps)
		}
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
