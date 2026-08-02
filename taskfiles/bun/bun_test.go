// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package bun_test

import (
	"os"
	"path/filepath"
	"regexp"
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
	scriptRef struct {
		root     string
		taskName string
		command  string
	}

	taskScriptsRef struct {
		root     string
		taskName string
		task     tasktestutil.TaskNode
	}

	dangerousCheck struct {
		taskName string
		command  string
		patterns []*regexp.Regexp
	}

	publicTaskCheck struct {
		spec *tasktestutil.PublicTaskSpec
		task tasktestutil.TaskNode
	}

	installerFlowArgs struct {
		root   string
		bunBin string
		env    []string
	}

	bunIntegration struct {
		root   string
		home   string
		bunBin string
		env    []string
	}

	bunStepFunc func(name string, fn func(t *testing.T))
	bunRunFunc  func(t *testing.T, args ...string) tasktestutil.CommandResult
)

const (
	constBunTestPrettier = "prettier"
	constBunTestListAll  = "--list-all"
	constBunTestWindows  = "windows"
	flagList             = "--list"
	flagSort             = "--sort"
	sortAlphanumeric     = "alphanumeric"
	packagesArgKey       = "PACKAGES"
	installTask          = "install"
	installUndoTask      = "install:undo"
	upgradeTask          = "upgrade"
	upgradeCanaryTask    = "upgrade:canary"
	upgradeStableTask    = "upgrade:stable"
	versionTask          = "version"
	taskfileLabel        = "Taskfile"
	jsonFlag             = "--json"
	descField            = "desc"
	outputField          = "output"
	skipUnixOnlyMsg      = "stub bun tests target Unix-like systems"
	yesFlag              = "--yes"
	homeEnvKey           = "HOME"
	bunDirName           = ".bun"
	binDirName           = "bin"
	bunBinName           = "bun"
	pathEnvKey           = "PATH"
	exitSuccess          = 0
	bunDirMode           = 0o700
	bunFileMode          = 0o600
	bunExecMode          = 0o500
	minDescLen           = 12
	minSummaryLen        = 25
	emptyString          = ""
)

func expectedPublicTasksPackageOps() []tasktestutil.PublicTaskSpec {
	spec := tasktestutil.NewPublicTaskSpec

	return []tasktestutil.PublicTaskSpec{
		spec("add", tasktestutil.WithArgs(map[string]string{packagesArgKey: constBunTestPrettier}),
			tasktestutil.WithDryRunArgs(), tasktestutil.WithGroupOutput()),
		spec("exec", tasktestutil.WithArgs(map[string]string{"BINARY": constBunTestPrettier}),
			tasktestutil.WithDryRunArgs(), tasktestutil.WithGroupOutput()),
		spec(
			"remove",
			tasktestutil.WithArgs(map[string]string{packagesArgKey: constBunTestPrettier}),
			tasktestutil.WithDryRunArgs(),
			tasktestutil.WithGroupOutput(),
		),
	}
}

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
	}
}

func upgradeTaskSpec(name string) tasktestutil.PublicTaskSpec {
	return tasktestutil.NewPublicTaskSpec(
		name,
		tasktestutil.WithDryRunArgs(),
		tasktestutil.WithGroupOutput(),
		tasktestutil.WithSummary(),
	)
}

func expectedPublicTasksUpgradeVariants() []tasktestutil.PublicTaskSpec {
	return []tasktestutil.PublicTaskSpec{
		upgradeTaskSpec(upgradeTask),
		upgradeTaskSpec(upgradeCanaryTask),
		upgradeTaskSpec(upgradeStableTask),
	}
}

func expectedPublicTasksUpgrade() []tasktestutil.PublicTaskSpec {
	spec := tasktestutil.NewPublicTaskSpec

	return append(
		expectedPublicTasksUpgradeVariants(),
		spec(versionTask, tasktestutil.WithDryRunArgs(), tasktestutil.WithSummary()),
	)
}

