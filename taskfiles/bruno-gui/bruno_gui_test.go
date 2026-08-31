// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package bruno_gui_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"help",
		"open",
	}
}

func publicVars() []string {
	return []string{
		"BRUNO_GUI_COLLECTION",
		"BRUNO_GUI_EXTRA_ARGS",
		"BRUNO_GUI_NIX_INSTALLABLE",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"bruno-gui",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
