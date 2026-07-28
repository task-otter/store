package rumdl_test

import (
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"fix",
	"fmt",
	"install",
	"install:undo",
	"lint",
	"upgrade",
	"version",
}

var publicVars = []string{
	"RUMDL_LINT_SKIP_PATTERN",
	"RUMDL_FMT_SKIP_PATTERN",
	"EXTRA_ARGS",
	"RUMDL_VERSION",
	"TARGETS",
	"UV_LOAD",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "rumdl", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "rumdl",
		[]string{"lint", "TARGETS=docs"},
		"rumdl check",
		"docs",
	)

	tasktest.AssertDryRunContains(t, "rumdl",
		[]string{"fix", "TARGETS=README.md"},
		"rumdl check --fix",
		"README.md",
	)

	tasktest.AssertDryRunContains(t, "rumdl",
		[]string{"fmt", "TARGETS=docs"},
		"rumdl fmt",
		"docs",
	)

	tasktest.AssertDryRunContains(t, "rumdl",
		[]string{"version"},
		"rumdl --version",
	)
}

func TestInstallDryRunUsesUv(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertInstallDryRun(t, "rumdl", "rumdl", "uv tool install")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestInstallHonorsVersionPin(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "rumdl",
			[]string{"install", "RUMDL_VERSION=0.0.0-test"},
			"rumdl==0.0.0-test",
		)
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}
