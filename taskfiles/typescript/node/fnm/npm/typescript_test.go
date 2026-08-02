// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package npm_test

import (
	"encoding/json"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

type (
	publicTaskSpec struct {
		name            string
		requiresPrompt  bool
		requiresSummary bool
	}

	dangerousCommand struct {
		taskName string
		command  string
	}

	loadedTaskfile struct {
		tasks map[string]*taskfile
	}

	taskfile struct {
		node *yaml.Node
	}

	commandResult struct {
		err    error
		output string
	}

	typeScriptFixtureDirs struct {
		projectDir string
		binDir     string
		nodeBinDir string
		env        []string
	}

	testFile struct {
		path    string
		content string
		mode    os.FileMode
	}

	readmeSectionScanner struct {
		header string
		active bool
		done   bool
	}

	readmeTaskScanner struct {
		row     *regexp.Regexp
		section *readmeSectionScanner
	}
)

const (
	constTypescriptTestCRLF                 = "\r\n"
	constTypescriptTestLF                   = "\n"
	constTypescriptTestListAll              = "--list-all"
	constTypescriptTestTypecheckFiles       = "typecheck:files"
	constTypescriptTestVersion              = "version"
	constTypescriptTestReadmeMD             = "README.md"
	constTypescriptTestParseTaskfileErr     = "parse Taskfile: %v"
	constTypescriptTestTasks                = "tasks"
	constTypescriptTestJSONFlag             = "--json"
	constTypescriptTestDesc                 = "desc"
	constTypescriptTestTaskfileYML          = "Taskfile.yml"
	constTypescriptTestPairStride           = 2
	constTypescriptTestHome                 = "HOME"
	constTypescriptTestTrue                 = "true"
	constTypescriptTestMode0644             = 0o644
	constTypescriptTestMode0600             = 0o600
	constTypescriptTestSrc                  = "src"
	constTypescriptTestDist                 = "dist"
	constTypescriptTestMode0500             = 0o500
	constTypescriptTestFnm                  = "fnm"
	constTypescriptTestBun                  = "bun"
	constTypescriptTestDotBun               = ".bun"
	constTypescriptTestBin                  = "bin"
	constTypescriptTestDotNvm               = ".nvm"
	constTypescriptTestDotLocal             = ".local"
	constTypescriptTestShare                = "share"
	constTypescriptTestPath                 = "PATH"
	constTypescriptTestReadErr              = "read %s: %v"
	constTypescriptTestZero                 = 0
	constTypescriptTestOne                  = 1
	constTypescriptTestEmptyString          = ""
	constTypescriptTestMode0755             = 0o755
	constTypescriptTestMode0700             = 0o700
	constTypescriptTestMinDescLen           = 12
	constTypescriptTestMinSummaryLen        = 25
	constTypescriptTestFlagList             = "--list"
	constTypescriptTestFlagSort             = "--sort"
	constTypescriptTestSortAlpha            = "alphanumeric"
	constTypescriptTestActionDelete         = "delete"
	constTypescriptTestActionRemove         = "remove"
	constTypescriptTestActionContinue       = "continue"
	constTypescriptTestPlaceholderTODO      = "TODO"
	constTypescriptTestPlaceholderFIXME     = "FIXME"
	constTypescriptTestPlaceholderCHANGEME  = "CHANGEME"
	constTypescriptTestPlaceholderCopyright = "Copyright"
	constTypescriptTestPlaceholderLorem     = "LOREM IPSUM"
	constTypescriptTestBinTsc               = "tsc"
	constTypescriptTestBinTsx               = "tsx"
	constTypescriptTestBinTsserver          = "tsserver"
	constTypescriptTestDecimalBase          = 10
	constTypescriptTestUint32BitSize        = 32
)

func publicTasksBuildAndConfig() []publicTaskSpec {
	return []publicTaskSpec{
		{name: "build", requiresSummary: true, requiresPrompt: false},
		{name: "build:clean", requiresSummary: true, requiresPrompt: false},
		{name: "build:watch", requiresSummary: true, requiresPrompt: false},
		{name: "ci", requiresSummary: true, requiresPrompt: false},
		{name: "clean", requiresPrompt: true, requiresSummary: true},
		{name: "clean:all", requiresPrompt: true, requiresSummary: true},
		{name: "config:diagnostics", requiresSummary: true, requiresPrompt: false},
		{name: "config:files", requiresSummary: true, requiresPrompt: false},
		{name: "config:init", requiresSummary: true, requiresPrompt: false},
		{name: "config:show", requiresSummary: true, requiresPrompt: false},
		{name: "config:trace", requiresSummary: true, requiresPrompt: false},
	}
}

func publicTasksRunAndTypecheck() []publicTaskSpec {
	return []publicTaskSpec{
		{name: "dev", requiresSummary: true, requiresPrompt: false},
		{name: "emit:dts", requiresSummary: true, requiresPrompt: false},
		{name: "install", requiresSummary: true, requiresPrompt: false},
		{name: "install:undo", requiresPrompt: true, requiresSummary: true},
		{name: "run", requiresSummary: true, requiresPrompt: false},
		{name: "start", requiresSummary: true, requiresPrompt: false},
		{name: "tsserver:info", requiresSummary: true, requiresPrompt: false},
		{name: "typecheck", requiresSummary: true, requiresPrompt: false},
		{
			name:            constTypescriptTestTypecheckFiles,
			requiresSummary: true,
			requiresPrompt:  false,
		},
		{name: "typecheck:watch", requiresSummary: true, requiresPrompt: false},
		{name: "upgrade", requiresSummary: true, requiresPrompt: false},
		{name: constTypescriptTestVersion, requiresSummary: true, requiresPrompt: false},
	}
}

func publicTasks() []publicTaskSpec {
	return append(publicTasksBuildAndConfig(), publicTasksRunAndTypecheck()...)
}

// TestTaskfileAndReadmePublicApi
func TestTaskfileAndReadmePublicApi(t *testing.T) {
	t.Parallel()

	taskfile := loadTaskfile(t)

	expected := publicTaskNames()

	actual := taskNames(taskfile.tasks)

	if !slices.Equal(expected, actual) {
		t.Fatalf("public task drift\nexpected: %v\nactual:   %v", expected, actual)
	}

	readmeTasks := readmeTaskNames(readTool(t, constTypescriptTestReadmeMD))

	if !slices.Equal(expected, readmeTasks) {
		t.Fatalf("README public task drift\nexpected: %v\nactual:   %v", expected, readmeTasks)
	}
}

// TestTaskfileYamlIsCleanAndValid
func TestTaskfileYamlIsCleanAndValid(t *testing.T) {
	t.Parallel()

	content := read(t)

	assertLFLineEndings(t, content)
	assertNoTrailingWhitespace(t, content)

	root := documentRoot(t, parseTaskfileYAML(t, content))

	assertTaskfileVersion(t, root)
	assertNonEmptyTasksMap(t, root)
}

func assertLFLineEndings(t *testing.T, content string) {
	t.Helper()

	if strings.Contains(content, constTypescriptTestCRLF) {
		t.Fatal("Taskfile must use LF line endings")
	}
}

func assertNoTrailingWhitespace(t *testing.T, content string) {
	t.Helper()

	trimmedAll := strings.TrimRight(content, " \t\r\n")
	trimmedLF := strings.TrimRight(content, constTypescriptTestCRLF)

	if trimmedAll != trimmedLF {
		t.Fatal("Taskfile has trailing whitespace")
	}
}

func parseTaskfileYAML(t *testing.T, content string) *yaml.Node {
	t.Helper()

	var doc yaml.Node

	err := yaml.Unmarshal([]byte(content), &doc)
	if err != nil {
		t.Fatalf(constTypescriptTestParseTaskfileErr, err)
	}

	return &doc
}

func assertTaskfileVersion(t *testing.T, root *yaml.Node) {
	t.Helper()

	version := scalarField(root, constTypescriptTestVersion)

	if version != "3" && !strings.HasPrefix(version, "3.") {
		t.Fatalf("Taskfile version must be 3 or 3.x, got %q", version)
	}
}

func assertNonEmptyTasksMap(t *testing.T, root *yaml.Node) {
	t.Helper()

	tasks := mappingField(root, constTypescriptTestTasks)

	if tasks == nil || len(tasks.Content) == constTypescriptTestZero {
		t.Fatal("Taskfile must contain a non-empty tasks map")
	}
}

func taskCliListArgVariants() [][]string {
	return [][]string{
		{constTypescriptTestFlagList},
		{constTypescriptTestListAll},
		{constTypescriptTestListAll, constTypescriptTestFlagSort, constTypescriptTestSortAlpha},
		{constTypescriptTestListAll, constTypescriptTestJSONFlag},
	}
}

func assertTaskCliListSucceeds(t *testing.T, args []string) {
	t.Helper()

	result := runTask(t, isolatedEnv(t), args...)
	assertExitCode(t, &result, constTypescriptTestZero)
	assertNotContains(t, strings.ToLower(result.output), "taskfile does not exist")
	assertNotContains(t, strings.ToLower(result.output), "unknown")
}

// TestTaskCliCanLoadTaskfile
func TestTaskCliCanLoadTaskfile(t *testing.T) {
	t.Parallel()

	variants := taskCliListArgVariants()

	for i := range variants {
		args := variants[i]
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()

			assertTaskCliListSucceeds(t, args)
		})
	}
}

