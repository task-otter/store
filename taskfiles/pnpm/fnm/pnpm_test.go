package pnpmfnm_test

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/task-otter/store/internal/tasktestutil"
	"gopkg.in/yaml.v3"
)

const (
	constPnpmTestYes = "--yes"
)

func publicTasks() []string {
	return []string{
		"add",
		"audit",
		"audit:fix",
		"audit:json",
		"audit:report",
		"build",
		"ci",
		"clean",
		"clean:all",
		"dev",
		"exec",
		"format",
		"install",
		"install:undo",
		"lint",
		"manager:pin",
		"manager:setup",
		"node:setup",
		"outdated",
		"outdated:strict",
		"remove",
		"run",
		"store:prune",
		"test",
		"typecheck",
		"update",
		"upgrade",
		"version",
	}
}

func TestTaskfileAndReadmePublicApi(t *testing.T) {
	t.Parallel()

	doc := loadTaskfile(t)

	var root map[string]any

	err := doc.Decode(&root)
	if err != nil {
		t.Fatalf("decode Taskfile: %v", err)
	}

	tasks, ok := root["tasks"].(map[string]any)
	if !ok || len(tasks) == 0 {
		t.Fatal("Taskfile tasks map is missing")
	}

	actual := tasktestutil.SimplePublicTaskNames(tasks)
	if !slices.Equal(publicTasks(), actual) {
		t.Fatalf("public task drift\nexpected: %v\nactual:   %v", publicTasks(), actual)
	}

	readmeTasks := tasktestutil.ReadmePublicTaskNames(
		tasktestutil.MustRead(t, tasktestutil.ModuleReadmePath(t)),
	)
	if !slices.Equal(publicTasks(), readmeTasks) {
		t.Fatalf("README public task drift\nexpected: %v\nactual:   %v", publicTasks(), readmeTasks)
	}
}

func TestStubbedPnpmFlows(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("Unix shell stubs cover these flows")
	}

	env := stubEnv(t)
	for _, args := range [][]string{
		{constPnpmTestYes, "version"},
		{constPnpmTestYes, "install"},
		{constPnpmTestYes, "ci"},
		{constPnpmTestYes, "run", "SCRIPT=test", "--", "--watch"},
	} {
		result := tasktestutil.RunSimpleTask(t, ".", env, args...)
		if result.Err != nil {
			t.Fatalf("task %v failed:\n%s", args, result.Output)
		}
	}

	result := tasktestutil.RunSimpleTask(t, ".", env, constPnpmTestYes, "run", "SCRIPT=dev; exit 1")
	if result.Err == nil {
		t.Fatalf("unsafe SCRIPT unexpectedly succeeded:\n%s", result.Output)
	}
}

func stubEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()

	binDir := filepath.Join(home, ".local", "bin")

	err := os.MkdirAll(binDir, 0o755)
	if err != nil {
		t.Fatalf("create stub bin dir: %v", err)
	}

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
	tasktestutil.WriteStub(
		t,
		binDir,
		"corepack",
		"#!/usr/bin/env bash\necho \"corepack $* stub\"\n",
	)

	env := os.Environ()
	env = tasktestutil.SetEnv(env, "HOME", home)
	env = tasktestutil.SetEnv(env, "PATH", binDir+":"+os.Getenv("PATH"))
	env = tasktestutil.SetEnv(env, "TASK_ASSUME_YES", "true")
	env = tasktestutil.SetEnv(env, "NO_COLOR", "1")

	return env
}

func loadTaskfile(t *testing.T) yaml.Node {
	t.Helper()

	var doc yaml.Node

	err := yaml.Unmarshal(
		[]byte(tasktestutil.MustRead(t, filepath.Join(".", "Taskfile.yml"))),
		&doc,
	)
	if err != nil {
		t.Fatalf("parse Taskfile YAML: %v", err)
	}

	return doc
}
