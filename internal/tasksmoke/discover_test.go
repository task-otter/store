// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"path/filepath"
	"testing"
)

// TestDiscoverModulesFindsPublicLeaves exercises DiscoverModulesFindsPublicLeaves.
func TestDiscoverModulesFindsPublicLeaves(t *testing.T) {
	t.Parallel()

	modules, err := discoverModules(testdataRepo(t))
	requireNoErr(t, err)
	requireSame(t, hasModule(modules, testEcho), true)
	requireSame(t, hasModule(modules, testFamilyLeaf), true)
	requireSame(t, hasModule(modules, testInternalShared), false)
	requireSame(t, hasModule(modules, internalDirName), false)
}

// TestDiscoverModulesRejectsInvalidTaskfile exercises DiscoverModulesRejectsInvalidTaskfile.
func TestDiscoverModulesRejectsInvalidTaskfile(t *testing.T) {
	t.Parallel()

	value, err := discoverModules(testdataAbs(t, testdataBadPath))
	keepValue(value)
	requireErr(t, err)
}

// TestDiscoverModulesRejectsMissingRoot exercises DiscoverModulesRejectsMissingRoot.
func TestDiscoverModulesRejectsMissingRoot(t *testing.T) {
	t.Parallel()

	value, err := discoverModules(t.TempDir())
	keepValue(value)
	requireErr(t, err)
}

// TestToolNameUsesFirstSegment exercises ToolNameUsesFirstSegment.
func TestToolNameUsesFirstSegment(t *testing.T) {
	t.Parallel()

	requireEqual(t, toolName(testEslintNPM), testEslint)
	requireEqual(t, toolName(testYamllint), testYamllint)
}

// TestDeclaredPublicTasksSkipHidden exercises DeclaredPublicTasksSkipHidden.
func TestDeclaredPublicTasksSkipHidden(t *testing.T) {
	t.Parallel()

	modules, err := discoverModules(testdataRepo(t))
	requireNoErr(t, err)

	echoModule := mustModule(t, modules, testEcho)
	requireSame(t, hasTask(echoModule, testPing), true)
	requireSame(t, hasTask(echoModule, defaultTaskName), false)
	requireSame(t, hasTask(echoModule, testHiddenTask), false)
	requireSame(t, hasTask(echoModule, testSecretTask), false)
}

func hasModule(modules []*module, name string) bool {
	for i := range modules {
		if modules[i].Name == name {
			return true
		}
	}

	return false
}

func hasTask(module *module, name string) bool {
	for i := range module.Tasks {
		if module.Tasks[i].Name == name {
			return true
		}
	}

	return false
}

func mustModule(t *testing.T, modules []*module, name string) *module {
	t.Helper()

	for i := range modules {
		if modules[i].Name == name {
			return modules[i]
		}
	}

	t.Fatalf("module %q not found", name)

	return nil
}

// TestStubSourceJoinsTool exercises StubSourceJoinsTool.
func TestStubSourceJoinsTool(t *testing.T) {
	t.Parallel()

	got := stubSource(filepath.FromSlash(testRepoSlash), testEslintNPM)
	want := filepath.Join(filepath.FromSlash(testRepoSlash), dataTestDirName, testEslint)

	requireEqual(t, got, want)
}
