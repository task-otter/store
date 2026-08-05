// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package yamllint_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"ci",
		"ci:fix",
		"config:init",
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
		"YAMLLINT_LINT_SKIP_PATTERN",
		"YAMLLINT_CONFIG",
		"YAMLLINT_EXTRA_ARGS",
		"YAMLLINT_TARGETS",
		"UV_LOAD",
		"YAMLFIX_VERSION",
		"YAMLLINT_VERSION",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"yamllint",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
