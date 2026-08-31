// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package bruno_cli_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"ci",
		"help",
		"run",
	}
}

func publicVars() []string {
	return []string{
		"BRUNO_CLI_COLLECTION",
		"BRUNO_CLI_ENV",
		"BRUNO_CLI_EXTRA_ARGS",
		"BRUNO_CLI_NIX_INSTALLABLE",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"bruno-cli",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
