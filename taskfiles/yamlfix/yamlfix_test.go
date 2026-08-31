// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package yamlfix_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"yamlfix",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{
		"ci:fix",
	}
}

func publicVars() []string {
	return []string{
		"YAMLFIX_EXTRA_ARGS",
		"YAMLFIX_FMT_SKIP_PATTERN",
		"YAMLFIX_NIX_INSTALLABLE",
		"YAMLFIX_TARGETS",
	}
}
