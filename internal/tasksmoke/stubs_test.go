// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"os"
	"path/filepath"
	"testing"
)

// TestBindMkdirTempCreatesDirectory exercises BindMkdirTempCreatesDirectory.
func TestBindMkdirTempCreatesDirectory(t *testing.T) {
	t.Parallel()

	dir, err := bindMkdirTemp(emptyString)()
	requireNoErr(t, err)

	t.Cleanup(func() {
		removeErr := os.RemoveAll(dir)
		requireNoErr(t, removeErr)
	})

	requireSame(t, pathExists(dir), true)
}

// TestBindMkdirTempError exercises BindMkdirTempError.
func TestBindMkdirTempError(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), testFileName)
	err := os.WriteFile(path, []byte(testFileBody), fileMode)
	requireNoErr(t, err)

	value, mkdirErr := bindMkdirTemp(path)()
	keepValue(value)
	requireErr(t, mkdirErr)
}

// TestCopyStubTreeCopiesFiles exercises CopyStubTreeCopiesFiles.
func TestCopyStubTreeCopiesFiles(t *testing.T) {
	t.Parallel()

	dst := t.TempDir()
	src := filepath.Join(testdataRepo(t), dataTestDirName, testEcho)
	err := copyStubTree(src, dst)
	requireNoErr(t, err)
	requireSame(t, pathExists(filepath.Join(dst, testHelloFile)), true)
}

// TestCopyStubTreeMissingSource exercises CopyStubTreeMissingSource.
func TestCopyStubTreeMissingSource(t *testing.T) {
	t.Parallel()

	dst := t.TempDir()
	err := copyStubTree(filepath.Join(t.TempDir(), testMissingTask), dst)
	requireNoErr(t, err)
}

// TestCopyStubTreeRejectsFileSource exercises CopyStubTreeRejectsFileSource.
func TestCopyStubTreeRejectsFileSource(t *testing.T) {
	t.Parallel()

	src := filepath.Join(t.TempDir(), testFileName)
	err := os.WriteFile(src, []byte(testFileBody), fileMode)
	requireNoErr(t, err)

	copyErr := copyStubTree(src, t.TempDir())
	requireErr(t, copyErr)
}

// TestCopyExistingTreeStatError exercises CopyExistingTreeStatError.
func TestCopyExistingTreeStatError(t *testing.T) {
	t.Parallel()

	err := copyExistingTree(
		filepath.Join(t.TempDir(), "missing-parent", testChildName),
		t.TempDir(),
	)
	requireNoErr(t, err)
}
