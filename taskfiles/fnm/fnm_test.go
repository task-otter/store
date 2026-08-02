// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package fnm_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"
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
		command  string
	}

	realInstallerRun struct {
		root string
		env  []string
		args []string
	}

	realInstallerFlow struct {
		root   string
		fnmBin string
		env    []string
	}

	fnmIntegration struct {
		root    string
		fnmBin  string
		fnmRoot string
		env     []string
	}

	taskScriptsRef struct {
		root     string
		taskName string
		task     tasktestutil.TaskNode
	}

	fnmInstallSteps struct {
		run     func(t *testing.T, args ...string) tasktestutil.CommandResult
		fnmBin  string
		fnmRoot string
	}

	fnmRunFunc func(t *testing.T, args ...string) tasktestutil.CommandResult

	publicTaskMetadataCheck struct {
		taskfile *tasktestutil.LoadedTaskfile
		spec     *tasktestutil.PublicTaskSpec
	}
)

const (
	nodeInstallTask     = "node:install"
	nodeVersion2400     = "24.0.0"
	nodeUninstallTask   = "node:uninstall"
	versionArg          = "VERSION"
	listAllFlag         = "--list-all"
	listFlag            = "--list"
	sortFlag            = "--sort"
	sortAlphanumeric    = "alphanumeric"
	fnmBinaryRemovedMsg = "expected fnm binary to be removed: %s"
	windowsOS           = "windows"
	installTask         = "install"
	installUndoTask     = "install:undo"
	lsTask              = "ls"
	ltsFlag             = "--lts"
	nodeUseTask         = "node:use"
	nodeVersionTask     = "node:version"
	shellSetupTask      = "shell:setup"
	versionTask         = "version"
	taskfileLabel       = "Taskfile"
	jsonFlag            = "--json"
	descField           = "desc"
	outputField         = "output"
	taskMissingFmt      = "task %q is missing"
	skipUnixOnlyMsg     = "stub fnm tests target Unix-like systems"
	yesFlag             = "--yes"
	homeEnvKey          = "HOME"
	localDirName        = ".local"
	binDirName          = "bin"
	fnmBinName          = "fnm"
	version18Arg        = "VERSION=18.0.0"
	version18           = "18.0.0"
	fnmEnvToken         = "fnm env"
	shareDirName        = "share"
	nodeVersionsDir     = "node-versions"
	bashrcFileName      = ".bashrc"
	filePerm0600        = 0o600
	pathEnvKey          = "PATH"

	exitCodeSuccess              = 0
	maxAllowedProfileOccurrences = 1
	stubBinDirPerm               = 0o700
	stubFnmBinaryPerm            = 0o500

	emptyString          = ""
	minTaskDescLength    = 12
	minTaskSummaryLength = 25
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
		spec(
			shellSetupTask,
			tasktestutil.WithDryRunArgs(),
			tasktestutil.WithGroupOutput(),
			tasktestutil.WithSummary(),
		),
		spec(versionTask, tasktestutil.WithDryRunArgs(), tasktestutil.WithSummary()),
	}
}

func expectedPublicTasks() []tasktestutil.PublicTaskSpec {
	tasks := expectedPublicTasksInstall()

	tasks = append(tasks, expectedPublicTasksNode()...)

	return append(tasks, expectedPublicTasksMisc()...)
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

	doc := parseTaskfileYamlDoc(t, content)

	tasktestutil.AssertNoDuplicateMappingKeys(t, doc, taskfileLabel)
	tasktestutil.AssertNoYamlAliases(t, doc, taskfileLabel)

	root := tasktestutil.DocumentRoot(t, doc)

	assertTaskfileVersionIsValid(t, root)
	assertTaskfileHasTasks(t, root)
}

func parseTaskfileYamlDoc(t *testing.T, content string) *yaml.Node {
	t.Helper()

	var doc yaml.Node

	err := yaml.Unmarshal([]byte(content), &doc)
	if err != nil {
		t.Fatalf("Taskfile YAML is invalid: %v", err)
	}

	return &doc
}

func assertTaskfileVersionIsValid(t *testing.T, root *yaml.Node) {
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
		tasktestutil.TaskRun{Root: root, Env: tasktestutil.IsolatedEnv(t), Args: args},
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
			Env:  tasktestutil.IsolatedEnv(t),
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

			assertTaskIsPublicOrInternal(t, name, taskfile.Tasks[name])
		})
	}
}

