// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasktest_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// TestParseTaskfileSuccess exercises ParseTaskfileSuccess.
func TestParseTaskfileSuccess(t *testing.T) {
	t.Parallel()

	taskfile, err := tasktest.ParseTaskfile(
		[]byte("version: '3'\ntasks:\n  ci:\n    cmds: [echo]\n"),
	)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if taskfile.Tasks["ci"] == nil {
		t.Fatal("missing ci")
	}
}

// TestParseTaskfileError exercises ParseTaskfileError.
func TestParseTaskfileError(t *testing.T) {
	t.Parallel()

	parsed, err := tasktest.ParseTaskfile([]byte(":\n"))
	if err == nil {
		t.Fatal("expected parse error")
	}

	if parsed != nil {
		t.Fatal("expected nil taskfile")
	}
}
