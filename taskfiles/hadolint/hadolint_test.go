package hadolint_test

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
	"HADOLINT_LINT_SKIP_PATTERN",
	"CONFIG",
	"DOCKERFILE",
	"EXTRA_ARGS",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "hadolint", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "hadolint",
		[]string{"lint"},
		"hadolint",
		"Dockerfile",
	)

	tasktest.AssertDryRunContains(t, "hadolint",
		[]string{"version"},
		"hadolint",
		"--version",
	)
}

func TestInstallDryRunUsesPlatformPackageManager(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertInstallDryRun(t, "hadolint", "hadolint", "brew", "hadolint")
	case "linux":
		tasktest.AssertInstallDryRun(t, "hadolint", "hadolint", "curl", "hadolint-")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}
