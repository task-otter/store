package biomenodenvmyarn_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"cache:clean",
	"check",
	"check:write",
	"ci",
	"config:init",
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
	tasktest.AssertModule(t, "biome/node/nvm/yarn", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	tasktest.AssertDryRunContains(t, "biome/node/nvm/yarn",
		[]string{"ci", "CONFIG=biome.json", "TARGETS=src"},
		"yarn:exec",
		"biome.json",
		"ci",
		"src",
	)
}
