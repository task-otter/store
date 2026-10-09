// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"runtime"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

type (
	osSkipCase = struct {
		skip   func(string, string) string
		module string
		goos   string
		want   string
	}
)

// TestSkipPrompt exercises SkipPrompt.
func TestSkipPrompt(t *testing.T) {
	t.Parallel()

	reason := skipReason(&skipInput{
		Config: nil,

		Module: testEcho,
		Name:   "prompted",
		Task:   testPromptTask(testContinue),
	})

	requireEqual(t, reason, reasonPrompt)
}

// TestSkipInteractive exercises SkipInteractive.
func TestSkipInteractive(t *testing.T) {
	t.Parallel()

	reason := skipReason(&skipInput{
		Config: nil,

		Module: "nix",
		Name:   "install:shell",
		Task:   testInteractiveTask(),
	})

	requireEqual(t, reason, reasonInteractive)
}

// TestSkipUninstall exercises SkipUninstall.
func TestSkipUninstall(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(testNamedSkipInput(testGo, nameUninstall)), reasonDestructive)
}

// TestSkipFmtKeepsFmtCheck exercises SkipFmtKeepsFmtCheck.
func TestSkipFmtKeepsFmtCheck(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(testNamedSkipInput(testGo, nameFmt)), reasonFmt)
	requireEqual(
		t,
		skipReason(testNamedSkipInput(testGo, nameFmt+colonSeparator+"check")),
		emptyString,
	)
}

// TestSkipCacheCleanStaysIn exercises SkipCacheCleanStaysIn.
func TestSkipCacheCleanStaysIn(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(testNamedSkipInput(testEslint, nameCacheClean)), emptyString)
}

// TestSkipDockerWorkTasks exercises SkipDockerWorkTasks.
func TestSkipDockerWorkTasks(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(testNamedSkipInput(toolDocker, nameBuild)), reasonDocker)
	requireEqual(t, skipReason(testNamedSkipInput(toolDocker, nameVerify)), reasonDocker)
	requireEqual(t, skipReason(testNamedSkipInput(toolDocker, nameVersion)), reasonDocker)
	requireEqual(t, skipReason(testNamedSkipInput(testGo, nameBuild)), emptyString)
}

// TestSkipNixInstallOnly exercises SkipNixInstallOnly.
func TestSkipNixInstallOnly(t *testing.T) {
	t.Parallel()

	expected := reasonNixInstall

	if unixOnly := unixOnlySkipOnOS(toolNix, runtime.GOOS); unixOnly != emptyString {
		expected = unixOnly
	}

	requireEqual(t, skipReason(testNamedSkipInput(toolNix, nameInstall)), expected)
	requireEqual(t, skipReason(testNamedSkipInput(testYamllint, nameInstall)), emptyString)
}

// TestSkipFuzzWithoutCLIArgs exercises SkipFuzzWithoutCLIArgs.
func TestSkipFuzzWithoutCLIArgs(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(testNamedSkipInput(testGo, nameFuzz)), reasonFuzz)
}

// TestSkipFuzzWithCLIArgs exercises SkipFuzzWithCLIArgs.
func TestSkipFuzzWithCLIArgs(t *testing.T) {
	t.Parallel()

	reason := skipReason(&skipInput{
		Task: nil,

		Module: testGo,
		Name:   nameFuzz,
		Config: &smokeConfig{
			Skip: nil,
			Vars: map[string]string{nameCLIArgs: "-fuzz FuzzName ."},
		},
	})

	requireEqual(t, reason, emptyString)
}

// TestSkipMissingRequiredVars exercises SkipMissingRequiredVars.
func TestSkipMissingRequiredVars(t *testing.T) {
	t.Parallel()

	reason := skipReason(&skipInput{
		Config: nil,

		Module: testGo,
		Name:   testInstallPkg,
		Task:   testRequiredTask(&tasktest.TaskRequires{Vars: []string{testGOPKG}}),
	})

	requireEqual(t, reason, reasonRequired)
}