// TestTaskListAllJsonIsValid
func TestTaskListAllJsonIsValid(t *testing.T) {
	t.Parallel()

	result := runTask(t, isolatedEnv(t), constTypescriptTestListAll, constTypescriptTestJSONFlag)
	assertExitCode(t, &result, constTypescriptTestZero)

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

// TestPublicTasksHaveMetadataAndCommands
func TestPublicTasksHaveMetadataAndCommands(t *testing.T) {
	t.Parallel()

	taskfile := loadTaskfile(t)

	for i := range publicTasks() {
		spec := publicTasks()[i]
		t.Run(spec.name, func(t *testing.T) {
			t.Parallel()

			assertPublicTaskMetadata(t, &spec, mustTask(t, taskfile, spec.name))
		})
	}
}

func assertPublicTaskMetadata(t *testing.T, spec *publicTaskSpec, task *taskfile) {
	t.Helper()

	if task.node.Kind != yaml.MappingNode {
		t.Fatalf("public task %q must use mapping syntax", spec.name)
	}

	assertTaskDescription(t, spec, task)
	assertTaskSummary(t, spec, task)
	assertTaskHasCommands(t, spec, task)
}

func assertTaskDescription(t *testing.T, spec *publicTaskSpec, task *taskfile) {
	t.Helper()

	desc := nodeText(mappingValue(task.node, constTypescriptTestDesc))

	if len(strings.TrimSpace(desc)) < constTypescriptTestMinDescLen {
		t.Fatalf("public task %q desc is missing or too short: %q", spec.name, desc)
	}
}

func assertTaskSummary(t *testing.T, spec *publicTaskSpec, task *taskfile) {
	t.Helper()

	summary := nodeText(mappingValue(task.node, "summary"))

	if spec.requiresSummary && len(strings.TrimSpace(summary)) < constTypescriptTestMinSummaryLen {
		t.Fatalf("public task %q summary is missing or too short:\n%s", spec.name, summary)
	}
}

func assertTaskHasCommands(t *testing.T, spec *publicTaskSpec, task *taskfile) {
	t.Helper()

	missingCmdsAndDeps := isEmptyNode(mappingValue(task.node, "cmds")) &&
		isEmptyNode(mappingValue(task.node, "deps"))

	if missingCmdsAndDeps {
		t.Fatalf("public task %q must have cmds or deps", spec.name)
	}
}

// TestDestructivePublicTasksHavePrompt
func TestDestructivePublicTasksHavePrompt(t *testing.T) {
	t.Parallel()

	taskfile := loadTaskfile(t)

	for i := range publicTasks() {
		spec := publicTasks()[i]
		t.Run(spec.name, func(t *testing.T) {
			t.Parallel()

			if !spec.requiresPrompt {
				return
			}

			assertDestructivePrompt(t, &spec, mustTask(t, taskfile, spec.name))
		})
	}
}

func assertDestructivePrompt(t *testing.T, spec *publicTaskSpec, task *taskfile) {
	t.Helper()

	prompt := strings.ToLower(nodeText(mappingValue(task.node, "prompt")))

	if promptMentionsConfirmation(prompt) {
		return
	}

	t.Fatalf("destructive task %q needs an explicit prompt:\n%s", spec.name, prompt)
}

func promptMentionsConfirmation(prompt string) bool {
	keywords := []string{
		constTypescriptTestActionDelete,
		constTypescriptTestActionRemove,
		constTypescriptTestActionContinue,
	}

	for i := range keywords {
		keyword := keywords[i]

		if strings.Contains(prompt, keyword) {
			return true
		}
	}

	return false
}

// TestTaskSummariesWork
func TestTaskSummariesWork(t *testing.T) {
	t.Parallel()

	for i := range publicTasks() {
		spec := publicTasks()[i]
		t.Run(spec.name, func(t *testing.T) {
			t.Parallel()

			if !spec.requiresSummary {
				return
			}

			result := runTask(t, isolatedEnv(t), "--summary", spec.name)
			assertExitCode(t, &result, constTypescriptTestZero)
			assertContains(t, result.output, spec.name)
			assertNotContains(t, strings.ToLower(result.output), "task not found")
			assertNotContains(t, strings.ToLower(result.output), "no summary")
		})
	}
}

// TestTypecheckFilesRequiresExplicitFiles
func TestTypecheckFilesRequiresExplicitFiles(t *testing.T) {
	t.Parallel()

	result := runTask(t, isolatedEnv(t), "--dry", "--yes", constTypescriptTestTypecheckFiles)

	if result.err == nil {
		t.Fatalf("typecheck:files without FILES unexpectedly succeeded:\n%s", result.output)
	}

	assertContains(t, strings.ToLower(result.output), "files")
}

// TestTsserverGuidanceStaysEditorManaged
func TestTsserverGuidanceStaysEditorManaged(t *testing.T) {
	t.Parallel()

	content := read(t)

	tokens := []string{
		"Never start tsserver manually",
		"not LSP directly",
		"TypeScript Server protocol",
		"Managed by your editor",
	}

	for i := range tokens {
		token := tokens[i]
		assertContains(t, content, token)
	}
}

// TestCommandsDoNotContainDangerousPatterns
func TestCommandsDoNotContainDangerousPatterns(t *testing.T) {
	t.Parallel()

	taskfile := loadTaskfile(t)
	patterns := dangerousCommandPatterns()

	for taskName := range taskfile.tasks {
		task := taskfile.tasks[taskName]
		commands := collectScalars(task.node)

		for i := range commands {
			command := commands[i]
			assertCommandSafe(t, &dangerousCommand{taskName: taskName, command: command}, patterns)
		}
	}
}

func dangerousCommandPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		regexp.MustCompile(`(?m)\brm\s+-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*\s+/(?:\s|$)`),
		regexp.MustCompile(`(?m)\bsudo\s+rm\s+-[a-zA-Z]*r[a-zA-Z]*f`),
		regexp.MustCompile(`(?m)\bchmod\s+-R\s+777\s+/`),
		regexp.MustCompile(`(?m)\bcurl\b.*\s-k(?:\s|$)`),
		regexp.MustCompile(`(?m)\bcurl\b.*--insecure`),
	}
}

