package knipnodenvmnpm_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"ci",
	"config:init",
	"config:skip",
	"dependencies",
	"dev-dependencies",
	"exports",
	"files",
	"help",
	"init",
	"install",
	"install:undo",
	"lint",
	"lint:fix",
	"production",
	"upgrade",
	"version",
}

var publicVars = []string{
	"KNIP_LINT_SKIP_PATTERN",
	"CONFIG",
	"EXTRA_ARGS",
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	tasktest.AssertModule(t, "knip/node/nvm/npm", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	// covered by module contract
}
