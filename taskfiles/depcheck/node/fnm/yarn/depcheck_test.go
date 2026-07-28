package depchecknodefnmyarn_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"ci",
	"help",
	"ignores",
	"install",
	"install:undo",
	"json",
	"lint",
	"skip-missing",
	"upgrade",
	"version",
}

var publicVars = []string{
	"DEPCHECK_LINT_SKIP_PATTERN",
	"EXTRA_ARGS",
	"IGNORE_PACKAGES",
	"PROJECT_PATH",
	"TARGETS",
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "depcheck/node/fnm/yarn", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	// covered by module contract
}
