package adrs_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"exec",
	"generate",
	"init",
	"install",
	"install:undo",
	"list",
	"new",
	"upgrade",
	"version",
}

var publicVars = []string{
	"ADRS_VERSION",
	"CARGO_BIN_UNIX",
	"EXTRA_ARGS",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "adrs", publicTasks, publicVars)
}
