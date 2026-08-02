// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package shellcheck_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"install",
		"install:undo",
		"lint",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"SHELLCHECK_LINT_SKIP_PATTERN",
		"EXTRA_ARGS",
		"TARGETS",
		"VERSION",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"shellcheck",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
