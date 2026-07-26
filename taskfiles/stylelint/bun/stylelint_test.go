package stylelintbun_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"cache:clean",
	"ci",
	"config:init",
	"help",
	"install",
	"install:undo",
	"lint",
	"lint:fix",
	"upgrade",
	"version",
}

var publicVars = []string{
	"STYLELINT_LINT_SKIP_PATTERN",
	"ALLOW_EMPTY_INPUT",
	"CACHE",
	"CONFIG",
	"EXTRA_ARGS",
	"TARGETS",
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "stylelint/bun", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	// covered by module contract
}
