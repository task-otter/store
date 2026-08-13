// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package buf_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"breaking",
		"ci",
		"ci:fix",
		"fmt:check",
		"generate",
		"install",
		"install:undo",
		"lint",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"BUF_LINT_SKIP_PATTERN",
		"BUF_FMT_SKIP_PATTERN",
		"BUF_AGAINST",
		"BUF_VERSION",
		"BUF_CONFIG",
		"BUF_EXTRA_ARGS",
		"BUF_INPUT",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"buf",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
