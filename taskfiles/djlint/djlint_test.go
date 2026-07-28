package djlint_test

import (
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"fmt",
	"fmt:check",
	"install",
	"install:undo",
	"lint",
	"upgrade",
	"version",
}

var publicVars = []string{
	"DJLINT_LINT_SKIP_PATTERN",
	"DJLINT_FMT_SKIP_PATTERN",
	"DJLINT_VERSION",
	"EXTRA_ARGS",
	"TARGETS",
	"UV_LOAD",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "djlint", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "djlint",
		[]string{"lint", "TARGETS=templates"},
		"djlint",
		"--lint",
		"templates",
	)

	tasktest.AssertDryRunContains(t, "djlint",
		[]string{"fmt", "TARGETS=templates"},
		"djlint",
		"--reformat",
		"templates",
	)

	tasktest.AssertDryRunContains(t, "djlint",
		[]string{"fmt:check", "TARGETS=templates"},
		"djlint",
		"--check",
		"templates",
	)

	tasktest.AssertDryRunContains(t, "djlint",
		[]string{"version"},
		"djlint --version",
	)
}

func TestInstallDryRunUsesUv(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertInstallDryRun(t, "djlint", "djlint", "uv tool install")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestInstallHonorsVersionPin(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "djlint",
			[]string{"install", "DJLINT_VERSION=0.0.0-test"},
			"djlint==0.0.0-test",
		)
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}
