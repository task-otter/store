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
		"knip/node/yarn",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"ci",
		"ci:fix",
		"config:init",
		"config:skip",
		"dependencies",
		"dev-dependencies",
		"exports",
		"files",
		"help",
		"install",
		"install:undo",
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

// covered by module contract
