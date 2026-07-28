package prettiernodefnmpnpm_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"ci",
	"config:init",
	"fix",
	"fmt",
	"fmt:check",
	"help",
	"install",
	"install:undo",
	"upgrade",
	"version",
}

var publicVars = []string{
	"PRETTIER_FMT_SKIP_PATTERN",
	"CONFIG",
	"EXTRA_ARGS",
	"IGNORE_PATH",
	"TARGETS",
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "prettier/node/fnm/pnpm", publicTasks, publicVars)
}

func TestConfigInitDryRun(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "prettier/node/fnm/pnpm", []string{"config:init"},
		"singleQuote",
		".prettierrc.json",
	)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "prettier/node/fnm/pnpm",
		[]string{"fmt:check", "TARGETS=src/**/*.ts", "CONFIG=.prettierrc.json"},
		"pnpm:exec",
		"src/**/*.ts --check",
		".prettierrc.json",
	)
}
