// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package bun_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"cache:clean",
		"check",
		"check:write",
		"ci",
		"config:init",
		"config:skip",
		"fix",
		"fmt",
		"fmt:check",
		"help",
		"init",
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
		"BIOME_LINT_SKIP_PATTERN",
		"BIOME_FMT_SKIP_PATTERN",
		"CONFIG",
		"EXTRA_ARGS",
		"TARGETS",
		"VERSION",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"biome/bun",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

// covered by module contract
