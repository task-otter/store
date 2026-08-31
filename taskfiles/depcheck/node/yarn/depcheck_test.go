// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package yarn_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"depcheck/node/yarn",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"ci",
		"help",
		"ignores",
		"install",
		"install:undo",
		"json",
		"skip-missing",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"DEPCHECK_LINT_SKIP_PATTERN",
		"DEPCHECK_EXTRA_ARGS",
		"DEPCHECK_IGNORE_PACKAGES",
		"DEPCHECK_PROJECT_PATH",
		"DEPCHECK_TARGETS",
		"DEPCHECK_VERSION",
	}
}

// covered by module contract
