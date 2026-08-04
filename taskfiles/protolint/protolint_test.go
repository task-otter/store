// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package protolint_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"fix",
		"install",
		"install:undo",
		"lint",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"PROTOLINT_LINT_SKIP_PATTERN",
		"EXTRA_ARGS",
		"GO_GLOBAL_BIN",
		"PROTOLINT_VERSION",
		"TARGETS",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"protolint",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
