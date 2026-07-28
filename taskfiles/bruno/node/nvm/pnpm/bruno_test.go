package brunonodenvmpnpm_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"ci",
	"help",
	"install",
	"install:undo",
	"run",
	"upgrade",
	"version",
}

var publicVars = []string{
	"COLLECTION",
	"ENV",
	"EXTRA_ARGS",
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "bruno/node/nvm/pnpm", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "bruno/node/nvm/pnpm",
		[]string{"run"},
		"pnpm:exec",
		`run .`,
	)
}
