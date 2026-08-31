// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskfiles_test

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
	"github.com/task-otter/store/internal/tasktestutil"
)

type (
	// folderScanner collects the module folders under taskfiles/.
	folderScanner struct {
		root    string
		folders []string
	}
)

const (
	integrationTestFileName = "integration_test.go"
	integrationSuiteCall    = "taskintegration.RunHere(t)"
)

// TestEveryTaskfileFolderHasAnIntegrationTest keeps the shared task CLI suite
// wired to every module. A new Taskfile folder added without an
// integration_test.go would otherwise ship without ever being run.
func TestEveryTaskfileFolderHasAnIntegrationTest(t *testing.T) {
	t.Parallel()

	folders := taskfileFolders(t)

	if len(folders) == constZero {
		t.Fatal(noModulesDiscoveredMsg)
	}

	for i := range folders {
		assertFolderRunsIntegrationSuite(t, folders[i])
	}
}

// record keeps folder unless it holds a shared Taskfile fragment. The fragments
// under taskfiles/internal are included by modules and expose no tasks of their own.
func (scanner *folderScanner) record(folder string) {
	module := filepath.ToSlash(strings.TrimPrefix(folder, scanner.root+string(filepath.Separator)))
	shared := module == internalDirName ||
		strings.HasPrefix(module, internalDirName+pathSeparator)

	if shared {
		return
	}

	scanner.folders = append(scanner.folders, folder)
}

func (scanner *folderScanner) visit(path string, entry fs.DirEntry, walkErr error) error {
	if entry != nil && !entry.IsDir() && entry.Name() == skipTaskfileYML {
		scanner.record(filepath.Dir(path))
	}

	return walkErr
}

func taskfileFolders(t *testing.T) []string {
	t.Helper()

	root := filepath.Join(tasktest.RepoRoot(t), taskfilesDirName)
	scanner := &folderScanner{root: root, folders: nil}

	err := filepath.WalkDir(root, scanner.visit)
	if err != nil {
		t.Fatalf(walkTaskfilesErrFormat, err)
	}

	return scanner.folders
}

func assertFolderRunsIntegrationSuite(t *testing.T, folder string) {
	t.Helper()

	path := filepath.Join(folder, integrationTestFileName)

	if !tasktestutil.FileExists(path) {
		t.Fatalf("%s must contain %s", folder, integrationTestFileName)
	}

	if strings.Contains(tasktestutil.ReadFile(t, path), integrationSuiteCall) {
		return
	}

	t.Fatalf("%s must call %s", path, integrationSuiteCall)
}
