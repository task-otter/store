// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"errors"
	"io"
	"path/filepath"
	"testing"
	"time"
)

type (
	failWriter struct{}

	reportSinkStub struct {
		flushErr error
		remain   int
	}

	zeroWriter struct{}
)

const (
	testdataRepoPath = "testdata/repo"
	testdataPingPath = "testdata/ping"
	testdataBadPath  = "testdata/badrepo"
)

var errSentinel = errors.New("sentinel")

func (failWriter) Write(payload []byte) (int, error) {
	keepValue(payload)

	return emptyLength, io.ErrClosedPipe
}

func (stub *reportSinkStub) Flush() error {
	return stub.flushErr
}

func (stub *reportSinkStub) Write(payload []byte) (int, error) {
	if stub.remain <= emptyLength {
		return emptyLength, io.ErrClosedPipe
	}

	stub.remain--

	return len(payload), nil
}

func (zeroWriter) Flush() error {
	return nil
}

func (zeroWriter) Write(payload []byte) (int, error) {
	return len(payload) * emptyLength, nil
}

func testdataRepo(t *testing.T) string {
	t.Helper()

	path, err := filepath.Abs(testdataRepoPath)
	if err != nil {
		t.Fatalf("abs testdata repo: %v", err)
	}

	return path
}

func testdataAbs(t *testing.T, rel string) string {
	t.Helper()

	path, err := filepath.Abs(rel)
	if err != nil {
		t.Fatalf("abs %s: %v", rel, err)
	}

	return path
}

func testEngine(t *testing.T) *engine {
	t.Helper()

	engine := newEngine()
	stubEngineHooks(t, engine)

	return engine
}

func stubEngineHooks(t *testing.T, engine *engine) {
	t.Helper()

	engine.applyHome = func(string) error { return nil }
	engine.mkdirTemp = func() (string, error) { return t.TempDir(), nil }
	engine.now = func() time.Time { return time.Unix(emptyLength, emptyLength) }
	engine.runTask = func(*runRequest) error { return nil }
}

func keepValue[T any](value T) {
	holder := any(value)

	if holder == nil {
		return
	}
}

func requireNoErr(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func requireErr(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error")
	}
}

func requireStatus(t *testing.T, result *taskResult, want string) {
	t.Helper()

	if result.Status != want {
		t.Fatalf("status = %q, want %q", result.Status, want)
	}
}

func requireEqual(t *testing.T, got, want string) {
	t.Helper()

	requireSame(t, got, want)
}

func requireSame[T comparable](t *testing.T, got, want T) {
	t.Helper()

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func sentinelErr() error {
	return errSentinel
}
