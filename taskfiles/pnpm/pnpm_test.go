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
	pnpmCmdShim        = "pnpm.cmd"
	cmdExeInvoke       = "cmd /c"
	barePnpmArgs       = "pnpm {{.ARGS}}"
	pnpmExecPrefix     = "exec --"
	nodeModulesBinPath = `node_modules\.bin`
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

// TestPnpmWindowsUsesCmdShim proves Windows pnpm runs via cmd /c pnpm.cmd,
// not a bare pnpm invocation that PowerShell would splat or wrap as .ps1.
func TestPnpmWindowsUsesCmdShim(t *testing.T) {
	t.Parallel()

	cmds := mustTaskCmds(t, pnpmWindowsTask)

	if !strings.Contains(cmds, pnpmCmdShim) {
		t.Fatalf("_pnpm:windows must invoke pnpm.cmd, got %s", cmds)
	}

	if !strings.Contains(cmds, cmdExeInvoke) {
		t.Fatalf("_pnpm:windows must invoke pnpm via cmd /c, got %s", cmds)
	}

	if strings.Contains(cmds, barePnpmArgs) {
		t.Fatalf("_pnpm:windows must not use bare pnpm {{.ARGS}}, got %s", cmds)
	}

	if !strings.Contains(cmds, "USERPROFILE") {
		t.Fatalf("_pnpm:windows must align HOME with USERPROFILE, got %s", cmds)
	}
}

// TestExecWindowsUsesPnpmExec proves Windows exec uses pnpm exec -- through
// the pnpm.cmd helper instead of prepending node_modules\.bin and & binary.
func TestExecWindowsUsesPnpmExec(t *testing.T) {
	t.Parallel()

	cmds := mustTaskCmds(t, execWindowsTask)

	if !strings.Contains(cmds, pnpmExecPrefix) {
		t.Fatalf("_exec:windows must run pnpm exec --, got %s", cmds)
	}

	if !strings.Contains(cmds, pnpmWindowsTask) {
		t.Fatalf("_exec:windows must use the _pnpm:windows helper, got %s", cmds)
	}

	if strings.Contains(cmds, nodeModulesBinPath) {
		t.Fatalf("_exec:windows must not prepend node_modules\\.bin, got %s", cmds)
	}
}

func mustTaskCmds(t *testing.T, taskName string) string {
	t.Helper()

	taskfile := tasktest.LoadTaskfile(t, pnpmModuleName)
	task, ok := taskfile.Tasks[taskName]

	if !ok {
		t.Fatalf("%s is missing", taskName)
	}

	return fmt.Sprintf("%v", task.Cmds)
}
