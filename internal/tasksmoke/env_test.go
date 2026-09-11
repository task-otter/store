// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"os"
	"path/filepath"
	"testing"
)

// TestApplyIsolatedHomeEnvError exercises ApplyIsolatedHomeEnvError.
func TestApplyIsolatedHomeEnvError(t *testing.T) {
	t.Parallel()

	err := applyIsolatedHomeEnv(t.TempDir(), func(string, string) error {
		return sentinelErr()
	})
	requireErr(t, err)
}

// TestApplyIsolatedHomeWritesProfile exercises ApplyIsolatedHomeWritesProfile.
func TestApplyIsolatedHomeWritesProfile(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	err := applyIsolatedHome(home)
	requireNoErr(t, err)
	requireSame(t, pathExists(filepath.Join(home, bashrcName)), true)
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
