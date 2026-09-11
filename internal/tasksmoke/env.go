// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func applyEnvPairs(pairs []envPair, setter func(string, string) error) error {
	for i := range pairs {
		err := setter(pairs[i][0], pairs[i][1])
		if err != nil {
			return fmt.Errorf("set %s: %w", pairs[i][0], err)
		}
	}

	return nil
}

func applyIsolatedHome(home string) error {
	err := applyIsolatedHomeEnv(home, os.Getenv(envHome), os.Setenv)
	if err != nil {
		return fmt.Errorf(errWrapFormat, errApplyIsolatedHome, err)
	}

	return nil
}

func applyIsolatedHomeEnv(home, hostHome string, setter func(string, string) error) error {
	err := writeIsolatedHomeFiles(home, hostHome)
	if err != nil {
		return fmt.Errorf("write isolated home files: %w", err)
	}

	envErr := applySmokeEnv(home, setter)
	if envErr != nil {
		return fmt.Errorf("set isolated env: %w", envErr)
	}

	return nil
}

func applySmokeEnv(home string, setter func(string, string) error) error {
	err := applyEnvPairs(smokeEnvPairs(home, filepath.Join(home, bashrcName)), setter)
	if err != nil {
		return fmt.Errorf("set smoke env: %w", err)
	}

	return nil
}

func ensureIsolatedConfigDir(home string) error {
	err := os.MkdirAll(filepath.Join(home, configDirName), dirMode)
	if err != nil {
		return fmt.Errorf("mkdir isolated config: %w", err)
	}

	return nil
}

func hostNixPath(hostHome string) string {
	return filepath.Join(hostHome, nixProfileDir, nixBinName) + colonSeparator + nixDefaultBin
}

func isolatedHomeDir(runner *engine) (string, error) {
	home, err := runner.mkdirTemp()
	if err != nil {
		return emptyString, fmt.Errorf("create isolated home: %w", err)
	}

	applyErr := runner.applyHome(home)
	if applyErr != nil {
		return emptyString, fmt.Errorf(errWrapFormat, errApplyIsolatedHome, applyErr)
	}

	return home, nil
}

func isolatedProfileBody(hostHome string) string {
	return nixDaemonBlock() + pathExportLine(hostHome)
}

func loginProfileNames() []string {
	return []string{bashProfileName, profileName, bashrcName}
}

func nixDaemonBlock() string {
	return shellIfExistsPrefix + nixDaemonSh + shellThenSource + nixDaemonSh + shellIfEnd
}

func pathExportLine(hostHome string) string {
	return exportPathPrefix + hostNixPath(hostHome) + exportPathSuffix
}

func setEnvPairs(pairs []envPair) error {
	err := applyEnvPairs(pairs, os.Setenv)
	if err != nil {
		return fmt.Errorf("set env pairs: %w", err)
	}

	return nil
}

func smokeEnvPairs(home, profile string) []envPair {
	return []envPair{
		{envHome, home},
		{envProfile, profile},
		{envZdotdir, home},
		{envCI, envTrue},
		{envTaskColor, strconv.Itoa(emptyLength)},
		{envNoColor, strconv.Itoa(exitFail)},
	}
}

func writeIsolatedHomeFiles(home, hostHome string) error {
	err := ensureIsolatedConfigDir(home)
	if err != nil {
		return fmt.Errorf("ensure isolated config: %w", err)
	}

	return writeLoginShellProfiles(home, hostHome)
}

func writeLoginShellProfiles(home, hostHome string) error {
	err := writeProfileNames(home, isolatedProfileBody(hostHome), loginProfileNames())
	if err != nil {
		return fmt.Errorf("write login profiles: %w", err)
	}

	return nil
}

func writeProfileFile(home, name, body string) error {
	path := filepath.Join(home, name)
	err := os.WriteFile(path, []byte(body), fileMode)
	if err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

func writeProfileNames(home, body string, names []string) error {
	for i := range names {
		err := writeProfileFile(home, names[i], body)
		if err != nil {
			return fmt.Errorf("write login profile: %w", err)
		}
	}

	return nil
}
