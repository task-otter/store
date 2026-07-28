package jq_test

import (
	"os/exec"

	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"install",
	"install:undo",
	"upgrade",
	"version",
}

func jqAvailable() bool {
	_, err := exec.LookPath("jq")
	return err == nil
}

var publicVars = []string{
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "jq", publicTasks, publicVars)
}