func expectedPublicTasksLifecycle() []tasktestutil.PublicTaskSpec {
	return append(expectedPublicTasksInstall(), expectedPublicTasksUpgrade()...)
}

func expectedPublicTasks() []tasktestutil.PublicTaskSpec {
	return append(expectedPublicTasksPackageOps(), expectedPublicTasksLifecycle()...)
}

// TestTaskBinaryIsAvailable
func TestTaskBinaryIsAvailable(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)
	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{Root: root, Env: nil, Args: []string{"--version"}},
	)
	tasktestutil.AssertExitCode(t, result, exitSuccess)
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

	assertTaskfileVersionValid(t, root)
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

func assertTaskfileVersionValid(t *testing.T, root *yaml.Node) {
	t.Helper()

	version := tasktestutil.ScalarField(root, versionTask)

	if version != "3" && !strings.HasPrefix(version, "3.") {
		t.Fatalf("Taskfile version must be 3 or 3.x, got %q", version)
	}
}

func assertTaskfileHasTasks(t *testing.T, root *yaml.Node) {
	t.Helper()

	tasks := tasktestutil.MappingField(root, "tasks")

	if tasks == nil || len(tasks.Content) == exitSuccess {
		t.Fatal("Taskfile must contain non-empty tasks map")
	}
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

func taskCliListArgVariants() [][]string {
	return [][]string{
		{flagList},
		{constBunTestListAll},
		{constBunTestListAll, flagSort, sortAlphanumeric},
		{constBunTestListAll, jsonFlag},
	}
}

func assertTaskCliListSucceeds(t *testing.T, root string, args []string) {
	t.Helper()

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{Root: root, Env: tasktestutil.IsolatedEnv(t), Args: args},
	)
	tasktestutil.AssertExitCode(t, result, exitSuccess)
	tasktestutil.AssertNotContains(
		t,
		strings.ToLower(result.Combined()),
		"taskfile does not exist",
	)
	tasktestutil.AssertNotContains(t, strings.ToLower(result.Combined()), "unknown")
}

// TestTaskListAllJsonIsValid
func TestTaskListAllJsonIsValid(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)
	result := tasktestutil.RunTask(
		t, tasktestutil.TaskRun{Root: root, Env: tasktestutil.IsolatedEnv(t), Args: []string{
			constBunTestListAll,
			jsonFlag,
		}})

	tasktestutil.AssertExitCode(t, result, exitSuccess)

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
		task := taskfile.Tasks[name]
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assertTaskIsPublicOrInternal(t, name, task)
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
func TestPublicTasksHaveMetadata(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)

	for i := range expectedPublicTasks() {
		spec := expectedPublicTasks()[i]
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()

			task := tasktestutil.MustTask(t, taskfile, spec.Name)
			check := &publicTaskCheck{spec: &spec, task: task}

			assertTaskUsesFullMappingSyntax(t, check)
			assertTaskDescValid(t, check)
			assertTaskSummaryValid(t, check)
		})
	}
}

func assertTaskUsesFullMappingSyntax(t *testing.T, check *publicTaskCheck) {
	t.Helper()

	if check.task.Node.Kind != yaml.MappingNode {
		t.Fatalf("public task %q must use full mapping syntax, not short syntax", check.spec.Name)
	}
}

func assertTaskDescValid(t *testing.T, check *publicTaskCheck) {
	t.Helper()

	desc := check.task.StringField(descField)

	if strings.TrimSpace(desc) == emptyString {
		t.Fatalf("public task %q is missing desc", check.spec.Name)
	}

	if len(strings.TrimSpace(desc)) < minDescLen {
		t.Fatalf("public task %q desc is too short: %q", check.spec.Name, desc)
	}

	tasktestutil.AssertNoPlaceholderText(t, check.spec.Name, desc)
}

