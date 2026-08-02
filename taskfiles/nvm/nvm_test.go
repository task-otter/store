// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package nvm_test

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/task-otter/store/internal/tasktestutil"
	yaml "go.yaml.in/yaml/v3"
)

type (
	undoPair struct {
		task       string
		undoTarget string
	}

	scriptCheck struct {
		root     string
		taskName string
	}

	realTaskRun struct {
		t    *testing.T
		root string
		env  []string
	}

	nvmIntegration struct {
		root   string
		nvmDir string
		env    []string
	}

	taskScriptsRef struct {
		root     string
		taskName string
		task     tasktestutil.TaskNode
	}

	nvmRunFunc func(t *testing.T, args ...string) tasktestutil.CommandResult

	publicTaskMetadataCheck struct {
		taskfile *tasktestutil.LoadedTaskfile
		spec     *tasktestutil.PublicTaskSpec
	}
)

const (
	nodeInstallTask   = "node:install"
	nodeVersion2400   = "24.0.0"
	nodeUninstallTask = "node:uninstall"
	versionArg        = "VERSION"
	listAllFlag       = "--list-all"
	listFlag          = "--list"
	sortFlag          = "--sort"
	sortAlphanumeric  = "alphanumeric"
	summaryField      = "summary"
	windowsOS         = "windows"
	installTask       = "install"
	installUndoTask   = "install:undo"
	lsTask            = "ls"
	ltsFlag           = "--lts"
	nodeUseTask       = "node:use"
	nodeVersionTask   = "node:version"
	versionTask       = "version"
	taskfileLabel     = "Taskfile"
	jsonFlag          = "--json"
	descField         = "desc"
	outputField       = "output"
	taskMissingFmt    = "task %q is missing"
	nvmDirEnvKey      = "NVM_DIR"
	skipUnixOnlyMsg   = "stub nvm tests target Unix-like systems"
	yesFlag           = "--yes"
	nvmShFileName     = "nvm.sh"
	versionsDirName   = "versions"
	nodeDirName       = "node"
	version18Arg      = "VERSION=18.0.0"
	version18         = "18.0.0"
	versionV18DirName = "v18.0.0"
	dirPerm0700       = 0o700
	filePerm0600      = 0o600
	stubVersionDirErr = "failed to create stub version dir: %v"
	exitCodeSuccess   = 0
	minDescLen        = 12
	minSummaryLen     = 25
	emptyString       = ""
)

func expectedPublicTasksInstall() []tasktestutil.PublicTaskSpec {
	spec := tasktestutil.NewPublicTaskSpec

	return []tasktestutil.PublicTaskSpec{
		spec(
			installTask,
			tasktestutil.WithDryRunArgs(),
			tasktestutil.WithGroupOutput(),
			tasktestutil.WithSummary(),
		),
		spec(installUndoTask, tasktestutil.WithDryRunArgs(), tasktestutil.WithGroupOutput(),
			tasktestutil.WithPrompt(), tasktestutil.WithSummary()),
		spec(lsTask, tasktestutil.WithDryRunArgs(), tasktestutil.WithSummary()),
	}
}

func expectedPublicTasksNode() []tasktestutil.PublicTaskSpec {
	spec := tasktestutil.NewPublicTaskSpec
	versionArgs := map[string]string{versionArg: nodeVersion2400}

	return []tasktestutil.PublicTaskSpec{
		spec(nodeInstallTask, tasktestutil.WithArgs(versionArgs), tasktestutil.WithDryRunArgs(),
			tasktestutil.WithDryRunNoArgs(), tasktestutil.WithExpectedDefaultTokens(ltsFlag),
			tasktestutil.WithGroupOutput(), tasktestutil.WithSummary()),
		spec(nodeUninstallTask, tasktestutil.WithArgs(versionArgs), tasktestutil.WithDryRunArgs(),
			tasktestutil.WithGroupOutput(), tasktestutil.WithPrompt(), tasktestutil.WithSummary()),
		spec(
			nodeUseTask,
			tasktestutil.WithArgs(versionArgs),
			tasktestutil.WithDryRunArgs(),
			tasktestutil.WithDryRunNoArgs(),
			tasktestutil.WithExpectedDefaultTokens(ltsFlag),
			tasktestutil.WithSummary(),
		),
		spec(nodeVersionTask, tasktestutil.WithDryRunArgs(), tasktestutil.WithSummary()),
	}
}

