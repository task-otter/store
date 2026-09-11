// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadSmokeConfigMissingFile exercises LoadSmokeConfigMissingFile.
func TestLoadSmokeConfigMissingFile(t *testing.T) {
	t.Parallel()

	config, err := loadSmokeConfig(filepath.Join(t.TempDir(), smokeConfigName))
	requireNoErr(t, err)
	requireSame(t, config != nil, true)
}

// TestLoadSmokeConfigParsesVarsAndSkip exercises LoadSmokeConfigParsesVarsAndSkip.
func TestLoadSmokeConfigParsesVarsAndSkip(t *testing.T) {
	t.Parallel()

	path := filepath.Join(testdataRepo(t), dataTestDirName, testSkipme, smokeConfigName)
	config, err := loadSmokeConfig(path)
	requireNoErr(t, err)
	requireEqual(t, config.Vars[testSample], testValue)
	requireEqual(t, config.Skip[0], testGalaxyInstall)
}

// TestLoadSmokeConfigRejectsInvalidYAML exercises LoadSmokeConfigRejectsInvalidYAML.
func TestLoadSmokeConfigRejectsInvalidYAML(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), smokeConfigName)
	err := os.WriteFile(path, []byte(":\n"), fileMode)
	requireNoErr(t, err)

	value, parseErr := loadSmokeConfig(path)
	keepValue(value)
	requireErr(t, parseErr)
}

// TestLoadSmokeConfigRejectsUnreadableFile exercises LoadSmokeConfigRejectsUnreadableFile.
func TestLoadSmokeConfigRejectsUnreadableFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), smokeConfigName)
	err := os.WriteFile(path, []byte("vars: {}\n"), fileMode)
	requireNoErr(t, err)

	chmodErr := os.Chmod(path, emptyLength)
	requireNoErr(t, chmodErr)

	t.Cleanup(func() {
		restoreErr := os.Chmod(path, fileMode)
		requireNoErr(t, restoreErr)
	})

	value, readErr := loadSmokeConfig(path)
	keepValue(value)
	requireErr(t, readErr)
}

// TestModuleSmokeConfigUsesToolName exercises ModuleSmokeConfigUsesToolName.
func TestModuleSmokeConfigUsesToolName(t *testing.T) {
	t.Parallel()

	config, err := moduleSmokeConfig(testdataRepo(t), testSkipme)
	requireNoErr(t, err)
	requireEqual(t, config.Skip[0], testGalaxyInstall)
}

// TestEmptySmokeConfigNilMaps exercises EmptySmokeConfigNilMaps.
func TestEmptySmokeConfigNilMaps(t *testing.T) {
	t.Parallel()

	config := emptySmokeConfig()
	requireSame(t, config.Vars == nil, true)
}
