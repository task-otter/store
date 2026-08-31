// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package yarn_test

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
		"htmlhint/node/yarn",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"config:init",
		"help",
		"install",
		"install:undo",
		"ci",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"HTMLHINT_CONFIG",
		"HTMLHINT_EXTRA_ARGS",
		"HTMLHINT_TARGETS",
		"HTMLHINT_VERSION",
	}
}
