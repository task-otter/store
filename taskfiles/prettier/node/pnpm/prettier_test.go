// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package pnpm_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"prettier/node/pnpm",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"config:init",
		"ci:fix",
		"fmt:check",
		"help",
		"install",
		"install:undo",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"PRETTIER_CONFIG",
		"PRETTIER_EXTRA_ARGS",
		"PRETTIER_IGNORE_PATH",
		"PRETTIER_TARGETS",
		"PRETTIER_VERSION",
	}
}
