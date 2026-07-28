package yamllint_test

import (
	"os/exec"

	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"ci",
	"config:init",
	"install",
	"install:undo",
	"lint",
	"lint:fix",
	"upgrade",
	"version",
}

var publicVars = []string{
	"YAMLLINT_LINT_SKIP_PATTERN",
	"CONFIG",
	"EXTRA_ARGS",
	"TARGETS",
	"UV_LOAD",
	"YAMLFIX_VERSION",
	"YAMLLINT_VERSION",
}

func yamllintAvailable() bool {
	_, err := exec.LookPath("yamllint")
	return err == nil
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "yamllint", publicTasks, publicVars)
}
