// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package sqlfluff_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"sqlfluff",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"ci",
		"ci:fix",
		"config:init",
		"parse",
	}
}

func publicVars() []string {
	return []string{
		"SQLFLUFF_NIX_INSTALLABLE",
	}
}
