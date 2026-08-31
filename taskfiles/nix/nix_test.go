// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package nix_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"nix",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"features:enable",
		"features:show",
		"install",
		"install:profile",
		"install:shell",
		"install:undo",
		"uninstall",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"NIX_CHANNEL",
		"NIX_COMMAND",
		"NIX_CONF",
		"NIX_EXPERIMENTAL_FEATURES",
		"NIX_EXPERIMENTAL_FEATURES_ALL",
		"NIX_EXTRA_ARGS",
		"NIX_INSTALL_URL",
		"NIX_INSTALLABLE",
		"NIX_LOAD",
		"NIX_NEEDED_FEATURES",
		"NIX_VERSION",
	}
}

// TestInstallProfileNotRunOnce
func TestInstallProfileNotRunOnce(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "nix")
	task, exists := taskfile.Tasks["install:profile"]

	if !exists {
		t.Fatal("install:profile task is missing")
	}

	// run: once would skip the second module's NIX_INSTALLABLE in one
	// `task ci` invocation, leaving that CLI off PATH.
	if task.Run == "once" {
		t.Fatal("install:profile must not use run: once")
	}
}
