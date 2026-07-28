package sqlfluff_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"config:init",
	"config:skip",
	"fix",
	"install",
	"install:undo",
	"lint",
	"parse",
	"upgrade",
	"version",
}

var publicVars = []string{
	"SQLFLUFF_LINT_SKIP_PATTERN",
	"SQLFLUFF_VERSION",
	"UV_LOAD",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "sqlfluff", publicTasks, publicVars)
}
