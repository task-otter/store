// Copyright 2026 task-otter
// SPDX-License-Identifier: Apache-2.0

package brunobun_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"ci",
		"help",
		"install",
		"install:undo",
		"run",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"COLLECTION",
		"ENV",
		"EXTRA_ARGS",
		"VERSION",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "bruno/bun", publicTasks(), publicVars())
}
