package shellcheck_test

import (
	"runtime"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"install",
	"install:undo",
	"lint",
	"upgrade",
	"version",
}

var publicVars = []string{
	"SHELLCHECK_LINT_SKIP_PATTERN",
	"EXTRA_ARGS",
	"TARGETS",
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "shellcheck", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "shellcheck",
		[]string{"lint"},
		"shellcheck",
	)

	tasktest.AssertDryRunContains(t, "shellcheck",
		[]string{"version"},
		"shellcheck",
		"--version",
	)
}

func TestInstallDryRunUsesPlatformPackageManager(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertInstallDryRun(t, "shellcheck", "shellcheck", "brew", "shellcheck")
	case "linux":
		tasktest.AssertInstallDryRun(t, "shellcheck", "shellcheck", "shellcheck")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUpgradeDryRunUsesPlatformPackageManager(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "shellcheck",
			[]string{"upgrade"},
			"brew",
			"shellcheck",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "shellcheck",
			[]string{"upgrade"},
			"shellcheck",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}
