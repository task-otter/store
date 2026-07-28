package bencher_test

import (
	"runtime"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"exec",
	"install",
	"run",
	"upgrade",
	"version",
}

var publicVars = []string{
	"BENCHER_INSTALL_URL",
	"BENCHER_INSTALL_URL_WINDOWS",
	"BENCHER_VERSION",
	"EXTRA_ARGS",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "bencher", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "bencher",
		[]string{"version"},
		"bencher --version",
	)

	tasktest.AssertDryRunContains(t, "bencher",
		[]string{"run", "--", "--project", "demo", "make bench"},
		"bencher run",
		"--project demo",
		"make bench",
	)

	tasktest.AssertDryRunContains(t, "bencher",
		[]string{"exec", "--", "mock"},
		"bencher",
		"mock",
	)
}

func TestInstallDryRunUsesOfficialInstaller(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertInstallDryRun(t, "bencher", "bencher", "curl", "bencher.dev/download/install-cli.sh")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUpgradeDryRunUsesOfficialInstaller(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "bencher",
			[]string{"upgrade"},
			"curl",
			"bencher.dev/download/install-cli.sh",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}
