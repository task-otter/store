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
	constPnpmTestYes          = "--yes"
	constPnpmTaskInstallClean = "install:clean"
	constPnpmTaskInstall      = "install"
	constPnpmTaskRun          = "run"
	constPnpmTaskVersion      = "version"
	constPnpmCurrentDir       = "."
	constPnpmPathEnvVar       = "PATH"
	constPnpmDirMode          = 0o700
	constPnpmMinTasks         = 0
	constPnpmScriptTest       = "SCRIPT=test"
	constPnpmDoubleDash       = "--"
	constPnpmWatchFlag        = "--watch"
)

// TestTaskfileAndReadmePublicApi
func TestTaskfileAndReadmePublicApi(t *testing.T) {
	t.Parallel()

	tasks := decodeTaskfileTasks(t)

	assertPublicTaskNamesMatch(t, tasks)
	assertReadmeTaskNamesMatch(t)
}

// TestStubbedPnpmFlows
func TestStubbedPnpmFlows(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("Unix shell stubs cover these flows")
	}

	env := stubEnv(t)

	runSafePnpmFlows(t, env)
	assertUnsafeScriptRejected(t, env)
}

func publicTasks() []string {
	return append(publicTasksCore(), publicTasksExtra()...)
}

func publicTasksCore() []string {
	return []string{
		"add",
		"audit",
		"audit:fix",
		"audit:json",
		"audit:report",
		"build",
		"ci:fix",
		"clean",
		"clean:all",
		"dev",
		"exec",
		constPnpmTaskInstall,
		constPnpmTaskInstallClean,
		"install:undo",
	}
}

func publicTasksExtra() []string {
	return []string{
		"lint",
		"manager:pin",
		"manager:setup",
		"node:setup",
		"outdated",
		"outdated:strict",
		"remove",
		constPnpmTaskRun,
		"store:prune",
		"test",
		"typecheck",
		"update",
		"upgrade",
		constPnpmTaskVersion,
	}
}

func decodeTaskfileTasks(t *testing.T) map[string]any {
	t.Helper()

	doc := loadTaskfile(t)

	var root map[string]any

	err := doc.Decode(&root)
	if err != nil {
		t.Fatalf("decode Taskfile: %v", err)
	}

	tasks, ok := root["tasks"].(map[string]any)

	if !ok || len(tasks) == constPnpmMinTasks {
		t.Fatal("Taskfile tasks map is missing")
	}

	return tasks
}

func assertPublicTaskNamesMatch(t *testing.T, tasks map[string]any) {
	t.Helper()

	actual := tasktestutil.SimplePublicTaskNames(tasks)

	if !slices.Equal(publicTasks(), actual) {
		t.Fatalf("public task drift\nexpected: %v\nactual:   %v", publicTasks(), actual)
	}
}

func assertReadmeTaskNamesMatch(t *testing.T) {
	t.Helper()

	readmeTasks := tasktestutil.ReadmePublicTaskNames(
		tasktestutil.MustRead(t, tasktestutil.ModuleReadmePath(t)),
	)

	if !slices.Equal(publicTasks(), readmeTasks) {
		t.Fatalf("README public task drift\nexpected: %v\nactual:   %v", publicTasks(), readmeTasks)
	}
}

func safePnpmFlowArgs() [][]string {
	return [][]string{
		{constPnpmTestYes, constPnpmTaskVersion},
		{constPnpmTestYes, constPnpmTaskInstall},
		{constPnpmTestYes, constPnpmTaskInstallClean},
		{
			constPnpmTestYes,
			constPnpmTaskRun,
			constPnpmScriptTest,
			constPnpmDoubleDash,
			constPnpmWatchFlag,
		},
	}
}

func runSafePnpmFlows(t *testing.T, env []string) {
	t.Helper()

	variants := safePnpmFlowArgs()

	for i := range variants {
		args := variants[i]
		result := tasktestutil.RunSimpleTask(
			t,
			tasktestutil.TaskRun{Root: constPnpmCurrentDir, Env: env, Args: args},
		)

		if result.Err != nil {
			t.Fatalf("task %v failed:\n%s", args, result.Stdout)
		}
	}
}

func assertUnsafeScriptRejected(t *testing.T, env []string) {
	t.Helper()

	result := tasktestutil.RunSimpleTask(t, tasktestutil.TaskRun{
		Root: constPnpmCurrentDir,
		Env:  env,
		Args: []string{constPnpmTestYes, constPnpmTaskRun, "SCRIPT=dev; exit 1"},
	})

	if result.Err == nil {
		t.Fatalf("unsafe SCRIPT unexpectedly succeeded:\n%s", result.Stdout)
	}
}

func stubEnv(t *testing.T) []string {
	t.Helper()

	home, binDir := stubEnvDirs(t)

	writePnpmFnmStubs(t, binDir)

	return stubEnvVars(home, binDir)
}

func stubEnvDirs(t *testing.T) (home, binDir string) {
	t.Helper()

	home = t.TempDir()
	binDir = filepath.Join(home, ".local", "bin")

	err := os.MkdirAll(binDir, constPnpmDirMode)
	if err != nil {
		t.Fatalf("create stub bin dir: %v", err)
	}

	return home, binDir
}

func stubEnvVars(home, binDir string) []string {
	env := os.Environ()

	env = tasktestutil.SetEnv(env, "HOME", home)
	env = tasktestutil.SetEnv(env, constPnpmPathEnvVar, binDir+":"+os.Getenv(constPnpmPathEnvVar))
	env = tasktestutil.SetEnv(env, "TASK_ASSUME_YES", "true")
	env = tasktestutil.SetEnv(env, "NO_COLOR", "1")

	return env
}

func writePnpmFnmToolStub(t *testing.T, binDir string) {
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

func writePnpmFnmStubs(t *testing.T, binDir string) {
	t.Helper()

	writePnpmFnmToolStub(t, binDir)
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
		[]byte(tasktestutil.MustRead(t, filepath.Join(constPnpmCurrentDir, "Taskfile.yml"))),
		&doc,
	)
	if err != nil {
		t.Fatalf("parse Taskfile YAML: %v", err)
	}

	return doc
}