func expectedPublicTasksMisc() []tasktestutil.PublicTaskSpec {
	spec := tasktestutil.NewPublicTaskSpec

	return []tasktestutil.PublicTaskSpec{
		spec(versionTask, tasktestutil.WithDryRunArgs(), tasktestutil.WithSummary()),
	}
}

func expectedPublicTasks() []tasktestutil.PublicTaskSpec {
	tasks := expectedPublicTasksInstall()

	tasks = append(tasks, expectedPublicTasksNode()...)

	return append(tasks, expectedPublicTasksMisc()...)
}

func isolatedEnv(t *testing.T) []string {
	t.Helper()

	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")

	return tasktestutil.SetEnv(env, nvmDirEnvKey, filepath.Join(home, ".nvm"))
}

// TestTaskBinaryIsAvailable
func TestTaskBinaryIsAvailable(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)
	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{Root: root, Env: nil, Args: []string{"--version"}},
	)
	tasktestutil.AssertExitCode(t, result, exitCodeSuccess)
	tasktestutil.AssertNotEmpty(t, result.Combined(), "task --version output is empty")
}

// TestTaskfileYamlIsCleanAndValid
func TestTaskfileYamlIsCleanAndValid(t *testing.T) {
	t.Parallel()

	path := tasktestutil.ModuleTaskfilePath(t)
	content := tasktestutil.ReadFile(t, path)
	tasktestutil.AssertTextFileClean(t, path, content)

	doc := parseTaskfileYaml(t, content)

	tasktestutil.AssertNoDuplicateMappingKeys(t, &doc, taskfileLabel)
	tasktestutil.AssertNoYamlAliases(t, &doc, taskfileLabel)

	root := tasktestutil.DocumentRoot(t, &doc)

	assertTaskfileVersion(t, root)
	assertTaskfileHasTasks(t, root)
}

func parseTaskfileYaml(t *testing.T, content string) yaml.Node {
	t.Helper()

	var doc yaml.Node

	err := yaml.Unmarshal([]byte(content), &doc)
	if err != nil {
		t.Fatalf("Taskfile YAML is invalid: %v", err)
	}

	return doc
}

func assertTaskfileVersion(t *testing.T, root *yaml.Node) {
	t.Helper()

	version := tasktestutil.ScalarField(root, versionTask)

	if version != "3" && !strings.HasPrefix(version, "3.") {
		t.Fatalf("Taskfile version must be 3 or 3.x, got %q", version)
	}
}

func assertTaskfileHasTasks(t *testing.T, root *yaml.Node) {
	t.Helper()

	tasks := tasktestutil.MappingField(root, "tasks")

	if tasks == nil || len(tasks.Content) == exitCodeSuccess {
		t.Fatal("Taskfile must contain non-empty tasks map")
	}
}

func taskCliListArgVariants() [][]string {
	return [][]string{
		{listFlag},
		{listAllFlag},
		{listAllFlag, sortFlag, sortAlphanumeric},
		{listAllFlag, jsonFlag},
	}
}

func assertTaskCliListSucceeds(t *testing.T, root string, args []string) {
	t.Helper()

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{Root: root, Env: isolatedEnv(t), Args: args},
	)
	tasktestutil.AssertExitCode(t, result, exitCodeSuccess)
	tasktestutil.AssertNotContains(
		t,
		strings.ToLower(result.Combined()),
		"taskfile does not exist",
	)
	tasktestutil.AssertNotContains(t, strings.ToLower(result.Combined()), "unknown")
}

// TestTaskCliCanLoadTaskfile
func TestTaskCliCanLoadTaskfile(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)

	variants := taskCliListArgVariants()

	for i := range variants {
		args := variants[i]
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()

			assertTaskCliListSucceeds(t, root, args)
		})
	}
}

