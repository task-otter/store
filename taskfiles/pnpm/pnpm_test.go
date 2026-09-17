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
	pnpmWindowsTask    = "_pnpm:windows"
	execWindowsTask    = "_exec:windows"
	pnpmWindowsExe     = "pnpm.exe"
	pnpmExecPrefix     = "exec --"
	nodeModulesBinPath = `node_modules\.bin`
	fmtPercentV        = "%v"
)

// TestModuleIntegration runs the shared task CLI integration suite for this module.
func TestModuleIntegration(t *testing.T) {
	t.Parallel()

	taskintegration.RunHere(t)
}

// TestInstallWindowsUsesNpx proves Windows installation uses npm's npx flow.
func TestInstallWindowsStatusUsesPnpmVersion(t *testing.T) {
	t.Parallel()

	status := mustTaskStatus(t, installWindowsTask)

	assertContains(t, status, "pnpm --version")

	cmds := mustTaskCmds(t, installWindowsTask)
	assertContains(t, cmds, "npx get-pnpm")
}

// TestPnpmWindowsUsesExe proves Windows pnpm runs via the executable shipped
// by the executable installed through npx, not an unavailable .cmd shim.
func TestPnpmWindowsUsesExe(t *testing.T) {
	t.Parallel()

	cmds := mustTaskCmds(t, pnpmWindowsTask)

	assertContains(t, cmds, pnpmWindowsExe)
	assertContains(t, cmds, "USERPROFILE")
}

// TestExecWindowsUsesPnpmExec proves Windows exec uses pnpm exec -- through
// _pnpm:windows instead of prepending node_modules\.bin and & binary.
func TestExecWindowsUsesPnpmExec(t *testing.T) {
	t.Parallel()

	cmds := mustTaskCmds(t, execWindowsTask)

	assertContains(t, cmds, pnpmExecPrefix)
	assertContains(t, cmds, pnpmWindowsTask)
	assertNotContains(t, cmds, nodeModulesBinPath)
}

func mustModuleTask(t *testing.T, name string) *tasktest.Task {
	t.Helper()

	taskfile := tasktest.LoadTaskfile(t, pnpmModuleName)
	task, ok := taskfile.Tasks[name]

	if !ok {
		t.Fatalf("%s is missing", name)
	}

	return task
}

func mustTaskCmds(t *testing.T, taskName string) string {
	t.Helper()

	return fmt.Sprintf(fmtPercentV, mustModuleTask(t, taskName).Cmds)
}

func mustTaskStatus(t *testing.T, taskName string) string {
	t.Helper()

	return fmt.Sprintf(fmtPercentV, mustModuleTask(t, taskName).Status)
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()

	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected output to contain %q:\n%s", needle, haystack)
	}
}

func assertNotContains(t *testing.T, haystack, needle string) {
	t.Helper()

	if strings.Contains(haystack, needle) {
		t.Fatalf("expected output not to contain %q:\n%s", needle, haystack)
	}
}
