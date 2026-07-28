package bash_exec_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"check",
	"exec",
	"run",
	"version",
}

var publicVars = []string{
	"ARGS",
	"BASH_FLAGS",
	"COMMAND",
	"SCRIPT",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "bash-exec", publicTasks, publicVars)
}