// TestTaskListAllJsonIsValid
func TestTaskListAllJsonIsValid(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)
	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: root,
			Env:  isolatedEnv(t),
			Args: []string{listAllFlag, jsonFlag},
		},
	)
	tasktestutil.AssertExitCode(t, result, exitCodeSuccess)

	err := tasktestutil.ValidateJSON(result.Stdout)
	if err != nil {
		t.Fatalf("task --list-all --json produced invalid JSON:\n%s\nerror: %v", result.Stdout, err)
	}
}

// TestPublicApiDoesNotDrift
func TestPublicApiDoesNotDrift(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)
	expected := tasktestutil.ExpectedPublicTaskNames(expectedPublicTasks())

	actual := tasktestutil.PublicTaskNamesFromTaskfile(t, taskfile)

	if !slices.Equal(expected, actual) {
		t.Fatalf(
			"public Taskfile API drift detected\n\nexpected:\n%s\n\nactual:\n%s\n\n"+
				"Fix either the Taskfile public tasks or expectedPublicTasks in the test.",
			tasktestutil.FormatList(expected), tasktestutil.FormatList(actual),
		)
	}
}

// TestEveryTaskIsEitherPublicOrInternal
func TestEveryTaskIsEitherPublicOrInternal(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)

	for name := range taskfile.Tasks {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assertPublicOrInternalTask(t, name, taskfile.Tasks[name])
		})
	}
}

func assertPublicOrInternalTask(t *testing.T, name string, task tasktestutil.TaskNode) {
	t.Helper()

	if strings.HasPrefix(name, "_") || task.BoolField("internal") {
		return
	}

	if task.StringField(descField) == emptyString {
		t.Fatalf(
			"task %q is not internal and has no desc. Either add desc/summary or mark it internal: true",
			name,
		)
	}
}

// TestPublicTasksHaveMetadata
func assertTaskHasValidMetadata(t *testing.T, check *publicTaskMetadataCheck) {
	t.Helper()

	task := tasktestutil.MustTask(t, check.taskfile, check.spec.Name)
	assertTaskUsesFullMapping(t, check.spec.Name, task)
	assertTaskDescValid(t, check.spec.Name, task.StringField(descField))
	assertTaskSummaryValid(t, check.spec, task.StringField(summaryField))

	desc := task.StringField(descField)
	summary := task.StringField(summaryField)

	tasktestutil.AssertNoPlaceholderText(t, check.spec.Name, desc)
	tasktestutil.AssertNoPlaceholderText(t, check.spec.Name, summary)
}

// TestPublicTasksHaveMetadata validates the behavior covered by this test case.
func TestPublicTasksHaveMetadata(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)
	specs := expectedPublicTasks()

	for i := range specs {
		spec := specs[i]
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()

			assertTaskHasValidMetadata(
				t,
				&publicTaskMetadataCheck{taskfile: &taskfile, spec: &spec},
			)
		})
	}
}

func assertTaskUsesFullMapping(t *testing.T, name string, task tasktestutil.TaskNode) {
	t.Helper()

	if task.Node.Kind != yaml.MappingNode {
		t.Fatalf("public task %q must use full mapping syntax, not short syntax", name)
	}
}

func assertTaskDescValid(t *testing.T, name, desc string) {
	t.Helper()

	if strings.TrimSpace(desc) == emptyString {
		t.Fatalf("public task %q is missing desc", name)
	}

	if len(strings.TrimSpace(desc)) < minDescLen {
		t.Fatalf("public task %q desc is too short: %q", name, desc)
	}
}

func assertTaskSummaryValid(t *testing.T, spec *tasktestutil.PublicTaskSpec, summary string) {
	t.Helper()

	if !spec.RequiresSummary {
		return
	}

	if strings.TrimSpace(summary) == emptyString {
		t.Fatalf("public task %q is missing summary", spec.Name)
	}

	if len(strings.TrimSpace(summary)) < minSummaryLen {
		t.Fatalf("public task %q summary is too short:\n%s", spec.Name, summary)
	}
}

// TestDestructivePublicTasksHavePrompt
func TestDestructivePublicTasksHavePrompt(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)

	for i := range expectedPublicTasks() {
		spec := expectedPublicTasks()[i]
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()

			if !spec.RequiresPrompt {
				return
			}

			task := tasktestutil.MustTask(t, taskfile, spec.Name)

			tasktestutil.AssertDestructivePrompt(t, spec.Name, task.Field("prompt"))
		})
	}
}