func assertTaskSummaryValid(t *testing.T, check *publicTaskCheck) {
	t.Helper()

	summary := check.task.StringField("summary")

	if check.spec.RequiresSummary && strings.TrimSpace(summary) == emptyString {
		t.Fatalf("public task %q is missing summary", check.spec.Name)
	}

	if check.spec.RequiresSummary && len(strings.TrimSpace(summary)) < minSummaryLen {
		t.Fatalf("public task %q summary is too short:\n%s", check.spec.Name, summary)
	}

	tasktestutil.AssertNoPlaceholderText(t, check.spec.Name, summary)
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

			assertTaskHasCommands(t, &taskfile, spec.Name)
		})
	}
}

func assertTaskHasCommands(t *testing.T, taskfile *tasktestutil.LoadedTaskfile, name string) {
	t.Helper()

	task := tasktestutil.MustTask(t, taskfile, name)
	missingCmdsAndDeps := tasktestutil.IsEmptyNode(task.Field("cmds")) &&
		tasktestutil.IsEmptyNode(task.Field("deps"))

	if missingCmdsAndDeps {
		t.Fatalf("public task %q must have cmds or deps", name)
	}
}

// TestTaskSummariesWork
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
	tasktestutil.AssertExitCode(t, result, exitSuccess)

	out := result.Combined()
	tasktestutil.AssertContains(t, out, spec.Name)
	tasktestutil.AssertNotContains(t, strings.ToLower(out), "task not found")
	tasktestutil.AssertNotContains(t, strings.ToLower(out), "unknown task")
	tasktestutil.AssertNotContains(t, strings.ToLower(out), "no summary")
}

// TestTaskSummariesWork validates the behavior covered by this test case.
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

		if _, ok := taskfile.Tasks[task]; !ok {
			t.Fatalf("task %q is missing", task)
		}

		if _, ok := taskfile.Tasks[undo]; !ok {
			t.Fatalf("undo task %q for %q is missing", undo, task)
		}
	}
}

// TestReferencedScriptsExist
func assertReferencedScriptsExistForTask(t *testing.T, ref *taskScriptsRef) {
	t.Helper()

	commands := tasktestutil.CollectCommandStrings(ref.task.Node)

	for i := range commands {
		command := commands[i]

		t.Run(ref.taskName, func(t *testing.T) {
			t.Parallel()

			assertReferencedScriptsExist(
				t,
				&scriptRef{root: ref.root, taskName: ref.taskName, command: command},
			)
		})
	}
}

// TestReferencedScriptsExist validates the behavior covered by this test case.
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

func assertReferencedScriptsExist(t *testing.T, ref *scriptRef) {
	t.Helper()

	for i := range tasktestutil.ReferencedLocalShellScripts(ref.command) {
		scriptPath := tasktestutil.ReferencedLocalShellScripts(ref.command)[i]
		abs := filepath.Join(ref.root, scriptPath)

		info, err := os.Stat(abs)
		if err != nil {
			t.Fatalf("task %q references missing script %q", ref.taskName, scriptPath)
		}

		if info.IsDir() {
			t.Fatalf(
				"task %q references script path but it is a directory: %q",
				ref.taskName,
				scriptPath,
			)
		}
	}
}

// TestCommandsDoNotContainDangerousPatterns
func TestCommandsDoNotContainDangerousPatterns(t *testing.T) {
	t.Parallel()

	dangerousPatterns := tasktestutil.DangerousCommandPatterns()

	taskfile := tasktestutil.LoadTaskfile(t)

	for taskName := range taskfile.Tasks {
		task := taskfile.Tasks[taskName]

		for i := range tasktestutil.CollectCommandStrings(task.Node) {
			command := tasktestutil.CollectCommandStrings(task.Node)[i]
			assertCommandHasNoDangerousPattern(t, &dangerousCheck{
				taskName: taskName,
				command:  command,
				patterns: dangerousPatterns,
			})
		}
	}
}

