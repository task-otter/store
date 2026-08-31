// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskintegration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/task-otter/store/internal/taskintegration"
	"github.com/task-otter/store/internal/tasktest"
)

const (
	fixtureModule   = "jq"
	shfmtModule     = "shfmt"
	biomeModule     = "biome"
	defaultTaskName = "default"
	stdoutText      = "listing"
	stderrText      = "warning"
	taskfilesDir    = "taskfiles"
	dirMode         = 0o700
	chdirLockRetry  = time.Millisecond
)

// TestRunExercisesEveryCheck runs the default suite against a real module so
// every check in the suite executes at least once.
func TestRunExercisesEveryCheck(t *testing.T) {
	t.Parallel()

	taskintegration.Run(t, fixtureModule)
}

// TestRunExercisesBiome runs the suite against a family with variants.
func TestRunExercisesBiome(t *testing.T) {
	t.Parallel()

	taskintegration.Run(t, biomeModule)
}

// TestRunExercisesShfmt runs the suite against a module with public tasks.
func TestRunExercisesShfmt(t *testing.T) {
	t.Parallel()

	taskintegration.Run(t, shfmtModule)
}

// TestRunHereExercisesCurrentModule runs the suite from the module working directory.
func TestRunHereExercisesCurrentModule(t *testing.T) {
	inDir(t, jqModuleDir(t), func() {
		taskintegration.RunHere(t)
	})

	t.Parallel()
}

// TestRunSpecDryRunsSelectedTasks covers the opt-in dry-run check, which the
// default suite skips because most modules need a tool that is not installed.
func TestRunSpecDryRunsSelectedTasks(t *testing.T) {
	t.Parallel()

	taskintegration.RunSpec(t, &taskintegration.Spec{
		Module:      fixtureModule,
		DryRunTasks: []string{defaultTaskName},
	})
}

// TestResultCombinedAppendsStderr checks the output helper the failure messages rely on.
func TestResultCombinedAppendsStderr(t *testing.T) {
	t.Parallel()

	result := taskintegration.Result{
		Stdout: stdoutText,
		Stderr: stderrText,
		Args:   nil,
		Failed: false,
	}

	combined := result.Combined()
	want := stdoutText + "\n" + stderrText

	if combined != want {
		t.Fatalf("Combined() = %q, want %q", combined, want)
	}
}

func changeDir(t *testing.T, dir string) {
	t.Helper()

	err := syscall.Chdir(dir)
	if err != nil {
		t.Fatalf("change directory to %s: %v", dir, err)
	}
}

func inDir(t *testing.T, dir string, callback func()) {
	t.Helper()

	unlock := lockWorkingDirectory(t)

	defer unlock()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current directory: %v", err)
	}

	changeDir(t, dir)

	defer restoreDir(t, previous)

	callback()
}

func jqModuleDir(t *testing.T) string {
	t.Helper()

	return filepath.Join(tasktest.RepoRoot(t), taskfilesDir, fixtureModule)
}

func lockAcquired(t *testing.T, lockPath string) bool {
	t.Helper()

	err := os.Mkdir(lockPath, dirMode)
	if err == nil {
		return true
	}

	if !os.IsExist(err) {
		t.Fatalf("lock working directory: %v", err)
	}

	return false
}

func lockWorkingDirectory(t *testing.T) func() {
	t.Helper()

	lockPath := workingDirectoryLockPath(t)

	for {
		if lockAcquired(t, lockPath) {
			return func() { removePath(t, lockPath) }
		}

		time.Sleep(chdirLockRetry)
	}
}

func removePath(t *testing.T, path string) {
	t.Helper()

	err := os.Remove(path)
	if err != nil {
		t.Fatal(err)
	}
}

func restoreDir(t *testing.T, previous string) {
	t.Helper()

	err := syscall.Chdir(previous)
	if err != nil {
		t.Fatalf("restore directory: %v", err)
	}
}

func workingDirectoryLockPath(t *testing.T) string {
	t.Helper()

	return filepath.Join(
		filepath.Dir(filepath.Dir(t.TempDir())),
		fmt.Sprintf("taskintegration-chdir-%d.lock", os.Getpid()),
	)
}