// TestInstallTasksUseGithubGroupOutput
func TestInstallTasksUseGithubGroupOutput(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)

	for i := range expectedPublicTasks() {
		spec := expectedPublicTasks()[i]
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()

			if !spec.RequiresGroupOutput {
				return
			}

			assertTaskUsesGroupOutput(t, &taskfile, spec.Name)
		})
	}
}

func assertTaskUsesGroupOutput(t *testing.T, taskfile *tasktestutil.LoadedTaskfile, name string) {
	t.Helper()

	task := tasktestutil.MustTask(t, taskfile, name)

	outputNode := task.Field(outputField)

	if outputNode == nil {
		outputNode = taskfile.Root.Field(outputField)
	}

	tasktestutil.AssertGithubGroupOutput(t, name, outputNode)
}

// TestPublicTasksHaveCommands
func TestPublicTasksHaveCommands(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)

	for i := range expectedPublicTasks() {
		spec := expectedPublicTasks()[i]
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()

			task := tasktestutil.MustTask(t, taskfile, spec.Name)
			assertTaskHasCommandsOrDeps(t, spec.Name, task)
		})
	}
}

func assertTaskHasCommandsOrDeps(t *testing.T, name string, task tasktestutil.TaskNode) {
	t.Helper()

	missingCmdsAndDeps := tasktestutil.IsEmptyNode(task.Field("cmds")) &&
		tasktestutil.IsEmptyNode(task.Field("deps"))

	if missingCmdsAndDeps {
		t.Fatalf("public task %q must have cmds or deps", name)
	}
}

func runTaskSummary(t *testing.T, root, name string) tasktestutil.CommandResult {
	t.Helper()

	return tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: root,
			Env:  isolatedEnv(t),
			Args: []string{"--summary", name},
		},
	)
}

func assertTaskSummaryWorks(t *testing.T, root string, spec *tasktestutil.PublicTaskSpec) {
	t.Helper()

	if !spec.RequiresSummary {
		return
	}

	result := runTaskSummary(t, root, spec.Name)
	tasktestutil.AssertExitCode(t, result, exitCodeSuccess)

	out := result.Combined()
	tasktestutil.AssertContains(t, out, spec.Name)
	tasktestutil.AssertNotContains(t, strings.ToLower(out), "task not found")
	tasktestutil.AssertNotContains(t, strings.ToLower(out), "unknown task")
	tasktestutil.AssertNotContains(t, strings.ToLower(out), "no summary")
}

// TestTaskSummariesWork
func TestTaskSummariesWork(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)
	specs := expectedPublicTasks()

	for i := range specs {
		spec := specs[i]
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()

			assertTaskSummaryWorks(t, root, &spec)
		})
	}
}

// TestUndoPairsExist
func TestUndoPairsExist(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)

	pairs := []undoPair{
		{task: installTask, undoTarget: installUndoTask},
		{task: nodeInstallTask, undoTarget: nodeUninstallTask},
		{task: nodeUninstallTask, undoTarget: nodeInstallTask},
	}

	for i := range pairs {
		pair := pairs[i]
		assertUndoPairPresent(t, &taskfile, &pair)
	}
}

func assertUndoPairPresent(t *testing.T, taskfile *tasktestutil.LoadedTaskfile, pair *undoPair) {
	t.Helper()

	if _, ok := taskfile.Tasks[pair.task]; !ok {
		t.Fatalf(taskMissingFmt, pair.task)
	}

	if _, ok := taskfile.Tasks[pair.undoTarget]; !ok {
		t.Fatalf("undo task %q for %q is missing", pair.undoTarget, pair.task)
	}
}

func assertReferencedScriptExistsForCommandRun(t *testing.T, check *scriptCheck, command string) {
	t.Helper()

	t.Run(check.taskName, func(t *testing.T) {
		t.Parallel()

		assertReferencedScriptsExist(t, check, command)
	})
}

func assertReferencedScriptsExistForTask(t *testing.T, ref *taskScriptsRef) {
	t.Helper()

	commands := tasktestutil.CollectCommandStrings(ref.task.Node)
	check := &scriptCheck{root: ref.root, taskName: ref.taskName}

	for i := range commands {
		assertReferencedScriptExistsForCommandRun(t, check, commands[i])
	}
}

