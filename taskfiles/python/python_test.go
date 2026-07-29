package python_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"install",
		"install:undo",
		"pip:install",
		"run",
		"upgrade",
		"verify",
		"version",
		"venv",
	}
}

func publicVars() []string {
	return []string{
		"ARGS",
		"EXTRA_ARGS",
		"FILE",
		"PYTHON_PIN_VERSION",
		"REQUIREMENTS",
		"UV_LOAD",
		"VENV",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "python", publicTasks(), publicVars())
}
