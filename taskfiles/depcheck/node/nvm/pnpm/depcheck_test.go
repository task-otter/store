package depchecknodenvmpnpm_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
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

	tasktest.AssertModule(t, "depcheck/node/nvm/pnpm", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "depcheck/node/nvm/pnpm",
		[]string{"ignores", "IGNORE_PACKAGES=@types/*,eslint-*"},
		"pnpm:exec",
		"@types/*,eslint-*",
	)
}
