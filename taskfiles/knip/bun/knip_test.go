// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package bun_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"knip/bun",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"ci",
		"ci:fix",
		"config:init",
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
		"KNIP_CONFIG",
		"KNIP_EXTRA_ARGS",
		"KNIP_VERSION",
	}
}
