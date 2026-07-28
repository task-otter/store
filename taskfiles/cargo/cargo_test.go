package cargo_test

import (
	"runtime"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"build",
	"check",
	"fmt",
	"fmt:check",
	"install",
	"install:undo",
	"lint",
	"lint:fix",
	"test",
	"upgrade",
	"verify",
	"version",
	"which",
}

var publicVars = []string{
	"CARGO_LINT_SKIP_PATTERN",
	"CARGO_FMT_SKIP_PATTERN",
	"CARGO_BIN_UNIX",
	"EXTRA_ARGS",
	"RUST_TOOLCHAIN",
	"RUSTUP_INSTALL_URL",
	"RUSTUP_INSTALL_URL_WINDOWS",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "cargo", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "cargo",
		[]string{"version"},
		"cargo",
		"--version",
	)

	tasktest.AssertDryRunContains(t, "cargo",
		[]string{"lint"},
		"clippy",
	)

	tasktest.AssertDryRunContains(t, "cargo",
		[]string{"fmt:check"},
		"fmt --check",
	)
}

func TestInstallDryRunUsesOfficialInstaller(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertInstallDryRun(t, "cargo", "cargo", "curl", "sh.rustup.rs")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUpgradeDryRunUsesRustup(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "cargo",
			[]string{"upgrade"},
			"rustup",
			"update",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}
