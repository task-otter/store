// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"runtime"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// TestSkipPrompt exercises SkipPrompt.
func TestSkipPrompt(t *testing.T) {
	t.Parallel()

	reason := skipReason(&skipInput{
		Module: testEcho,
		Name:   "prompted",
		Task:   &tasktest.Task{Prompt: testContinue},
	})

	requireEqual(t, reason, reasonPrompt)
}

// TestSkipInteractive exercises SkipInteractive.
func TestSkipInteractive(t *testing.T) {
	t.Parallel()

	reason := skipReason(&skipInput{
		Module: "nix",
		Name:   "install:shell",
		Task:   &tasktest.Task{Interactive: true},
	})

	requireEqual(t, reason, reasonInteractive)
}

// TestSkipUninstall exercises SkipUninstall.
func TestSkipUninstall(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{Module: testGo, Name: nameUninstall}), reasonDestructive)
}

// TestSkipFmtKeepsFmtCheck exercises SkipFmtKeepsFmtCheck.
func TestSkipFmtKeepsFmtCheck(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{Module: testGo, Name: nameFmt}), reasonFmt)
	requireEqual(
		t,
		skipReason(&skipInput{Module: testGo, Name: nameFmt + colonSeparator + "check"}),
		emptyString,
	)
}

// TestSkipCacheCleanStaysIn exercises SkipCacheCleanStaysIn.
func TestSkipCacheCleanStaysIn(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{Module: testEslint, Name: nameCacheClean}), emptyString)
}

// TestSkipDockerWorkTasks exercises SkipDockerWorkTasks.
func TestSkipDockerWorkTasks(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{Module: toolDocker, Name: nameBuild}), reasonDocker)
	requireEqual(t, skipReason(&skipInput{Module: toolDocker, Name: nameVerify}), reasonDocker)
	requireEqual(t, skipReason(&skipInput{Module: toolDocker, Name: nameVersion}), reasonDocker)
	requireEqual(t, skipReason(&skipInput{Module: testGo, Name: nameBuild}), emptyString)
}

// TestSkipNixInstallOnly exercises SkipNixInstallOnly.
func TestSkipNixInstallOnly(t *testing.T) {
	t.Parallel()

	expected := reasonNixInstall
	if unixOnly := unixOnlySkipOnOS(toolNix, runtime.GOOS); unixOnly != emptyString {
		expected = unixOnly
	}

	requireEqual(t, skipReason(&skipInput{Module: toolNix, Name: nameInstall}), expected)
	requireEqual(t, skipReason(&skipInput{Module: testYamllint, Name: nameInstall}), emptyString)
}

// TestSkipFuzzWithoutCLIArgs exercises SkipFuzzWithoutCLIArgs.
func TestSkipFuzzWithoutCLIArgs(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{Module: testGo, Name: nameFuzz}), reasonFuzz)
}

// TestSkipFuzzWithCLIArgs exercises SkipFuzzWithCLIArgs.
func TestSkipFuzzWithCLIArgs(t *testing.T) {
	t.Parallel()

	reason := skipReason(&skipInput{
		Module: testGo,
		Name:   nameFuzz,
		Config: &smokeConfig{Vars: map[string]string{nameCLIArgs: "-fuzz FuzzName ."}},
	})

	requireEqual(t, reason, emptyString)
}

// TestSkipMissingRequiredVars exercises SkipMissingRequiredVars.
func TestSkipMissingRequiredVars(t *testing.T) {
	t.Parallel()

	reason := skipReason(&skipInput{
		Module: testGo,
		Name:   testInstallPkg,
		Task:   &tasktest.Task{Requires: &tasktest.TaskRequires{Vars: []string{testGOPKG}}},
	})

	requireEqual(t, reason, reasonRequired)
}

// TestSkipYAMLList exercises SkipYAMLList.
func TestSkipYAMLList(t *testing.T) {
	t.Parallel()

	reason := skipReason(&skipInput{
		Module: testYamllint,
		Name:   testGalaxyInstall,
		Config: &smokeConfig{Skip: []string{testGalaxyInstall}},
	})

	requireEqual(t, reason, reasonYAMLSkip)
}

// TestSkipPromptListAndStringList exercises SkipPromptListAndStringList.
func TestSkipPromptListAndStringList(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{
		Module: testGit,
		Name:   "install:undo",
		Task:   &tasktest.Task{Prompt: []any{testContinue}},
	}), reasonPrompt)
	requireEqual(t, skipReason(&skipInput{
		Module: testGit,
		Name:   testOther,
		Task:   &tasktest.Task{Prompt: []string{testContinue}},
	}), reasonPrompt)
}

// TestSkipEmptyPromptIgnored exercises SkipEmptyPromptIgnored.
func TestSkipEmptyPromptIgnored(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{
		Module: testEcho,
		Name:   nameCI,
		Task:   &tasktest.Task{Prompt: "  "},
	}), emptyString)
}

// TestSkipUnknownPromptType exercises SkipUnknownPromptType.
func TestSkipUnknownPromptType(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{
		Module: testEcho,
		Name:   nameCI,
		Task:   &tasktest.Task{Prompt: 1},
	}), reasonPrompt)
}

// TestSkipGitMutators exercises SkipGitMutators.
func TestSkipGitMutators(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{Module: testGit, Name: namePushForce}), reasonDestructive)
	requireEqual(t, skipReason(&skipInput{Module: testGit, Name: nameClean}), reasonDestructive)
}

// TestNameMatchesSuffix exercises NameMatchesSuffix.
func TestNameMatchesSuffix(t *testing.T) {
	t.Parallel()

	requireSame(t, nameMatches("tool:uninstall", nameUninstall), true)
	requireSame(t, !nameMatches("uninstall:pkg", nameUninstall), true)
}