func assertCommandSafe(t *testing.T, cmd *dangerousCommand, patterns []*regexp.Regexp) {
	t.Helper()

	for i := range patterns {
		pattern := patterns[i]

		if pattern.MatchString(cmd.command) {
			t.Fatalf(
				"task %q contains dangerous command pattern %q:\n%s",
				cmd.taskName,
				pattern.String(),
				cmd.command,
			)
		}
	}
}

// TestNoPlaceholderTextInTaskfileOrReadme
func TestNoPlaceholderTextInTaskfileOrReadme(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		constTypescriptTestTaskfileYML: read(t),
		constTypescriptTestReadmeMD:    readTool(t, constTypescriptTestReadmeMD),
	}

	for name := range files {
		content := files[name]
		assertNoPlaceholders(t, name, content)
	}
}

func assertNoPlaceholders(t *testing.T, name, content string) {
	t.Helper()

	upper := strings.ToUpper(content)

	placeholders := []string{
		constTypescriptTestPlaceholderTODO,
		constTypescriptTestPlaceholderFIXME,
		constTypescriptTestPlaceholderCHANGEME,
		constTypescriptTestPlaceholderCopyright,
		constTypescriptTestPlaceholderLorem,
	}

	for i := range placeholders {
		placeholder := placeholders[i]

		if strings.Contains(upper, placeholder) {
			t.Fatalf("%s contains placeholder text: %s", name, placeholder)
		}
	}
}

