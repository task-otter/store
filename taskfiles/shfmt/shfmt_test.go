// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package shfmt_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"fmt",
		"fmt:check",
		"install",
		"install:undo",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"SHFMT_FMT_SKIP_PATTERN",
		"EXTRA_ARGS",
		"GO_GLOBAL_BIN",
		"SHFMT_VERSION",
		"TARGETS",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"shfmt",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