// TestHasNonEmptyVarNilMap exercises HasNonEmptyVarNilMap.
func TestHasNonEmptyVarNilMap(t *testing.T) {
	t.Parallel()

	requireSame(t, !hasNonEmptyVar(nil, nameCLIArgs), true)
}

// TestSkipWatchSuffix exercises SkipWatchSuffix.
func TestSkipWatchSuffix(t *testing.T) {
	t.Parallel()

	requireEqual(
		t,
		skipReason(&skipInput{Module: testGo, Name: nameBuild + suffixWatch}),
		reasonWatch,
	)
	requireEqual(t, skipReason(&skipInput{Module: testGo, Name: nameBuild}), emptyString)
	requireEqual(t, skipReason(&skipInput{Module: testGo, Name: testTypecheck}), emptyString)
	requireEqual(
		t,
		skipReason(&skipInput{Module: testGo, Name: testTypecheck + suffixWatch}),
		reasonWatch,
	)
}

// TestSkipWingetOnUnix exercises SkipWingetOnUnix.
func TestSkipWingetOnUnix(t *testing.T) {
	t.Parallel()

	requireEqual(t, wingetSkipOnOS(toolWinget, osDarwin), reasonWinget)
	requireEqual(t, wingetSkipOnOS(toolWinget, osLinux), reasonWinget)
	requireEqual(t, wingetSkipOnOS(toolWinget, osWindows), emptyString)
	requireEqual(t, wingetSkipOnOS(testGo, osLinux), emptyString)
}

// TestSkipWingetViaPolicy exercises SkipWingetViaPolicy.
func TestSkipWingetViaPolicy(t *testing.T) {
	t.Parallel()

	requireEqual(
		t,
		skipReason(&skipInput{Module: toolWinget, Name: nameInstall}),
		wingetSkipOnOS(toolWinget, runtime.GOOS),
	)
}

// TestSkipUnixOnlyOnWindows exercises SkipUnixOnlyOnWindows.
func TestSkipUnixOnlyOnWindows(t *testing.T) {
	t.Parallel()

	requireEqual(t, unixOnlySkipOnOS(toolAnsible, osWindows), reasonUnixOnly)
	requireEqual(t, unixOnlySkipOnOS(toolAnsibleLint, osWindows), reasonUnixOnly)
	requireEqual(t, unixOnlySkipOnOS(toolNix, osWindows), reasonUnixOnly)
	requireEqual(t, unixOnlySkipOnOS(toolAnsible, osDarwin), emptyString)
	requireEqual(t, unixOnlySkipOnOS(toolAnsible, osLinux), emptyString)
	requireEqual(t, unixOnlySkipOnOS(testGo, osWindows), emptyString)
}

// TestSkipUnixOnlyViaPolicy exercises SkipUnixOnlyViaPolicy.
func TestSkipUnixOnlyViaPolicy(t *testing.T) {
	t.Parallel()

	requireEqual(
		t,
		skipReason(&skipInput{Module: toolAnsible, Name: nameVersion}),
		unixOnlySkipOnOS(toolAnsible, runtime.GOOS),
	)
	requireEqual(
		t,
		skipReason(&skipInput{Module: toolAnsibleLint, Name: nameVersion}),
		unixOnlySkipOnOS(toolAnsibleLint, runtime.GOOS),
	)
	requireEqual(
		t,
		skipReason(&skipInput{Module: toolNix, Name: nameVersion}),
		unixOnlySkipOnOS(toolNix, runtime.GOOS),
	)
}

// TestSkipCargoSourceOnWindows exercises SkipCargoSourceOnWindows.
func TestSkipCargoSourceOnWindows(t *testing.T) {
	t.Parallel()

	requireEqual(t, cargoSourceSkipOnOS(toolAdrs, osWindows), reasonCargoSource)
	requireEqual(t, cargoSourceSkipOnOS(toolAdrs, osDarwin), emptyString)
	requireEqual(t, cargoSourceSkipOnOS(toolAdrs, osLinux), emptyString)
	requireEqual(t, cargoSourceSkipOnOS(testGo, osWindows), emptyString)
}

// TestSkipCargoSourceViaPolicy exercises SkipCargoSourceViaPolicy.
func TestSkipCargoSourceViaPolicy(t *testing.T) {
	t.Parallel()

	requireEqual(
		t,
		skipReason(&skipInput{Module: toolAdrs, Name: nameVersion}),
		cargoSourceSkipOnOS(toolAdrs, runtime.GOOS),
	)
}

// TestSkipGHKeepsAllowed exercises SkipGHKeepsAllowed.
func TestSkipGHKeepsAllowed(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{Module: toolGH, Name: nameInstall}), emptyString)
	requireEqual(t, skipReason(&skipInput{Module: toolGH, Name: nameVersion}), emptyString)
	requireEqual(t, skipReason(&skipInput{Module: toolGH, Name: nameWhich}), emptyString)
	requireEqual(t, skipReason(&skipInput{Module: toolGH, Name: nameHelp}), emptyString)
	requireEqual(t, skipReason(&skipInput{Module: toolGH, Name: nameVerify}), emptyString)
	requireEqual(t, skipReason(&skipInput{Module: toolGH, Name: nameConfigList}), emptyString)
	requireEqual(t, skipReason(&skipInput{Module: toolGH, Name: nameAliasList}), emptyString)
}

// TestSkipGHAuthNetwork exercises SkipGHAuthNetwork.
func TestSkipGHAuthNetwork(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{Module: toolGH, Name: testOther}), reasonGH)
	requireEqual(t, skipReason(&skipInput{Module: testGit, Name: testOther}), emptyString)
}
