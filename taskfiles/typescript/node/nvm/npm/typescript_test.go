// Copyright 2026 task-otter
// SPDX-License-Identifier: Apache-2.0

package typescriptnodenvmnpm_test

import (
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

const (
	constTypescriptTestListAll = "--list-all"
)

type publicTaskSpec struct {
	name            string
	args            []string
	requiresPrompt  bool
	requiresSummary bool
}

func publicTasks() []publicTaskSpec {
	return []publicTaskSpec{
		{name: "build", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "build:clean", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "build:watch", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "ci", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "clean", requiresPrompt: true, requiresSummary: true, args: nil},
		{name: "clean:all", requiresPrompt: true, requiresSummary: true, args: nil},
		{name: "config:diagnostics", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "config:files", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "config:init", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "config:show", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "config:trace", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "dev", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "emit:dts", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "install", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "install:undo", requiresPrompt: true, requiresSummary: true, args: nil},
		{name: "run", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "start", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "tsserver:info", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "typecheck", requiresSummary: true, args: nil, requiresPrompt: false},
		{
			name:            "typecheck:files",
			args:            []string{"FILES=src/index.ts"},
			requiresSummary: true,
			requiresPrompt:  false,
		},
		{name: "typecheck:watch", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "upgrade", requiresSummary: true, args: nil, requiresPrompt: false},
		{name: "version", requiresSummary: true, args: nil, requiresPrompt: false},
	}
}

func TestTaskfileAndReadmePublicApi(t *testing.T) {
	t.Parallel()

	taskfile := loadTaskfile(t)

	expected := publicTaskNames()

	actual := taskNames(taskfile.tasks)

	if !slices.Equal(expected, actual) {
		t.Fatalf("public task drift\nexpected: %v\nactual:   %v", expected, actual)
	}

	readmeTasks := readmeTaskNames(readTool(t, "README.md"))

	if !slices.Equal(expected, readmeTasks) {
		t.Fatalf("README public task drift\nexpected: %v\nactual:   %v", expected, readmeTasks)
	}
}

func TestTaskfileYamlIsCleanAndValid(t *testing.T) {
	t.Parallel()

	content := read(t)

	if strings.Contains(content, "\r\n") {
		t.Fatal("Taskfile must use LF line endings")
	}

	if strings.TrimRight(content, " \t\r\n") != strings.TrimRight(content, "\r\n") {
		t.Fatal("Taskfile has trailing whitespace")
	}

	var doc yaml.Node

	err := yaml.Unmarshal([]byte(content), &doc)
	if err != nil {
		t.Fatalf("parse Taskfile: %v", err)
	}

	root := documentRoot(t, &doc)

	if version := scalarField(
		root,
		"version",
	); version != "3" &&
		!strings.HasPrefix(version, "3.") {

		t.Fatalf("Taskfile version must be 3 or 3.x, got %q", version)
	}

	if tasks := mappingField(root, "tasks"); tasks == nil || len(tasks.Content) == 0 {
		t.Fatal("Taskfile must contain a non-empty tasks map")
	}
}

func TestTaskCliCanLoadTaskfile(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		{"--list"},
		{constTypescriptTestListAll},
		{constTypescriptTestListAll, "--sort", "alphanumeric"},
		{constTypescriptTestListAll, "--json"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()

			result := runTask(t, isolatedEnv(t), args...)
			assertExitCode(t, result, 0)
			assertNotContains(t, strings.ToLower(result.output), "taskfile does not exist")
			assertNotContains(t, strings.ToLower(result.output), "unknown")
		})
	}
}

func TestTaskListAllJsonIsValid(t *testing.T) {
	t.Parallel()

	result := runTask(t, isolatedEnv(t), constTypescriptTestListAll, "--json")
	assertExitCode(t, result, 0)

	var payload any

	err := json.Unmarshal([]byte(result.output), &payload)
	if err != nil {
		t.Fatalf(
			"task --list-all --json did not produce valid JSON:\n%s\nerror: %v",
			result.output,
			err,
		)
	}
}

