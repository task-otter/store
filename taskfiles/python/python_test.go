// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package python_test

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
		"python",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"install",
		"pip:install",
		"run",
		"venv",
		"verify",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"PYTHON_ARGS",
		"PYTHON_EXTRA_ARGS",
		"PYTHON_FILE",
		"PYTHON_NIX_INSTALLABLE",
		"PYTHON_REQUIREMENTS",
		"PYTHON_VENV",
	}
}
