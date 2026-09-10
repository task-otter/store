// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package cargo_test

import (
	"testing"

	"github.com/task-otter/store/internal/taskintegration"
	"github.com/task-otter/store/internal/tasktest"
)

const (
	cargoModuleName  = "cargo"
	installCrateTask = "install:crate"
)

// TestModuleIntegration runs the shared task CLI integration suite for this module.
func TestModuleIntegration(t *testing.T) {
	t.Parallel()

	taskintegration.RunHere(t)
}

// TestInstallCrateNotRunOnce
func TestInstallCrateNotRunOnce(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, cargoModuleName)
	task, exists := taskfile.Tasks[installCrateTask]

	if !exists {
		t.Fatal("install:crate task is missing")
	}

	// run: once would skip the second module's CARGO_CRATE in one
	// `task ci` invocation, leaving that crate uninstalled.
	if task.Run == "once" {
		t.Fatal("install:crate must not use run: once")
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		cargoModuleName,
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
		installCrateTask,
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
		"CARGO_LOAD",
		"CARGO_NIX_INSTALLABLE",
		"CARGO_WINGET_INSTALLABLE",
		"RUST_TOOLCHAIN",
	}
}
