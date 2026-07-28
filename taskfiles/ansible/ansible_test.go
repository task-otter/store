package ansible_test

import (
	"os/exec"

	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"galaxy:install",
	"install",
	"install:undo",
	"lint",
	"lint:fix",
	"list:hosts",
	"ping",
	"run",
	"syntax:check",
	"upgrade",
	"vault:decrypt",
	"vault:encrypt",
	"version",
}

var publicVars = []string{
	"ANSIBLE_LINT_SKIP_PATTERN",
	"ANSIBLE_LINT_VERSION",
	"ANSIBLE_VERSION",
	"EXTRA_ARGS",
	"FILE",
	"INVENTORY",
	"PATTERN",
	"PLAYBOOK",
	"REQUIREMENTS",
	"TARGETS",
	"UV_LOAD",
}

func ansibleAvailable() bool {
	_, err := exec.LookPath("ansible")
	return err == nil
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "ansible", publicTasks, publicVars)
}