func loadTaskfile(t *testing.T) loadedTaskfile {
	t.Helper()

	root := documentRoot(t, parseTaskfileYAML(t, read(t)))
	tasksNode := mappingField(root, constTypescriptTestTasks)

	if tasksNode == nil {
		t.Fatal("Taskfile has no tasks map")
	}

	return loadedTaskfile{tasks: buildTaskMap(tasksNode)}
}

func buildTaskMap(tasksNode *yaml.Node) map[string]*taskfile {
	tasks := map[string]*taskfile{}

	for i := constTypescriptTestZero; i < len(tasksNode.Content); i += constTypescriptTestPairStride {
		tasks[tasksNode.Content[i].Value] = &taskfile{
			node: tasksNode.Content[i+constTypescriptTestOne],
		}
	}

	return tasks
}

func mustTask(t *testing.T, taskfile loadedTaskfile, name string) *taskfile {
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
	fullArgs := append(
		[]string{"--taskfile", filepath.Join(dir(t), constTypescriptTestTaskfileYML)},
		args...)

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

	env = setEnv(env, constTypescriptTestHome, home)
	env = setEnv(env, "ZDOTDIR", home)
	env = setEnv(env, "CI", constTypescriptTestTrue)
	env = setEnv(env, "TASK_COLOR", "0")
	env = setEnv(env, "NO_COLOR", "1")
	env = setEnv(env, "TASK_ASSUME_YES", constTypescriptTestTrue)

	return env
}

