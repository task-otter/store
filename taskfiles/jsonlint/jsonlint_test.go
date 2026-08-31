// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package jsonlint_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"jsonlint",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"ci",
	}
}

func publicVars() []string {
	return []string{
		"JSONLINT_EXTRA_ARGS",
		"JSONLINT_NIX_INSTALLABLE",
		"JSONLINT_TARGETS",
	}
}
