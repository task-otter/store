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
	err := applyIsolatedHomeEnv(home, os.Setenv)
	if err != nil {
		return fmt.Errorf("apply isolated home: %w", err)
	}

	return nil
}

func applyIsolatedHomeEnv(home string, setter func(string, string) error) error {
	profile := filepath.Join(home, bashrcName)

	err := os.WriteFile(profile, []byte(emptyString), fileMode)
	if err != nil {
		return fmt.Errorf("write %s: %w", profile, err)
	}

	envErr := applyEnvPairs(smokeEnvPairs(home, profile), setter)
	if envErr != nil {
		return fmt.Errorf("set smoke env: %w", envErr)
	}

	return nil
}

func isolatedHomeDir(runner *engine) (string, error) {
	home, err := runner.mkdirTemp()
	if err != nil {
		return emptyString, fmt.Errorf("create isolated home: %w", err)
	}

	applyErr := runner.applyHome(home)
	if applyErr != nil {
		return emptyString, fmt.Errorf("apply isolated home: %w", applyErr)
	}

	return home, nil
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
