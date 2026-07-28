package jsonlint_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"install",
	"install:undo",
	"lint",
	"upgrade",
	"version",
}

var publicVars = []string{
	"JSONLINT_LINT_SKIP_PATTERN",
	"EXTRA_ARGS",
	"JSONLINT_VERSION",
	"TARGETS",
	"UV_LOAD",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "jsonlint", publicTasks, publicVars)
}