func assertTaskIsPublicOrInternal(t *testing.T, name string, task tasktestutil.TaskNode) {
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

	assertTaskUsesFullMappingSyntax(t, check.spec.Name, task)

	desc := task.StringField(descField)
	summary := task.StringField("summary")

	assertTaskDescIsValid(t, check.spec.Name, desc)
	assertTaskSummaryIsValid(t, check.spec, summary)

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

func assertTaskUsesFullMappingSyntax(t *testing.T, name string, task tasktestutil.TaskNode) {
	t.Helper()

	if task.Node.Kind != yaml.MappingNode {
		t.Fatalf("public task %q must use full mapping syntax, not short syntax", name)
	}
}

func assertTaskDescIsValid(t *testing.T, name, desc string) {
	t.Helper()

	if strings.TrimSpace(desc) == emptyString {
		t.Fatalf("public task %q is missing desc", name)
	}

	if len(strings.TrimSpace(desc)) < minTaskDescLength {
		t.Fatalf("public task %q desc is too short: %q", name, desc)
	}
}

func assertTaskSummaryIsValid(t *testing.T, spec *tasktestutil.PublicTaskSpec, summary string) {
	t.Helper()

	if !spec.RequiresSummary {
		return
	}

	if strings.TrimSpace(summary) == emptyString {
		t.Fatalf("public task %q is missing summary", spec.Name)
	}

	if len(strings.TrimSpace(summary)) < minTaskSummaryLength {
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

			assertTaskUsesGithubGroupOutput(
				t,
				&publicTaskMetadataCheck{taskfile: &taskfile, spec: &spec},
			)
		})
	}
}

func assertTaskUsesGithubGroupOutput(t *testing.T, check *publicTaskMetadataCheck) {
	t.Helper()

	if !check.spec.RequiresGroupOutput {
		return
	}

	task := tasktestutil.MustTask(t, check.taskfile, check.spec.Name)

	outputNode := task.Field(outputField)

	if outputNode == nil {
		outputNode = check.taskfile.Root.Field(outputField)
	}

	tasktestutil.AssertGithubGroupOutput(t, check.spec.Name, outputNode)
}

// TestPublicTasksHaveCommands
func TestPublicTasksHaveCommands(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)

	for i := range expectedPublicTasks() {
		spec := expectedPublicTasks()[i]
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()

			assertTaskHasCommandsOrDeps(t, &taskfile, spec.Name)
		})
	}
}

func assertTaskHasCommandsOrDeps(t *testing.T, taskfile *tasktestutil.LoadedTaskfile, name string) {
	t.Helper()

	task := tasktestutil.MustTask(t, taskfile, name)
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
		tasktestutil.TaskRun{Root: root, Env: tasktestutil.IsolatedEnv(t), Args: []string{
			"--summary",
			name,
		}},
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

	for task := range map[string]string{installTask: installUndoTask} {
		undo := map[string]string{installTask: installUndoTask}[task]
		assertUndoPairExists(t, &taskfile, &undoPair{task: task, undoTarget: undo})
	}

	for i := range []undoPair{
		{nodeInstallTask, nodeUninstallTask},
		{nodeUninstallTask, nodeInstallTask},
	} {
		pair := []undoPair{
			{nodeInstallTask, nodeUninstallTask},
			{nodeUninstallTask, nodeInstallTask},
		}[i]
		assertUndoPairExists(t, &taskfile, &pair)
	}
}

func assertUndoPairExists(t *testing.T, taskfile *tasktestutil.LoadedTaskfile, pair *undoPair) {
	t.Helper()

	if _, ok := taskfile.Tasks[pair.task]; !ok {
		t.Fatalf(taskMissingFmt, pair.task)
	}

	if _, ok := taskfile.Tasks[pair.undoTarget]; !ok {
		t.Fatalf("undo task %q for %q is missing", pair.undoTarget, pair.task)
	}
}

func assertReferencedScriptExistsForCommandRun(t *testing.T, check *scriptCheck) {
	t.Helper()

	t.Run(check.taskName, func(t *testing.T) {
		t.Parallel()

		assertReferencedScriptsExistForCommand(t, check)
	})
}

