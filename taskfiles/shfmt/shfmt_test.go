package shfmt_test

import (
	"runtime"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"fmt",
	"fmt:check",
	"install",
	"install:undo",
	"upgrade",
	"version",
}

var publicVars = []string{
	"SHFMT_FMT_SKIP_PATTERN",
	"EXTRA_ARGS",
	"GLOBAL_GO_BIN",
	"SHFMT_VERSION",
	"TARGETS",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "shfmt", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "shfmt",
		[]string{"fmt", "TARGETS=scripts", "EXTRA_ARGS=-i 2 -ci"},
		"shfmt\" -w",
		"scripts",
		"-i 2 -ci",
	)

	tasktest.AssertDryRunContains(t, "shfmt",
		[]string{"fmt:check", "TARGETS=scripts"},
		"shfmt\" -d",
		"scripts",
	)

	tasktest.AssertDryRunContains(t, "shfmt",
		[]string{"version"},
		"shfmt",
		"-version",
	)
}

func TestInstallDryRunUsesOfficialGoModule(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertInstallDryRun(t, "shfmt", "shfmt", "go install", "mvdan.cc/sh/v3/cmd/shfmt@latest")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUpgradeDryRunUsesOfficialGoModule(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "shfmt",
			[]string{"upgrade"},
			"go install",
			"mvdan.cc/sh/v3/cmd/shfmt@latest",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}
