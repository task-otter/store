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
