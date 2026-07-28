package htmlhintnodefnmpnpm_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"config:init",
	"help",
	"install",
	"install:undo",
	"lint",
	"upgrade",
	"version",
}

var publicVars = []string{
	"HTMLHINT_LINT_SKIP_PATTERN",
	"CONFIG",
	"EXTRA_ARGS",
	"TARGETS",
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "htmlhint/node/fnm/pnpm", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "htmlhint/node/fnm/pnpm",
		[]string{"lint", "TARGETS=src/**/*.html"},
		"pnpm:exec",
		"htmlhint",
		"src/**/*.html",
	)

	tasktest.AssertDryRunContains(t, "htmlhint/node/fnm/pnpm",
		[]string{"version"},
		"htmlhint",
		"--version",
	)
}
