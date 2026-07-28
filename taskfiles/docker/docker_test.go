package docker_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"build",
	"images",
	"install",
	"install:undo",
	"prune",
	"prune:all",
	"ps",
	"ps:all",
	"pull",
	"stop:all",
	"upgrade",
	"verify",
	"version",
}

var publicVars = []string{
	"CONTEXT",
	"EXTRA_ARGS",
	"FILE",
	"IMAGE",
	"VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "docker", publicTasks, publicVars)
}
