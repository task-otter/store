package python_test

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"install",
	"install:undo",
	"pip:install",
	"run",
	"upgrade",
	"verify",
	"version",
	"venv",
}

var publicVars = []string{
	"ARGS",
	"EXTRA_ARGS",
	"FILE",
	"PYTHON_PIN_VERSION",
	"REQUIREMENTS",
	"UV_LOAD",
	"VENV",
}

func pythonAvailable() bool {
	_, err := exec.LookPath("python3")
	if err == nil {
		return true
	}
	_, err = exec.LookPath("python")
	return err == nil
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "python", publicTasks, publicVars)
}

func TestInstallDryRun(t *testing.T) {
	t.Parallel()

	if pythonAvailable() {
		t.Skip("python already installed; status check short-circuits install body")
	}

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "python", []string{"install"}, "uv python install", "--default")
	default:
		t.Skip("install dry-run is covered on macOS and Linux")
	}
}

func TestVersionDryRun(t *testing.T) {
	t.Parallel()

	if !pythonAvailable() {
		t.Skip("python is not installed")
	}

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "python", []string{"version"}, "python3 --version")
	default:
		tasktest.AssertDryRunContains(t, "python", []string{"version"}, "python --version")
	}
}

func TestVenvDryRun(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "python", []string{"venv"}, "python3 -m venv")
	default:
		tasktest.AssertDryRunContains(t, "python", []string{"venv"}, "python -m venv")
	}
}

func TestPipInstallDryRun(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "python", []string{"pip:install"}, "pip3 install -r")
	default:
		tasktest.AssertDryRunContains(t, "python", []string{"pip:install"}, "pip install -r")
	}
}

func TestRunDryRun(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
		tasktest.AssertDryRunContains(t, "python", []string{"run", "FILE=hello.py"},
			"python3",
			"hello.py",
		)
	default:
		tasktest.AssertDryRunContains(t, "python", []string{"run", "FILE=hello.py"},
			"python",
			"hello.py",
		)
	}
}
