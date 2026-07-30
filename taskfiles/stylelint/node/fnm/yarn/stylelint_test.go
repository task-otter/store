// REPLACE_ME 2026
// SPDX-License-Identifier: Apache-2.0

package stylelintnodefnmyarn_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"cache:clean",
		"ci",
		"config:init",
		"help",
		"install",
		"install:undo",
		"lint",
		"lint:fix",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"STYLELINT_LINT_SKIP_PATTERN",
		"ALLOW_EMPTY_INPUT",
		"CACHE",
		"CONFIG",
		"EXTRA_ARGS",
		"TARGETS",
		"VERSION",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "stylelint/node/fnm/yarn", publicTasks(), publicVars())
}
