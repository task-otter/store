// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package bencher_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"exec",
		"run",
	}
}

func publicVars() []string {
	return []string{
		"BENCHER_NIX_INSTALLABLE",
		"BENCHER_EXTRA_ARGS",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"bencher",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
