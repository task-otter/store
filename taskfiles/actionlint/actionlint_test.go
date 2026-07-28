package actionlint_test

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
	"ACTIONLINT_LINT_SKIP_PATTERN",
	"ACTIONLINT_EXTRA_ARGS",
	"ACTIONLINT_TARGETS",
	"ACTIONLINT_VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "actionlint", publicTasks, publicVars)
}
