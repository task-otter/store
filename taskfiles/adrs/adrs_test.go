// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package adrs_test

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
		"adrs",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"exec",
		"generate",
		"init",
		"list",
		"new",
	}
}

func publicVars() []string {
	return []string{
		"ADRS_EXTRA_ARGS",
		"ADRS_NIX_INSTALLABLE",
	}
}