func assertReferencedScriptsExistForTask(t *testing.T, ref *taskScriptsRef) {
	t.Helper()

	commands := tasktestutil.CollectCommandStrings(ref.task.Node)

	for i := range commands {
		assertReferencedScriptExistsForCommandRun(t, &scriptCheck{
			root:     ref.root,
			taskName: ref.taskName,
			command:  commands[i],
		})
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

func assertReferencedScriptsExistForCommand(t *testing.T, check *scriptCheck) {
	t.Helper()

	for i := range tasktestutil.ReferencedLocalShellScripts(check.command) {
		scriptPath := tasktestutil.ReferencedLocalShellScripts(check.command)[i]
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

// TestVersionTaskExitsSuccessfully
func TestVersionTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	assertStubbedFnmTaskExits(t, versionTask)
}

// TestLsTaskExitsSuccessfully
func TestLsTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	assertStubbedFnmTaskExits(t, lsTask)
}

// TestInstallIsIdempotentWithStubFnm
func TestInstallIsIdempotentWithStubFnm(t *testing.T) {
	t.Parallel()

	assertStubbedFnmTaskExits(t, installTask)
	assertStubbedFnmTaskExits(t, installTask)
}

// TestInstallUndoRemovesFnmBinary
func TestInstallUndoRemovesFnmBinary(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip(skipUnixOnlyMsg)
	}

	root := tasktestutil.ModuleRoot(t)
	env := fnmStubEnv(t)
	stubBin := stubFnmBinPath(env)
	tasktestutil.AssertFileExists(t, stubBin)
	tasktestutil.AssertExitCode(
		t,
		tasktestutil.RunTask(
			t,
			tasktestutil.TaskRun{Root: root, Env: env, Args: []string{yesFlag, installUndoTask}},
		),
		exitCodeSuccess,
	)

	assertFileRemoved(t, stubBin, "expected fnm binary at %s to be removed after install:undo")
}

func stubFnmBinPath(env []string) string {
	return filepath.Join(
		tasktestutil.EnvValue(env, homeEnvKey),
		localDirName,
		binDirName,
		fnmBinName,
	)
}

func assertFileRemoved(t *testing.T, path, msgFmt string) {
	t.Helper()

	info, err := os.Stat(path)

	if err == nil && info != nil {
		t.Fatalf("expected path to not exist: %s", path)
	}

	if !os.IsNotExist(err) {
		t.Fatalf(msgFmt, path)
	}
}

// TestNodeInstallWithVersionPrintsVersionInOutput
func TestNodeInstallWithVersionPrintsVersionInOutput(t *testing.T) {
	t.Parallel()

	result := runStubbedFnmTask(t, nodeInstallTask, version18Arg)
	tasktestutil.AssertContains(t, result.Combined(), version18)
}

// TestNodeInstallDefaultVersionUsesLts
func TestNodeInstallDefaultVersionUsesLts(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip(skipUnixOnlyMsg)
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{Root: tasktestutil.ModuleRoot(t), Env: fnmStubEnv(t), Args: []string{
			yesFlag,
			nodeInstallTask,
		}},
	)

	tasktestutil.AssertExitCode(t, result, exitCodeSuccess)
	tasktestutil.AssertContains(t, result.Combined(), "--lts")
}

// TestNodeUninstallWithInstalledVersionPrintsVersionInOutput
func TestNodeUninstallWithInstalledVersionPrintsVersionInOutput(t *testing.T) {
	t.Parallel()

	result := runStubbedFnmTask(t, nodeUninstallTask, version18Arg)
	tasktestutil.AssertContains(t, result.Combined(), version18)
}

func assertStubbedFnmTaskExits(t *testing.T, task string) {
	t.Helper()

	tasktestutil.AssertExitCode(t, runStubbedFnmTask(t, task), exitCodeSuccess)
}

func runStubbedFnmTask(t *testing.T, task string, args ...string) tasktestutil.CommandResult {
	t.Helper()

	if runtime.GOOS == windowsOS {
		t.Skip(skipUnixOnlyMsg)
	}

	taskArgs := append([]string{yesFlag, task}, args...)

	return tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{Root: tasktestutil.ModuleRoot(t), Env: fnmStubEnv(t), Args: taskArgs},
	)
}

// TestShellSetupAddsActivationToProfile
func TestShellSetupAddsActivationToProfile(t *testing.T) {
	t.Parallel()

	home := runFreshProfileTask(t, shellSetupTask)

	if !profileContains(home, fnmEnvToken) {
		t.Fatal("expected at least one shell profile to contain 'fnm env' after task shell:setup")
	}
}

