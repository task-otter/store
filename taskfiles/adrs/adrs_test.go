package adrs_test

import (
	"runtime"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"exec",
	"generate",
	"init",
	"install",
	"install:undo",
	"list",
	"new",
	"upgrade",
	"version",
}

var publicVars = []string{
	"ADRS_VERSION",
	"CARGO_BIN_UNIX",
	"EXTRA_ARGS",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "adrs", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "adrs",
		[]string{"list"},
		"adrs",
		"list",
	)

	tasktest.AssertDryRunContains(t, "adrs",
		[]string{"exec", "--", "doctor"},
		"adrs",
		"doctor",
	)

	tasktest.AssertDryRunContains(t, "adrs",
		[]string{"version"},
		"--version",
	)
}

func TestInstallDryRunUsesCargo(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertInstallDryRun(t, "adrs", "adrs", "install")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUpgradeHonorsVersionPin(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "adrs",
			[]string{"upgrade", "ADRS_VERSION=0.0.0-test"},
			"--version 0.0.0-test",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}