func assertCommandHasNoDangerousPattern(t *testing.T, check *dangerousCheck) {
	t.Helper()

	for i := range check.patterns {
		pattern := check.patterns[i]

		if pattern.MatchString(check.command) {
			t.Fatalf(
				"task %q contains dangerous command pattern %q:\n%s",
				check.taskName,
				pattern.String(),
				check.command,
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

	assertStubbedBunTaskExits(t, versionTask)
}

// TestVersionTaskPrintsBunVersion
func TestVersionTaskPrintsBunVersion(t *testing.T) {
	t.Parallel()

	result := runStubbedBunTask(t, versionTask)
	tasktestutil.AssertContains(t, result.Combined(), "1.")
}

// TestInstallIsIdempotentWithStubBun
func TestInstallIsIdempotentWithStubBun(t *testing.T) {
	t.Parallel()

	assertStubbedBunTaskExits(t, installTask)
	assertStubbedBunTaskExits(t, installTask)
}

// TestInstallSkipsWhenBunIsAlreadyPresent
func TestInstallSkipsWhenBunIsAlreadyPresent(t *testing.T) {
	t.Parallel()

	result := runStubbedBunTask(t, installTask)
	tasktestutil.AssertNotContains(t, result.Combined(), "Installing Bun")
}

// TestInstallUndoRemovesBunDir
func TestInstallUndoRemovesBunDir(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constBunTestWindows {
		t.Skip(skipUnixOnlyMsg)
	}

	root := tasktestutil.ModuleRoot(t)
	env := bunStubEnv(t)
	bunDir := filepath.Join(tasktestutil.EnvValue(env, homeEnvKey), bunDirName)
	tasktestutil.AssertDirExists(t, bunDir)
	tasktestutil.AssertExitCode(
		t,
		tasktestutil.RunTask(
			t,
			tasktestutil.TaskRun{Root: root, Env: env, Args: []string{yesFlag, installUndoTask}},
		),
		exitSuccess,
	)
	tasktestutil.AssertDirNotExists(t, bunDir)
}

// TestInstallUndoIsIdempotent
func TestInstallUndoIsIdempotent(t *testing.T) {
	t.Parallel()

	assertStubbedBunTaskExits(t, installUndoTask)
	assertStubbedBunTaskExits(t, installUndoTask)
}

// TestUpgradeExitsSuccessfully
func TestUpgradeExitsSuccessfully(t *testing.T) {
	t.Parallel()

	assertStubbedBunTaskExits(t, upgradeTask)
}

// TestUpgradeCanaryExitsSuccessfully
func TestUpgradeCanaryExitsSuccessfully(t *testing.T) {
	t.Parallel()

	assertStubbedBunTaskExits(t, upgradeCanaryTask)
}

// TestUpgradeStableExitsSuccessfully
func TestUpgradeStableExitsSuccessfully(t *testing.T) {
	t.Parallel()

	assertStubbedBunTaskExits(t, upgradeStableTask)
}

func assertStubbedBunTaskExits(t *testing.T, task string) {
	t.Helper()

	tasktestutil.AssertExitCode(t, runStubbedBunTask(t, task), exitSuccess)
}

func runStubbedBunTask(t *testing.T, task string) tasktestutil.CommandResult {
	t.Helper()

	if runtime.GOOS == constBunTestWindows {
		t.Skip(skipUnixOnlyMsg)
	}

	return tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{Root: tasktestutil.ModuleRoot(t), Env: bunStubEnv(t), Args: []string{
			yesFlag,
			task,
		}},
	)
}

// TestRealInstallerFlowOnlyWhenExplicitlyEnabled
func TestRealInstallerFlowOnlyWhenExplicitlyEnabled(t *testing.T) {
	t.Parallel()

	skipUnlessRealInstallerTestsEnabled(t)

	root := tasktestutil.ModuleRoot(t)
	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, homeEnvKey)
	bunBin := filepath.Join(home, bunDirName, binDirName, bunBinName)

	runRealInstallerFlow(t, &installerFlowArgs{root: root, env: env, bunBin: bunBin})
	assertBunDirRemoved(t, home)
}

func skipUnlessRealInstallerTestsEnabled(t *testing.T) {
	t.Helper()

	if os.Getenv("RUN_INSTALLER_TESTS") != "1" {
		t.Skip("set RUN_INSTALLER_TESTS=1 to run real install/uninstall tests")
	}

	if runtime.GOOS == constBunTestWindows {
		t.Skip("real bun installer tests are intended for Unix-like systems")
	}
}

func runRealInstallerInstallStep(t *testing.T, args *installerFlowArgs) {
	t.Helper()

	tasktestutil.AssertExitCode(
		t,
		tasktestutil.RunTaskTimeout(
			t,
			tasktestutil.TaskRun{
				Root: args.root,
				Env:  args.env,
				Args: []string{yesFlag, installTask},
			},
			10*time.Minute,
		),
		exitSuccess,
	)
	tasktestutil.AssertFileExists(t, args.bunBin)
}

func runRealInstallerVersionStep(t *testing.T, args *installerFlowArgs) {
	t.Helper()

	tasktestutil.AssertExitCode(
		t,
		tasktestutil.RunTaskTimeout(
			t,
			tasktestutil.TaskRun{Root: args.root, Env: args.env, Args: []string{versionTask}},
			10*time.Minute,
		),
		exitSuccess,
	)
}

func runRealInstallerUndoStep(t *testing.T, args *installerFlowArgs) {
	t.Helper()

	tasktestutil.AssertExitCode(
		t,
		tasktestutil.RunTaskTimeout(
			t,
			tasktestutil.TaskRun{
				Root: args.root,
				Env:  args.env,
				Args: []string{yesFlag, installUndoTask},
			},
			10*time.Minute,
		),
		exitSuccess,
	)
}

func runRealInstallerFlow(t *testing.T, args *installerFlowArgs) {
	t.Helper()

	runRealInstallerInstallStep(t, args)
	runRealInstallerVersionStep(t, args)
	runRealInstallerUndoStep(t, args)
}

func assertBunDirRemoved(t *testing.T, home string) {
	t.Helper()

	info, err := os.Stat(filepath.Join(home, bunDirName))

	if err == nil && info != nil {
		t.Fatalf("expected path to not exist: %s", filepath.Join(home, bunDirName))
	}

	if !os.IsNotExist(err) {
		t.Fatalf("expected .bun directory to be removed after install:undo: %s", home)
	}
}

func skipUnlessIntegrationTestsEnabled(t *testing.T) {
	t.Helper()

	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("set RUN_INTEGRATION_TESTS=1 to run integration tests (downloads and installs Bun)")
	}

	if runtime.GOOS == constBunTestWindows {
		t.Skip("integration tests target Unix-like systems")
	}
}