func fakeTypeScriptProject(t *testing.T, env []string) (dir string, extraEnv []string) {
	t.Helper()

	dirs := newTypeScriptFixtureDirs(t, env)

	createTypeScriptFixtureDirs(t, &dirs)
	writeTypeScriptProjectFiles(t, &dirs)

	stubBody := toolStubScript()

	writeToolStubBinaries(t, &dirs, stubBody)
	writeToolchainStubs(t, &dirs, stubBody)

	path := dirs.binDir + ":" + dirs.nodeBinDir + ":" + envValue(dirs.env, constTypescriptTestPath)

	return dirs.projectDir, setEnv(env, constTypescriptTestPath, path)
}

func newTypeScriptFixtureDirs(t *testing.T, env []string) typeScriptFixtureDirs {
	t.Helper()

	dirs := typeScriptFixtureDirs{
		projectDir: t.TempDir(),
		binDir:     "",
		nodeBinDir: "",
		env:        env,
	}

	dirs.binDir = filepath.Join(dirs.projectDir, ".stub-bin")
	dirs.nodeBinDir = filepath.Join(dirs.projectDir, "node_modules", ".bin")

	return dirs
}

func toolStubScript() string {
	return "#!/usr/bin/env bash\n" +
		"case \"$1\" in\n" +
		"  --version) echo \"$0 1.0.0\" ;;\n" +
		"  *) exit 0 ;;\n" +
		"esac\n"
}

