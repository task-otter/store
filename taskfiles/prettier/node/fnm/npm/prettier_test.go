// REPLACE_ME 2026
// SPDX-License-Identifier: Apache-2.0

package prettiernodefnmnpm_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"ci",
		"config:init",
		"fix",
		"fmt",
		"fmt:check",
		"help",
		"install",
		"install:undo",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"PRETTIER_FMT_SKIP_PATTERN",
		"CONFIG",
		"EXTRA_ARGS",
		"IGNORE_PATH",
		"TARGETS",
		"VERSION",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "prettier/node/fnm/npm", publicTasks(), publicVars())
}

// covered by module contract
