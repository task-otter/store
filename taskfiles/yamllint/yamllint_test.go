// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package yamllint_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"yamllint",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"ci",
		"config:init",
	}
}

func publicVars() []string {
	return []string{
		"YAMLLINT_CONFIG",
		"YAMLLINT_EXTRA_ARGS",
		"YAMLLINT_NIX_INSTALLABLE",
		"YAMLLINT_TARGETS",
	}
}
