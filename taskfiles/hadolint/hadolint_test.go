package hadolint_test

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
	"HADOLINT_LINT_SKIP_PATTERN",
	"CONFIG",
	"DOCKERFILE",
	"EXTRA_ARGS",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "hadolint", publicTasks, publicVars)
}
