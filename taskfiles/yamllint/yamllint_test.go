// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package yamllint_test

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
		"yamllint",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"ci",
		"config:init",
		"install",
		"version",
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
