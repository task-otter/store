// Copyright 2026 task-otter
// SPDX-License-Identifier: Apache-2.0

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
		"ARGS",
		"EXTRA_ARGS",
		"FILE",
		"PYTHON_VERSION",
		"REQUIREMENTS",
		"TOOL",
		"UV_INSTALL_URL",
		"UV_INSTALL_URL_WINDOWS",
		"UV_LOAD",
		"UV_VERSION",
		"VENV",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "uv", publicTasks(), publicVars())
}
