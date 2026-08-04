// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package uv_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"install",
		"install:undo",
		"pip:install",
		"python:install",
		"run",
		"tool:install",
		"tool:upgrade",
		"upgrade",
		"venv",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"UV_ARGS",
		"UV_EXTRA_ARGS",
		"UV_FILE",
		"PYTHON_VERSION",
		"UV_REQUIREMENTS",
		"UV_TOOL",
		"UV_INSTALL_URL",
		"UV_INSTALL_URL_WINDOWS",
		"UV_LOAD",
		"UV_VERSION",
		"UV_VENV",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"uv",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