// TestReferencedScriptsExist
func TestReferencedScriptsExist(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)
	taskfile := tasktestutil.LoadTaskfile(t)

	for taskName := range taskfile.Tasks {
		assertReferencedScriptsExistForTask(
			t,
			&taskScriptsRef{root: root, taskName: taskName, task: taskfile.Tasks[taskName]},
		)
	}
}

func assertReferencedScriptsExist(t *testing.T, check *scriptCheck, command string) {
	t.Helper()

	for i := range tasktestutil.ReferencedLocalShellScripts(command) {
		scriptPath := tasktestutil.ReferencedLocalShellScripts(command)[i]
		assertScriptFileExists(t, check, scriptPath)
	}
}

func assertScriptFileExists(t *testing.T, check *scriptCheck, scriptPath string) {
	t.Helper()

	abs := filepath.Join(check.root, scriptPath)

	info, err := os.Stat(abs)
	if err != nil {
		t.Fatalf("task %q references missing script %q", check.taskName, scriptPath)
	}

	if info.IsDir() {
		t.Fatalf(
			"task %q references script path but it is a directory: %q",
			check.taskName,
			scriptPath,
		)
	}
}

// TestCommandsDoNotContainDangerousPatterns
func TestCommandsDoNotContainDangerousPatterns(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)

	for taskName := range taskfile.Tasks {
		task := taskfile.Tasks[taskName]

		for i := range tasktestutil.CollectCommandStrings(task.Node) {
			command := tasktestutil.CollectCommandStrings(task.Node)[i]
			assertCommandHasNoDangerousPattern(t, taskName, command)
		}
	}
}

