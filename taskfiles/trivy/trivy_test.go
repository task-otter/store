// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package trivy_test

import (
	"testing"

	"github.com/task-otter/store/internal/taskintegration"
	"github.com/task-otter/store/internal/tasktest"
)

// TestModuleIntegration runs the shared task CLI integration suite for this module.
func TestModuleIntegration(t *testing.T) {
	t.Parallel()

	taskintegration.RunHere(t)
}

// TestTaskfileModuleContract checks the documented public surface.
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"trivy",
		&tasktest.ModuleExpectations{
			Tasks: []string{"ci:fs", "ci:image", "install", "version"},
			Vars: []string{
				"TRIVY_IMAGE", "TRIVY_FS_TARGET",
				"TRIVY_SCANNERS",
				"TRIVY_IMAGE_ARGS", "TRIVY_FS_ARGS",
				"TRIVY_IMAGE_EXTRA_ARGS", "TRIVY_FS_EXTRA_ARGS",
				"TRIVY_NIX_INSTALLABLE", "TRIVY_WINGET_INSTALLABLE",
			},
		},
	)
}
