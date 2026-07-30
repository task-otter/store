// REPLACE_ME 2026
// SPDX-License-Identifier: Apache-2.0

package spectralnodenvmnpm_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"config:init",
		"help",
		"install",
		"install:undo",
		"lint",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"SPECTRAL_LINT_SKIP_PATTERN",
		"EXTRA_ARGS",
		"RULESET",
		"TARGETS",
		"VERSION",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "spectral/node/nvm/npm", publicTasks(), publicVars())
}
