// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/task-otter/store/internal/tasktest"
)

func discoverModules(repoRoot string) ([]*module, error) {
	scanner := newFolderScanner(repoRoot)

	err := walkTaskfiles(scanner)
	if err != nil {
		return nil, fmt.Errorf("walk modules: %w", err)
	}

	modules, loadErr := loadScannedModules(scanner)
	if loadErr != nil {
		return nil, fmt.Errorf(errWrapFormat, errLoadScanned, loadErr)
	}

	return modules, nil
}

func loadModule(taskfilesRoot, folder string) (*module, error) {
	name := filepath.ToSlash(strings.TrimPrefix(folder, taskfilesRoot+string(filepath.Separator)))

	taskfile, err := readModuleTaskfile(folder)
	if err != nil {
		return nil, fmt.Errorf("load module taskfile: %w", err)
	}

	return newLoadedModule(folder, name, taskfile), nil
}

func loadScannedModules(scanner *folderScan) ([]*module, error) {
	modules := make([]*module, emptyLength, len(scanner.folders))

	for i := range scanner.folders {
		loaded, err := loadModule(scanner.root, scanner.folders[i])
		if err != nil {
			return nil, fmt.Errorf("load scanned module: %w", err)
		}

		modules = append(modules, loaded)
	}

	return modules, nil
}

func maybeRecordFolder(scanner *folderScan, path string, entry fs.DirEntry) {
	if entry == nil || entry.IsDir() || entry.Name() != taskfileName {
		return
	}

	recordFolder(scanner, filepath.Dir(path))
}

func newFolderScanner(repoRoot string) *folderScan {
	return &folderScan{
		root:    filepath.Join(repoRoot, taskfilesDirName),
		folders: nil,
	}
}

func newLoadedModule(folder, name string, taskfile *tasktest.Taskfile) *module {
	return &module{
		Config: nil,
		Dir:    folder,
		Name:   name,
		Tasks:  declaredPublicTasks(taskfile),
	}
}

func parseModuleTaskfile(path string, content []byte) (*tasktest.Taskfile, error) {
	taskfile, err := tasktest.ParseTaskfile(content)
	if err != nil {
		return nil, fmt.Errorf(errParseFormat, path, err)
	}

	return taskfile, nil
}

func readFile(path string) ([]byte, error) {
	clean := filepath.Clean(path)

	content, err := fs.ReadFile(os.DirFS(filepath.Dir(clean)), filepath.Base(clean))
	if err != nil {
		return nil, fmt.Errorf(errReadFormat, clean, err)
	}

	return content, nil
}

func readModuleTaskfile(folder string) (*tasktest.Taskfile, error) {
	path := filepath.Join(folder, taskfileName)

	content, err := readFile(path)
	if err != nil {
		return nil, fmt.Errorf(errWrapFormat, errReadModule, err)
	}

	taskfile, parseErr := parseModuleTaskfile(path, content)
	if parseErr != nil {
		return nil, fmt.Errorf(errWrapFormat, errParseModule, parseErr)
	}

	return taskfile, nil
}

func recordFolder(scanner *folderScan, folder string) {
	name := relativeModule(scanner, folder)

	if isSharedFragment(name) {
		return
	}

	scanner.folders = append(scanner.folders, folder)
}

func relativeModule(scanner *folderScan, folder string) string {
	prefix := scanner.root + string(filepath.Separator)

	return filepath.ToSlash(strings.TrimPrefix(folder, prefix))
}

func isSharedFragment(name string) bool {
	if name == internalDirName {
		return true
	}

	return strings.HasPrefix(name, internalDirName+pathSeparator)
}

func toolName(name string) string {
	parts := strings.SplitN(name, pathSeparator, nameSplitParts)

	return parts[emptyLength]
}

func visitFolder(visit *walkVisit) error {
	if visit.Err != nil {
		return fmt.Errorf("walk taskfile folder: %w", visit.Err)
	}

	maybeRecordFolder(visit.Scan, visit.Path, visit.Entry)

	return nil
}

func walkTaskfiles(scanner *folderScan) error {
	err := filepath.WalkDir(
		scanner.root,
		func(path string, entry fs.DirEntry, walkErr error) error {
			return visitFolder(&walkVisit{Entry: entry, Err: walkErr, Path: path, Scan: scanner})
		},
	)
	if err != nil {
		return fmt.Errorf(errWalkFormat, err)
	}

	return nil
}
