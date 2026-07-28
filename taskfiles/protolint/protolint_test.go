package protolint_test

import (
	"runtime"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"fix",
	"install",
	"install:undo",
	"lint",
	"upgrade",
	"version",
}

var publicVars = []string{
	"PROTOLINT_LINT_SKIP_PATTERN",
	"EXTRA_ARGS",
	"GLOBAL_GO_BIN",
	"PROTOLINT_VERSION",
	"TARGETS",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "protolint", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "protolint",
		[]string{"lint", "TARGETS=api", "EXTRA_ARGS=-reporter json"},
		"protolint\" lint",
		"api",
		"-reporter json",
	)

	tasktest.AssertDryRunContains(t, "protolint",
		[]string{"fix", "TARGETS=api"},
		"protolint\" lint -fix",
		"api",
	)

	tasktest.AssertDryRunContains(t, "protolint",
		[]string{"version"},
		"protolint\" version",
	)
}

func TestInstallDryRunUsesOfficialGoModule(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertInstallDryRun(t, "protolint", "protolint",
			"go install", "github.com/yoheimuta/protolint/cmd/protolint@latest")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestUpgradeDryRunUsesOfficialGoModule(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "protolint",
			[]string{"upgrade"},
			"go install",
			"github.com/yoheimuta/protolint/cmd/protolint@latest",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}

func TestInstallHonorsVersionPin(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "protolint",
			[]string{"upgrade", "PROTOLINT_VERSION=v0.0.0-test"},
			"github.com/yoheimuta/protolint/cmd/protolint@v0.0.0-test",
		)
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}
