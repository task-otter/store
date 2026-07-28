package brunonodenvmpnpm_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"ci",
	"help",
	"install",
	"install:undo",
	"run",
	"upgrade",
	"version",
}

var publicVars = []string{
	"COLLECTION",
	"ENV",
	"EXTRA_ARGS",
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "bruno/node/nvm/pnpm", publicTasks, publicVars)
}
