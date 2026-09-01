// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package buf_test

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
		"buf",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"breaking",
		"ci",
		"ci:fix",
		"fmt:check",
		"generate",
		"install",
		"lint",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"BUF_AGAINST",
		"BUF_NIX_INSTALLABLE",
		"BUF_CONFIG",
		"BUF_EXTRA_ARGS",
		"BUF_INPUT",
	}
}
