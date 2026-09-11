// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"io"
)

type (
	mainRuntime = struct {
		args func() []string
		exit func(int)
		run  func([]string, io.Writer, io.Writer) int
	}

	mainRuntimeError func() mainRuntime
)