// TestSkipYAMLList exercises SkipYAMLList.
func TestSkipYAMLList(t *testing.T) {
	t.Parallel()

	reason := skipReason(&skipInput{
		Task: nil,

		Module: testYamllint,
		Name:   testGalaxyInstall,
		Config: &smokeConfig{
			Vars: nil,
			Skip: []string{testGalaxyInstall},
		},
	})

	requireEqual(t, reason, reasonYAMLSkip)
}

// TestSkipPromptListAndStringList exercises SkipPromptListAndStringList.
func TestSkipPromptListAndStringList(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{
		Config: nil,

		Module: testGit,
		Name:   "install:undo",
		Task:   testPromptTask([]any{testContinue}),
	}), reasonPrompt)
	requireEqual(t, skipReason(&skipInput{
		Config: nil,

		Module: testGit,
		Name:   testOther,
		Task:   testPromptTask([]string{testContinue}),
	}), reasonPrompt)
}

// TestSkipEmptyPromptIgnored exercises SkipEmptyPromptIgnored.
func TestSkipEmptyPromptIgnored(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{
		Config: nil,

		Module: testEcho,
		Name:   nameCI,
		Task:   testPromptTask("  "),
	}), emptyString)
}

// TestSkipUnknownPromptType exercises SkipUnknownPromptType.
func TestSkipUnknownPromptType(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{
		Config: nil,

		Module: testEcho,
		Name:   nameCI,
		Task:   testPromptTask(1),
	}), reasonPrompt)
}

// TestSkipGitMutators exercises SkipGitMutators.
func TestSkipGitMutators(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(testNamedSkipInput(testGit, namePushForce)), reasonDestructive)
	requireEqual(t, skipReason(testNamedSkipInput(testGit, nameClean)), reasonDestructive)
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
		skipReason(testNamedSkipInput(testGo, nameBuild+suffixWatch)),
		reasonWatch,
	)
	requireEqual(t, skipReason(testNamedSkipInput(testGo, nameBuild)), emptyString)
	requireEqual(t, skipReason(testNamedSkipInput(testGo, testTypecheck)), emptyString)
	requireEqual(
		t,
		skipReason(testNamedSkipInput(testGo, testTypecheck+suffixWatch)),
		reasonWatch,
	)
}

// TestSkipOnOS exercises SkipOnOS.
func TestSkipOnOS(t *testing.T) {
	t.Parallel()

	assertOSSkipCases(t, osSkipCases())
}

// TestSkipWingetViaPolicy exercises SkipWingetViaPolicy.
func TestSkipWingetViaPolicy(t *testing.T) {
	t.Parallel()

	requireEqual(
		t,
		skipReason(testNamedSkipInput(toolWinget, nameInstall)),
		wingetSkipOnOS(toolWinget, runtime.GOOS),
	)
}

// TestSkipUnixOnlyViaPolicy exercises SkipUnixOnlyViaPolicy.
func TestSkipUnixOnlyViaPolicy(t *testing.T) {
	t.Parallel()

	requireEqual(
		t,
		skipReason(testNamedSkipInput(toolAnsible, nameVersion)),
		unixOnlySkipOnOS(toolAnsible, runtime.GOOS),
	)
	requireEqual(
		t,
		skipReason(testNamedSkipInput(toolAnsibleLint, nameVersion)),
		unixOnlySkipOnOS(toolAnsibleLint, runtime.GOOS),
	)
	requireEqual(
		t,
		skipReason(testNamedSkipInput(toolNix, nameVersion)),
		unixOnlySkipOnOS(toolNix, runtime.GOOS),
	)
}

// TestSkipCargoSourceViaPolicy exercises SkipCargoSourceViaPolicy.
func TestSkipCargoSourceViaPolicy(t *testing.T) {
	t.Parallel()

	requireEqual(
		t,
		skipReason(testNamedSkipInput(toolAdrs, nameVersion)),
		cargoSourceSkipOnOS(toolAdrs, runtime.GOOS),
	)
}

