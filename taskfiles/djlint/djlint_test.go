// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package djlint_test

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
		"lint",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"DJLINT_LINT_SKIP_PATTERN",
		"DJLINT_FMT_SKIP_PATTERN",
		"DJLINT_VERSION",
		"EXTRA_ARGS",
		"TARGETS",
		"UV_LOAD",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"djlint",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
