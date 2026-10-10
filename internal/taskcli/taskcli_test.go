// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskcli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

type (
	invocationCase = struct {
		name string
		want string
		args []string
		code int
	}
)

const (
	testSuccessCode  = 0
	testFailureCode  = 7
	testUnknownCode  = 200
	testVersionFlag  = "--version"
	testOutputTask   = "output"
	testStdout       = "stdout"
	testStderr       = "stderr"
	testUnknownTask  = "unknown"
	testEmptyString  = ""
	testTimeout      = 10 * time.Second
	shortTimeout     = 30 * time.Millisecond
	testFileMode     = 0o600
	testTaskfileName = "Taskfile.yml"
	testValue        = "spaces ; $(echo injected) & 'quoted' = value"
	testTaskfile     = `version: '3'
tasks:
  output:
    cmds:
      - printf stdout
      - printf stderr >&2
  environment:
    cmds:
      - printf '%s' "$TASKCLI_TEST_VALUE"
  variables:
    env:
      VALUE: '{{.VALUE}}'
    cmds:
      - printf '%s' "$VALUE"
  fail:
    cmds:
      - exit 7
  wait:
    cmds:
      - sleep 30
`
)

// TestRun exercises real Task execution and exit errors.
func TestRun(t *testing.T) {
	t.Parallel()

	scenarios := invocationCases()

	for i := range scenarios {
		scenario := scenarios[i]
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			assertInvocation(t, scenario)
		})
	}
}

// TestRunCanceled preserves a parent context's cancellation.
func TestRunCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := Run(ctx, testRequest(testEmptyString, []string{testVersionFlag}))
	assertErrorIs(t, err, context.Canceled)
}

// TestRunCombinedOutput safely captures stdout and stderr in one buffer.
func TestRunCombinedOutput(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	request := fixtureRequest(t, []string{testOutputTask})

	request.Stdout = &output
	request.Stderr = &output

	assertNoError(t, Run(t.Context(), request))
	assertContains(t, output.String(), testStdout)
	assertContains(t, output.String(), testStderr)
}

// TestRunDeadline bounds execution and preserves deadline errors.
func TestRunDeadline(t *testing.T) {
	t.Parallel()

	request := fixtureRequest(t, []string{"wait"})

	request.Timeout = shortTimeout

	err := Run(t.Context(), request)
	assertErrorIs(t, err, context.DeadlineExceeded)
}

// TestRunEnvironment passes the requested child environment verbatim.
func TestRunEnvironment(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	request := fixtureRequest(t, []string{"environment"})

	request.Env = append(os.Environ(), "TASKCLI_TEST_VALUE="+testValue)
	request.Stdout = &output

	assertNoError(t, Run(t.Context(), request))
	assertText(t, output.String(), testValue)
}

// TestRunInvalidTaskfile reports malformed Taskfiles as process errors.
func TestRunInvalidTaskfile(t *testing.T) {
	t.Parallel()

	request := fixtureRequest(t, []string{testOutputTask})
	assertNoError(
		t,
		os.WriteFile(
			filepath.Join(request.Dir, testTaskfileName),
			[]byte("[invalid"),
			testFileMode,
		),
	)
	assertNonzeroExit(t, Run(t.Context(), request))
}

// TestRunMissingExecutable gives installation guidance and keeps the lookup error.
func TestRunMissingExecutable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := Run(t.Context(), testRequest(testEmptyString, []string{testVersionFlag}))
	assertErrorIs(t, err, exec.ErrNotFound)
	assertContains(t, err.Error(), "install Task")
}

// TestRunMissingTaskfile reports absent Taskfiles as process errors.
func TestRunMissingTaskfile(t *testing.T) {
	t.Parallel()

	err := Run(
		t.Context(),
		testRequest(t.TempDir(), []string{testOutputTask}),
	)
	assertNonzeroExit(t, err)
}

// TestRunNilWriters discards command output when no writers are provided.
func TestRunNilWriters(t *testing.T) {
	t.Parallel()

	assertNoError(t, Run(t.Context(), fixtureRequest(t, []string{testOutputTask})))
}

// TestRunSeparateOutput keeps stdout and stderr separate.
func TestRunSeparateOutput(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	request := fixtureRequest(t, []string{testOutputTask})

	request.Stdout = &stdout
	request.Stderr = &stderr

	assertNoError(t, Run(t.Context(), request))
	assertText(t, stdout.String(), testStdout)
	assertContains(t, stderr.String(), testStderr)
}

// TestRunVariables preserves spaces and metacharacters without invoking a shell.
func TestRunVariables(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	request := fixtureRequest(
		t,
		append([]string{"variables"}, VarsArgs(map[string]string{"VALUE": testValue})...),
	)

	request.Stdout = &output

	assertNoError(t, Run(t.Context(), request))
	assertText(t, output.String(), testValue)
}

// TestVarsArgs checks deterministic argument order and empty inputs.
func TestVarsArgs(t *testing.T) {
	t.Parallel()

	if VarsArgs(nil) != nil || VarsArgs(map[string]string{}) != nil {
		t.Fatal("empty variables must produce nil arguments")
	}

	want := []string{"A=first", "Z=" + testValue}

	if !slices.Equal(VarsArgs(map[string]string{"Z": testValue, "A": "first"}), want) {
		t.Fatal("variable arguments must be sorted and preserve values")
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()

	if !strings.Contains(got, want) {
		t.Fatalf("%q does not contain %q", got, want)
	}
}

func assertErrorIs(t *testing.T, err, want error) {
	t.Helper()

	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

func assertExitError(t *testing.T, err error) *exec.ExitError {
	t.Helper()

	var exitErr *exec.ExitError

	if !errors.As(err, &exitErr) {
		t.Fatalf("error = %v, want process exit error", err)
	}

	return exitErr
}

func assertInvocation(t *testing.T, scenario *invocationCase) {
	t.Helper()

	var output bytes.Buffer

	request := fixtureRequest(t, scenario.args)

	request.Stdout = &output

	err := Run(t.Context(), request)

	if scenario.code != testSuccessCode {
		assertExitCode(t, err, scenario.code)

		return
	}

	assertNoError(t, err)
	assertContains(t, output.String(), scenario.want)
}

func assertExitCode(t *testing.T, err error, want int) {
	t.Helper()

	if got := assertExitError(t, err).ExitCode(); got != want {
		t.Fatalf("exit code = %d, want %d", got, want)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

func assertNonzeroExit(t *testing.T, err error) {
	t.Helper()

	if assertExitError(t, err).ExitCode() == testSuccessCode {
		t.Fatal("expected a nonzero process exit code")
	}
}

func assertText(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func fixtureRequest(t *testing.T, args []string) *Request {
	t.Helper()

	dir := t.TempDir()
	assertNoError(
		t,
		os.WriteFile(filepath.Join(dir, testTaskfileName), []byte(testTaskfile), testFileMode),
	)

	return testRequest(dir, append([]string{"--silent"}, args...))
}

func invocationCases() []*invocationCase {
	return []*invocationCase{
		{name: "success", args: []string{testOutputTask}, want: testStdout, code: testSuccessCode},
		{
			name: "failure",
			args: []string{"--exit-code", "fail"},
			want: testEmptyString,
			code: testFailureCode,
		},
		{
			name: testUnknownTask,
			args: []string{testUnknownTask},
			want: testEmptyString,
			code: testUnknownCode,
		},
	}
}

func testRequest(dir string, args []string) *Request {
	return &Request{
		Dir: dir, Args: args, Timeout: testTimeout,
		Stdout: nil, Stderr: nil, Env: nil,
	}
}
