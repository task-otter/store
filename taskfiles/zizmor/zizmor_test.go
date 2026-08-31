// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package zizmor_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"zizmor",
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
		"ZIZMOR_EXTRA_ARGS",
		"ZIZMOR_TARGETS",
		"ZIZMOR_NIX_INSTALLABLE",
	}
}
