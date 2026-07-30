// Copyright 2026 task-otter
// SPDX-License-Identifier: Apache-2.0

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
		"CONTEXT",
		"EXTRA_ARGS",
		"FILE",
		"IMAGE",
		"VERSION",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "docker", publicTasks(), publicVars())
}
