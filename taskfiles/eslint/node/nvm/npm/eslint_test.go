package eslintnodenvmnpm_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"cache:clean",
	"ci",
	"config:init",
	"help",
	"init",
	"install",
	"install:undo",
	"lint",
	"lint:fix",
	"upgrade",
	"version",
}

var publicVars = []string{
	"ESLINT_LINT_SKIP_PATTERN",
	"CACHE",
	"CONFIG",
	"EXTRA_ARGS",
	"TARGETS",
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "eslint/node/nvm/npm", publicTasks, publicVars)
}