// TestAllPublicTasksIntegration
func TestAllPublicTasksIntegration(t *testing.T) {
	t.Parallel()

	skipUnlessIntegrationTestsEnabled(t)

	root := tasktestutil.ModuleRoot(t)
	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, homeEnvKey)
	bunBin := filepath.Join(home, bunDirName, binDirName, bunBinName)

	runBunIntegrationSteps(t, &bunIntegration{
		root:   root,
		env:    env,
		home:   home,
		bunBin: bunBin,
	})
}

func runBunIntegrationSteps(t *testing.T, integration *bunIntegration) {
	t.Helper()

	step, run := bunIntegrationHelpers(t, integration)

	registerBunIntegrationSteps(step, run, integration)
}

func bunIntegrationHelpers(t *testing.T, integration *bunIntegration) (bunStepFunc, bunRunFunc) {
	t.Helper()

	return bunStepRunner(t), bunTaskRunner(t, integration)
}

func bunStepRunner(t *testing.T) bunStepFunc {
	t.Helper()

	return func(name string, fn func(t *testing.T)) {
		t.Helper()
		t.Run(name, fn)

		if t.Failed() {
			t.FailNow()
		}
	}
}

func bunTaskRunner(t *testing.T, integration *bunIntegration) bunRunFunc {
	t.Helper()

	return func(t *testing.T, args ...string) tasktestutil.CommandResult {
		t.Helper()

		result := tasktestutil.RunTaskTimeout(
			t,
			tasktestutil.TaskRun{Root: integration.root, Env: integration.env, Args: args},
			10*time.Minute,
		)
		tasktestutil.AssertExitCode(t, result, exitSuccess)

		return result
	}
}

