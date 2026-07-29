package dotenv_linter_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"diff",
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
		"DOTENV_LINTER_LINT_SKIP_PATTERN",
		"CARGO_BIN_UNIX",
		"DOTENV_LINTER_VERSION",
		"EXTRA_ARGS",
		"TARGETS",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "dotenv-linter", publicTasks(), publicVars())
}
