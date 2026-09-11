// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestApplyEnvPairsError exercises ApplyEnvPairsError.
func TestApplyEnvPairsError(t *testing.T) {
	t.Parallel()

	err := applyEnvPairs([]envPair{{testKey, testVal}}, func(string, string) error {
		return sentinelErr()
	})
	requireErr(t, err)
}

// TestApplyIsolatedHomeWriteError exercises ApplyIsolatedHomeWriteError.
func TestApplyIsolatedHomeWriteError(t *testing.T) {
	t.Parallel()

	home := filepath.Join(t.TempDir(), testFileName)
	err := os.WriteFile(home, []byte(testFileBody), fileMode)
	requireNoErr(t, err)

	applyErr := applyIsolatedHome(home)
	requireErr(t, applyErr)
}

// TestExecuteGoTaskCapturesOutput exercises ExecuteGoTaskCapturesOutput.
func TestExecuteGoTaskCapturesOutput(t *testing.T) {
	t.Parallel()

	output := new(bytes.Buffer)
	err := executeGoTask(&runRequest{
		Dir:     testdataAbs(t, testdataPingPath),
		Home:    t.TempDir(),
		Name:    testPing,
		Output:  output,
		Timeout: defaultTimeout,
		WorkDir: t.TempDir(),
	})
	requireNoErr(t, err)
	requireSame(t, bytes.Contains(output.Bytes(), []byte(testPing)), true)
}

// TestCollectResultsWorkDirError exercises CollectResultsWorkDirError.
func TestCollectResultsWorkDirError(t *testing.T) {
	t.Parallel()

	engine := testEngine(t)

	engine.mkdirTemp = failAfterFirstMkdir(t)

	value, err := runSuite(engine, &runOptions{RepoRoot: testdataRepo(t), Module: testEcho})
	keepValue(value)
	requireErr(t, err)
}

func failAfterFirstMkdir(t *testing.T) mkdirTempFunc {
	t.Helper()

	calls := emptyLength

	return func() (string, error) {
		calls++

		if calls == exitFail {
			return t.TempDir(), nil
		}

		return emptyString, sentinelErr()
	}
}

// TestCopyDirectoryCopyFSError exercises CopyDirectoryCopyFSError.
func TestCopyDirectoryCopyFSError(t *testing.T) {
	t.Parallel()

	src := filepath.Join(testdataRepo(t), dataTestDirName, testEcho)
	dst := t.TempDir()
	chmodErr := os.Chmod(dst, emptyLength)
	requireNoErr(t, chmodErr)

	t.Cleanup(func() {
		restoreErr := os.Chmod(dst, dirMode)
		requireNoErr(t, restoreErr)
	})

	info, err := os.Stat(src)
	requireNoErr(t, err)

	copyErr := copyDirectory(src, dst, info)
	requireErr(t, copyErr)
}

// TestCopyExistingTreePermission exercises CopyExistingTreePermission.
func TestCopyExistingTreePermission(t *testing.T) {
	t.Parallel()

	parent := filepath.Join(t.TempDir(), testLocked)
	err := os.Mkdir(parent, emptyLength)
	requireNoErr(t, err)

	t.Cleanup(func() {
		chmodErr := os.Chmod(parent, dirMode)
		requireNoErr(t, chmodErr)
	})

	copyErr := copyExistingTree(filepath.Join(parent, testChildName), t.TempDir())
	requireErr(t, copyErr)
}

// TestDetectRepoRootFromError exercises DetectRepoRootFromError.
func TestDetectRepoRootFromError(t *testing.T) {
	t.Parallel()

	value, err := detectRepoRootFrom(func() (string, error) {
		return emptyString, sentinelErr()
	})
	keepValue(value)
	requireErr(t, err)
}

// TestMkdirTempInError exercises MkdirTempInError.
func TestMkdirTempInError(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), testFileName)
	err := os.WriteFile(path, []byte(testFileBody), fileMode)
	requireNoErr(t, err)

	value, mkdirErr := mkdirTempIn(path, tempWorkPrefix)
	keepValue(value)
	requireErr(t, mkdirErr)
}

