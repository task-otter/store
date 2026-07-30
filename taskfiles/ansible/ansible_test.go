// Copyright 2026 task-otter
// SPDX-License-Identifier: Apache-2.0

package ansible_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"galaxy:install",
		"install",
		"install:undo",
		"lint",
		"lint:fix",
		"list:hosts",
		"ping",
		"run",
		"syntax:check",
		"upgrade",
		"vault:decrypt",
		"vault:encrypt",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"ANSIBLE_LINT_SKIP_PATTERN",
		"ANSIBLE_LINT_VERSION",
		"ANSIBLE_VERSION",
		"EXTRA_ARGS",
		"FILE",
		"INVENTORY",
		"PATTERN",
		"PLAYBOOK",
		"REQUIREMENTS",
		"TARGETS",
		"UV_LOAD",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "ansible", publicTasks(), publicVars())
}
