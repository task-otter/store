// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestFolderScannerVisitError exercises FolderScannerVisitError.
func TestFolderScannerVisitError(t *testing.T) {
	t.Parallel()

	scanner := newFolderScanner(t.TempDir())
	err := visitFolder(&walkVisit{Err: sentinelErr(), Path: testShortPath, Scan: scanner})
	requireErr(t, err)
}

// TestReadModuleTaskfileMissing exercises ReadModuleTaskfileMissing.
func TestReadModuleTaskfileMissing(t *testing.T) {
	t.Parallel()

	value, err := readModuleTaskfile(t.TempDir())
	keepValue(value)
	requireErr(t, err)
}

// TestCopyStubTreeMkdirError exercises CopyStubTreeMkdirError.
func TestCopyStubTreeMkdirError(t *testing.T) {
	t.Parallel()

	dest := filepath.Join(t.TempDir(), testFileName)
	err := os.WriteFile(dest, []byte(testFileBody), fileMode)
	requireNoErr(t, err)

	copyErr := copyStubTree(t.TempDir(), dest)
	requireErr(t, copyErr)
}

// TestRunWithRootDiscoverError exercises RunWithRootDiscoverError.
func TestRunWithRootDiscoverError(t *testing.T) {
	t.Parallel()

	code := runWithRoot(&cliRun{
		Engine: testEngine(t),
		Flags:  &cliFlags{List: true},
		Root:   t.TempDir(),
		Stdout: new(bytes.Buffer),
		Stderr: new(bytes.Buffer),
	})
	requireSame(t, code == exitFail, true)
}

// TestMainListYamllint exercises MainListYamllint.
func TestMainListYamllint(t *testing.T) {
	t.Parallel()

	stdout := new(bytes.Buffer)
	code := Main(
		[]string{appName, flagListArg, flagModuleArg, testYamllint},
		stdout,
		new(bytes.Buffer),
	)
	requireSame(t, code == emptyLength, true)
	requireSame(t, bytes.Contains(stdout.Bytes(), []byte(testYamllint)), true)
}

// TestSkipRemainingDestructiveNames exercises SkipRemainingDestructiveNames.
func TestSkipRemainingDestructiveNames(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipName(nameCIFix), reasonDestructive)
	requireEqual(t, skipName(nameLintFix), reasonDestructive)
	requireEqual(t, skipName(nameUpgrade), reasonDestructive)
	requireEqual(t, skipName(nameVaultEncrypt), reasonDestructive)
}

func skipName(name string) string {
	return skipReason(&skipInput{Module: testShortPath, Name: name})
}

// TestSkipDockerNameList exercises SkipDockerNameList.
func TestSkipDockerNameList(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{Module: toolDocker, Name: namePruneAll}), reasonDocker)
	requireEqual(t, skipReason(&skipInput{Module: toolDocker, Name: nameStopAll}), reasonDocker)
}

// TestSmokeSkipsNilConfig exercises SmokeSkipsNilConfig.
func TestSmokeSkipsNilConfig(t *testing.T) {
	t.Parallel()

	requireSame(t, smokeSkips(nil) == nil, true)
	requireSame(t, smokeVars(nil) == nil, true)
}

// TestYamlSkipNoMatch exercises YamlSkipNoMatch.
func TestYamlSkipNoMatch(t *testing.T) {
	t.Parallel()

	requireEqual(t, yamlSkipReason(testCI, []string{testOther}), emptyString)
}

// TestHasPromptNil exercises HasPromptNil.
func TestHasPromptNil(t *testing.T) {
	t.Parallel()

	requireSame(t, !hasPrompt(nil), true)
}
