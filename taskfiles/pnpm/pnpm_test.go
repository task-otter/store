// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package pnpm_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/taskintegration"
	"github.com/task-otter/store/internal/tasktest"
)

const (
	pnpmModuleName     = "pnpm"
	installWindowsTask = "_install:windows"
)

// TestModuleIntegration runs the shared task CLI integration suite for this module.
func TestModuleIntegration(t *testing.T) {
	t.Parallel()

	taskintegration.RunHere(t)
}

// TestInstallWindowsStatusUsesPnpmVersion proves Corepack shims cannot skip
// winget: the Windows install status runs pnpm --version after WINGET_LOAD.
func TestInstallWindowsStatusUsesPnpmVersion(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, pnpmModuleName)
	task, ok := taskfile.Tasks[installWindowsTask]

	if !ok {
		t.Fatal("_install:windows is missing")
	}

	status := fmt.Sprintf("%v", task.Status)

	if !strings.Contains(status, "WINGET_LOAD") {
		t.Fatalf("Windows install status must use WINGET_LOAD, got %s", status)
	}

	if !strings.Contains(status, "pnpm --version") {
		t.Fatalf("Windows install status must run pnpm --version, got %s", status)
	}
}
