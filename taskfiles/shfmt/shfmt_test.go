// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package shfmt_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"ci:fix",
		"fmt:check",
	}
}

func publicVars() []string {
	return []string{
		"SHFMT_EXTRA_ARGS",
		"SHFMT_NIX_INSTALLABLE",
		"SHFMT_TARGETS",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"shfmt",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
