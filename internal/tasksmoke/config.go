// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"fmt"
	"path/filepath"

	yaml "go.yaml.in/yaml/v3"
)

func emptySmokeConfig() *smokeConfig {
	return &smokeConfig{
		Skip: nil,
		Vars: nil,
	}
}

func loadSmokeConfig(path string) (*smokeConfig, error) {
	if !pathExists(path) {
		return emptySmokeConfig(), nil
	}

	config, err := decodeSmokeFile(path)
	if err != nil {
		return nil, fmt.Errorf(errWrapFormat, errDecodeSmoke, err)
	}

	return config, nil
}

func decodeSmokeFile(path string) (*smokeConfig, error) {
	content, err := readFile(path)
	if err != nil {
		return nil, fmt.Errorf("read smoke config: %w", err)
	}

	config, parseErr := parseLoadedConfig(path, content)
	if parseErr != nil {
		return nil, fmt.Errorf(errWrapFormat, errParseLoaded, parseErr)
	}

	return config, nil
}

func loadToolSmokeConfig(path string) (*smokeConfig, error) {
	config, err := loadSmokeConfig(path)
	if err != nil {
		return nil, fmt.Errorf("load tool smoke config: %w", err)
	}

	return config, nil
}

func moduleSmokeConfig(repoRoot, module string) (*smokeConfig, error) {
	path := filepath.Join(repoRoot, dataTestDirName, toolName(module), smokeConfigName)

	config, err := loadToolSmokeConfig(path)
	if err != nil {
		return nil, fmt.Errorf("module smoke config: %w", err)
	}

	return config, nil
}

func parseLoadedConfig(path string, content []byte) (*smokeConfig, error) {
	config, err := parseSmokeConfig(path, content)
	if err != nil {
		return nil, fmt.Errorf("parse smoke config: %w", err)
	}

	return config, nil
}

func parseSmokeConfig(path string, content []byte) (*smokeConfig, error) {
	config := emptySmokeConfig()

	err := yaml.Unmarshal(content, config)
	if err != nil {
		return nil, fmt.Errorf(errParseFormat, path, err)
	}

	return config, nil
}
