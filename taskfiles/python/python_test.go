package python_test

import (
	"os/exec"

	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"install",
	"install:undo",
	"pip:install",
	"run",
	"upgrade",
	"verify",
	"version",
	"venv",
}

var publicVars = []string{
	"ARGS",
	"EXTRA_ARGS",
	"FILE",
	"PYTHON_PIN_VERSION",
	"REQUIREMENTS",
	"UV_LOAD",
	"VENV",
}

func pythonAvailable() bool {
	_, err := exec.LookPath("python3")
	if err == nil {
		return true
	}
	_, err = exec.LookPath("python")
	return err == nil
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "python", publicTasks, publicVars)
}
