package htmlhintnodenvmpnpm_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"config:init",
	"help",
	"install",
	"install:undo",
	"lint",
	"upgrade",
	"version",
}

var publicVars = []string{
	"HTMLHINT_LINT_SKIP_PATTERN",
	"CONFIG",
	"EXTRA_ARGS",
	"TARGETS",
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "htmlhint/node/nvm/pnpm", publicTasks, publicVars)
}
