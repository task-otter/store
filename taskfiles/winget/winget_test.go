// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package winget_test

import (
	"testing"

	"github.com/task-otter/store/internal/taskintegration"
	"github.com/task-otter/store/internal/tasktest"
)

const (
	wingetModuleName   = "winget"
	installPackageTask = "install:package"
)

// TestModuleIntegration runs the shared task CLI integration suite for this module.
func TestModuleIntegration(t *testing.T) {
	t.Parallel()

	taskintegration.RunHere(t)
}

// TestInstallPackageNotRunOnce
func TestInstallPackageNotRunOnce(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, wingetModuleName)
	task, exists := taskfile.Tasks[installPackageTask]

	if !exists {
		t.Fatal("install:package task is missing")
	}

	// run: once would skip the second module's WINGET_INSTALLABLE in one
	// `task ci` invocation, leaving that CLI off PATH.
	if task.Run == "once" {
		t.Fatal("install:package must not use run: once")
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		wingetModuleName,
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"install",
		installPackageTask,
		"install:undo",
		"uninstall",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"WINGET_EXTRA_ARGS",
		"WINGET_INSTALLABLE",
		"WINGET_LOAD",
		"WINGET_SCOPE",
		"WINGET_SOURCE",
		"WINGET_VERSION",
	}
}
