package bencher_test

import (
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
