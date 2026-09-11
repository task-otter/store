// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"fmt"
	"os"
	"path/filepath"
)

func detectRepoRootFrom(getwd func() (string, error)) (string, error) {
	workingDirectory, err := getwd()
	if err != nil {
		return emptyString, fmt.Errorf(errGetwdFormat, err)
	}

	root, findErr := findRepoRoot(workingDirectory)
	if findErr != nil {
		return emptyString, fmt.Errorf("find repo root: %w", findErr)
	}

	return root, nil
}

func findRepoRoot(directory string) (string, error) {
	if pathExists(filepath.Join(directory, goModFileName)) {
		return directory, nil
	}

	root, err := findRepoRootParent(directory)
	if err != nil {
		return emptyString, fmt.Errorf("find repo root parent: %w", err)
	}

	return root, nil
}

func findRepoRootParent(directory string) (string, error) {
	parent := filepath.Dir(directory)

	if parent == directory {
		return emptyString, errRepoMissing
	}

	root, err := findRepoRoot(parent)
	if err != nil {
		return emptyString, fmt.Errorf("walk repo parent: %w", err)
	}

	return root, nil
}

func pathExists(path string) bool {
	info, err := os.Stat(path)

	if info == nil {
		return false
	}

	return err == nil
}
