// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package yarn_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"ci",
		"config:init",
		"config:skip",
		"dependencies",
		"dev-dependencies",
		"exports",
		"files",
		"help",
		"init",
		"install",
		"install:undo",
		"lint",
		"lint:fix",
		"production",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"KNIP_LINT_SKIP_PATTERN",
		"KNIP_CONFIG",
		"KNIP_EXTRA_ARGS",
		"KNIP_VERSION",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"knip/node/fnm/yarn",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

// covered by module contract
