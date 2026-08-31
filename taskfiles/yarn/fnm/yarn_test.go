// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package fnm_test

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/task-otter/store/internal/tasktestutil"
	yaml "go.yaml.in/yaml/v3"
)

const (
	constYarnTestYes          = "--yes"
	constYarnTaskInstallClean = "install:clean"
	constYarnTaskInstall      = "install"
	constYarnTaskRun          = "run"
	constYarnTaskVersion      = "version"
	constYarnCurrentDir       = "."
	constYarnPathEnvVar       = "PATH"
	constYarnDirMode          = 0o700
	constYarnMinTasks         = 0
	constYarnScriptTest       = "SCRIPT=test"
	constYarnDoubleDash       = "--"
	constYarnWatchFlag        = "--watch"
)

// TestTaskfileAndReadmePublicApi
func TestTaskfileAndReadmePublicApi(t *testing.T) {
	t.Parallel()

	tasks := loadTaskfileTasks(t)
	actual := tasktestutil.SimplePublicTaskNames(tasks)

	if !slices.Equal(publicTasks(), actual) {
		t.Fatalf("public task drift\nexpected: %v\nactual:   %v", publicTasks(), actual)
	}

	assertReadmeTasksMatch(t)
}

// TestStubbedYarnFlows
func TestStubbedYarnFlows(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("Unix shell stubs cover these flows")
	}

	env := stubEnv(t)

	runStubbedYarnTasks(t, env)
	assertUnsafeScriptRejected(t, env)
}

func publicTasks() []string {
	return append(publicTasksCore(), publicTasksExtra()...)
}

func publicTasksCore() []string {
	return []string{
		"add",
		"audit",
		"audit:json",
		"audit:report",
		"build",
		"cache:clean",
		"ci:fix",
		"clean",
		"clean:all",
		"dev",
		"exec",
		constYarnTaskInstall,
		constYarnTaskInstallClean,
		"install:undo",
	}
}

func publicTasksExtra() []string {
	return []string{
		"lint",
		"manager:pin",
		"manager:setup",
		"node:setup",
		"remove",
		constYarnTaskRun,
		"test",
		"typecheck",
		"update",
		"upgrade",
		constYarnTaskVersion,
	}
}

func loadTaskfileTasks(t *testing.T) map[string]any {
	t.Helper()

	doc := loadTaskfile(t)

	var root map[string]any

	err := doc.Decode(&root)
	if err != nil {
		t.Fatalf("decode Taskfile: %v", err)
	}

	tasks, ok := root["tasks"].(map[string]any)

	if !ok || len(tasks) == constYarnMinTasks {
		t.Fatal("Taskfile tasks map is missing")
	}

	return tasks
}

func assertReadmeTasksMatch(t *testing.T) {
	t.Helper()

	readmeTasks := tasktestutil.ReadmePublicTaskNames(
		tasktestutil.MustRead(t, tasktestutil.ModuleReadmePath(t)),
	)

	if !slices.Equal(publicTasks(), readmeTasks) {
		t.Fatalf("README public task drift\nexpected: %v\nactual:   %v", publicTasks(), readmeTasks)
	}
}

func stubbedYarnTaskArgs() [][]string {
	return [][]string{
		{constYarnTestYes, constYarnTaskVersion},
		{constYarnTestYes, constYarnTaskInstall},
		{constYarnTestYes, constYarnTaskInstallClean},
		{
			constYarnTestYes,
			constYarnTaskRun,
			constYarnScriptTest,
			constYarnDoubleDash,
			constYarnWatchFlag,
		},
	}
}

func runStubbedYarnTasks(t *testing.T, env []string) {
	t.Helper()

	variants := stubbedYarnTaskArgs()

	for i := range variants {
		args := variants[i]
		result := tasktestutil.RunSimpleTask(
			t,
			tasktestutil.TaskRun{Root: constYarnCurrentDir, Env: env, Args: args},
		)

		if result.Err != nil {
			t.Fatalf("task %v failed:\n%s", args, result.Stdout)
		}
	}
}

func assertUnsafeScriptRejected(t *testing.T, env []string) {
	t.Helper()

	result := tasktestutil.RunSimpleTask(t, tasktestutil.TaskRun{
		Root: constYarnCurrentDir,
		Env:  env,
		Args: []string{constYarnTestYes, constYarnTaskRun, "SCRIPT=dev; exit 1"},
	})

	if result.Err == nil {
		t.Fatalf("unsafe SCRIPT unexpectedly succeeded:\n%s", result.Stdout)
	}
}

func stubEnv(t *testing.T) []string {
	t.Helper()

	home, binDir := stubEnvDirs(t)

	writeFnmStubs(t, binDir)

	return stubEnvVars(home, binDir)
}

func stubEnvDirs(t *testing.T) (home, binDir string) {
	t.Helper()

	home = t.TempDir()
	binDir = filepath.Join(home, ".local", "bin")

	err := os.MkdirAll(binDir, constYarnDirMode)
	if err != nil {
		t.Fatalf("create stub bin dir: %v", err)
	}

	return home, binDir
}

func stubEnvVars(home, binDir string) []string {
	env := os.Environ()

	env = tasktestutil.SetEnv(env, "HOME", home)
	env = tasktestutil.SetEnv(env, constYarnPathEnvVar, binDir+":"+os.Getenv(constYarnPathEnvVar))
	env = tasktestutil.SetEnv(env, "TASK_ASSUME_YES", "true")
	env = tasktestutil.SetEnv(env, "NO_COLOR", "1")

	return env
}

func writeFnmNodeStub(t *testing.T, binDir string) {
	t.Helper()

	tasktestutil.WriteStub(
		t,
		binDir,
		"fnm",
		"#!/usr/bin/env bash\n"+
			"case \"$1\" in env) echo '# fnm env stub' ;; use) echo 'Using Node stub' ;; *) exit 0 ;; esac\n",
	)
	tasktestutil.WriteStub(
		t,
		binDir,
		"node",
		"#!/usr/bin/env bash\nif [ \"$1\" = '--version' ]; then echo 'v22.0.0 stub'; fi\n",
	)
}

func writeFnmStubs(t *testing.T, binDir string) {
	t.Helper()

	writeFnmNodeStub(t, binDir)
	tasktestutil.WriteStub(
		t,
		binDir,
		"corepack",
		"#!/usr/bin/env bash\necho \"corepack $* stub\"\n",
	)
}

func loadTaskfile(t *testing.T) yaml.Node {
	t.Helper()

	var doc yaml.Node

	err := yaml.Unmarshal(
		[]byte(tasktestutil.MustRead(t, filepath.Join(constYarnCurrentDir, "Taskfile.yml"))),
		&doc,
	)
	if err != nil {
		t.Fatalf("parse Taskfile YAML: %v", err)
	}

	return doc
}
