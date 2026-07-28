package staticcheck_test

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
	"STATICCHECK_LINT_SKIP_PATTERN",
	"GLOBAL_GO_BIN",
	"STATICCHECK_RELEASE_BASE_URL",
	"STATICCHECK_VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "staticcheck", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "staticcheck",
		[]string{"lint", "--", "./cmd/..."},
		"staticcheck",
		"./cmd/...",
	)

	tasktest.AssertDryRunContains(t, "staticcheck",
		[]string{"version"},
		"-version",
	)
}

func TestInstallDryRunUsesPlatformArchive(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "staticcheck",
			[]string{"install"},
			"staticcheck_darwin_",
			".tar.gz",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "staticcheck",
			[]string{"install"},
			"staticcheck_linux_",
			".tar.gz",
		)
	default:
		t.Skip("archive dry-run is covered on macOS and Linux")
	}
}
