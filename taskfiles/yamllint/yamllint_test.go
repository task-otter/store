package yamllint_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"ci",
		"config:init",
		"install",
		"install:undo",
		"lint",
		"lint:fix",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"YAMLLINT_LINT_SKIP_PATTERN",
		"CONFIG",
		"EXTRA_ARGS",
		"TARGETS",
		"UV_LOAD",
		"YAMLFIX_VERSION",
		"YAMLLINT_VERSION",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "yamllint", publicTasks(), publicVars())
}
