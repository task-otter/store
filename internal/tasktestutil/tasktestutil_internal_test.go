// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasktestutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type (
	internalFatalCall struct {
		message string
	}

	internalFakeTest struct {
		fatal    *internalFatalCall
		tempDirs []string
		nextDir  int
		stay     stayAfterFatal
	}

	stayAfterFatal bool

	internalFatalResult struct {
		fatal *internalFatalCall
		done  bool
	}

	internalFatalFunc func(*internalFakeTest)
)

const (
	stayAlive             stayAfterFatal = true
	internalStubName                     = "stub"
	internalExtraArg                     = "extra"
	internalReadmeName                   = "README.md"
	internalTaskfilesName                = "taskfiles"
	internalExpectedFatal                = "expected fatal call"
	internalDirMode                      = 0o700
)

func (fake *internalFakeTest) Fatal(args ...any) {
	fake.fatal = &internalFatalCall{message: fmt.Sprint(args...)}

	if !fake.stay {
		runtime.Goexit()
	}
}

func (fake *internalFakeTest) Fatalf(format string, args ...any) {
	fake.fatal = &internalFatalCall{message: fmt.Sprintf(format, args...)}

	if !fake.stay {
		runtime.Goexit()
	}
}

func (*internalFakeTest) Helper() {}

func (fake *internalFakeTest) TempDir() string {
	if fake.nextDir < len(fake.tempDirs) {
		dir := fake.tempDirs[fake.nextDir]
		fake.nextDir++

		return dir
	}

	dir, err := os.MkdirTemp(emptyString, "tasktestutil-internal-")
	if err != nil {
		fake.Fatalf("create temp dir: %v", err)
	}

	fake.tempDirs = append(fake.tempDirs, dir)
	fake.nextDir++

	return dir
}

// TestReadmeSearchHelpers validates the behavior covered by this test case.
func TestReadmeSearchHelpers(t *testing.T) {
	t.Parallel()

	assertReadmeSearchBoundaries(t)
	assertFindModuleReadmeStops(t)
	assertFindModuleTaskfilePathFatal(t)
}

// TestWorkingDirFromError validates the behavior covered by this test case.
func TestWorkingDirFromError(t *testing.T) {
	t.Parallel()

	assertWorkingDirFromError(t)
	assertModuleRootFromGetwdError(t)
}

// TestWriteStubFileChmodError validates the behavior covered by this test case.
func TestWriteStubFileChmodError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	expectInternalFatal(t, "mark", func(fake *internalFakeTest) {
		writeStubFile(fake, &stub{Dir: dir, Name: internalStubName, Body: "body"}, failingChmod)
	})
}

// TestWriteStubValue validates the behavior covered by this test case.
func TestWriteStubValue(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	WriteStub(t, stub{Dir: dir, Name: internalStubName, Body: "#!/bin/sh\n"})
	assertStubFileExists(t, dir)
	assertWriteStubValueExtraArgs(t, dir)
}

func newInternalFake(stay stayAfterFatal) *internalFakeTest {
	return &internalFakeTest{fatal: nil, tempDirs: nil, nextDir: zeroIndex, stay: stay}
}

func assertFindModuleReadmeStops(t *testing.T) {
	t.Helper()

	root := t.TempDir()
	module := filepath.Join(root, internalTaskfilesName, "module")

	writeInternalFile(t, filepath.Join(root, internalReadmeName), "# Root\n")
	writeInternalFile(t, filepath.Join(module, taskfileYML), "version: \"3\"\n")

	path, found := findModuleReadme(module)

	if found || path != emptyString {
		t.Fatalf("findModuleReadme = %q, %v", path, found)
	}
}

func assertFindModuleTaskfilePathFatal(t *testing.T) {
	t.Helper()

	expectInternalFatalStay(t, errTaskfileNotFound, func(fake *internalFakeTest) {
		if got := findModuleTaskfilePath(fake, t.TempDir()); got != emptyString {
			t.Fatalf("findModuleTaskfilePath = %q", got)
		}
	})
}

func assertInternalFatal(t *testing.T, want string, result internalFatalResult) {
	t.Helper()

	if !result.done || result.fatal == nil {
		t.Fatal(internalExpectedFatal)
	}

	if !strings.Contains(result.fatal.message, want) {
		t.Fatalf("fatal message %q does not contain %q", result.fatal.message, want)
	}
}

func assertModuleRootFromGetwdError(t *testing.T) {
	t.Helper()

	expectInternalFatal(t, "failed to get working directory", func(fake *internalFakeTest) {
		moduleRootFrom(fake, failingWorkingDir)
	})
}

func assertReadmeSearchBoundaries(t *testing.T) {
	t.Helper()

	if !reachedReadmeSearchBoundary(filepath.Dir(string(os.PathSeparator))) {
		t.Fatal("filesystem root must stop README search")
	}

	if !reachedReadmeSearchBoundary(filepath.Join(t.TempDir(), internalTaskfilesName)) {
		t.Fatal("taskfiles directory must stop README search")
	}
}

func assertStubFileExists(t *testing.T, dir string) {
	t.Helper()

	if !FileExists(filepath.Join(dir, internalStubName)) {
		t.Fatal("stub file was not written")
	}
}

func assertWorkingDirFromError(t *testing.T) {
	t.Helper()

	workingDirectory, err := workingDirFrom(failingWorkingDir)

	if err == nil || workingDirectory != emptyString {
		t.Fatal("expected working directory error")
	}
}

func assertWriteStubValueExtraArgs(t *testing.T, dir string) {
	t.Helper()

	expectInternalFatal(t, "positional arguments", func(fake *internalFakeTest) {
		WriteStub(fake, stub{Dir: dir, Name: internalStubName, Body: "x"}, internalExtraArg)
	})
}

func expectInternalFatal(t *testing.T, want string, fatalFunc internalFatalFunc) {
	t.Helper()

	result := runInternalFatal(newInternalFake(false), fatalFunc)

	assertInternalFatal(t, want, result)
}

func expectInternalFatalStay(t *testing.T, want string, fatalFunc internalFatalFunc) {
	t.Helper()

	fake := newInternalFake(stayAlive)

	fatalFunc(fake)
	assertInternalFatal(t, want, internalFatalResult{fatal: fake.fatal, done: true})
}

func failingChmod(path string, fileMode os.FileMode) error {
	return fmt.Errorf("%s %v: %w", path, fileMode, os.ErrPermission)
}

func failingWorkingDir() (string, error) {
	return emptyString, os.ErrPermission
}

func runInternalFatal(fake *internalFakeTest, callback internalFatalFunc) internalFatalResult {
	done := make(chan internalFatalResult)

	go func() {
		defer func() {
			done <- internalFatalResult{fatal: fake.fatal, done: true}
		}()

		callback(fake)
	}()

	return <-done
}

func writeInternalFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), internalDirMode)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path, []byte(content), privateFileMode)
	if err != nil {
		t.Fatal(err)
	}
}
