// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
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
	requireEqual(t, skipReason(&skipInput{Module: testGo, Name: nameBuild}), emptyString)
}

// TestSkipNixInstallOnly exercises SkipNixInstallOnly.
func TestSkipNixInstallOnly(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{Module: toolNix, Name: nameInstall}), reasonNixInstall)
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
		Module: testAnsible,
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
		Name:   testCI,
		Task:   &tasktest.Task{Prompt: "  "},
	}), emptyString)
}

// TestSkipUnknownPromptType exercises SkipUnknownPromptType.
func TestSkipUnknownPromptType(t *testing.T) {
	t.Parallel()

	requireEqual(t, skipReason(&skipInput{
		Module: testEcho,
		Name:   testCI,
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
