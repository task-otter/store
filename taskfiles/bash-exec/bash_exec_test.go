// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package bashexec_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"check",
		"exec",
		"run",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"ARGS",
		"BASH_FLAGS",
		"COMMAND",
		"SCRIPT",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"bash-exec",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
