// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package jq_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"jq",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return []string{}
}

func publicVars() []string {
	return []string{
		"JQ_NIX_INSTALLABLE",
	}
}
