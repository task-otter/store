// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package djlint_test

import (
	"testing"

	"github.com/task-otter/store/internal/taskintegration"
	"github.com/task-otter/store/internal/tasktest"
)

// TestModuleIntegration runs the shared task CLI integration suite for this module.
func TestModuleIntegration(t *testing.T) {
	t.Parallel()

	taskintegration.RunHere(t)
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
		"DJLINT_NIX_INSTALLABLE",
		"DJLINT_TARGETS",
	}
}