func writeTypeScriptProjectManifestFiles(t *testing.T, dirs *typeScriptFixtureDirs) {
	t.Helper()

	writeFile(
		t,
		filepath.Join(dirs.projectDir, "package.json"),
		`{"scripts":{"build":"tsc"}}`+"\n",
		constTypescriptTestMode0644,
	)
	writeFile(
		t,
		filepath.Join(dirs.projectDir, "package-lock.json"),
		"{}\n",
		constTypescriptTestMode0600,
	)
	writeFile(
		t,
		filepath.Join(dirs.projectDir, "tsconfig.json"),
		`{"compilerOptions":{"outDir":constTypescriptTestDist}}`+"\n",
		constTypescriptTestMode0644,
	)
}

func writeTypeScriptProjectFiles(t *testing.T, dirs *typeScriptFixtureDirs) {
	t.Helper()

	writeTypeScriptProjectManifestFiles(t, dirs)
	writeFile(
		t,
		filepath.Join(dirs.projectDir, constTypescriptTestSrc, "index.ts"),
		"export {}\n",
		constTypescriptTestMode0600,
	)
	writeFile(
		t,
		filepath.Join(dirs.projectDir, constTypescriptTestDist, "index.js"),
		"console.log('ok')\n",
		constTypescriptTestMode0600,
	)
}

func writeNodeToolStubBinaries(t *testing.T, dirs *typeScriptFixtureDirs, stubBody string) {
	t.Helper()

	nodeBinNames := []string{
		constTypescriptTestBinTsc, constTypescriptTestBinTsx, constTypescriptTestBinTsserver,
	}

	for i := range nodeBinNames {
		writeFile(
			t,
			filepath.Join(dirs.nodeBinDir, nodeBinNames[i]),
			stubBody,
			constTypescriptTestMode0500,
		)
	}
}

func writePathToolStubBinaries(t *testing.T, dirs *typeScriptFixtureDirs, stubBody string) {
	t.Helper()

	toolNames := []string{
		constTypescriptTestFnm, "node", "npx", "npm", "pnpm", "yarn",
		constTypescriptTestBun,
	}

	for i := range toolNames {
		name := toolNames[i]
		writeFile(t, filepath.Join(dirs.binDir, name), stubBody, constTypescriptTestMode0500)
	}
}

func writeToolStubBinaries(t *testing.T, dirs *typeScriptFixtureDirs, stubBody string) {
	t.Helper()

	writeNodeToolStubBinaries(t, dirs, stubBody)
	writePathToolStubBinaries(t, dirs, stubBody)
}

func writeBunToolchainStub(t *testing.T, home, stubBody string) {
	t.Helper()

	writeFile(
		t,
		filepath.Join(
			home,
			constTypescriptTestDotBun,
			constTypescriptTestBin,
			constTypescriptTestBun,
		),
		stubBody,
		constTypescriptTestMode0500,
	)
}

func writeNvmToolchainStub(t *testing.T, home string) {
	t.Helper()

	writeFile(
		t,
		filepath.Join(home, constTypescriptTestDotNvm, "nvm.sh"),
		"# nvm stub\n",
		constTypescriptTestMode0600,
	)
}

func writeFnmToolchainStub(t *testing.T, home, stubBody string) {
	t.Helper()

	writeFile(
		t,
		filepath.Join(
			home,
			constTypescriptTestDotLocal,
			constTypescriptTestShare,
			constTypescriptTestFnm,
			constTypescriptTestFnm,
		),
		stubBody,
		constTypescriptTestMode0755,
	)
}

func writeToolchainStubs(t *testing.T, dirs *typeScriptFixtureDirs, stubBody string) {
	t.Helper()

	home := envValue(dirs.env, constTypescriptTestHome)

	writeBunToolchainStub(t, home, stubBody)
	writeNvmToolchainStub(t, home)
	writeFnmToolchainStub(t, home, stubBody)
}

