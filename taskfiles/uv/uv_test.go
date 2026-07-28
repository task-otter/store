package uv_test

import (
	"os/exec"

	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"install",
	"install:undo",
	"pip:install",
	"python:install",
	"run",
	"tool:install",
	"tool:upgrade",
	"upgrade",
	"venv",
	"version",
}

var publicVars = []string{
	"ARGS",
	"EXTRA_ARGS",
	"FILE",
	"PYTHON_VERSION",
	"REQUIREMENTS",
	"TOOL",
	"UV_INSTALL_URL",
	"UV_INSTALL_URL_WINDOWS",
	"UV_LOAD",
	"UV_VERSION",
	"VENV",
}

func uvAvailable() bool {
	_, err := exec.LookPath("uv")
	return err == nil
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "uv", publicTasks, publicVars)
}
