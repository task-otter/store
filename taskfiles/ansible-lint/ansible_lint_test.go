// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package ansiblelint_test

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
		"ansible-lint",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"ci",
		"ci:fix",
		"install",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"ANSIBLE_LINT_CONFIG",
		"ANSIBLE_LINT_EXTRA_ARGS",
		"ANSIBLE_LINT_NIX_INSTALLABLE",
		"ANSIBLE_LINT_TARGETS",
	}
}
