// REPLACE_ME 2026
// SPDX-License-Identifier: Apache-2.0

package knipnodenvmpnpm_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"ci",
		"config:init",
		"config:skip",
		"dependencies",
		"dev-dependencies",
		"exports",
		"files",
		"help",
		"init",
		"install",
		"install:undo",
		"lint",
		"lint:fix",
		"production",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"KNIP_LINT_SKIP_PATTERN",
		"CONFIG",
		"EXTRA_ARGS",
		"VERSION",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "knip/node/nvm/pnpm", publicTasks(), publicVars())
}
