// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package cargo_test

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
		"cargo",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"build",
		"check",
		"ci",
		"ci:fix",
		"fmt",
		"fmt:check",
		"install",
		"install:crate",
		"lint",
		"lint:fix",
		"test",
		"verify",
		"version",
		"which",
	}
}

func publicVars() []string {
	return []string{
		"CARGO_CRATE",
		"CARGO_EXTRA_ARGS",
		"CARGO_NIX_INSTALLABLE",
		"CARGO_WINGET_INSTALLABLE",
		"RUST_TOOLCHAIN",
	}
}