func TestPublicTasksHaveMetadataAndCommands(t *testing.T) {
	t.Parallel()

	taskfile := loadTaskfile(t)

	for _, spec := range publicTasks() {
		t.Run(spec.name, func(t *testing.T) {
			t.Parallel()

			task := mustTask(t, taskfile, spec.name)

			if task.node.Kind != yaml.MappingNode {
				t.Fatalf("public task %q must use mapping syntax", spec.name)
			}

			desc := nodeText(mappingValue(task.node, "desc"))

			if len(strings.TrimSpace(desc)) < 12 {
				t.Fatalf("public task %q desc is missing or too short: %q", spec.name, desc)
			}

			summary := nodeText(mappingValue(task.node, "summary"))

			if spec.requiresSummary && len(strings.TrimSpace(summary)) < 25 {
				t.Fatalf("public task %q summary is missing or too short:\n%s", spec.name, summary)
			}

			if isEmptyNode(mappingValue(task.node, "cmds")) &&
				isEmptyNode(mappingValue(task.node, "deps")) {

				t.Fatalf("public task %q must have cmds or deps", spec.name)
			}
		})
	}
}

func TestDestructivePublicTasksHavePrompt(t *testing.T) {
	t.Parallel()

	taskfile := loadTaskfile(t)

	for _, spec := range publicTasks() {
		t.Run(spec.name, func(t *testing.T) {
			t.Parallel()

			if !spec.requiresPrompt {
				return
			}

			task := mustTask(t, taskfile, spec.name)

			prompt := strings.ToLower(nodeText(mappingValue(task.node, "prompt")))

			if !strings.Contains(prompt, "delete") &&
				!strings.Contains(prompt, "remove") &&
				!strings.Contains(prompt, "continue") {

				t.Fatalf("destructive task %q needs an explicit prompt:\n%s", spec.name, prompt)
			}
		})
	}
}

func TestTaskSummariesWork(t *testing.T) {
	t.Parallel()

	for _, spec := range publicTasks() {
		t.Run(spec.name, func(t *testing.T) {
			t.Parallel()

			if !spec.requiresSummary {
				return
			}

			result := runTask(t, isolatedEnv(t), "--summary", spec.name)
			assertExitCode(t, result, 0)
			assertContains(t, result.output, spec.name)
			assertNotContains(t, strings.ToLower(result.output), "task not found")
			assertNotContains(t, strings.ToLower(result.output), "no summary")
		})
	}
}

func TestTypecheckFilesRequiresExplicitFiles(t *testing.T) {
	t.Parallel()

	result := runTask(t, isolatedEnv(t), "--dry", "--yes", "typecheck:files")

	if result.err == nil {
		t.Fatalf("typecheck:files without FILES unexpectedly succeeded:\n%s", result.output)
	}

	assertContains(t, strings.ToLower(result.output), "files")
}

func TestTsserverGuidanceStaysEditorManaged(t *testing.T) {
	t.Parallel()

	content := read(t)

	for _, token := range []string{
		"Never start tsserver manually",
		"not LSP directly",
		"TypeScript Server protocol",
		"Managed by your editor",
	} {
		assertContains(t, content, token)
	}
}

