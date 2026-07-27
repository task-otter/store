package biomebun_test

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
	tasktest.AssertModule(t, "biome/bun", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	// covered by module contract
}
