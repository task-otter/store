// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package ansible_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"galaxy:install",
		"list:hosts",
		"ping",
		"run",
		"syntax:check",
		"vault:decrypt",
		"vault:encrypt",
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

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"ansible",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
