package buf_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"breaking",
	"fmt",
	"fmt:check",
	"generate",
	"install",
	"install:undo",
	"lint",
	"upgrade",
	"version",
}

var publicVars = []string{
	"BUF_LINT_SKIP_PATTERN",
	"BUF_FMT_SKIP_PATTERN",
	"AGAINST",
	"BUF_VERSION",
	"CONFIG",
	"EXTRA_ARGS",
	"INPUT",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "buf", publicTasks, publicVars)
}
