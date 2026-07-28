package spectralnodenvmnpm_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
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

	tasktest.AssertModule(t, "spectral/node/nvm/npm", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "spectral/node/nvm/npm",
		[]string{"lint", "TARGETS=openapi.yaml"},
		"spectral",
		"lint",
		"openapi.yaml",
	)

	tasktest.AssertDryRunContains(t, "spectral/node/nvm/npm",
		[]string{"version"},
		"spectral",
		"--version",
	)
}