func typeScriptFixturePaths(dirs *typeScriptFixtureDirs, home string) []string {
	return []string{
		filepath.Join(dirs.projectDir, constTypescriptTestSrc),
		filepath.Join(dirs.projectDir, constTypescriptTestDist),
		dirs.binDir,
		dirs.nodeBinDir,
		filepath.Join(home, constTypescriptTestDotBun, constTypescriptTestBin),
		filepath.Join(home, constTypescriptTestDotNvm),
		filepath.Join(
			home,
			constTypescriptTestDotLocal,
			constTypescriptTestShare,
			constTypescriptTestFnm,
		),
	}
}

func createTypeScriptFixtureDirs(t *testing.T, dirs *typeScriptFixtureDirs) {
	t.Helper()

	home := envValue(dirs.env, constTypescriptTestHome)
	paths := typeScriptFixturePaths(dirs, home)

	for i := range paths {
		err := os.MkdirAll(paths[i], constTypescriptTestMode0700)
		if err != nil {
			t.Fatalf("create test project dir %s: %v", paths[i], err)
		}
	}
}

func writeFile(t *testing.T, fileValue any, parts ...any) {
	t.Helper()

	file := normalizeTestFile(t, fileValue, parts)

	err := os.WriteFile(file.path, []byte(file.content), file.mode)
	if err != nil {
		t.Fatalf("write %s: %v", file.path, err)
	}
}

func normalizeTestFile(t *testing.T, fileValue any, parts []any) testFile {
	t.Helper()

	if file, ok := fileValue.(testFile); ok {
		return normalizeExplicitTestFile(t, &file, parts)
	}

	return normalizeTestFileFromParts(t, fileValue, parts)
}

func normalizeExplicitTestFile(t *testing.T, file *testFile, parts []any) testFile {
	t.Helper()

	if len(parts) != constTypescriptTestZero {
		t.Fatalf("testFile does not accept positional arguments: %v", parts)
	}

	return *file
}

func normalizeTestFileFromParts(t *testing.T, fileValue any, parts []any) testFile {
	t.Helper()

	path, ok := fileValue.(string)

	if !ok {
		t.Fatalf("file path must be string or testFile, got %T", fileValue)
	}

	if len(parts) != constTypescriptTestPairStride {
		t.Fatalf("writeFile requires content and mode, got %d values", len(parts))
	}

	content, ok := parts[constTypescriptTestZero].(string)

	if !ok {
		t.Fatalf("file content must be string, got %T", parts[constTypescriptTestZero])
	}

	return testFile{path: path, content: content, mode: fileMode(t, parts[constTypescriptTestOne])}
}

func fileMode(t *testing.T, value any) os.FileMode {
	t.Helper()

	switch mode := value.(type) {
	case os.FileMode:
		return mode
	case int:
		return fileModeFromInt(t, mode)
	default:
		t.Fatalf("file mode must be os.FileMode, got %T", value)
	}

	return constTypescriptTestZero
}

func fileModeFromInt(t *testing.T, mode int) os.FileMode {
	t.Helper()

	if mode < constTypescriptTestZero {
		t.Fatalf("file mode must be non-negative, got %d", mode)
	}

	if mode > math.MaxUint32 {
		t.Fatalf("file mode too large: %d", mode)
	}

	parsed, err := strconv.ParseUint(
		strconv.Itoa(mode),
		constTypescriptTestDecimalBase,
		constTypescriptTestUint32BitSize,
	)
	if err != nil {
		t.Fatalf("parse file mode: %v", err)
	}

	return os.FileMode(parsed)
}

func envValue(env []string, key string) string {
	prefix := key + "="

	for i := range env {
		item := env[i]

		if after, ok := strings.CutPrefix(item, prefix); ok {
			return after
		}
	}

	return os.Getenv(key)
}

func publicTaskNames() []string {
	names := make([]string, constTypescriptTestZero, len(publicTasks()))

	for i := range publicTasks() {
		spec := publicTasks()[i]

		names = append(names, spec.name)
	}

	slices.Sort(names)

	return names
}

func taskNames(tasks map[string]*taskfile) []string {
	names := []string{}

	for name := range tasks {
		task := tasks[name]

		if isPublicNamedTask(name, task) {
			names = append(names, name)
		}
	}

	slices.Sort(names)

	return names
}