// TestShellSetupIsIdempotent
func TestShellSetupIsIdempotent(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip(skipUnixOnlyMsg)
	}

	root := tasktestutil.ModuleRoot(t)
	env := fnmFreshProfileEnv(t)
	home := tasktestutil.EnvValue(env, homeEnvKey)
	runShellSetup(t, root, env)
	runShellSetup(t, root, env)

	assertProfilesHaveAtMostOneFnmEnvToken(t, home)
}

func runShellSetup(t *testing.T, root string, env []string) {
	t.Helper()

	tasktestutil.AssertExitCode(
		t,
		tasktestutil.RunTask(
			t,
			tasktestutil.TaskRun{Root: root, Env: env, Args: []string{yesFlag, shellSetupTask}},
		),
		exitCodeSuccess,
	)
}

func assertProfileHasAtMostOneFnmEnvToken(t *testing.T, profilePath string) {
	t.Helper()

	content, err := readProfile(profilePath)
	if err != nil {
		return
	}

	if count := strings.Count(
		string(content),
		fnmEnvToken,
	); count > maxAllowedProfileOccurrences {
		t.Fatalf(
			"profile %s contains fnm env %d times after two shell:setup runs; want at most 1",
			profilePath,
			count,
		)
	}
}

func assertProfilesHaveAtMostOneFnmEnvToken(t *testing.T, home string) {
	t.Helper()

	paths := shellProfilePaths(home)

	for i := range paths {
		assertProfileHasAtMostOneFnmEnvToken(t, paths[i])
	}
}

// TestInstallAlsoConfiguresShellActivation
func TestInstallAlsoConfiguresShellActivation(t *testing.T) {
	t.Parallel()

	home := runFreshProfileTask(t, installTask)

	if !profileContains(home, fnmEnvToken) {
		t.Fatal(
			"expected task install to configure shell activation but no profile contains 'fnm env'",
		)
	}
}

func runFreshProfileTask(t *testing.T, task string) string {
	t.Helper()

	if runtime.GOOS == windowsOS {
		t.Skip(skipUnixOnlyMsg)
	}

	root := tasktestutil.ModuleRoot(t)
	env := fnmFreshProfileEnv(t)
	home := tasktestutil.EnvValue(env, homeEnvKey)
	tasktestutil.AssertExitCode(
		t,
		tasktestutil.RunTask(
			t,
			tasktestutil.TaskRun{Root: root, Env: env, Args: []string{yesFlag, task}},
		),
		exitCodeSuccess,
	)

	return home
}

func runRealInstallerFlow(t *testing.T, flow *realInstallerFlow) {
	t.Helper()

	runRealInstallerTaskTimeout(
		t,
		&realInstallerRun{root: flow.root, env: flow.env, args: []string{yesFlag, installTask}},
	)
	tasktestutil.AssertFileExists(t, flow.fnmBin)
	runRealInstallerTaskTimeout(
		t,
		&realInstallerRun{root: flow.root, env: flow.env, args: []string{versionTask}},
	)
	runRealInstallerTaskTimeout(
		t,
		&realInstallerRun{root: flow.root, env: flow.env, args: []string{yesFlag, installUndoTask}},
	)
}

// TestRealInstallerFlowOnlyWhenExplicitlyEnabled
func TestRealInstallerFlowOnlyWhenExplicitlyEnabled(t *testing.T) {
	t.Parallel()

	skipUnlessRealInstallerTestsEnabled(t)

	root, env, fnmBin := setUpRealInstallerEnv(t)

	runRealInstallerFlow(t, &realInstallerFlow{root: root, env: env, fnmBin: fnmBin})

	assertFileRemoved(t, fnmBin, "expected fnm binary to be removed after install:undo: %s")
}

func skipUnlessRealInstallerTestsEnabled(t *testing.T) {
	t.Helper()

	if os.Getenv("RUN_INSTALLER_TESTS") != "1" {
		t.Skip("set RUN_INSTALLER_TESTS=1 to run real install/uninstall tests")
	}

	if runtime.GOOS == windowsOS {
		t.Skip("real fnm shell installer tests are intended for Unix-like systems")
	}
}

func setUpRealInstallerEnv(t *testing.T) (root string, env []string, fnmBin string) {
	t.Helper()

	root = tasktestutil.ModuleRoot(t)
	env = tasktestutil.IsolatedEnv(t)

	home := tasktestutil.EnvValue(env, homeEnvKey)

	fnmBin = filepath.Join(home, localDirName, binDirName, fnmBinName)

	return root, env, fnmBin
}

