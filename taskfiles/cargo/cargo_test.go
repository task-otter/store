// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package cargo_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

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
		"lint",
		"lint:fix",
		"test",
		"verify",
		"which",
	}
}

func publicVars() []string {
	return []string{
		"CARGO_EXTRA_ARGS",
		"CARGO_NIX_INSTALLABLE",
		"RUST_TOOLCHAIN",
	}
}
