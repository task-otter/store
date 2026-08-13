// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package yamlfix_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"ci:fix",
		"fmt",
		"install",
		"install:undo",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"YAMLFIX_FMT_SKIP_PATTERN",
		"YAMLFIX_EXTRA_ARGS",
		"YAMLFIX_TARGETS",
		"UV_LOAD",
		"YAMLFIX_VERSION",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"yamlfix",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
