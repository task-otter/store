// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package nvm_test

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
	taskResult struct {
		err    error
		output string
	}

	taskNameComparison struct {
		label    string
		expected []string
		actual   []string
	}

	stubFile struct {
		path string
		name string
		body string
	}

	readmeScanner struct {
		names  []string
		active bool
	}
)

const (
	constCorepackTestYes = "--yes"
	flagList             = "--list"
	flagListAll          = "--list-all"
	flagJSON             = "--json"
	flagDry              = "--dry"
	packageManagerPnpm   = "PACKAGE_MANAGER=pnpm"
	enableTask           = "enable"
	setupTask            = "setup"
	useTask              = "use"
	versionTask          = "version"
	taskfileYML          = "Taskfile.yml"
	versionLatestArg     = "VERSION=latest"
	pathEnvVar           = "PATH"
	twoLiteral           = 2
	emptyLen             = 0
	secondArgIndex       = 1
	dirPerm              = fs.FileMode(0o700)
	filePerm             = fs.FileMode(0o600)
	execPerm             = 0o500
)

var readmeTaskRow = regexp.MustCompile("^\\|\\s*`([^`]+)`\\s*\\|")

func publicTasks() []string {
	return []string{
		"cache:clean",
		"disable",
		enableTask,
		"install",
		"install:undo",
		"node:setup",
		setupTask,
		"upgrade",
		useTask,
		versionTask,
	}
}

// TestTaskfileAndReadmePublicApi
func TestTaskfileAndReadmePublicApi(t *testing.T) {
	t.Parallel()

	tasks := parseTaskfileTasks(t, read(t, taskfileYML))
	actual := taskNames(tasks)

	assertTaskNamesEqual(
		t,
		&taskNameComparison{label: "public task", expected: publicTasks(), actual: actual},
	)

	readmeTasks := readmeTaskNames(read(t, filepath.Join("..", "README.md")))

	assertTaskNamesEqual(
		t,
		&taskNameComparison{
			label:    "README public task",
			expected: publicTasks(),
			actual:   readmeTasks,
		},
	)
}

func parseTaskfileTasks(t *testing.T, content string) map[string]any {
	t.Helper()

	root := decodeTaskfileRoot(t, content)

	tasks, ok := root["tasks"].(map[string]any)

	if !ok || len(tasks) == emptyLen {
		t.Fatal("Taskfile tasks map is missing")
	}

	return tasks
}

func decodeTaskfileRoot(t *testing.T, content string) map[string]any {
	t.Helper()

	var (
		root map[string]any
		doc  yaml.Node
	)

	err := yaml.Unmarshal([]byte(content), &doc)
	if err != nil {
		t.Fatalf("parse Taskfile: %v", err)
	}

	err = doc.Decode(&root)
	if err != nil {
		t.Fatalf("decode Taskfile: %v", err)
	}

	return root
}

func assertTaskNamesEqual(t *testing.T, cmp *taskNameComparison) {
	t.Helper()

	if !slices.Equal(cmp.expected, cmp.actual) {
		t.Fatalf("%s drift\nexpected: %v\nactual:   %v", cmp.label, cmp.expected, cmp.actual)
	}
}

func corepackFlowArgVariants() [][]string {
	return [][]string{
		{flagList},
		{flagListAll, flagJSON},
		{flagDry, constCorepackTestYes, setupTask},
		{constCorepackTestYes, versionTask},
		{constCorepackTestYes, enableTask},
		{constCorepackTestYes, useTask, packageManagerPnpm, versionLatestArg},
	}
}

func assertCorepackFlowsSucceed(t *testing.T, env []string) {
	t.Helper()

	variants := corepackFlowArgVariants()

	for i := range variants {
		args := variants[i]
		result := runTask(t, env, args...)

		if result.err != nil {
			t.Fatalf("task %v failed:\n%s", args, result.output)
		}
	}
}

func assertInvalidPackageManagerFails(t *testing.T, env []string) {
	t.Helper()

	result := runTask(
		t,
		env,
		constCorepackTestYes,
		useTask,
		"PACKAGE_MANAGER=bad",
		versionLatestArg,
	)

	if result.err == nil {
		t.Fatalf("invalid package manager unexpectedly succeeded:\n%s", result.output)
	}
}

// TestTaskCliAndCorepackFlows
func TestTaskCliAndCorepackFlows(t *testing.T) {
	t.Parallel()

	env := stubEnv(t)

	assertCorepackFlowsSucceed(t, env)
	assertInvalidPackageManagerFails(t, env)
}

// TestCorepackVersionDefaultIsPinned
func TestCorepackVersionDefaultIsPinned(t *testing.T) {
	t.Parallel()

	content := read(t, taskfileYML)

	if !strings.Contains(content, "COREPACK_VERSION: 0.34.0") {
		t.Fatalf("COREPACK_VERSION default should stay pinned for reproducibility:\n%s", content)
	}

	if !strings.Contains(content, "override with COREPACK_VERSION=latest") {
		t.Fatal("COREPACK_VERSION pin should include an override comment")
	}
}

