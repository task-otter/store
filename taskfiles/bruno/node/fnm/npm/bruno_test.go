// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package npm_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"ci",
		"help",
		"install",
		"install:undo",
		"run",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"COLLECTION",
		"ENV",
		"EXTRA_ARGS",
		"VERSION",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"bruno/node/fnm/npm",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
