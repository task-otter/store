// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasktestutil

import (
	"regexp"
)

type (
	envPartsMessages struct {
		required  string
		wrongType string
	}

	legacyTaskRunMessages struct {
		env      envPartsMessages
		rootType string
	}

	legacyTaskRunArgs struct {
		run      any
		messages *legacyTaskRunMessages
		parts    []any
	}

	exitCodeMismatch struct {
		result   *CommandResult
		expected int
		actual   int
	}

	readmeTableScanner struct {
		row     *regexp.Regexp
		names   []string
		inTable bool
	}
)
