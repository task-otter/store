package spectralnodefnmpnpm_test

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
	"SPECTRAL_LINT_SKIP_PATTERN",
	"EXTRA_ARGS",
	"RULESET",
	"TARGETS",
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "spectral/node/fnm/pnpm", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "spectral/node/fnm/pnpm",
		[]string{"lint", "TARGETS=openapi.yaml"},
		"pnpm:exec",
		"spectral",
		"lint",
		"openapi.yaml",
	)

	tasktest.AssertDryRunContains(t, "spectral/node/fnm/pnpm",
		[]string{"version"},
		"spectral",
		"--version",
	)
}
