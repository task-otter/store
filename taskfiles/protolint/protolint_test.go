// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package protolint_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"ci",
		"ci:fix",
	}
}

func publicVars() []string {
	return []string{
		"PROTOLINT_EXTRA_ARGS",
		"PROTOLINT_NIX_INSTALLABLE",
		"PROTOLINT_TARGETS",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"protolint",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
