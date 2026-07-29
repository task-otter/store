package biomebun_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
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
}

func publicVars() []string {
	return []string{
		"BIOME_LINT_SKIP_PATTERN",
		"BIOME_FMT_SKIP_PATTERN",
		"CONFIG",
		"EXTRA_ARGS",
		"TARGETS",
		"VERSION",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "biome/bun", publicTasks(), publicVars())
}

// covered by module contract
