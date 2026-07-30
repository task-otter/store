// REPLACE_ME 2026
// SPDX-License-Identifier: Apache-2.0

package proto_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"gen",
		"install",
		"install:undo",
		"upgrade",
		"ungen",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"GO_CMD",
		"GLOBAL_GO_BIN",
		"PROTO_PATH",
		"PROTO_PATTERN",
		"PROTOC_GEN_GO_GRPC_VERSION",
		"PROTOC_GEN_GO_VERSION",
		"PROTOC_VERSION",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "proto", publicTasks(), publicVars())
}

func TestPluginWorkflowsInstallGoFirst(t *testing.T) {
	t.Parallel()

	tf := tasktest.LoadTaskfile(t, "proto")

	for _, taskName := range []string{"install", "upgrade"} {
		deps, ok := tf.Tasks[taskName].Deps.([]any)

		if !ok {
			t.Fatalf("%s deps have type %T, want []any", taskName, tf.Tasks[taskName].Deps)
		}

		if !containsTaskDependency(deps, "go:install") {
			t.Errorf("%s must depend on go:install; deps: %v", taskName, deps)
		}
	}
}

func containsTaskDependency(deps []any, expected string) bool {
	for _, rawDep := range deps {
		switch dep := rawDep.(type) {
		case string:
			if dep == expected {
				return true
			}
		case map[string]any:
			if dep["task"] == expected {
				return true
			}
		}
	}

	return false
}
