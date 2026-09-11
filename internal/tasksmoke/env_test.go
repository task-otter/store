// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestApplyIsolatedHomeEnvError exercises ApplyIsolatedHomeEnvError.
func TestApplyIsolatedHomeEnvError(t *testing.T) {
	t.Parallel()

	err := applyIsolatedHomeEnv(t.TempDir(), testHostHome, func(string, string) error {
		return sentinelErr()
	})
	requireErr(t, err)
}

// TestApplyIsolatedHomeEnvWritesHostNixPath exercises ApplyIsolatedHomeEnvWritesHostNixPath.
func TestApplyIsolatedHomeEnvWritesHostNixPath(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	err := applyIsolatedHomeEnv(home, testHostHome, discardEnv)
	requireNoErr(t, err)
	assertLoginProfiles(t, home)
	assertProfileUsesHostNix(t, home)
	requireSame(t, pathExists(filepath.Join(home, configDirName)), true)
	requireSame(t, pathExists(filepath.Join(home, nixProfileDir)), false)
}

// TestApplyIsolatedHomeWritesProfile exercises ApplyIsolatedHomeWritesProfile.
func TestApplyIsolatedHomeWritesProfile(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	err := applyIsolatedHome(home)
	requireNoErr(t, err)
	assertLoginProfiles(t, home)
}

// TestDetectRepoRoot exercises DetectRepoRoot.
func TestDetectRepoRoot(t *testing.T) {
	t.Parallel()

	root, err := detectRepoRootFrom(os.Getwd)
	requireNoErr(t, err)
	requireSame(t, pathExists(filepath.Join(root, goModFileName)), true)
}

// TestIsolatedHomeDirMkdirError exercises IsolatedHomeDirMkdirError.
func TestIsolatedHomeDirMkdirError(t *testing.T) {
	t.Parallel()

	engine := testEngine(t)

	engine.mkdirTemp = func() (string, error) {
		return emptyString, sentinelErr()
	}

	value, err := isolatedHomeDir(engine)
	keepValue(value)
	requireErr(t, err)
}

// TestIsolatedHomeDirApplyError exercises IsolatedHomeDirApplyError.
func TestIsolatedHomeDirApplyError(t *testing.T) {
	t.Parallel()

	engine := testEngine(t)

	engine.applyHome = func(string) error {
		return sentinelErr()
	}

	value, err := isolatedHomeDir(engine)
	keepValue(value)
	requireErr(t, err)
}

func assertLoginProfiles(t *testing.T, home string) {
	t.Helper()

	names := loginProfileNames()

	for i := range names {
		requireSame(t, pathExists(filepath.Join(home, names[i])), true)
	}
}

func assertProfileUsesHostNix(t *testing.T, home string) {
	t.Helper()

	body := readIsolatedProfile(t, home, bashProfileName)
	requireSame(t, strings.Contains(body, hostNixPath(testHostHome)), true)
	requireSame(t, strings.Contains(body, nixDaemonSh), true)
	requireSame(t, strings.Contains(body, filepath.Join(home, nixProfileDir)), false)
}

func discardEnv(key, value string) error {
	keepValue(key)
	keepValue(value)

	return nil
}

func readIsolatedProfile(t *testing.T, home, name string) string {
	t.Helper()

	body, err := os.ReadFile(filepath.Join(home, name))
	requireNoErr(t, err)

	return string(body)
}
