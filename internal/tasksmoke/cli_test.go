// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"bytes"
	"path/filepath"
	"testing"
)

// TestFindRepoRootFromModule exercises FindRepoRootFromModule.
func TestFindRepoRootFromModule(t *testing.T) {
	t.Parallel()

	root, err := findRepoRoot(filepath.Join(testdataAbs(t, "."), "testdata"))
	requireNoErr(t, err)
	requireSame(t, pathExists(filepath.Join(root, goModFileName)), true)
}

// TestFindRepoRootMissing exercises FindRepoRootMissing.
func TestFindRepoRootMissing(t *testing.T) {
	t.Parallel()

	value, err := findRepoRoot(t.TempDir())
	keepValue(value)
	requireErr(t, err)
}

// TestPathExistsMissing exercises PathExistsMissing.
func TestPathExistsMissing(t *testing.T) {
	t.Parallel()

	requireSame(t, pathExists(filepath.Join(t.TempDir(), testMissingName)), false)
}

// TestParseFlagsModuleAndList exercises ParseFlagsModuleAndList.
func TestParseFlagsModuleAndList(t *testing.T) {
	t.Parallel()

	stderr := new(bytes.Buffer)
	flags, err := parseFlags([]string{appName, flagModuleArg, testYamllint, flagListArg}, stderr)
	requireNoErr(t, err)
	requireEqual(t, flags.Module, testYamllint)
	requireSame(t, flags.List, true)
}

// TestParseFlagsRejectsUnknown exercises ParseFlagsRejectsUnknown.
func TestParseFlagsRejectsUnknown(t *testing.T) {
	t.Parallel()

	value, err := parseFlags([]string{appName, testBogusFlag}, new(bytes.Buffer))
	keepValue(value)
	requireErr(t, err)
}

// TestParseFlagsEmptyArgs exercises ParseFlagsEmptyArgs.
func TestParseFlagsEmptyArgs(t *testing.T) {
	t.Parallel()

	flags, err := parseFlags(nil, new(bytes.Buffer))
	requireNoErr(t, err)
	requireSame(t, !flags.List, true)
}

// TestReportExitCode exercises ReportExitCode.
func TestReportExitCode(t *testing.T) {
	t.Parallel()

	requireSame(
		t,
		reportExitCode(&smokeReport{Results: []*taskResult{{Status: statusFail}}}),
		exitFail,
	)
	requireSame(
		t,
		reportExitCode(&smokeReport{Results: []*taskResult{{Status: statusPass}}}),
		emptyLength,
	)
}

// TestPrintErrReturnsCode exercises PrintErrReturnsCode.
func TestPrintErrReturnsCode(t *testing.T) {
	t.Parallel()

	code := printErr(new(bytes.Buffer), sentinelErr(), exitFail)
	requireSame(t, code == exitFail, true)
}

// TestPrintErrIgnoresWriterError exercises PrintErrIgnoresWriterError.
func TestPrintErrIgnoresWriterError(t *testing.T) {
	t.Parallel()

	code := printErr(failWriter{}, sentinelErr(), nameSplitParts)
	requireSame(t, code == nameSplitParts, true)
}

// TestWriteAndExitReportsFailure exercises WriteAndExitReportsFailure.
func TestWriteAndExitReportsFailure(t *testing.T) {
	t.Parallel()

	code := writeAndExit(new(bytes.Buffer), new(bytes.Buffer), &smokeReport{
		Results: []*taskResult{{Module: testEcho, Task: testPing, Status: statusFail}},
	})
	requireSame(t, code == exitFail, true)
}

// TestWriteAndExitWriterError exercises WriteAndExitWriterError.
func TestWriteAndExitWriterError(t *testing.T) {
	t.Parallel()

	code := writeAndExit(failWriter{}, new(bytes.Buffer), &smokeReport{
		Results: []*taskResult{{Module: testEcho, Task: testPing, Status: statusPass}},
	})
	requireSame(t, code == exitFail, true)
}

// TestRunWithRootList exercises RunWithRootList.
func TestRunWithRootList(t *testing.T) {
	t.Parallel()

	code := runWithRoot(&cliRun{
		Engine: testEngine(t),
		Flags:  &cliFlags{List: true, Module: testEcho},
		Root:   testdataRepo(t),
		Stdout: new(bytes.Buffer),
		Stderr: new(bytes.Buffer),
	})
	requireSame(t, code == emptyLength, true)
}

// TestMainUnknownFlag exercises MainUnknownFlag.
func TestMainUnknownFlag(t *testing.T) {
	t.Parallel()

	code := Main([]string{appName, testBogusFlag}, new(bytes.Buffer), new(bytes.Buffer))
	requireSame(t, code == nameSplitParts, true)
}