func runRealInstallerTaskTimeout(t *testing.T, run *realInstallerRun) {
	t.Helper()

	tasktestutil.AssertExitCode(
		t,
		tasktestutil.RunTaskTimeout(
			t,
			tasktestutil.TaskRun{Root: run.root, Env: run.env, Args: run.args},
			10*time.Minute,
		),
		exitCodeSuccess,
	)
}

// TestAllPublicTasksIntegration
func TestAllPublicTasksIntegration(t *testing.T) {
	t.Parallel()

	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip(
			"set RUN_INTEGRATION_TESTS=1 to run integration tests (downloads and installs fnm and Node.js)",
		)
	}

	if runtime.GOOS == windowsOS {
		t.Skip("integration tests target Unix-like systems")
	}

	runFnmIntegrationSteps(t, newFnmIntegration(t))
}

func newFnmIntegration(t *testing.T) *fnmIntegration {
	t.Helper()

	root := tasktestutil.ModuleRoot(t)
	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, homeEnvKey)

	return &fnmIntegration{
		root:    root,
		env:     env,
		fnmBin:  filepath.Join(home, localDirName, binDirName, fnmBinName),
		fnmRoot: filepath.Join(home, localDirName, shareDirName, fnmBinName),
	}
}

func runFnmUninstallStep(t *testing.T, run fnmRunFunc, fnmBin string) {
	t.Helper()

	runIntegrationStep(t, "install:undo — fnm binary is removed", func(t *testing.T) {
		t.Helper()
		run(t, yesFlag, installUndoTask)

		info, err := os.Stat(fnmBin)

		if err == nil && info != nil {
			t.Fatalf(fnmBinaryRemovedMsg, fnmBin)
		}

		if !os.IsNotExist(err) {
			t.Fatalf(fnmBinaryRemovedMsg, fnmBin)
		}
	})
}

func runFnmIntegrationSteps(t *testing.T, integration *fnmIntegration) {
	t.Helper()

	run := successfulIntegrationRun(integration.root, integration.env)

	runFnmInstallSteps(t, &fnmInstallSteps{
		run:     run,
		fnmBin:  integration.fnmBin,
		fnmRoot: integration.fnmRoot,
	})
	runFnmNodeSteps(t, run, integration.fnmRoot)
	runFnmUninstallStep(t, run, integration.fnmBin)
}

func runFnmBinaryInstallSteps(t *testing.T, steps *fnmInstallSteps) {
	t.Helper()

	runIntegrationStep(t, "install — fnm binary is present on disk", func(t *testing.T) {
		t.Helper()
		steps.run(t, yesFlag, installTask)
		tasktestutil.AssertFileExists(t, steps.fnmBin)
	})
	runIntegrationStep(t, "version — fnm version string is printed", func(t *testing.T) {
		t.Helper()

		result := steps.run(t, versionTask)
		tasktestutil.AssertNotEmpty(t, result.Combined(), "version output is empty")
	})
}

func runFnmInstallSteps(t *testing.T, steps *fnmInstallSteps) {
	t.Helper()

	runFnmBinaryInstallSteps(t, steps)
	runIntegrationStep(
		t,
		"node:install — default LTS version directory is created",
		func(t *testing.T) {
			t.Helper()
			steps.run(t, yesFlag, nodeInstallTask)
			tasktestutil.AssertDirHasEntries(t, filepath.Join(steps.fnmRoot, nodeVersionsDir))
		},
	)
	runIntegrationStep(t, "ls — installed versions appear in output", func(t *testing.T) {
		t.Helper()

		result := steps.run(t, lsTask)
		tasktestutil.AssertNotEmpty(t, result.Combined(), "ls output is empty")
	})
}

