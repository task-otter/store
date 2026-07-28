package dotenv_linter_test

import (
	"runtime"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"diff",
	"fix",
	"install",
	"install:undo",
	"lint",
	"upgrade",
	"version",
}

var publicVars = []string{
	"DOTENV_LINTER_LINT_SKIP_PATTERN",
	"CARGO_BIN_UNIX",
	"DOTENV_LINTER_VERSION",
	"EXTRA_ARGS",
	"TARGETS",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "dotenv-linter", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "dotenv-linter",
		[]string{"lint", "TARGETS=.env.example"},
		"dotenv-linter",
		"check",
		".env.example",
	)

	tasktest.AssertDryRunContains(t, "dotenv-linter",
		[]string{"fix", "TARGETS=.env.example"},
		"fix",
		".env.example",
	)

	tasktest.AssertDryRunContains(t, "dotenv-linter",
		[]string{"diff", "TARGETS=.env .env.example"},
		"diff",
		".env .env.example",
	)

	tasktest.AssertDryRunContains(t, "dotenv-linter",
		[]string{"version"},
		"--version",
	)
}

func TestInstallDryRunUsesCargo(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertInstallDryRun(t, "dotenv-linter", "dotenv-linter", "install")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUpgradeHonorsVersionPin(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "dotenv-linter",
			[]string{"upgrade", "DOTENV_LINTER_VERSION=0.0.0-test"},
			"--version 0.0.0-test",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}
