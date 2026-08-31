// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package brunogui_test

import (
	"testing"

	"github.com/task-otter/store/internal/taskintegration"
)

// TestModuleIntegration runs the shared task CLI integration suite for this module.
func TestModuleIntegration(t *testing.T) {
	t.Parallel()

	taskintegration.RunHere(t)
}
