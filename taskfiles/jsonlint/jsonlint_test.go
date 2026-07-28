package jsonlint_test

import (
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"install",
	"install:undo",
	"lint",
	"upgrade",
	"version",
}

var publicVars = []string{
	"JSONLINT_LINT_SKIP_PATTERN",
	"EXTRA_ARGS",
	"JSONLINT_VERSION",
	"TARGETS",
	"UV_LOAD",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "jsonlint", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "jsonlint",
		[]string{"lint", "TARGETS=config.json"},
		"jsonlint",
		"config.json",
	)

	tasktest.AssertDryRunContains(t, "jsonlint",
		[]string{"version"},
		"jsonlint --version",
	)
}

func TestInstallDryRunUsesDemjson3(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertInstallDryRun(t, "jsonlint", "jsonlint", "demjson3")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestInstallHonorsVersionPin(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "jsonlint",
			[]string{"install", "JSONLINT_VERSION=0.0.0-test"},
			"demjson3==0.0.0-test",
		)
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}
