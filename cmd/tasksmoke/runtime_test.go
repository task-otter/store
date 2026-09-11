// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

var errRuntimeBoom = errors.New("boom")

// TestDefaultRuntime exercises DefaultRuntime.
func TestDefaultRuntime(t *testing.T) {
	t.Parallel()

	requireRuntime(t, defaultRuntime())
}

// TestExecuteUnknownFlag exercises ExecuteUnknownFlag.
func TestExecuteUnknownFlag(t *testing.T) {
	t.Parallel()

	code := execute([]string{testAppName, "-bogus"}, new(bytes.Buffer), new(bytes.Buffer))

	if code == testZero {
		t.Fatal("expected non-zero exit")
	}
}

// TestMainDelegatesToCLI exercises MainDelegatesToCLI.
func TestMainDelegatesToCLI(t *testing.T) {
	t.Parallel()

	gotArgs, gotCode := runStubbedRuntime()
	requireMainResult(t, gotArgs, gotCode)
}

// TestMainRuntimeErrorMethods exercises MainRuntimeErrorMethods.
func TestMainRuntimeErrorMethods(t *testing.T) {
	t.Parallel()

	provider := mainRuntimeError(defaultRuntime)

	if provider.Error() != mainRuntimeErrorName || provider.Name() != mainRuntimeErrorName {
		t.Fatalf("Error/Name = %q/%q", provider.Error(), provider.Name())
	}

	if provider.Unwrap() != nil {
		t.Fatal("Unwrap() want nil")
	}

	requireRuntime(t, provider.Runtime())
}

// TestMainRuntimeFromFallback exercises MainRuntimeFromFallback.
func TestMainRuntimeFromFallback(t *testing.T) {
	t.Parallel()

	requireRuntime(t, mainRuntimeFrom(errRuntimeBoom))
	requireRuntime(t, mainRuntimeFrom(mainRuntimeError(nil)))
}

// TestMainUsesRuntimeProvider exercises MainUsesRuntimeProvider.
func TestMainUsesRuntimeProvider(t *testing.T) {
	t.Parallel()

	gotArgs, gotCode := invokeMainWithStub()
	requireMainResult(t, gotArgs, gotCode)
}

// TestOsArgsReturnsProcessArgs exercises OsArgsReturnsProcessArgs.
func TestOsArgsReturnsProcessArgs(t *testing.T) {
	t.Parallel()

	if len(osArgs()) == testZero {
		t.Fatal("osArgs empty")
	}
}

func invokeMainWithStub() (args []string, code int) {
	previous := errMainRuntime
	gotArgs, gotCode, runtime := newStubRuntime()

	errMainRuntime = mainRuntimeError(func() mainRuntime {
		return runtime
	})

	main()

	errMainRuntime = previous

	return *gotArgs, *gotCode
}

func newStubRuntime() (args *[]string, code *int, runtime mainRuntime) {
	gotArgs := new([]string)
	gotCode := new(int)

	runtime = mainRuntime{
		args: func() []string { return []string{testAppName} },
		exit: func(code int) { *gotCode = code },
		run:  stubRun(gotArgs),
	}

	return gotArgs, gotCode, runtime
}

func requireMainResult(t *testing.T, args []string, code int) {
	t.Helper()

	if len(args) != testOne || args[testZero] != testAppName {
		t.Fatalf("args = %v", args)
	}

	if code != testSeven {
		t.Fatalf("exit code = %d, want %d", code, testSeven)
	}
}

func requireRuntime(t *testing.T, runtime mainRuntime) {
	t.Helper()

	if runtime.args == nil || runtime.exit == nil || runtime.run == nil {
		t.Fatal("runtime missing args, exit, or run")
	}
}

func runStubbedRuntime() (args []string, code int) {
	gotArgs, gotCode, runtime := newStubRuntime()
	startRuntime(runtime)

	return *gotArgs, *gotCode
}

func stubRun(gotArgs *[]string) func([]string, io.Writer, io.Writer) int {
	return func(args []string, stdout io.Writer, stderr io.Writer) int {
		*gotArgs = append([]string(nil), args...)

		if stdout == nil || stderr == nil {
			return testSeven
		}

		return testSeven
	}
}