func isPublicNamedTask(name string, task *taskfile) bool {
	if name == "default" || strings.HasPrefix(name, "_") {
		return false
	}

	return nodeText(
		mappingValue(task.node, constTypescriptTestDesc),
	) != constTypescriptTestEmptyString
}

func (scanner *readmeSectionScanner) advance(trimmed string) bool {
	if scanner.done {
		return false
	}

	if trimmed == scanner.header {
		scanner.active = true

		return false
	}

	if scanner.active && strings.HasPrefix(trimmed, "## ") {
		scanner.done = true
		scanner.active = false

		return false
	}

	return scanner.active
}

func (scanner *readmeTaskScanner) taskName(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)

	if !scanner.section.advance(trimmed) {
		return constTypescriptTestEmptyString, false
	}

	return readmeTaskRowName(scanner.row, trimmed)
}

func readmeTaskRowName(row *regexp.Regexp, trimmed string) (string, bool) {
	match := row.FindStringSubmatch(trimmed)

	if len(match) != constTypescriptTestPairStride {
		return constTypescriptTestEmptyString, false
	}

	return match[constTypescriptTestOne], true
}

func readmeTaskNames(content string) []string {
	scanner := readmeTaskScanner{
		section: &readmeSectionScanner{header: "## Public Tasks", active: false, done: false},
		row:     regexp.MustCompile("^\\|\\s*`([^`]+)`\\s*\\|"),
	}
	names := []string{}

	for line := range strings.SplitSeq(content, constTypescriptTestLF) {
		if name, ok := scanner.taskName(line); ok {
			names = append(names, name)
		}
	}

	slices.Sort(names)

	return names
}

func documentRoot(t *testing.T, doc *yaml.Node) *yaml.Node {
	t.Helper()

	if doc.Kind != yaml.DocumentNode || len(doc.Content) != constTypescriptTestOne {
		t.Fatalf(
			"expected YAML document node, got kind=%directory children=%directory",
			doc.Kind,
			len(doc.Content),
		)
	}

	return doc.Content[constTypescriptTestZero]
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

	for i := constTypescriptTestZero; i < len(node.Content); i += constTypescriptTestPairStride {
		if node.Content[i].Value == name {
			return node.Content[i+constTypescriptTestOne]
		}
	}

	return nil
}

func nodeText(node *yaml.Node) string {
	if node == nil {
		return constTypescriptTestEmptyString
	}

	if node.Kind == yaml.ScalarNode {
		return node.Value
	}

	parts := []string{}

	for i := range node.Content {
		child := node.Content[i]

		if text := nodeText(child); text != constTypescriptTestEmptyString {
			parts = append(parts, text)
		}
	}

	return strings.Join(parts, constTypescriptTestLF)
}

func isEmptyNode(node *yaml.Node) bool {
	return node == nil || strings.TrimSpace(nodeText(node)) == constTypescriptTestEmptyString
}

func collectScalars(node *yaml.Node) []string {
	if node == nil {
		return nil
	}

	if node.Kind == yaml.ScalarNode {
		return []string{node.Value}
	}

	values := make([]string, constTypescriptTestZero, len(node.Content))

	for i := range node.Content {
		child := node.Content[i]

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
		t.Fatalf(constTypescriptTestReadErr, name, err)
	}

	return string(content)
}

func read(t *testing.T) string {
	t.Helper()

	path := filepath.Join(dir(t), constTypescriptTestTaskfileYML)

	content, err := fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		t.Fatalf(constTypescriptTestReadErr, path, err)
	}

	return string(content)
}

func dir(t *testing.T) string {
	t.Helper()

	programCounter, file, line, ok := runtime.Caller(constTypescriptTestZero)

	if !ok || programCounter == constTypescriptTestZero || line == constTypescriptTestZero {
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

func assertExitCode(t *testing.T, result *commandResult, want int) {
	t.Helper()

	wantSuccess := want == constTypescriptTestZero

	if wantSuccess && result.err != nil {
		t.Fatalf("command failed: %v\n%s", result.err, result.output)
	}

	if !wantSuccess && result.err == nil {
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
