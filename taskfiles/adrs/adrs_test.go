// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package adrs_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"exec",
		"generate",
		"init",
		"list",
		"new",
	}
}

func publicVars() []string {
	return []string{
		"ADRS_EXTRA_ARGS",
		"ADRS_NIX_INSTALLABLE",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"adrs",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
