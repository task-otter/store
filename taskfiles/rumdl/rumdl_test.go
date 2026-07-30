// REPLACE_ME 2026
// SPDX-License-Identifier: Apache-2.0

package rumdl_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"fix",
		"fmt",
		"install",
		"install:undo",
		"lint",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"RUMDL_LINT_SKIP_PATTERN",
		"RUMDL_FMT_SKIP_PATTERN",
		"EXTRA_ARGS",
		"RUMDL_VERSION",
		"TARGETS",
		"UV_LOAD",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "rumdl", publicTasks(), publicVars())
}