func assertCommandHasNoDangerousPattern(t *testing.T, taskName, command string) {
	t.Helper()

	for i := range tasktestutil.DangerousCommandPatterns() {
		pattern := tasktestutil.DangerousCommandPatterns()[i]

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

// TestNoPlaceholderTextInTaskfile
func TestNoPlaceholderTextInTaskfile(t *testing.T) {
	t.Parallel()

	content := tasktestutil.ReadFile(t, tasktestutil.ModuleTaskfilePath(t))

	upper := strings.ToUpper(content)

	placeholders := []string{
		"TODO",
		"FIXME",
		"CHANGEME",
		"Copyright",
		"YOUR VALUE HERE",
		"LOREM IPSUM",
	}

	for i := range placeholders {
		if strings.Contains(upper, placeholders[i]) {
			t.Fatalf("Taskfile contains placeholder text: %s", placeholders[i])
		}
	}
}

// TestRealInstallerFlowOnlyWhenExplicitlyEnabled
func TestRealInstallerFlowOnlyWhenExplicitlyEnabled(t *testing.T) {
	t.Parallel()
	skipUnlessRealInstallerTestsEnabled(t)

	root := tasktestutil.ModuleRoot(t)
	env := isolatedEnv(t)
	nvmDir := requireNvmDirFromEnv(t, env)
	realRun := realTaskRun{t: t, root: root, env: env}

	runRealInstallUninstallFlow(t, &realRun, nvmDir)
}

func runRealInstallUninstallFlow(t *testing.T, realRun *realTaskRun, nvmDir string) {
	t.Helper()

	realRun.run(yesFlag, installTask)
	assertNvmDirExists(t, nvmDir)

	realRun.run(versionTask)
	realRun.run(yesFlag, installUndoTask)

	assertNvmDirRemoved(t, nvmDir)
}

func skipUnlessRealInstallerTestsEnabled(t *testing.T) {
	t.Helper()

	if os.Getenv("RUN_INSTALLER_TESTS") != "1" {
		t.Skip("set RUN_INSTALLER_TESTS=1 to run real install/uninstall tests")
	}

	if runtime.GOOS == windowsOS {
		t.Skip("real nvm shell installer tests are intended for Unix-like systems")
	}
}

func requireNvmDirFromEnv(t *testing.T, env []string) string {
	t.Helper()

	nvmDir := tasktestutil.EnvValue(env, nvmDirEnvKey)

	if nvmDir == emptyString {
		t.Fatal("NVM_DIR was not set in isolated environment")
	}

	return nvmDir
}

func (run *realTaskRun) run(args ...string) {
	run.t.Helper()

	tasktestutil.AssertExitCode(
		run.t,
		tasktestutil.RunTaskTimeout(
			run.t,
			tasktestutil.TaskRun{Root: run.root, Env: run.env, Args: args},
			10*time.Minute,
		),
		exitCodeSuccess,
	)
}

func assertNvmDirExists(t *testing.T, nvmDir string) {
	t.Helper()

	info, err := os.Stat(nvmDir)
	if err != nil {
		t.Fatalf("expected NVM_DIR to exist after install: %s\nerror: %v", nvmDir, err)
	}

	if !info.IsDir() {
		t.Fatalf("expected NVM_DIR to be a directory after install: %s", nvmDir)
	}
}

func assertNvmDirRemoved(t *testing.T, nvmDir string) {
	t.Helper()

	info, err := os.Stat(nvmDir)

	if err == nil && info != nil {
		t.Fatalf("expected path to not exist: %s", nvmDir)
	}

	if !os.IsNotExist(err) {
		t.Fatalf("expected NVM_DIR to be removed after install:undo: %s", nvmDir)
	}
}

// TestAllPublicTasksIntegration
func TestAllPublicTasksIntegration(t *testing.T) {
	t.Parallel()

	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip(
			"set RUN_INTEGRATION_TESTS=1 to run integration tests (downloads and installs NVM and Node.js)",
		)
	}

	if runtime.GOOS == windowsOS {
		t.Skip("integration tests target Unix-like systems")
	}

	root := tasktestutil.ModuleRoot(t)
	env := isolatedEnv(t)
	nvmDir := tasktestutil.EnvValue(env, nvmDirEnvKey)

	runNvmIntegrationSteps(t, &nvmIntegration{root: root, env: env, nvmDir: nvmDir})
}

func runNvmIntegrationSteps(t *testing.T, integration *nvmIntegration) {
	t.Helper()

	run := successfulIntegrationRun(integration.root, integration.env)

	runNvmInstallSteps(t, run, integration.nvmDir)
	runNvmNodeSteps(t, run, integration.nvmDir)
	runIntegrationStep(t, "install:undo — NVM directory is removed", func(t *testing.T) {
		t.Helper()
		run(t, yesFlag, installUndoTask)
		tasktestutil.AssertDirNotExists(t, integration.nvmDir)
	})
}

func runNvmBinaryInstallSteps(t *testing.T, run nvmRunFunc, nvmDir string) {
	t.Helper()

	runIntegrationStep(t, "install — nvm.sh is present on disk", func(t *testing.T) {
		t.Helper()
		run(t, yesFlag, installTask)
		tasktestutil.AssertFileExists(t, filepath.Join(nvmDir, nvmShFileName))
	})
	runIntegrationStep(t, "version — nvm version string is printed", func(t *testing.T) {
		t.Helper()

		result := run(t, versionTask)
		tasktestutil.AssertNotEmpty(t, result.Combined(), "version output is empty")
	})
}

func runNvmInstallSteps(t *testing.T, run nvmRunFunc, nvmDir string) {
	t.Helper()

	runNvmBinaryInstallSteps(t, run, nvmDir)
	runIntegrationStep(
		t,
		"node:install — default LTS version directory is created",
		func(t *testing.T) {
			t.Helper()
			run(t, yesFlag, nodeInstallTask)
			tasktestutil.AssertDirHasEntries(t, filepath.Join(nvmDir, versionsDirName, nodeDirName))
		},
	)
	runIntegrationStep(t, "ls — installed versions appear in output", func(t *testing.T) {
		t.Helper()

		result := run(t, lsTask)
		tasktestutil.AssertNotEmpty(t, result.Combined(), "ls output is empty")
	})
}

func runNvmNodeVersionSteps(t *testing.T, run nvmRunFunc, nvmDir string) {
	t.Helper()

	versionDir := filepath.Join(nvmDir, versionsDirName, nodeDirName, "v"+version18)

	runIntegrationStep(t, "node:install VERSION=18.0.0 — specific version directory is created",
		func(t *testing.T) {
			t.Helper()
			run(t, yesFlag, nodeInstallTask, "VERSION="+version18)
			tasktestutil.AssertDirExists(t, versionDir)
		},
	)

	runIntegrationStep(t, "node:uninstall VERSION=18.0.0 — specific version directory is removed",
		func(t *testing.T) {
			t.Helper()
			run(t, yesFlag, nodeUninstallTask, "VERSION="+version18)
			tasktestutil.AssertDirNotExists(t, versionDir)
		},
	)
}

func runNvmNodeUseAndVersionSteps(t *testing.T, run nvmRunFunc) {
	t.Helper()

	runIntegrationStep(t, "node:use — LTS is activated without error", func(t *testing.T) {
		t.Helper()
		run(t, yesFlag, nodeUseTask)
	})
	runIntegrationStep(
		t,
		"node:version — active node and npm version strings are printed",
		func(t *testing.T) {
			t.Helper()

			result := run(t, nodeVersionTask)
			tasktestutil.AssertContains(t, result.Combined(), "v")
		},
	)
}

func runNvmNodeSteps(t *testing.T, run nvmRunFunc, nvmDir string) {
	t.Helper()

	runNvmNodeVersionSteps(t, run, nvmDir)
	runNvmNodeUseAndVersionSteps(t, run)
}

func runIntegrationStep(t *testing.T, name string, fn func(t *testing.T)) {
	t.Helper()
	t.Run(name, fn)

	if t.Failed() {
		t.FailNow()
	}
}

func successfulIntegrationRun(root string, env []string) nvmRunFunc {
	return func(t *testing.T, args ...string) tasktestutil.CommandResult {
		t.Helper()

		result := tasktestutil.RunTaskTimeout(
			t,
			tasktestutil.TaskRun{Root: root, Env: env, Args: args},
			10*time.Minute,
		)
		tasktestutil.AssertExitCode(t, result, exitCodeSuccess)

		return result
	}
}

// TestVersionTaskExitsSuccessfully
func TestVersionTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	assertDryRunNvmTaskExits(t, versionTask)
}

