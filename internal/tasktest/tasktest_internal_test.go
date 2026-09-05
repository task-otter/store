// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasktest

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"
)

type (
	fatalSink struct {
		messages chan string
	}

	getwdError struct{}
)

const (
	once               = 1
	workingDirErrMsg   = "forced getwd error"
	timeoutFatalPrefix = "task command timed out"
	emptyDir           = ""
	expectedFatalCall  = "expected fatal call"
)

func (sink *fatalSink) Fatal(args ...any) {
	sink.record(fmt.Sprint(args...))
}

func (sink *fatalSink) Fatalf(format string, args ...any) {
	sink.record(fmt.Sprintf(format, args...))
}

func (*fatalSink) Helper() {}

func (*fatalSink) TempDir() string {
	return emptyDir
}

func (sink *fatalSink) record(message string) {
	sink.messages <- message

	runtime.Goexit()
}

func (getwdError) Error() string {
	return workingDirErrMsg
}

// TestAssertTaskCommandDidNotTimeout verifies an expired context is fatal.
func TestAssertTaskCommandDidNotTimeout(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(t.Context(), time.Time{})
	t.Cleanup(cancel)
	expectSinkFatal(t, timeoutFatalPrefix, func(sink *fatalSink) {
		assertTaskCommandDidNotTimeout(ctx, sink, []string{taskJSONFlag})
	})
}

// TestFindRepoRootFromGetwdError verifies repo discovery fails when Getwd fails.
func TestFindRepoRootFromGetwdError(t *testing.T) {
	t.Parallel()
	expectSinkFatal(t, workingDirErrMsg, fatalFindRepoRootFromGetwd)
}

func expectSinkFatal(t *testing.T, want string, fatalFunc func(*fatalSink)) {
	t.Helper()

	messages := make(chan string, once)
	done := make(chan struct{})

	runSinkFatal(done, messages, fatalFunc)
	assertSinkFatal(t, want, messages)
}

func fatalFindRepoRootFromGetwd(sink *fatalSink) {
	findRepoRootFrom(sink, failingWorkingDirectory)
}

func failingWorkingDirectory() (string, error) {
	return emptyDir, getwdError{}
}

func runSinkFatal(done chan struct{}, messages chan string, fatalFunc func(*fatalSink)) {
	go func() {
		defer close(done)

		fatalFunc(&fatalSink{messages: messages})
	}()

	<-done
}

func assertSinkFatal(t *testing.T, want string, messages <-chan string) {
	t.Helper()

	message, ok := receiveSinkFatal(messages)

	if !ok {
		t.Fatal(expectedFatalCall)
	}

	if !strings.Contains(message, want) {
		t.Fatalf("fatal message %q does not contain %q", message, want)
	}
}

func receiveSinkFatal(messages <-chan string) (message string, ok bool) {
	select {
	case message = <-messages:
		return message, true
	default:
		return emptyDir, false
	}
}
