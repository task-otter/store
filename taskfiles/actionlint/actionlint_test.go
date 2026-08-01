// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package actionlint_test

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
		"ACTIONLINT_LINT_SKIP_PATTERN",
		"ACTIONLINT_EXTRA_ARGS",
		"ACTIONLINT_TARGETS",
		"ACTIONLINT_VERSION",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"actionlint",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
