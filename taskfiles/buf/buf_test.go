package buf_test

import (
	"runtime"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"breaking",
	"fmt",
	"fmt:check",
	"generate",
	"install",
	"install:undo",
	"lint",
	"upgrade",
	"version",
}

var publicVars = []string{
	"BUF_LINT_SKIP_PATTERN",
	"BUF_FMT_SKIP_PATTERN",
	"AGAINST",
	"BUF_VERSION",
	"CONFIG",
	"EXTRA_ARGS",
	"INPUT",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "buf", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "buf",
		[]string{"lint"},
		"buf lint",
	)

	tasktest.AssertDryRunContains(t, "buf",
		[]string{"version"},
		"buf --version",
	)

	tasktest.AssertDryRunContains(t, "buf",
		[]string{"fmt:check"},
		"buf format",
		"--diff",
	)

	tasktest.AssertDryRunContains(t, "buf",
		[]string{"breaking"},
		"buf breaking",
		"--against",
	)
}

func TestInstallDryRunUsesPlatformPackageManager(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertInstallDryRun(t, "buf", "buf", "brew", "bufbuild/buf/buf")
	case "linux":
		tasktest.AssertInstallDryRun(t, "buf", "buf", "curl", "buf-Linux-")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUpgradeDryRunUsesPlatformPackageManager(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "buf",
			[]string{"upgrade"},
			"brew",
			"bufbuild/buf/buf",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "buf",
			[]string{"upgrade"},
			"curl",
			"buf-Linux-",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}
