// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package tasksmoke discovers public Taskfile tasks and runs them as a smoke suite.
package tasksmoke

import (
	"time"
)

const (
	defaultTimeout = 3 * time.Minute
)

func newEngine() *engine {
	return &engine{
		applyHome: applyIsolatedHome,
		mkdirTemp: bindMkdirTemp(emptyString),
		now:       time.Now,
		runTask:   executeGoTask,
		timeout:   defaultTimeout,
	}
}