func registerBunIntegrationSteps(step bunStepFunc, run bunRunFunc, integration *bunIntegration) {
	registerBunInstallSteps(step, run, integration)
	registerBunUpgradeSteps(step, run, integration)
}

func registerBunInstallSteps(step bunStepFunc, run bunRunFunc, integration *bunIntegration) {
	step("install — bun binary is present on disk", func(t *testing.T) {
		t.Helper()
		run(t, yesFlag, installTask)
		tasktestutil.AssertFileExists(t, integration.bunBin)
	})
	step("version — bun version string is printed", func(t *testing.T) {
		t.Helper()

		result := run(t, versionTask)
		tasktestutil.AssertNotEmpty(t, result.Combined(), "version output is empty")
	})
	step("install:undo — .bun directory is removed", func(t *testing.T) {
		t.Helper()
		run(t, yesFlag, installUndoTask)
		tasktestutil.AssertDirNotExists(t, filepath.Join(integration.home, bunDirName))
	})
}

func registerBunUpgradeSteps(step bunStepFunc, run bunRunFunc, integration *bunIntegration) {
	step("upgrade — bun upgrades without error", func(t *testing.T) {
		t.Helper()
		run(t, yesFlag, upgradeTask)
		tasktestutil.AssertFileExists(t, integration.bunBin)
	})
	step("upgrade:canary — bun switches to canary without error", func(t *testing.T) {
		t.Helper()
		run(t, yesFlag, upgradeCanaryTask)
		tasktestutil.AssertFileExists(t, integration.bunBin)
	})
	step("upgrade:stable — bun switches back to stable without error", func(t *testing.T) {
		t.Helper()
		run(t, yesFlag, upgradeStableTask)
		tasktestutil.AssertFileExists(t, integration.bunBin)
	})
}

// BunStubEnv returns an isolated environment with a stub bun binary that
// satisfies precondition checks without performing real operations.

func bunStubEnv(t *testing.T) []string {
	t.Helper()

	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, homeEnvKey)
	bunBinDir := filepath.Join(home, bunDirName, binDirName)

	writeBunStubBinary(t, bunBinDir)

	path := tasktestutil.EnvValue(env, pathEnvKey)

	return tasktestutil.SetEnv(env, pathEnvKey, bunBinDir+":"+path)
}

func writeBunStubBinary(t *testing.T, bunBinDir string) {
	t.Helper()

	err := os.MkdirAll(bunBinDir, bunDirMode)
	if err != nil {
		t.Fatalf("failed to create stub bun dir: %v", err)
	}

	bunPath := filepath.Join(bunBinDir, bunBinName)

	writeBunStubScript(t, bunPath)
}

func writeBunStubScript(t *testing.T, bunPath string) {
	t.Helper()

	stub := "#!/usr/bin/env bash\n" +
		"case \"$1\" in\n" +
		"  --version) echo \"1.2.3\" ;;\n" +
		"  --revision) echo \"abc1234\" ;;\n" +
		"  upgrade) echo \"Bun is already at the latest version\" ;;\n" +
		"  *) exit 0 ;;\n" +
		"esac\n"

	err := os.WriteFile(bunPath, []byte(stub), bunFileMode)
	if err != nil {
		t.Fatalf("failed to create stub bun binary: %v", err)
	}

	err = syscall.Chmod(bunPath, bunExecMode)
	if err != nil {
		t.Fatalf("make stub bun executable: %v", err)
	}
}
