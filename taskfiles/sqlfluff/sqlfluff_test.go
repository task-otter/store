package sqlfluff_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
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
}

func publicVars() []string {
	return []string{
		"SQLFLUFF_LINT_SKIP_PATTERN",
		"SQLFLUFF_VERSION",
		"UV_LOAD",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "sqlfluff", publicTasks(), publicVars())
}
