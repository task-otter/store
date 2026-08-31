// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package dotenvlinter_test

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
		"dotenv-linter",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"ci",
		"ci:fix",
		"diff",
	}
}

func publicVars() []string {
	return []string{
		"DOTENV_LINTER_EXTRA_ARGS",
		"DOTENV_LINTER_NIX_INSTALLABLE",
		"DOTENV_LINTER_TARGETS",
	}
}
