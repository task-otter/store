package zizmor_test

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
	"ZIZMOR_LINT_SKIP_PATTERN",
	"ZIZMOR_EXTRA_ARGS",
	"ZIZMOR_TARGETS",
	"ZIZMOR_VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "zizmor", publicTasks, publicVars)
}