// TestSetEnvPairsEmpty exercises SetEnvPairsEmpty.
func TestSetEnvPairsEmpty(t *testing.T) {
	t.Parallel()

	requireNoErr(t, setEnvPairs(nil))
}

// TestStartCLIError exercises StartCLIError.
func TestStartCLIError(t *testing.T) {
	t.Parallel()

	code := startCLI(&startedCLI{
		Err:    sentinelErr(),
		Flags:  &cliFlags{List: true},
		Stderr: new(bytes.Buffer),
		Stdout: new(bytes.Buffer),
	})
	requireSame(t, code == exitFail, true)
}

// TestAttachConfigsReadError exercises AttachConfigsReadError.
func TestAttachConfigsReadError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, dataTestDirName, testEcho, smokeConfigName)
	err := os.MkdirAll(path, dirMode)
	requireNoErr(t, err)

	attachErr := attachConfigs(root, []*module{{Name: testEcho}})
	requireErr(t, attachErr)
}

// TestRunSelectedConfigError exercises RunSelectedConfigError.
func TestRunSelectedConfigError(t *testing.T) {
	t.Parallel()

	engine := testEngine(t)
	root := blockedSmokeConfigRoot(t)
	suite := bindSuite(engine, &runOptions{RepoRoot: root, ListOnly: true}, emptyString)

	suite.modules = []*module{{Name: testEcho}}

	value, runErr := runSelected(suite)
	keepValue(value)
	requireErr(t, runErr)
}

func blockedSmokeConfigRoot(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	path := filepath.Join(root, dataTestDirName, testEcho, smokeConfigName)
	err := os.MkdirAll(path, dirMode)
	requireNoErr(t, err)

	return root
}

// TestDetectRepoRootFromMissing exercises DetectRepoRootFromMissing.
func TestDetectRepoRootFromMissing(t *testing.T) {
	t.Parallel()

	value, err := detectRepoRootFrom(func() (string, error) {
		return t.TempDir(), nil
	})
	keepValue(value)
	requireErr(t, err)
}

// TestPrepareWorkDirCopyError exercises PrepareWorkDirCopyError.
func TestPrepareWorkDirCopyError(t *testing.T) {
	t.Parallel()

	engine := testEngine(t)
	file := filepath.Join(t.TempDir(), testFileName)
	err := os.WriteFile(file, []byte(testFileBody), fileMode)
	requireNoErr(t, err)

	engine.mkdirTemp = func() (string, error) {
		return file, nil
	}

	value, copyErr := prepareWorkDir(engine, testdataRepo(t), testEcho)
	keepValue(value)
	requireErr(t, copyErr)
}

// TestPrintErrZeroWriter exercises PrintErrZeroWriter.
func TestPrintErrZeroWriter(t *testing.T) {
	t.Parallel()

	code := printErr(zeroWriter{}, sentinelErr(), exitFail)
	requireSame(t, code == exitFail, true)
}

// TestWriteHeaderZeroWriter exercises WriteHeaderZeroWriter.
func TestWriteHeaderZeroWriter(t *testing.T) {
	t.Parallel()

	err := writeHeader(zeroWriter{})
	requireErr(t, err)
}

// TestWriteResultRowZeroWriter exercises WriteResultRowZeroWriter.
func TestWriteResultRowZeroWriter(t *testing.T) {
	t.Parallel()

	err := writeResultRow(zeroWriter{}, &taskResult{
		Module: testEcho,
		Status: statusPass,
		Task:   testPing,
	})
	requireErr(t, err)
}

// TestSetEnvPairsInvalidKey exercises SetEnvPairsInvalidKey.
func TestSetEnvPairsInvalidKey(t *testing.T) {
	t.Parallel()

	err := setEnvPairs([]envPair{{emptyString, testValue}})
	requireErr(t, err)
}

// TestRunOneSkipsDestructiveTask exercises RunOneSkipsDestructiveTask.
func TestRunOneSkipsDestructiveTask(t *testing.T) {
	t.Parallel()

	result := runOne(testEngine(t), &moduleRun{
		Module: &module{Name: testGo},
		Opts:   &runOptions{},
	}, &taskSpec{Name: nameUninstall})
	requireStatus(t, result, statusSkip)
	requireEqual(t, result.Output, reasonDestructive)
}
