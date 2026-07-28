package cargo_test

import (
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