func runFnmNodeVersionSteps(t *testing.T, run fnmRunFunc, fnmRoot string) {
	t.Helper()

	versionDir := filepath.Join(fnmRoot, nodeVersionsDir, "v"+version18)

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

func runFnmNodeUseAndVersionSteps(t *testing.T, run fnmRunFunc) {
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

func runFnmNodeSteps(t *testing.T, run fnmRunFunc, fnmRoot string) {
	t.Helper()

	runFnmNodeVersionSteps(t, run, fnmRoot)
	runFnmNodeUseAndVersionSteps(t, run)
}

func runIntegrationStep(t *testing.T, name string, fn func(t *testing.T)) {
	t.Helper()
	t.Run(name, fn)

	if t.Failed() {
		t.FailNow()
	}
}

func successfulIntegrationRun(root string, env []string) fnmRunFunc {
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

// FnmStubEnv returns an isolated environment with a stub fnm binary.
// .bashrc is pre-populated so the shell:setup status check exits 0 and
// shell:setup is skipped in tests that don't specifically test it.

// fnmFreshProfileEnv returns the same stub env as fnmStubEnv but with an
// empty .bashrc so shell:setup actually runs. Use only in shell:setup tests.
func fnmFreshProfileEnv(t *testing.T) []string {
	t.Helper()

	env := fnmStubEnv(t)

	home := tasktestutil.EnvValue(env, homeEnvKey)

	err := os.WriteFile(filepath.Join(home, bashrcFileName), []byte(""), filePerm0600)
	if err != nil {
		t.Fatalf("failed to clear shell profile: %v", err)
	}

	return env
}

func shellProfilePaths(home string) []string {
	return []string{
		filepath.Join(home, bashrcFileName),
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".profile"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".zprofile"),
		filepath.Join(home, ".config", "fish", "config.fish"),
	}
}

func readProfile(profilePath string) ([]byte, error) {
	content, err := fs.ReadFile(os.DirFS(filepath.Dir(profilePath)), filepath.Base(profilePath))
	if err != nil {
		return nil, fmt.Errorf("read profile %s: %w", profilePath, err)
	}

	return content, nil
}

func profileContains(home, token string) bool {
	for i := range shellProfilePaths(home) {
		profilePath := shellProfilePaths(home)[i]

		content, err := readProfile(profilePath)
		if err != nil {
			continue
		}

		if strings.Contains(string(content), token) {
			return true
		}
	}

	return false
}

func fnmStubEnv(t *testing.T) []string {
	t.Helper()

	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, homeEnvKey)

	binDir := filepath.Join(home, localDirName, binDirName)

	writeStubFnmBinary(t, home, binDir)
	writeStubBashrc(t, home)

	path := tasktestutil.EnvValue(env, pathEnvKey)

	return tasktestutil.SetEnv(env, pathEnvKey, binDir+":"+path)
}

func writeStubFnmBinary(t *testing.T, home, binDir string) {
	t.Helper()

	err := os.MkdirAll(binDir, stubBinDirPerm)
	if err != nil {
		t.Fatalf("failed to create stub bin dir: %v", err)
	}

	writeStubFnmBinaryFile(t, home, binDir)
}

func writeStubFnmBinaryFile(t *testing.T, home, binDir string) {
	t.Helper()

	fnmPath := filepath.Join(binDir, fnmBinName)

	err := os.WriteFile(fnmPath, []byte(stubFnmBinaryScript(home)), filePerm0600)
	if err != nil {
		t.Fatalf("failed to create stub fnm binary: %v", err)
	}

	err = syscall.Chmod(fnmPath, stubFnmBinaryPerm)
	if err != nil {
		t.Fatalf("make stub fnm executable: %v", err)
	}
}

func stubFnmBinaryScript(home string) string {
	fnmRoot := filepath.Join(home, localDirName, shareDirName, fnmBinName)

	return "#!/usr/bin/env bash\n" +
		"case \"$1\" in\n" +
		"  --version) echo \"fnm 1.37.1 stub\" ;;\n" +
		"  root) echo \"" + fnmRoot + "\" ;;\n" +
		"  list|ls) echo \"* v20.0.0 default\" ;;\n" +
		"  current) echo \"v20.0.0\" ;;\n" +
		"  env) echo \"# fnm stub env\" ;;\n" +
		"  *) exit 0 ;;\n" +
		"esac\n"
}

func writeStubBashrc(t *testing.T, home string) {
	t.Helper()

	bashrc := filepath.Join(home, bashrcFileName)

	bashrcContent := "# fnm (Fast Node Manager)\n" +
		"export PATH=\"$HOME/.local/share/fnm:$HOME/.local/bin:$PATH\"\n" +
		"eval \"$(fnm env --use-on-cd --shell bash)\"\n"

	err := os.WriteFile(bashrc, []byte(bashrcContent), filePerm0600)
	if err != nil {
		t.Fatalf("failed to pre-populate shell profile: %v", err)
	}
}
