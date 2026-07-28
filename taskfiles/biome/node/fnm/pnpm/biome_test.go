package biomenodefnmpnpm_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"cache:clean",
	"check",
	"check:write",
	"ci",
	"config:init",
	"config:skip",
	"fix",
	"fmt",
	"fmt:check",
	"help",
	"init",
	"install",
	"install:undo",
	"lint",
	"lint:fix",
	"upgrade",
	"version",
}

var publicVars = []string{
	"BIOME_LINT_SKIP_PATTERN",
	"BIOME_FMT_SKIP_PATTERN",
	"CONFIG",
	"EXTRA_ARGS",
	"TARGETS",
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "biome/node/fnm/pnpm", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "biome/node/fnm/pnpm",
		[]string{"fmt", "--", "--no-errors-on-unmatched"},
		"pnpm:exec",
		"format --write",
		"--no-errors-on-unmatched",
	)
}
