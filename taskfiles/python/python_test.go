// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package python_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"install",
		"install:undo",
		"pip:install",
		"run",
		"upgrade",
		"verify",
		"version",
		"venv",
	}
}

func publicVars() []string {
	return []string{
		"PYTHON_ARGS",
		"PYTHON_EXTRA_ARGS",
		"PYTHON_FILE",
		"PYTHON_PIN_VERSION",
		"PYTHON_REQUIREMENTS",
		"UV_LOAD",
		"PYTHON_VENV",
	}
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
