// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package dotenvlinter_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"diff",
		"install",
		"install:undo",
		"ci",
		"ci:fix",
		"lint:fix",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"DOTENV_LINTER_LINT_SKIP_PATTERN",
		"CARGO_BIN_UNIX",
		"DOTENV_LINTER_VERSION",
		"DOTENV_LINTER_EXTRA_ARGS",
		"DOTENV_LINTER_TARGETS",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"dotenv-linter",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
