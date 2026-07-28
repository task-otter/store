package actionlint_test

import (
	"runtime"
	"strings"
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
	"ACTIONLINT_LINT_SKIP_PATTERN",
	"ACTIONLINT_EXTRA_ARGS",
	"ACTIONLINT_TARGETS",
	"ACTIONLINT_VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "actionlint", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "actionlint",
		[]string{"lint"},
		"actionlint",
	)

	tasktest.AssertDryRunContains(t, "actionlint",
		[]string{"lint", "ACTIONLINT_TARGETS=.github/workflows/main.yml"},
		"actionlint",
		".github/workflows/main.yml",
	)

	tasktest.AssertDryRunContains(t, "actionlint",
		[]string{"version"},
		"actionlint",
		"--version",
	)
}

func TestLintIgnoresSharedTargetVariable(t *testing.T) {
	t.Parallel()

	output := tasktest.DryRun(t, "actionlint", "lint", "TARGETS=.")
	if !strings.Contains(output, "actionlint") {
		t.Fatalf("actionlint dry-run command not found:\n%s", output)
	}
	if strings.Contains(output, "targets='.'") {
		t.Fatalf("actionlint command should not receive shared TARGETS value:\n%s", output)
	}
}

func TestInstallDryRunUsesPlatformPackageManager(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertInstallDryRun(t, "actionlint", "actionlint", "brew", "actionlint")
	case "linux":
		tasktest.AssertInstallDryRun(t, "actionlint", "actionlint", "curl", "actionlint_")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUpgradeDryRunUsesPlatformPackageManager(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "actionlint",
			[]string{"upgrade"},
			"brew",
			"actionlint",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "actionlint",
			[]string{"upgrade"},
			"curl",
			"actionlint_",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}
