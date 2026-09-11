// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"errors"
	"io"
	"os"

	"github.com/task-otter/store/internal/tasksmoke"
)

func defaultRuntime() mainRuntime {
	return mainRuntime{args: osArgs, exit: os.Exit, run: execute}
}

func execute(args []string, stdout, stderr io.Writer) int {
	return tasksmoke.Main(args, stdout, stderr)
}

func main() {
	startRuntime(mainRuntimeFrom(errMainRuntime))
}

func mainRuntimeFrom(err error) mainRuntime {
	var provider mainRuntimeError

	if !errors.As(err, &provider) {
		return defaultRuntime()
	}

	return provider.Runtime()
}

func osArgs() []string {
	return os.Args
}

func startRuntime(runtime mainRuntime) {
	runtime.exit(runtime.run(runtime.args(), os.Stdout, os.Stderr))
}

// Error implements the error interface for main runtime selection.
func (mainRuntimeError) Error() string {
	return mainRuntimeErrorName
}

// Name returns the stable error name for the main runtime provider.
func (provider mainRuntimeError) Name() string {
	return provider.Error()
}

// Runtime returns the injectable main runtime, or the default when unset.
func (provider mainRuntimeError) Runtime() mainRuntime {
	if provider == nil {
		return defaultRuntime()
	}

	return provider()
}

// Unwrap implements error unwrapping for mainRuntimeError.
func (mainRuntimeError) Unwrap() error {
	return nil
}
