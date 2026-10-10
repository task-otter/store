// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package taskcli runs the separately installed Task CLI.
package taskcli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os/exec"
	"slices"
	"time"
)

// Request describes a Task invocation. Timeout must be positive.
// A nil Env inherits the process environment; nil writers discard output.
type (
	Request = struct {
		Stdout  io.Writer
		Stderr  io.Writer
		Dir     string
		Env     []string
		Args    []string
		Timeout time.Duration
	}
)

const (
	emptyLength = 0
	waitDelay   = 5 * time.Second
)

// Run executes Task directly and preserves process and context errors.
func Run(ctx context.Context, request *Request) error {
	ctx, cancel := context.WithTimeout(ctx, request.Timeout)

	defer cancel()

	err := newCommand(ctx, request).Run()
	if err == nil {
		return nil
	}

	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf("task CLI is missing: install Task and add it to PATH: %w", err)
	}

	return fmt.Errorf("run Task CLI: %w", errors.Join(err, ctx.Err()))
}

// VarsArgs returns sorted KEY=VALUE arguments without shell interpretation.
func VarsArgs(values map[string]string) []string {
	if len(values) == emptyLength {
		return nil
	}

	keys := slices.Sorted(maps.Keys(values))
	args := make([]string, emptyLength, len(keys))

	for i := range keys {
		key := keys[i]

		args = append(args, key+"="+values[key])
	}

	return args
}

func newCommand(ctx context.Context, request *Request) *exec.Cmd {
	commandContext := exec.CommandContext
	cmd := commandContext(ctx, "task", request.Args...)

	cmd.Dir = request.Dir
	cmd.Env = request.Env
	cmd.Stdout = request.Stdout
	cmd.Stderr = request.Stderr
	cmd.WaitDelay = waitDelay

	return cmd
}
