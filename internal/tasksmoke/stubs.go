// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"fmt"
	"os"
	"path/filepath"
)

func bindMkdirTemp(dir string) mkdirTempFunc {
	return func() (string, error) {
		created, err := mkdirTempIn(dir, tempWorkPrefix)
		if err != nil {
			return emptyString, fmt.Errorf("create work temp dir: %w", err)
		}

		return created, nil
	}
}

func copyDirectory(src, dst string, info os.FileInfo) error {
	if !info.IsDir() {
		return fmt.Errorf(errWrapFormat, src, os.ErrInvalid)
	}

	err := os.CopyFS(dst, os.DirFS(src))
	if err != nil {
		return fmt.Errorf(errCopyFormat, src, err)
	}

	return nil
}

func copyExistingTree(src, dst string) error {
	info, err := os.Stat(src)

	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf(errStatFormat, src, err)
	}

	copyErr := copyDirectory(src, dst, info)
	if copyErr != nil {
		return fmt.Errorf("copy existing tree: %w", copyErr)
	}

	return nil
}

func copyStubTree(src, dst string) error {
	err := os.MkdirAll(dst, dirMode)
	if err != nil {
		return fmt.Errorf(errMkdirFormat, dst, err)
	}

	copyErr := copyExistingTree(src, dst)
	if copyErr != nil {
		return fmt.Errorf("copy stub tree: %w", copyErr)
	}

	return nil
}

func mkdirTempIn(dir, prefix string) (string, error) {
	created, err := os.MkdirTemp(dir, prefix)
	if err != nil {
		return emptyString, fmt.Errorf("create temp dir: %w", err)
	}

	return created, nil
}

func stubSource(repoRoot, module string) string {
	return filepath.Join(repoRoot, dataTestDirName, toolName(module))
}
