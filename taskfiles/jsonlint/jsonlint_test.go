// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package jsonlint_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"install",
		"install:undo",
		"lint",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"JSONLINT_LINT_SKIP_PATTERN",
		"EXTRA_ARGS",
		"JSONLINT_VERSION",
		"TARGETS",
		"UV_LOAD",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"jsonlint",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
