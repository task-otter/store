// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package docker_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"build",
		"images",
		"install",
		"install:undo",
		"prune",
		"prune:all",
		"ps",
		"ps:all",
		"pull",
		"stop:all",
		"upgrade",
		"verify",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"DOCKER_CONTEXT",
		"DOCKER_EXTRA_ARGS",
		"DOCKER_FILE",
		"DOCKER_IMAGE",
		"DOCKER_VERSION",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"docker",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
