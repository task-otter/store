package bash_exec_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"check",
	"exec",
	"run",
	"version",
}

var publicVars = []string{
	"ARGS",
	"BASH_FLAGS",
	"COMMAND",
	"SCRIPT",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "bash-exec", publicTasks, publicVars)
}

func TestRepresentativeDryRuns(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "bash-exec",
		[]string{"run", "SCRIPT=scripts/build.sh", "ARGS=--release"},
		"bash",
		"scripts/build.sh",
		"--release",
	)

	tasktest.AssertDryRunContains(t, "bash-exec",
		[]string{"check", "SCRIPT=scripts/build.sh"},
		"bash -n",
		"scripts/build.sh",
	)

	tasktest.AssertDryRunContains(t, "bash-exec",
		[]string{"exec", "COMMAND=printf hello"},
		"bash -c",
	)

	tasktest.AssertDryRunContains(t, "bash-exec",
		[]string{"version"},
		"bash --version",
	)
}