// TestUnixOnlyAndGHSkipOnOS exercises UnixOnlyAndGHSkipOnOS.
func TestUnixOnlyAndGHSkipOnOS(t *testing.T) {
	t.Parallel()

	requireEqual(
		t,
		unixOnlyAndGHSkipOnOS(testNamedSkipInput(toolAnsible, emptyString), osWindows),
		reasonUnixOnly,
	)
	requireEqual(
		t,
		unixOnlyAndGHSkipOnOS(testNamedSkipInput(toolAdrs, emptyString), osWindows),
		reasonCargoSource,
	)
	requireEqual(
		t,
		unixOnlyAndGHSkipOnOS(testNamedSkipInput(testGo, emptyString), osWindows),
		emptyString,
	)
}

// TestSkipGHKeepsAllowed exercises SkipGHKeepsAllowed.
func TestSkipGHKeepsAllowed(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(testNamedSkipInput(toolGH, nameInstall)), emptyString)
	requireEqual(t, skipReason(testNamedSkipInput(toolGH, nameVersion)), emptyString)
	requireEqual(t, skipReason(testNamedSkipInput(toolGH, nameWhich)), emptyString)
	requireEqual(t, skipReason(testNamedSkipInput(toolGH, nameHelp)), emptyString)
	requireEqual(t, skipReason(testNamedSkipInput(toolGH, nameVerify)), emptyString)
	requireEqual(t, skipReason(testNamedSkipInput(toolGH, nameConfigList)), emptyString)
	requireEqual(t, skipReason(testNamedSkipInput(toolGH, nameAliasList)), emptyString)
}

// TestSkipGHAuthNetwork exercises SkipGHAuthNetwork.
func TestSkipGHAuthNetwork(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(testNamedSkipInput(toolGH, testOther)), reasonGH)
	requireEqual(t, skipReason(testNamedSkipInput(testGit, testOther)), emptyString)
}

func assertOSSkipCases(t *testing.T, cases []osSkipCase) {
	t.Helper()

	for i := range cases {
		requireEqual(t, cases[i].skip(cases[i].module, cases[i].goos), cases[i].want)
	}
}

func osSkipCases() []osSkipCase {
	return []osSkipCase{
		{skip: wingetSkipOnOS, module: toolWinget, goos: osDarwin, want: reasonWinget},
		{skip: wingetSkipOnOS, module: toolWinget, goos: osLinux, want: reasonWinget},
		{skip: wingetSkipOnOS, module: toolWinget, goos: osWindows, want: emptyString},
		{skip: wingetSkipOnOS, module: testGo, goos: osLinux, want: emptyString},
		{skip: unixOnlySkipOnOS, module: toolAnsible, goos: osWindows, want: reasonUnixOnly},
		{skip: unixOnlySkipOnOS, module: toolAnsibleLint, goos: osWindows, want: reasonUnixOnly},
		{skip: unixOnlySkipOnOS, module: toolNix, goos: osWindows, want: reasonUnixOnly},
		{skip: unixOnlySkipOnOS, module: toolAnsible, goos: osDarwin, want: emptyString},
		{skip: unixOnlySkipOnOS, module: toolAnsible, goos: osLinux, want: emptyString},
		{skip: unixOnlySkipOnOS, module: testGo, goos: osWindows, want: emptyString},
		{skip: cargoSourceSkipOnOS, module: toolAdrs, goos: osWindows, want: reasonCargoSource},
		{skip: cargoSourceSkipOnOS, module: toolAdrs, goos: osDarwin, want: emptyString},
		{skip: cargoSourceSkipOnOS, module: toolAdrs, goos: osLinux, want: emptyString},
		{skip: cargoSourceSkipOnOS, module: testGo, goos: osWindows, want: emptyString},
	}
}
