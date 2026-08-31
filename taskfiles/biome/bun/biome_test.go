// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package bun_test

import (
	"testing"

	"github.com/task-otter/store/internal/taskintegration"
	"github.com/task-otter/store/internal/tasktest"
)

// TestModuleIntegration runs the shared task CLI integration suite for this module.
func TestModuleIntegration(t *testing.T) {
	t.Parallel()

	taskintegration.RunHere(t)
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

func publicTasks() []string {
	return []string{
		"cache:clean",
		"ci",
		"ci:fix",
		"config:init",
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
		"BIOME_CONFIG",
		"BIOME_EXTRA_ARGS",
		"BIOME_TARGETS",
		"BIOME_VERSION",
	}
}

// covered by module contract
