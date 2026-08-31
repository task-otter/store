// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package npm_test

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
		"eslint/node/npm",
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
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"ESLINT_CACHE",
		"ESLINT_CONFIG",
		"ESLINT_EXTRA_ARGS",
		"ESLINT_TARGETS",
		"ESLINT_VERSION",
	}
}
