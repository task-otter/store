// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskintegration_test

import (
	"testing"

	"github.com/task-otter/store/internal/taskintegration"
)

const (
	fixtureModule   = "jq"
	defaultTaskName = "default"
	stdoutText      = "listing"
	stderrText      = "warning"
)

// TestRunExercisesEveryCheck runs the default suite against a real module so
// every check in the suite executes at least once.
func TestRunExercisesEveryCheck(t *testing.T) {
	t.Parallel()

	taskintegration.Run(t, fixtureModule)
}

// TestRunSpecDryRunsSelectedTasks covers the opt-in dry-run check, which the
// default suite skips because most modules need a tool that is not installed.
func TestRunSpecDryRunsSelectedTasks(t *testing.T) {
	t.Parallel()

	taskintegration.RunSpec(t, &taskintegration.Spec{
		Module:      fixtureModule,
		DryRunTasks: []string{defaultTaskName},
	})
}

// TestResultCombinedAppendsStderr checks the output helper the failure messages rely on.
func TestResultCombinedAppendsStderr(t *testing.T) {
	t.Parallel()

	result := taskintegration.Result{
		Stdout: stdoutText,
		Stderr: stderrText,
		Args:   nil,
		Failed: false,
	}

	combined := result.Combined()
	want := stdoutText + "\n" + stderrText

	if combined != want {
		t.Fatalf("Combined() = %q, want %q", combined, want)
	}
}
