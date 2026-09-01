// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package ansible_test

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
		"ansible",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"galaxy:install",
		"install",
		"list:hosts",
		"ping",
		"run",
		"syntax:check",
		"vault:decrypt",
		"vault:encrypt",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"ANSIBLE_CONFIG",
		"ANSIBLE_EXTRA_ARGS",
		"ANSIBLE_FILE",
		"ANSIBLE_INVENTORY",
		"ANSIBLE_NIX_INSTALLABLE",
		"ANSIBLE_PATTERN",
		"ANSIBLE_PLAYBOOK",
		"ANSIBLE_REQUIREMENTS",
	}
}