func TestCommandsDoNotContainDangerousPatterns(t *testing.T) {
	t.Parallel()

	taskfile := loadTaskfile(t)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?m)\brm\s+-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*\s+/(?:\s|$)`),
		regexp.MustCompile(`(?m)\bsudo\s+rm\s+-[a-zA-Z]*r[a-zA-Z]*f`),
		regexp.MustCompile(`(?m)\bchmod\s+-R\s+777\s+/`),
		regexp.MustCompile(`(?m)\bcurl\b.*\s-k(?:\s|$)`),
		regexp.MustCompile(`(?m)\bcurl\b.*--insecure`),
	}

	for taskName, task := range taskfile.tasks {
		for _, command := range collectScalars(task.node) {
			for _, pattern := range patterns {
				if pattern.MatchString(command) {
					t.Fatalf(
						"task %q contains dangerous command pattern %q:\n%s",
						taskName,
						pattern.String(),
						command,
					)
				}
			}
		}
	}
}

func TestNoPlaceholderTextInTaskfileOrReadme(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"Taskfile.yml": read(t),
		"README.md":    readTool(t, "README.md"),
	}

	for name, content := range files {
		content = strings.ToUpper(content)

		for _, placeholder := range []string{"TODO", "FIXME", "CHANGEME", "Copyright", "LOREM IPSUM"} {
			if strings.Contains(content, placeholder) {
				t.Fatalf("%s contains placeholder text: %s", name, placeholder)
			}
		}
	}
}

type loadedTaskfile struct {
	tasks map[string]task
}

type task struct {
	node *yaml.Node
}

type commandResult struct {
	err    error
	output string
}

func loadTaskfile(t *testing.T) loadedTaskfile {
	t.Helper()

	var doc yaml.Node

	err := yaml.Unmarshal([]byte(read(t)), &doc)
	if err != nil {
		t.Fatalf("parse Taskfile: %v", err)
	}

	tasksNode := mappingField(documentRoot(t, &doc), "tasks")

	if tasksNode == nil {
		t.Fatal("Taskfile has no tasks map")
	}

	tasks := map[string]task{}

	for i := 0; i < len(tasksNode.Content); i += 2 {
		tasks[tasksNode.Content[i].Value] = task{node: tasksNode.Content[i+1]}
	}

	return loadedTaskfile{tasks: tasks}
}

func mustTask(t *testing.T, taskfile loadedTaskfile, name string) task {
	t.Helper()

	task, ok := taskfile.tasks[name]

	if !ok {
		t.Fatalf("expected public task %q is missing", name)
	}

	return task
}

func runTask(t *testing.T, env []string, args ...string) commandResult {
	t.Helper()

	projectDir, projectEnv := fakeTypeScriptProject(t, env)
	fullArgs := append([]string{"--taskfile", filepath.Join(dir(t), "Taskfile.yml")}, args...)

	commandContext := exec.CommandContext
	cmd := commandContext(t.Context(), "task", fullArgs...)

	cmd.Dir = projectDir
	cmd.Env = projectEnv

	out, err := cmd.CombinedOutput()

	return commandResult{output: string(out), err: err}
}

func isolatedEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()
	env := os.Environ()

	env = setEnv(env, "HOME", home)
	env = setEnv(env, "ZDOTDIR", home)
	env = setEnv(env, "CI", "true")
	env = setEnv(env, "TASK_COLOR", "0")
	env = setEnv(env, "NO_COLOR", "1")
	env = setEnv(env, "TASK_ASSUME_YES", "true")

	return env
}

func fakeTypeScriptProject(t *testing.T, env []string) (string, []string) {
	t.Helper()

	projectDir := t.TempDir()
	binDir := filepath.Join(projectDir, ".stub-bin")
	nodeBinDir := filepath.Join(projectDir, "node_modules", ".bin")

	createTypeScriptFixtureDirs(t, projectDir, binDir, nodeBinDir, env)

	writeFile(
		t,
		filepath.Join(projectDir, "package.json"),
		`{"scripts":{"build":"tsc"}}`+"\n",
		0o644,
	)
	writeFile(t, filepath.Join(projectDir, "package-lock.json"), "{}\n", 0o600)
	writeFile(
		t,
		filepath.Join(projectDir, "tsconfig.json"),
		`{"compilerOptions":{"outDir":"dist"}}`+"\n",
		0o644,
	)
	writeFile(t, filepath.Join(projectDir, "src", "index.ts"), "export {}\n", 0o600)
	writeFile(t, filepath.Join(projectDir, "dist", "index.js"), "console.log('ok')\n", 0o600)

	stubBody := "#!/usr/bin/env bash\n" +
		"case \"$1\" in\n" +
		"  --version) echo \"$0 1.0.0\" ;;\n" +
		"  *) exit 0 ;;\n" +
		"esac\n"

	for _, name := range []string{"tsc", "tsx", "tsserver"} {
		writeFile(t, filepath.Join(nodeBinDir, name), stubBody, 0o500)
	}

	for _, name := range []string{"fnm", "node", "npx", "npm", "pnpm", "yarn", "bun"} {
		writeFile(t, filepath.Join(binDir, name), stubBody, 0o500)
	}

	writeFile(t, filepath.Join(envValue(env, "HOME"), ".bun", "bin", "bun"), stubBody, 0o500)
	writeFile(t, filepath.Join(envValue(env, "HOME"), ".nvm", "nvm.sh"), "# nvm stub\n", 0o600)
	writeFile(
		t,
		filepath.Join(envValue(env, "HOME"), ".local", "share", "fnm", "fnm"),
		stubBody,
		0o755,
	)

	path := binDir + ":" + nodeBinDir + ":" + envValue(env, "PATH")

	env = setEnv(env, "PATH", path)

	return projectDir, env
}

func createTypeScriptFixtureDirs(
	t *testing.T,
	projectDir string,
	binDir string,
	nodeBinDir string,
	env []string,
) {

	t.Helper()

	for _, path := range []string{
		filepath.Join(projectDir, "src"),
		filepath.Join(projectDir, "dist"),
		binDir,
		nodeBinDir,
		filepath.Join(envValue(env, "HOME"), ".bun", "bin"),
		filepath.Join(envValue(env, "HOME"), ".nvm"),
		filepath.Join(envValue(env, "HOME"), ".local", "share", "fnm"),
	} {
		err := os.MkdirAll(path, 0o700)
		if err != nil {
			t.Fatalf("create test project dir %s: %v", path, err)
		}
	}
}

func writeFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()

	err := os.WriteFile(path, []byte(content), mode)
	if err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func envValue(env []string, key string) string {
	prefix := key + "="

	for _, item := range env {
		if after, ok := strings.CutPrefix(item, prefix); ok {
			return after
		}
	}

	return os.Getenv(key)
}

func publicTaskNames() []string {
	names := make([]string, 0, len(publicTasks()))

	for _, spec := range publicTasks() {
		names = append(names, spec.name)
	}

	slices.Sort(names)

	return names
}

func taskNames(tasks map[string]task) []string {
	names := []string{}

	for name, task := range tasks {
		if name == "default" || strings.HasPrefix(name, "_") {
			continue
		}

		if nodeText(mappingValue(task.node, "desc")) != "" {
			names = append(names, name)
		}
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

func documentRoot(t *testing.T, doc *yaml.Node) *yaml.Node {
	t.Helper()

	if doc.Kind != yaml.DocumentNode || len(doc.Content) != 1 {
		t.Fatalf(
			"expected YAML document node, got kind=%directory children=%directory",
			doc.Kind,
			len(doc.Content),
		)
	}

	return doc.Content[0]
}

func mappingField(root *yaml.Node, name string) *yaml.Node {
	if root == nil || root.Kind != yaml.MappingNode {
		return nil
	}

	return mappingValue(root, name)
}

func scalarField(root *yaml.Node, name string) string {
	return nodeText(mappingField(root, name))
}

func mappingValue(node *yaml.Node, name string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == name {
			return node.Content[i+1]
		}
	}

	return nil
}

func nodeText(node *yaml.Node) string {
	if node == nil {
		return ""
	}

	if node.Kind == yaml.ScalarNode {
		return node.Value
	}

	parts := []string{}

	for _, child := range node.Content {
		if text := nodeText(child); text != "" {
			parts = append(parts, text)
		}
	}

	return strings.Join(parts, "\n")
}

func isEmptyNode(node *yaml.Node) bool {
	return node == nil || strings.TrimSpace(nodeText(node)) == ""
}

func collectScalars(node *yaml.Node) []string {
	if node == nil {
		return nil
	}

	if node.Kind == yaml.ScalarNode {
		return []string{node.Value}
	}

	values := []string{}

	for _, child := range node.Content {
		values = append(values, collectScalars(child)...)
	}

	return values
}

func toolRoot(t *testing.T) string {
	t.Helper()

	directory := dir(t)

	for filepath.Base(filepath.Dir(directory)) != "taskfiles" {
		parent := filepath.Dir(directory)

		if parent == directory {
			t.Fatal("locate tool root")
		}

		directory = parent
	}

	return directory
}

func readTool(t *testing.T, name string) string {
	t.Helper()

	content, err := fs.ReadFile(os.DirFS(toolRoot(t)), filepath.ToSlash(name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}

	return string(content)
}

func read(t *testing.T) string {
	t.Helper()

	path := filepath.Join(dir(t), "Taskfile.yml")

	content, err := fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
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

func assertExitCode(t *testing.T, result commandResult, want int) {
	t.Helper()

	if want == 0 {
		if result.err != nil {
			t.Fatalf("command failed: %v\n%s", result.err, result.output)
		}

		return
	}

	if result.err == nil {
		t.Fatalf("command unexpectedly succeeded:\n%s", result.output)
	}
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()

	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected output to contain %q:\n%s", needle, haystack)
	}
}

func assertNotContains(t *testing.T, haystack, needle string) {
	t.Helper()

	if strings.Contains(haystack, needle) {
		t.Fatalf("expected output not to contain %q:\n%s", needle, haystack)
	}
}
