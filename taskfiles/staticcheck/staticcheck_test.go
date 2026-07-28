package staticcheck_test

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
	"STATICCHECK_LINT_SKIP_PATTERN",
	"GLOBAL_GO_BIN",
	"STATICCHECK_RELEASE_BASE_URL",
	"STATICCHECK_VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "staticcheck", publicTasks, publicVars)
}