// TestLsTaskExitsSuccessfully
func TestLsTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	assertDryRunNvmTaskExits(t, lsTask)
}

// TestInstallIsIdempotentWithStubNvm
func TestInstallIsIdempotentWithStubNvm(t *testing.T) {
	t.Parallel()

	assertDryRunNvmTaskExits(t, installTask)
	assertDryRunNvmTaskExits(t, installTask)
}

func assertDryRunNvmTaskExits(t *testing.T, task string) {
	t.Helper()

	if runtime.GOOS == windowsOS {
		t.Skip(skipUnixOnlyMsg)
	}

	tasktestutil.AssertExitCode(
		t,
		tasktestutil.RunTask(
			t,
			tasktestutil.TaskRun{
				Root: tasktestutil.ModuleRoot(t),
				Env:  dryRunEnv(t),
				Args: []string{yesFlag, task},
			},
		),
		exitCodeSuccess,
	)
}

// TestInstallUndoRemovesNvmDir
func TestInstallUndoRemovesNvmDir(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip(skipUnixOnlyMsg)
	}

	root := tasktestutil.ModuleRoot(t)
	env := dryRunEnv(t)
	nvmDir := tasktestutil.EnvValue(env, nvmDirEnvKey)
	tasktestutil.AssertDirExists(t, nvmDir)
	tasktestutil.AssertExitCode(
		t,
		tasktestutil.RunTask(
			t,
			tasktestutil.TaskRun{Root: root, Env: env, Args: []string{yesFlag, installUndoTask}},
		),
		exitCodeSuccess,
	)
	tasktestutil.AssertDirNotExists(t, nvmDir)
}

// TestNodeInstallWithVersionPrintsVersionInOutput
func TestNodeInstallWithVersionPrintsVersionInOutput(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip(skipUnixOnlyMsg)
	}

	result := tasktestutil.RunTask(
		t, tasktestutil.TaskRun{Root: tasktestutil.ModuleRoot(t), Env: dryRunEnv(t), Args: []string{
			yesFlag,
			nodeInstallTask,
			version18Arg,
		}})

	tasktestutil.AssertExitCode(t, result, exitCodeSuccess)
	tasktestutil.AssertContains(t, result.Combined(), version18)
}

