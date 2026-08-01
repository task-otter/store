// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package corepacknvm_test

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"syscall"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

type (
	result struct {
		err    error
		output string
	}
)

const (
	constCorepackTestYes = "--yes"
)

func publicTasks() []string {
	return []string{
		"cache:clean",
		"disable",
		"enable",
		"install",
		"install:undo",
		"node:setup",
		"setup",
		"upgrade",
		"use",
		"version",
	}
}

func TestTaskfileAndReadmePublicApi(t *testing.T) {
	t.Parallel()

	var (
		root map[string]any
		doc  yaml.Node
	)

	err := yaml.Unmarshal([]byte(read(t, "Taskfile.yml")), &doc)
	if err != nil {
		t.Fatalf("parse Taskfile: %v", err)
	}

	err = doc.Decode(&root)
	if err != nil {
		t.Fatalf("decode Taskfile: %v", err)
	}

	tasks, ok := root["tasks"].(map[string]any)

	if !ok || len(tasks) == 0 {
		t.Fatal("Taskfile tasks map is missing")
	}

	actual := taskNames(tasks)

	if !slices.Equal(publicTasks(), actual) {
		t.Fatalf("public task drift\nexpected: %v\nactual:   %v", publicTasks(), actual)
	}

	readmeTasks := readmeTaskNames(read(t, filepath.Join("..", "README.md")))

	if !slices.Equal(publicTasks(), readmeTasks) {
		t.Fatalf("README public task drift\nexpected: %v\nactual:   %v", publicTasks(), readmeTasks)
	}
}

func TestTaskCliAndCorepackFlows(t *testing.T) {
	t.Parallel()

	env := stubEnv(t)

	for _, args := range [][]string{
		{"--list"},
		{"--list-all", "--json"},
		{"--dry", constCorepackTestYes, "setup"},
		{constCorepackTestYes, "version"},
		{constCorepackTestYes, "enable"},
		{constCorepackTestYes, "use", "PACKAGE_MANAGER=pnpm", "VERSION=latest"},
	} {
		result := runTask(t, env, args...)

		if result.err != nil {
			t.Fatalf("task %v failed:\n%s", args, result.output)
		}
	}

	result := runTask(t, env, constCorepackTestYes, "use", "PACKAGE_MANAGER=bad", "VERSION=latest")

	if result.err == nil {
		t.Fatalf("invalid package manager unexpectedly succeeded:\n%s", result.output)
	}
}

func TestCorepackVersionDefaultIsPinned(t *testing.T) {
	t.Parallel()

	content := read(t, "Taskfile.yml")

	if !strings.Contains(content, "COREPACK_VERSION: 0.34.0") {
		t.Fatalf("COREPACK_VERSION default should stay pinned for reproducibility:\n%s", content)
	}

	if !strings.Contains(content, "override with COREPACK_VERSION=latest") {
		t.Fatal("COREPACK_VERSION pin should include an override comment")
	}
}

func runTask(t *testing.T, env []string, args ...string) result {
	t.Helper()

	commandContext := exec.CommandContext
	cmd := commandContext(t.Context(), "task", args...)

	cmd.Dir = dir(t)
	cmd.Env = env

	out, err := cmd.CombinedOutput()

	return result{output: string(out), err: err}
}

func stubEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()

	nvmDir := filepath.Join(home, ".nvm")

	err := os.MkdirAll(nvmDir, 0o700)
	if err != nil {
		t.Fatalf("create nvm dir: %v", err)
	}

	err = os.WriteFile(filepath.Join(nvmDir, "nvm.sh"), []byte("# nvm stub\n"), 0o600)
	if err != nil {
		t.Fatalf("write nvm.sh stub: %v", err)
	}

	bin := filepath.Join(home, ".local", "bin")

	err = os.MkdirAll(bin, 0o700)
	if err != nil {
		t.Fatalf("create stub bin: %v", err)
	}

	stub(t, bin, "nvm", "#!/usr/bin/env bash\ncase \"$1\" in use) exit 0 ;; *) exit 0 ;; esac\n")
	stub(
		t,
		bin,
		"corepack",
		"#!/usr/bin/env bash\ncase \"$1\" in --version) echo \"0.34.0\" ;; *) echo \"corepack $* stub\" ;; esac\n",
	)
	stub(t, bin, "npm", "#!/usr/bin/env bash\necho \"npm $* stub\"\n")

	env := os.Environ()

	env = setEnv(env, "HOME", home)
	env = setEnv(env, "PATH", bin+":"+os.Getenv("PATH"))
	env = setEnv(env, "NVM_DIR", nvmDir)
	env = setEnv(env, "NO_COLOR", "1")
	env = setEnv(env, "TASK_ASSUME_YES", "true")

	return env
}

type stubFile struct {
	path string
	name string
	body string
}

func stub(t *testing.T, fileValue any, parts ...string) {
	file := normalizeStubFile(t, fileValue, parts)
	t.Helper()

	t.Helper()

	stubPath := filepath.Join(file.path, file.name)

	err := os.WriteFile(stubPath, []byte(file.body), 0o600)
	if err != nil {
		t.Fatalf("write %s stub: %v", file.name, err)
	}

	err = syscall.Chmod(stubPath, 0o500)
	if err != nil {
		t.Fatalf("make %s stub executable: %v", file.name, err)
	}
}

func normalizeStubFile(t *testing.T, fileValue any, parts []string) stubFile {
	t.Helper()

	if file, ok := fileValue.(stubFile); ok {
		if len(parts) != 0 {
			t.Fatalf("stubFile does not accept positional arguments: %v", parts)
		}

		return file
	}

	path, ok := fileValue.(string)

	if !ok {
		t.Fatalf("stub path must be string or stubFile, got %T", fileValue)
	}

	if len(parts) != 2 {
		t.Fatalf("stub requires name and body, got %d values", len(parts))
	}

	return stubFile{path: path, name: parts[0], body: parts[1]}
}

func taskNames(tasks map[string]any) []string {
	names := []string{}

	for name, raw := range tasks {
		if name == "default" || strings.HasPrefix(name, "_") {
			continue
		}

		if task, ok := raw.(map[string]any); ok && task["internal"] == true {
			continue
		}

		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

func readmeTaskNames(content string) []string {
	row := regexp.MustCompile("^\\|\\s*`([^`]+)`\\s*\\|")
	names := []string{}
	active := false

	for line := range strings.SplitSeq(content, "\n") {
		trimmed := strings.TrimSpace(line)

		if trimmed == "## Public Tasks" {
			active = true

			continue
		}

		if active && strings.HasPrefix(trimmed, "## ") {
			break
		}

		if active {
			if match := row.FindStringSubmatch(trimmed); len(match) == 2 {
				names = append(names, match[1])
			}
		}
	}

	slices.Sort(names)

	return names
}

func read(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join(dir(t), name)

	content, err := fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}

	return string(content)
}

func dir(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)

	if !ok {
		t.Fatal("locate test file")
	}

	return filepath.Dir(file)
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="

	for i, item := range env {
		if strings.HasPrefix(item, prefix) {
			env[i] = prefix + value

			return env
		}
	}

	return append(env, prefix+value)
}
