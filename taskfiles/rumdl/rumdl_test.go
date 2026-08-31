// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package rumdl_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"lint:fix",
		"ci:fix",
		"fmt",
		"ci",
	}
}

func publicVars() []string {
	return []string{
		"RUMDL_LINT_SKIP_PATTERN",
		"RUMDL_FMT_SKIP_PATTERN",
		"RUMDL_EXTRA_ARGS",
		"RUMDL_NIX_INSTALLABLE",
		"RUMDL_TARGETS",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"rumdl",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
