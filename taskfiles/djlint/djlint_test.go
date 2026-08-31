// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package djlint_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"ci",
		"ci:fix",
		"fmt:check",
		"lint",
	}
}

func publicVars() []string {
	return []string{
		"DJLINT_EXTRA_ARGS",
		"DJLINT_FMT_SKIP_PATTERN",
		"DJLINT_LINT_SKIP_PATTERN",
		"DJLINT_NIX_INSTALLABLE",
		"DJLINT_TARGETS",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"djlint",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
