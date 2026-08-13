// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package npm_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"config:init",
		"help",
		"install",
		"install:undo",
		"ci",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"SPECTRAL_LINT_SKIP_PATTERN",
		"SPECTRAL_EXTRA_ARGS",
		"SPECTRAL_RULESET",
		"SPECTRAL_TARGETS",
		"SPECTRAL_VERSION",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"spectral/node/nvm/npm",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
