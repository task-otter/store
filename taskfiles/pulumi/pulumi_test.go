// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package pulumi_test

import (
	"testing"

	"github.com/task-otter/store/internal/taskintegration"
	"github.com/task-otter/store/internal/tasktest"
)

const (
	constPulumiModule = "pulumi"
)

// TestModuleIntegration runs the shared task CLI integration suite for this module.
func TestModuleIntegration(t *testing.T) {
	t.Parallel()

	taskintegration.RunHere(t)
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

func publicTasks() []string {
	return []string{
		"login",
		"new",
		"up",
	}
}

func publicVars() []string {
	return []string{
		"PULUMI_ARGS",
		"PULUMI_EXTRA_ARGS",
		"PULUMI_NIX_INSTALLABLE",
		"PULUMI_LOGIN_URL",
		"PULUMI_STACK",
		"PULUMI_TEMPLATE",
	}
}
