// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasktestutil

type (
	// TaskRun describes one task command invocation.
	TaskRun = struct {
		Root string
		Env  []string
		Args []string
	}

	// stub describes a shell stub to write into a test directory.
	stub = struct {
		Dir  string
		Name string
		Body string
	}
)
