package protolint_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"fix",
		"install",
		"install:undo",
		"lint",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"PROTOLINT_LINT_SKIP_PATTERN",
		"EXTRA_ARGS",
		"GLOBAL_GO_BIN",
		"PROTOLINT_VERSION",
		"TARGETS",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "protolint", publicTasks(), publicVars())
}
