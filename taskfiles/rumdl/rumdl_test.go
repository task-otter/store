// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package rumdl_test

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
		"rumdl",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

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
		"RUMDL_EXTRA_ARGS",
		"RUMDL_NIX_INSTALLABLE",
		"RUMDL_TARGETS",
	}
}
