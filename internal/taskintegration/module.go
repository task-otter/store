// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskintegration

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
	"github.com/task-otter/store/internal/tasktestutil"
)

type (
	// Module is one Taskfile folder resolved for integration testing.
	Module struct {
		// Name is the folder path relative to taskfiles/.
		Name string

		// Dir is the absolute path of the folder the task CLI runs in.
		Dir string

		// Env is the isolated environment every task run inherits.
		Env []string

		// Listed holds the task names reported by `task --list-all --json`.
		Listed []string

		// Declared holds the public task names the folder's Taskfile defines.
		Declared []string

		// Exported holds metadata.yml exported_tasks, nil when absent.
		Exported []string

		// Variants holds metadata.yml variants, nil when absent.
		Variants []string

		// Includes holds the namespaces the folder's Taskfile includes.
		Includes []string

		// Vars holds the top-level var names the folder's Taskfile declares.
		Vars []string

		// DryRun holds the tasks that must survive `task --dry`.
		DryRun []string
	}
)

const (
	taskfilesDirName = "taskfiles"
	defaultTaskName  = "default"
	namespaceSep     = ":"
	privatePrefix    = "_"
	pathSep          = "/"
	emptyString      = ""
	zeroLength       = 0
)

func newModule(t *testing.T, spec *Spec) *Module {
	t.Helper()

	module := newModuleShell(t, spec)

	module.Listed = listTaskNames(t, module)

	return module
}

func newModuleShell(t *testing.T, spec *Spec) *Module {
	t.Helper()

	taskfile := tasktest.LoadTaskfile(t, spec.Module)
	metadata := loadMetadata(t, moduleDir(t, spec.Module))

	return &Module{
		Name:     spec.Module,
		Dir:      moduleDir(t, spec.Module),
		Env:      tasktestutil.IsolatedEnv(t),
		Listed:   nil,
		Declared: declaredPublicTasks(taskfile),
		Exported: metadata.ExportedTasks,
		Variants: metadata.Variants,
		Includes: includeNamespaces(taskfile),
		Vars:     varNames(taskfile),
		DryRun:   spec.DryRunTasks,
	}
}

func moduleDir(t *testing.T, module string) string {
	t.Helper()

	return filepath.Join(tasktest.RepoRoot(t), taskfilesDirName, module)
}

// currentModule reports the taskfiles-relative path of the folder the calling
// test lives in, which go test runs as the working directory.
func currentModule(t *testing.T) string {
	t.Helper()

	return modulePathFrom(t, tasktest.RepoRoot(t), tasktestutil.ModuleRoot(t))
}

func modulePathFrom(t suiteT, repoRoot, moduleRoot string) string {
	t.Helper()

	root := filepath.Join(repoRoot, taskfilesDirName)

	module, err := filepath.Rel(root, moduleRoot)
	if err != nil {
		t.Fatalf("resolve the module folder under %s: %v", taskfilesDirName, err)
	}

	return filepath.ToSlash(module)
}

func declaredPublicTasks(taskfile *tasktest.Taskfile) []string {
	names := make([]string, zeroLength, len(taskfile.Tasks))

	for name := range taskfile.Tasks {
		if isPublicTask(name, taskfile.Tasks[name]) {
			names = append(names, name)
		}
	}

	slices.Sort(names)

	return names
}

func isPublicTask(name string, task *tasktest.Task) bool {
	if name == defaultTaskName || strings.HasPrefix(name, privatePrefix) {
		return false
	}

	return !task.Internal
}

func varNames(taskfile *tasktest.Taskfile) []string {
	names := make([]string, zeroLength, len(taskfile.Vars))

	for name := range taskfile.Vars {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

func includeNamespaces(taskfile *tasktest.Taskfile) []string {
	names := make([]string, zeroLength, len(taskfile.Includes))

	for name := range taskfile.Includes {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}