// TestNodeInstallDefaultVersionUsesLts
func TestNodeInstallDefaultVersionUsesLts(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip(skipUnixOnlyMsg)
	}

	result := tasktestutil.RunTask(
		t, tasktestutil.TaskRun{Root: tasktestutil.ModuleRoot(t), Env: dryRunEnv(t), Args: []string{
			yesFlag,
			nodeInstallTask,
		}})

	tasktestutil.AssertExitCode(t, result, exitCodeSuccess)
	tasktestutil.AssertContains(t, result.Combined(), ltsFlag)
}

// TestNodeInstallSkipsAlreadyInstalledVersion
func TestNodeInstallSkipsAlreadyInstalledVersion(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip(skipUnixOnlyMsg)
	}

	root := tasktestutil.ModuleRoot(t)
	env := dryRunEnv(t)
	stubInstalledVersionDir(t, env)

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: root,
			Env:  env,
			Args: []string{yesFlag, nodeInstallTask, version18Arg},
		},
	)
	tasktestutil.AssertExitCode(t, result, exitCodeSuccess)
	tasktestutil.AssertNotContains(t, result.Combined(), "Installing Node.js 18.0.0")
}

// stubInstalledVersionDir creates a stub installed-node-version directory
// under the isolated NVM_DIR referenced by env.
func stubInstalledVersionDir(t *testing.T, env []string) {
	t.Helper()

	nvmDir := tasktestutil.EnvValue(env, nvmDirEnvKey)
	versionDir := filepath.Join(nvmDir, versionsDirName, nodeDirName, versionV18DirName)

	err := os.MkdirAll(versionDir, dirPerm0700)
	if err != nil {
		t.Fatalf(stubVersionDirErr, err)
	}
}

// TestNodeUninstallSkipsWhenVersionNotInstalled
func TestNodeUninstallSkipsWhenVersionNotInstalled(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip(skipUnixOnlyMsg)
	}

	result := tasktestutil.RunTask(
		t, tasktestutil.TaskRun{Root: tasktestutil.ModuleRoot(t), Env: dryRunEnv(t), Args: []string{
			yesFlag, nodeUninstallTask, version18Arg,
		}})

	tasktestutil.AssertExitCode(t, result, exitCodeSuccess)
	tasktestutil.AssertNotContains(t, result.Combined(), "Uninstalling Node.js 18.0.0")
}

// TestNodeUninstallWithInstalledVersionPrintsVersionInOutput
func TestNodeUninstallWithInstalledVersionPrintsVersionInOutput(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip(skipUnixOnlyMsg)
	}

	root := tasktestutil.ModuleRoot(t)
	env := dryRunEnv(t)
	stubInstalledVersionDir(t, env)

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: root,
			Env:  env,
			Args: []string{yesFlag, nodeUninstallTask, version18Arg},
		},
	)
	tasktestutil.AssertExitCode(t, result, exitCodeSuccess)
	tasktestutil.AssertContains(t, result.Combined(), version18)
}

// dryRunEnv returns an isolated environment with a stub nvm.sh so that nvm
// preconditions resolve without a real nvm installation.
func dryRunEnv(t *testing.T) []string {
	t.Helper()

	env := isolatedEnv(t)
	nvmDir := tasktestutil.EnvValue(env, nvmDirEnvKey)

	writeStubNvmSh(t, nvmDir)

	return env
}

func writeStubNvmSh(t *testing.T, nvmDir string) {
	t.Helper()

	err := os.MkdirAll(nvmDir, dirPerm0700)
	if err != nil {
		t.Fatalf("failed to create fake NVM dir: %v", err)
	}

	stub := "nvm() { case \"$1\" in --version) echo 0.40.1 ;; version|current) echo stub ;; *) return 0 ;; esac; }\n"

	err = os.WriteFile(filepath.Join(nvmDir, nvmShFileName), []byte(stub), filePerm0600)
	if err != nil {
		t.Fatalf("failed to create fake nvm.sh: %v", err)
	}
}
