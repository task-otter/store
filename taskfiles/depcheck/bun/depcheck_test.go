// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package bun_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"ci",
		"help",
		"ignores",
		"install",
		"install:undo",
		"json",
		"lint",
		"skip-missing",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"DEPCHECK_LINT_SKIP_PATTERN",
		"EXTRA_ARGS",
		"IGNORE_PACKAGES",
		"PROJECT_PATH",
		"TARGETS",
		"VERSION",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"depcheck/bun",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

// covered by module contract
