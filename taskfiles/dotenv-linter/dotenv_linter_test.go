package dotenv_linter_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"diff",
	"fix",
	"install",
	"install:undo",
	"lint",
	"upgrade",
	"version",
}

var publicVars = []string{
	"DOTENV_LINTER_LINT_SKIP_PATTERN",
	"CARGO_BIN_UNIX",
	"DOTENV_LINTER_VERSION",
	"EXTRA_ARGS",
	"TARGETS",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "dotenv-linter", publicTasks, publicVars)
}
