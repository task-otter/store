// Copyright 2026 task-otter
// SPDX-License-Identifier: Apache-2.0

package adrs_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"exec",
		"generate",
		"init",
		"install",
		"install:undo",
		"list",
		"new",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"ADRS_VERSION",
		"CARGO_BIN_UNIX",
		"EXTRA_ARGS",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "adrs", publicTasks(), publicVars())
}
