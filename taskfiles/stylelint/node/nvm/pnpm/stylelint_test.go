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
		"config:skip",
		"help",
		"install",
		"install:undo",
		"lint:fix",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"STYLELINT_LINT_SKIP_PATTERN",
		"STYLELINT_ALLOW_EMPTY_INPUT",
		"STYLELINT_CACHE",
		"STYLELINT_CONFIG",
		"STYLELINT_EXTRA_ARGS",
		"STYLELINT_TARGETS",
		"STYLELINT_VERSION",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"stylelint/node/nvm/pnpm",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
