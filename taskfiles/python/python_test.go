// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package python_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/taskintegration"
	"github.com/task-otter/store/internal/tasktest"
)

const (
	pythonModuleName   = "python"
	installWindowsTask = "_install:windows"
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
		pythonModuleName,
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

// TestInstallWindowsStatusUsesPythonVersion proves Store python.exe aliases
// cannot skip winget: the Windows install status runs python --version.
func TestInstallWindowsStatusUsesPythonVersion(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, pythonModuleName)
	task, ok := taskfile.Tasks[installWindowsTask]

	if !ok {
		t.Fatal("_install:windows is missing")
	}

	status := fmt.Sprintf("%v", task.Status)

	if !strings.Contains(status, "python --version") {
		t.Fatalf("Windows install status must run python --version, got %s", status)
	}

	if strings.Contains(status, "Get-Command python") {
		t.Fatal("Windows install status must not use Get-Command python")
	}
}

func publicTasks() []string {
	return []string{
		"install",
		"pip:install",
		"run",
		"venv",
		"verify",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"PYTHON_ARGS",
		"PYTHON_EXTRA_ARGS",
		"PYTHON_FILE",
		"PYTHON_NIX_INSTALLABLE",
		"PYTHON_WINGET_INSTALLABLE",
		"PYTHON_REQUIREMENTS",
		"PYTHON_VENV",
	}
}
