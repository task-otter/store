// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package hadolint_test

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
		"HADOLINT_LINT_SKIP_PATTERN",
		"CONFIG",
		"DOCKERFILE",
		"EXTRA_ARGS",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"hadolint",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
