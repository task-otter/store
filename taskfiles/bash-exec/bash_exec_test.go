// REPLACE_ME 2026
// SPDX-License-Identifier: Apache-2.0

package bash_exec_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"check",
		"exec",
		"run",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"ARGS",
		"BASH_FLAGS",
		"COMMAND",
		"SCRIPT",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "bash-exec", publicTasks(), publicVars())
}
