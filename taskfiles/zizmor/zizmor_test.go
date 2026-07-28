package zizmor_test

import (
	"runtime"
	"strings"
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
	"ZIZMOR_LINT_SKIP_PATTERN",
	"ZIZMOR_EXTRA_ARGS",
	"ZIZMOR_TARGETS",
	"ZIZMOR_VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "zizmor", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "zizmor",
		[]string{"lint"},
		"zizmor",
	)

	tasktest.AssertDryRunContains(t, "zizmor",
		[]string{"lint", "ZIZMOR_TARGETS=.github/workflows/main.yml"},
		"zizmor",
		".github/workflows/main.yml",
	)

	tasktest.AssertDryRunContains(t, "zizmor",
		[]string{"version"},
		"zizmor",
		"--version",
	)
}

func TestLintIgnoresSharedTargetVariable(t *testing.T) {
	t.Parallel()

	output := tasktest.DryRun(t, "zizmor", "lint", "TARGETS=**/*.html")
	if !strings.Contains(output, "zizmor") {
		t.Fatalf("zizmor dry-run command not found:\n%s", output)
	}
	if strings.Contains(output, "**/*.html") {
		t.Fatalf("zizmor command should not receive shared TARGETS value:\n%s", output)
	}
}

func TestInstallDryRunDownloadsBinary(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertInstallDryRun(t, "zizmor", "zizmor", "curl", "apple-darwin")
	case "linux":
		tasktest.AssertInstallDryRun(t, "zizmor", "zizmor", "curl", "linux-gnu")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUpgradeDryRunDownloadsBinary(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "zizmor",
			[]string{"upgrade"},
			"curl",
			"zizmor",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}
