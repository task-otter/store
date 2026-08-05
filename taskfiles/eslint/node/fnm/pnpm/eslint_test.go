// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package pnpm_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"cache:clean",
		"ci",
		"ci:fix",
		"config:init",
		"help",
		"install",
		"install:undo",
		"lint",
		"lint:fix",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"ESLINT_LINT_SKIP_PATTERN",
		"ESLINT_CACHE",
		"ESLINT_CONFIG",
		"ESLINT_EXTRA_ARGS",
		"ESLINT_TARGETS",
		"ESLINT_VERSION",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"eslint/node/fnm/pnpm",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