func runTask(t *testing.T, env []string, args ...string) *taskResult {
	t.Helper()

	commandContext := exec.CommandContext
	cmd := commandContext(t.Context(), "task", args...)

	cmd.Dir = dir(t)
	cmd.Env = env

	out, err := cmd.CombinedOutput()

	return &taskResult{output: string(out), err: err}
}

func stubEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()
	nvmDir := setupNvmDir(t, home)
	bin := setupStubBin(t, home)

	return buildStubEnv(home, bin, nvmDir)
}

func setupNvmDir(t *testing.T, home string) string {
	t.Helper()

	nvmDir := filepath.Join(home, ".nvm")

	err := os.MkdirAll(nvmDir, dirPerm)
	if err != nil {
		t.Fatalf("create nvm dir: %v", err)
	}

	err = os.WriteFile(filepath.Join(nvmDir, "nvm.sh"), []byte("# nvm stub\n"), filePerm)
	if err != nil {
		t.Fatalf("write nvm.sh stub: %v", err)
	}

	return nvmDir
}

func setupStubBin(t *testing.T, home string) string {
	t.Helper()

	bin := filepath.Join(home, ".local", "bin")

	err := os.MkdirAll(bin, dirPerm)
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

	return bin
}

func buildStubEnv(home, bin, nvmDir string) []string {
	env := os.Environ()

	env = setEnv(env, "HOME", home)
	env = setEnv(env, pathEnvVar, bin+":"+os.Getenv(pathEnvVar))
	env = setEnv(env, "NVM_DIR", nvmDir)
	env = setEnv(env, "NO_COLOR", "1")
	env = setEnv(env, "TASK_ASSUME_YES", "true")

	return env
}

func stub(t *testing.T, fileValue any, parts ...string) {
	t.Helper()

	file := normalizeStubFile(t, fileValue, parts)

	stubPath := filepath.Join(file.path, file.name)

	err := os.WriteFile(stubPath, []byte(file.body), filePerm)
	if err != nil {
		t.Fatalf("write %s stub: %v", file.name, err)
	}

	err = syscall.Chmod(stubPath, execPerm)
	if err != nil {
		t.Fatalf("make %s stub executable: %v", file.name, err)
	}
}

func normalizeStubFile(t *testing.T, fileValue any, parts []string) stubFile {
	t.Helper()

	if file, ok := fileValue.(stubFile); ok {
		requireNoExtraParts(t, parts)

		return file
	}

	return newStubFileFromParts(t, fileValue, parts)
}

func requireNoExtraParts(t *testing.T, parts []string) {
	t.Helper()

	if len(parts) != emptyLen {
		t.Fatalf("stubFile does not accept positional arguments: %v", parts)
	}
}

func newStubFileFromParts(t *testing.T, fileValue any, parts []string) stubFile {
	t.Helper()

	path, ok := fileValue.(string)

	if !ok {
		t.Fatalf("stub path must be string or stubFile, got %T", fileValue)
	}

	if len(parts) != twoLiteral {
		t.Fatalf("stub requires name and body, got %d values", len(parts))
	}

	return stubFile{path: path, name: parts[emptyLen], body: parts[secondArgIndex]}
}

func taskNames(tasks map[string]any) []string {
	names := make([]string, emptyLen, len(tasks))

	for name := range tasks {
		raw := tasks[name]

		if isSkippableTask(name, raw) {
			continue
		}

		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

func isSkippableTask(name string, raw any) bool {
	if name == "default" || strings.HasPrefix(name, "_") {
		return true
	}

	task, ok := raw.(map[string]any)

	if !ok {
		return false
	}

	internal, ok := task["internal"].(bool)

	return ok && internal
}

func (scanner *readmeScanner) processLine(trimmed string) (stop bool) {
	if trimmed == "## Public Tasks" {
		scanner.active = true

		return false
	}

	if !scanner.active {
		return false
	}

	if strings.HasPrefix(trimmed, "## ") {
		return true
	}

	scanner.recordRow(trimmed)

	return false
}

func (scanner *readmeScanner) recordRow(trimmed string) {
	if name, ok := readmeTaskRowName(trimmed); ok {
		scanner.names = append(scanner.names, name)
	}
}

func readmeTaskNames(content string) []string {
	scanner := &readmeScanner{names: nil, active: false}

	for line := range strings.SplitSeq(content, "\n") {
		if scanner.processLine(strings.TrimSpace(line)) {
			break
		}
	}

	slices.Sort(scanner.names)

	return scanner.names
}

func readmeTaskRowName(trimmed string) (string, bool) {
	match := readmeTaskRow.FindStringSubmatch(trimmed)

	if len(match) != twoLiteral {
		return "", false
	}

	return match[secondArgIndex], true
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

	programCounter, file, line, ok := runtime.Caller(emptyLen)

	if !ok || programCounter == emptyLen || line == emptyLen {
		t.Fatal("locate test file")
	}

	return filepath.Dir(file)
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="

	for i := range env {
		item := env[i]

		if strings.HasPrefix(item, prefix) {
			env[i] = prefix + value

			return env
		}
	}

	return append(env, prefix+value)
}
