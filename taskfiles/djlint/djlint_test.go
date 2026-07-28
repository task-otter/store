package djlint_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"fmt",
	"fmt:check",
	"install",
	"install:undo",
	"lint",
	"upgrade",
	"version",
}

var publicVars = []string{
	"DJLINT_LINT_SKIP_PATTERN",
	"DJLINT_FMT_SKIP_PATTERN",
	"DJLINT_VERSION",
	"EXTRA_ARGS",
	"TARGETS",
	"UV_LOAD",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "djlint", publicTasks, publicVars)
}
