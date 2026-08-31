// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package nodejs_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{}
}

func publicVars() []string {
	return []string{
		"NODEJS_NIX_INSTALLABLE",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"nodejs",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
