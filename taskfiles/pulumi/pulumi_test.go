// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package pulumi_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

const (
	constPulumiModule = "pulumi"
)

func publicTasks() []string {
	return []string{
		"install",
		"install:undo",
		"login",
		"new",
		"up",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"PULUMI_ARGS",
		"PULUMI_EXTRA_ARGS",
		"PULUMI_INSTALL_PS1_URL",
		"PULUMI_INSTALL_URL",
		"PULUMI_LOAD",
		"PULUMI_LOGIN_URL",
		"PULUMI_STACK",
		"PULUMI_TEMPLATE",
		"PULUMI_VERSION",
	}
}

// TestTaskfileModuleContract validates the behavior covered by this test case.
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		constPulumiModule,
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}
