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
		"ci",
		"ci:fix",
		"config:init",
		"config:skip",
		"help",
		"install",
		"install:undo",
		"fmt",
		"fmt:check",
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
		"BIOME_CONFIG",
		"BIOME_EXTRA_ARGS",
		"BIOME_TARGETS",
		"BIOME_VERSION",
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
