// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package npm_test

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/task-otter/store/internal/taskintegration"
	"github.com/task-otter/store/internal/tasktestutil"
	yaml "go.yaml.in/yaml/v3"
)

type (
	errCheck struct {
		noErrMsg string
		substrA  string
		substrB  string
		msgFmt   string
	}

	publicTaskCase struct {
		taskfile *tasktestutil.LoadedTaskfile
		spec     *tasktestutil.PublicTaskSpec
	}
)

const (
	constNpmTestPrettier           = "prettier"
	constNpmTestListAll            = "--list-all"
	constNpmTestWindows            = "windows"
	constNpmTestPackages           = "PACKAGES"
	constNpmTestAuditReport        = "audit:report"
	constNpmTestBuild              = "build"
	constNpmTestInstallClean       = "install:clean"
	constNpmTestClean              = "clean"
	constNpmTestDev                = "dev"
	constNpmTestInstall            = "install"
	constNpmTestOutdated           = "outdated"
	constNpmTestOutdatedStrict     = "outdated:strict"
	constNpmTestRun                = "run"
	constNpmTestVersion            = "version"
	constNpmTestTaskfile           = "Taskfile"
	constNpmTestJSON               = "--json"
	constNpmTestDesc               = "desc"
	constNpmTestOutput             = "output"
	constNpmTestCmds               = "cmds"
	constNpmTestUnixSkipMsg        = "stub npm tests target Unix-like systems"
	constNpmTestYes                = "--yes"
	constNpmTestTaskfileFlag       = "--taskfile"
	constNpmTestTaskfileYml        = "Taskfile.yml"
	constNpmTestPackageJSON        = "package.json"
	constNpmTestFileMode0600       = 0o600
	constNpmTestFileMode0700       = 0o700
	constNpmTestPath               = "PATH"
	constNpmTestExitSuccess        = 0
	constNpmTestEmpty              = ""
	constNpmTestFlagList           = "--list"
	constNpmTestFlagSort           = "--sort"
	constNpmTestSortAlpha          = "alphanumeric"
	constNpmTestMinDescLen         = 12
	constNpmTestMinSummaryLen      = 25
	constNpmTestVersionOutputEmpty = "version output is empty"
)

// TestModuleIntegration runs the shared task CLI integration suite for this module.
func TestModuleIntegration(t *testing.T) {
	t.Parallel()

	taskintegration.RunHere(t)
}

// TestTaskBinaryIsAvailable
func TestTaskBinaryIsAvailable(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)
	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{Root: root, Env: nil, Args: []string{"--version"}},
	)
	tasktestutil.AssertExitCode(t, result, constNpmTestExitSuccess)
	tasktestutil.AssertNotEmpty(t, result.Combined(), "task --version output is empty")
}

// TestTaskfileYamlIsCleanAndValid
func TestTaskfileYamlIsCleanAndValid(t *testing.T) {
	t.Parallel()

	path := tasktestutil.ModuleTaskfilePath(t)
	content := tasktestutil.ReadFile(t, path)
	tasktestutil.AssertTextFileClean(t, path, content)

	doc := parseTaskfileYamlDoc(t, content)

	tasktestutil.AssertNoDuplicateMappingKeys(t, &doc, constNpmTestTaskfile)
	tasktestutil.AssertNoYamlAliases(t, &doc, constNpmTestTaskfile)

	root := tasktestutil.DocumentRoot(t, &doc)

	assertTaskfileVersionIsV3(t, root)
	assertTaskfileHasTasks(t, root)
}

// TestTaskCliCanLoadTaskfile
func TestTaskCliCanLoadTaskfile(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)

	variants := taskCliLoadTaskfileArgs()

	for i := range variants {
		runTaskCliCanLoadTaskfileCase(t, root, variants[i])
	}
}

// TestTaskListAllJsonIsValid
func TestTaskListAllJsonIsValid(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)
	result := tasktestutil.RunTask(
		t, tasktestutil.TaskRun{Root: root, Env: tasktestutil.IsolatedEnv(t), Args: []string{
			constNpmTestListAll,
			constNpmTestJSON,
		}})

	tasktestutil.AssertExitCode(t, result, constNpmTestExitSuccess)

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

// TestReadmePublicTaskTableDoesNotDrift
func TestReadmePublicTaskTableDoesNotDrift(t *testing.T) {
	t.Parallel()

	content := tasktestutil.ReadFile(t, tasktestutil.ModuleReadmePath(t))
	expected := tasktestutil.ExpectedPublicTaskNames(expectedPublicTasks())

	actual := readmePublicTaskNames(content)

	if !slices.Equal(expected, actual) {
		t.Fatalf(
			"README public task table drift detected\n\nexpected:\n%s\n\nactual:\n%s\n\n"+
				"Keep README.md Public Tasks aligned with expectedPublicTasks.",
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

// TestPublicTasksHaveMetadata validates the behavior covered by this test case.
func TestPublicTasksHaveMetadata(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)
	specs := expectedPublicTasks()

	for i := range specs {
		runPublicTaskMetadataCase(t, &publicTaskCase{taskfile: &taskfile, spec: &specs[i]})
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

	forEachPublicTask(t, assertTaskUsesGithubGroupOutput)
}

// TestPublicTasksHaveCommands
func TestPublicTasksHaveCommands(t *testing.T) {
	t.Parallel()

	forEachPublicTask(t, assertTaskHasCommands)
}

// TestTaskSummariesWork validates the behavior covered by this test case.
func TestTaskSummariesWork(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)

	specs := expectedPublicTasks()

	for i := range specs {
		runTaskSummaryCase(t, root, &specs[i])
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

// TestRunTaskRequiresScriptVariable
func TestRunTaskRequiresScriptVariable(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip(constNpmTestUnixSkipMsg)
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmStubEnv(t),
			Args: []string{constNpmTestYes, constNpmTestRun},
		},
	)

	if result.Err == nil {
		t.Fatal("expected task run to fail without SCRIPT variable but it succeeded")
	}
}

// TestVersionTaskExitsSuccessfully
func TestVersionTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip(constNpmTestUnixSkipMsg)
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmStubEnv(t),
			Args: []string{constNpmTestYes, constNpmTestVersion},
		},
	)
	tasktestutil.AssertExitCode(t, result, constNpmTestExitSuccess)
	tasktestutil.AssertNotEmpty(t, result.Combined(), constNpmTestVersionOutputEmpty)
}

// TestInstallTaskExitsSuccessfully
func TestInstallTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	assertStubbedNpmTaskExits(t, constNpmTestInstall)
}

// TestInstallCleanTaskExitsSuccessfully
func TestInstallCleanTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	assertStubbedNpmTaskExits(t, constNpmTestInstallClean)
}

// TestBuildTaskExitsSuccessfully
func TestBuildTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	assertStubbedNpmTaskExits(t, constNpmTestBuild)
}

// TestRunTaskExitsSuccessfully
func TestRunTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	result := runStubbedNpmTask(t, constNpmTestRun, "SCRIPT=build")
	tasktestutil.AssertContains(t, result.Combined(), constNpmTestBuild)
}

// TestCleanTaskSkipsWhenNodeModulesAbsent
func TestCleanTaskSkipsWhenNodeModulesAbsent(t *testing.T) {
	t.Parallel()

	assertStubbedNpmTaskExits(t, constNpmTestClean)
}

// TestOutdatedTaskExitsSuccessfully
func TestOutdatedTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	assertStubbedNpmTaskExits(t, constNpmTestOutdated)
}

// TestOutdatedStrictTaskExitsSuccessfully
func TestOutdatedStrictTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	assertStubbedNpmTaskExits(t, constNpmTestOutdatedStrict)
}

// TestAuditReportTaskExitsSuccessfully
func TestAuditReportTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	assertStubbedNpmTaskExits(t, constNpmTestAuditReport)
}

// TestRunTaskForwardsCliArgs
func TestRunTaskForwardsCliArgs(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip(constNpmTestUnixSkipMsg)
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.ModuleRoot(t),
		npmStubEnv(t),
		constNpmTestYes,
		constNpmTestRun,
		"SCRIPT=test",
		"--",
		"--watch",
	)
	tasktestutil.AssertExitCode(t, result, constNpmTestExitSuccess)
}

// TestRunTaskCliArgsWiredInYaml
func TestRunTaskCliArgsWiredInYaml(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)
	task := tasktestutil.MustTask(t, taskfile, "_run:unix")

	cmds := task.Field(constNpmTestCmds)

	if cmds == nil {
		t.Fatal("_run:unix task has no cmds")
	}

	if !strings.Contains(tasktestutil.NodeText(cmds), "CLI_ARGS") {
		t.Fatal(
			"_run:unix cmds do not reference CLI_ARGS; extra arguments after -- will not be forwarded to npm run",
		)
	}
}

// TestDevTaskExitsSuccessfully
func TestDevTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip(constNpmTestUnixSkipMsg)
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmStubEnv(t),
			Args: []string{constNpmTestYes, constNpmTestDev},
		},
	)
	tasktestutil.AssertExitCode(t, result, constNpmTestExitSuccess)
}

// TestInstallFailsOutsideProjectRoot validates the behavior covered by this test case.
func TestInstallFailsOutsideProjectRoot(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip(constNpmTestUnixSkipMsg)
	}

	projectDir := t.TempDir()
	result := runNpmTaskFor(t, projectDir, constNpmTestInstall)

	if result.Err == nil {
		t.Fatal("expected task install to fail outside a project root but it succeeded")
	}

	if !strings.Contains(strings.ToLower(result.Combined()), constNpmTestPackageJSON) {
		t.Fatalf("expected error mentioning package.json, got:\n%s", result.Combined())
	}
}

// TestInstallCleanFailsWithoutLockfile validates the behavior covered by this test case.
func TestInstallCleanFailsWithoutLockfile(t *testing.T) {
	t.Parallel()
	skipUnlessUnixShell(t)

	projectDir := t.TempDir()

	writeProjectPackageJSON(t, projectDir)

	result := runNpmTaskFor(t, projectDir, constNpmTestInstallClean)

	assertResultErrContains(t, &result, &errCheck{
		noErrMsg: "expected task install:clean to fail without package-lock.json but it succeeded",
		substrA:  "package-lock.json",
		substrB:  "lockfile",
		msgFmt:   "expected error mentioning lockfile, got:\n%s",
	})
}

// TestRunTaskRejectsUnsafeScript validates the behavior covered by this test case.
func TestRunTaskRejectsUnsafeScript(t *testing.T) {
	t.Parallel()
	skipUnlessUnixShell(t)

	result := tasktestutil.RunTask(
		t,
		tasktestutil.ModuleRoot(t),
		npmStubEnv(t),
		constNpmTestYes,
		constNpmTestRun,
		"SCRIPT=dev; rm -rf /",
	)

	assertResultErrContains(t, &result, &errCheck{
		noErrMsg: "expected task run to reject unsafe SCRIPT but it succeeded",
		substrA:  "invalid",
		substrB:  "script",
		msgFmt:   "expected error about invalid SCRIPT characters, got:\n%s",
	})
}

// TestRealNpmFlowOnlyWhenExplicitlyEnabled validates the behavior covered by this test case.
func TestRealNpmFlowOnlyWhenExplicitlyEnabled(t *testing.T) {
	t.Parallel()
	skipUnlessRealNpmFlowEnabled(t)

	root := tasktestutil.ModuleRoot(t)
	env := tasktestutil.IsolatedEnv(t)
	result := tasktestutil.RunTaskTimeout(
		t,
		tasktestutil.TaskRun{
			Root: root,
			Env:  env,
			Args: []string{constNpmTestYes, constNpmTestVersion},
		},
		10*time.Minute,
	)
	tasktestutil.AssertExitCode(t, result, constNpmTestExitSuccess)
	tasktestutil.AssertNotEmpty(t, result.Combined(), constNpmTestVersionOutputEmpty)
}

func dryGroupSummaryOptions() []tasktestutil.PublicTaskSpecOption {
	return []tasktestutil.PublicTaskSpecOption{
		tasktestutil.WithDryRunArgs(),
		tasktestutil.WithGroupOutput(),
		tasktestutil.WithSummary(),
	}
}

func expectedPublicTasksA() []tasktestutil.PublicTaskSpec {
	spec := tasktestutil.NewPublicTaskSpec
	withArgs := tasktestutil.WithArgs
	dryGroupSummary := dryGroupSummaryOptions()

	return []tasktestutil.PublicTaskSpec{
		spec("add", withArgs(map[string]string{constNpmTestPackages: constNpmTestPrettier}),
			tasktestutil.WithDryRunArgs(), tasktestutil.WithGroupOutput()),
		spec("audit", dryGroupSummary...),
		spec("audit:fix", dryGroupSummary...),
		spec("audit:json", dryGroupSummary...),
		spec(constNpmTestAuditReport, dryGroupSummary...),
		spec(constNpmTestBuild, dryGroupSummary...),
		spec("cache:clean", dryGroupSummary...),
		spec(constNpmTestInstallClean, dryGroupSummary...),
		spec("ci:fix", dryGroupSummary...),
	}
}

func expectedPublicTasksB() []tasktestutil.PublicTaskSpec {
	spec := tasktestutil.NewPublicTaskSpec
	withArgs := tasktestutil.WithArgs
	dryGroupSummary := dryGroupSummaryOptions()

	return []tasktestutil.PublicTaskSpec{
		spec(constNpmTestClean, append(dryGroupSummary, tasktestutil.WithPrompt())...),
		spec("clean:all", append(dryGroupSummary, tasktestutil.WithPrompt())...),
		spec(constNpmTestDev, dryGroupSummary...),
		spec("exec", withArgs(map[string]string{"BINARY": constNpmTestPrettier}),
			tasktestutil.WithDryRunArgs(), tasktestutil.WithGroupOutput()),
		spec("doctor", dryGroupSummary...),
		spec(constNpmTestInstall, dryGroupSummary...),
		spec("install:undo", dryGroupSummary...),
		spec("lint", dryGroupSummary...),
	}
}

func expectedPublicTasksC1() []tasktestutil.PublicTaskSpec {
	spec := tasktestutil.NewPublicTaskSpec
	withArgs := tasktestutil.WithArgs
	dryGroupSummary := dryGroupSummaryOptions()

	return []tasktestutil.PublicTaskSpec{
		spec(constNpmTestOutdated, dryGroupSummary...),
		spec(constNpmTestOutdatedStrict, dryGroupSummary...),
		spec("remove", withArgs(map[string]string{constNpmTestPackages: constNpmTestPrettier}),
			tasktestutil.WithDryRunArgs(), tasktestutil.WithGroupOutput()),
	}
}

func expectedPublicTasksC2() []tasktestutil.PublicTaskSpec {
	spec := tasktestutil.NewPublicTaskSpec
	withArgs := tasktestutil.WithArgs
	dryGroupSummary := dryGroupSummaryOptions()

	return []tasktestutil.PublicTaskSpec{
		spec(constNpmTestRun, append(
			[]tasktestutil.PublicTaskSpecOption{
				withArgs(map[string]string{"SCRIPT": constNpmTestBuild}),
			},
			dryGroupSummary...,
		)...),
		spec("test", dryGroupSummary...),
		spec("typecheck", dryGroupSummary...),
		spec("update", dryGroupSummary...),
		spec("upgrade", dryGroupSummary...),
		spec(constNpmTestVersion, tasktestutil.WithDryRunArgs(), tasktestutil.WithSummary()),
	}
}

func expectedPublicTasksC() []tasktestutil.PublicTaskSpec {
	return append(expectedPublicTasksC1(), expectedPublicTasksC2()...)
}

func expectedPublicTasks() []tasktestutil.PublicTaskSpec {
	return slices.Concat(expectedPublicTasksA(), expectedPublicTasksB(), expectedPublicTasksC())
}

func parseTaskfileYamlDoc(t *testing.T, content string) yaml.Node {
	t.Helper()

	var doc yaml.Node

	err := yaml.Unmarshal([]byte(content), &doc)
	if err != nil {
		t.Fatalf("Taskfile YAML is invalid: %v", err)
	}

	return doc
}

func assertTaskfileVersionIsV3(t *testing.T, root *yaml.Node) {
	t.Helper()

	version := tasktestutil.ScalarField(root, constNpmTestVersion)

	if version != "3" && !strings.HasPrefix(version, "3.") {
		t.Fatalf("Taskfile version must be 3 or 3.x, got %q", version)
	}
}

func assertTaskfileHasTasks(t *testing.T, root *yaml.Node) {
	t.Helper()

	tasks := tasktestutil.MappingField(root, "tasks")

	if tasks == nil || len(tasks.Content) == constNpmTestExitSuccess {
		t.Fatal("Taskfile must contain non-empty tasks map")
	}
}

func taskCliLoadTaskfileArgs() [][]string {
	return [][]string{
		{constNpmTestFlagList},
		{constNpmTestListAll},
		{constNpmTestListAll, constNpmTestFlagSort, constNpmTestSortAlpha},
		{constNpmTestListAll, constNpmTestJSON},
	}
}

func runTaskCliCanLoadTaskfileCase(t *testing.T, root string, args []string) {
	t.Helper()
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		t.Parallel()

		result := tasktestutil.RunTask(
			t,
			tasktestutil.TaskRun{Root: root, Env: tasktestutil.IsolatedEnv(t), Args: args},
		)
		tasktestutil.AssertExitCode(t, result, constNpmTestExitSuccess)
		tasktestutil.AssertNotContains(
			t,
			strings.ToLower(result.Combined()),
			"taskfile does not exist",
		)
		tasktestutil.AssertNotContains(t, strings.ToLower(result.Combined()), "unknown")
	})
}

func assertTaskIsPublicOrInternal(t *testing.T, name string, task tasktestutil.TaskNode) {
	t.Helper()

	if strings.HasPrefix(name, "_") || task.BoolField("internal") {
		return
	}

	if task.StringField(constNpmTestDesc) == "" {
		t.Fatalf(
			"task %q is not internal and has no desc. Either add desc/summary or mark it internal: true",
			name,
		)
	}
}

func assertTaskUsesMappingSyntax(t *testing.T, task tasktestutil.TaskNode, name string) {
	t.Helper()

	if task.Node.Kind != yaml.MappingNode {
		t.Fatalf("public task %q must use full mapping syntax, not short syntax", name)
	}
}

func assertTaskDescMeetsRequirements(t *testing.T, desc, name string) {
	t.Helper()

	if strings.TrimSpace(desc) == constNpmTestEmpty {
		t.Fatalf("public task %q is missing desc", name)
	}

	if len(strings.TrimSpace(desc)) < constNpmTestMinDescLen {
		t.Fatalf("public task %q desc is too short: %q", name, desc)
	}
}

func assertTaskSummaryValid(t *testing.T, spec *tasktestutil.PublicTaskSpec, summary string) {
	t.Helper()

	if !spec.RequiresSummary {
		return
	}

	if strings.TrimSpace(summary) == constNpmTestEmpty {
		t.Fatalf("public task %q is missing summary", spec.Name)
	}

	if len(strings.TrimSpace(summary)) < constNpmTestMinSummaryLen {
		t.Fatalf("public task %q summary is too short:\n%s", spec.Name, summary)
	}
}

// TestPublicTasksHaveMetadata
func runPublicTaskMetadataCase(t *testing.T, check *publicTaskCase) {
	t.Helper()
	t.Run(check.spec.Name, func(t *testing.T) {
		t.Parallel()

		task := tasktestutil.MustTask(t, check.taskfile, check.spec.Name)
		assertTaskUsesMappingSyntax(t, task, check.spec.Name)

		desc := task.StringField(constNpmTestDesc)
		summary := task.StringField("summary")

		assertTaskDescMeetsRequirements(t, desc, check.spec.Name)
		assertTaskSummaryValid(t, check.spec, summary)

		tasktestutil.AssertNoPlaceholderText(t, check.spec.Name, desc)
		tasktestutil.AssertNoPlaceholderText(t, check.spec.Name, summary)
	})
}

func assertTaskUsesGithubGroupOutput(t *testing.T, check *publicTaskCase) {
	t.Helper()

	if !check.spec.RequiresGroupOutput {
		return
	}

	task := tasktestutil.MustTask(t, check.taskfile, check.spec.Name)

	outputNode := task.Field(constNpmTestOutput)

	if outputNode == nil {
		outputNode = check.taskfile.Root.Field(constNpmTestOutput)
	}

	tasktestutil.AssertGithubGroupOutput(t, check.spec.Name, outputNode)
}

func forEachPublicTask(t *testing.T, callback func(*testing.T, *publicTaskCase)) {
	t.Helper()

	taskfile := tasktestutil.LoadTaskfile(t)
	specs := expectedPublicTasks()

	for i := range specs {
		spec := specs[i]

		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			callback(t, &publicTaskCase{taskfile: &taskfile, spec: &spec})
		})
	}
}

func assertTaskHasCommands(t *testing.T, check *publicTaskCase) {
	t.Helper()

	task := tasktestutil.MustTask(t, check.taskfile, check.spec.Name)
	missingCmdsAndDeps := tasktestutil.IsEmptyNode(task.Field(constNpmTestCmds)) &&
		tasktestutil.IsEmptyNode(task.Field("deps"))

	if missingCmdsAndDeps {
		t.Fatalf("public task %q must have cmds or deps", check.spec.Name)
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

func runTaskSummaryCase(t *testing.T, root string, spec *tasktestutil.PublicTaskSpec) {
	t.Helper()
	t.Run(spec.Name, func(t *testing.T) {
		t.Parallel()

		if !spec.RequiresSummary {
			return
		}

		result := runTaskSummary(t, root, spec.Name)
		tasktestutil.AssertExitCode(t, result, constNpmTestExitSuccess)

		out := result.Combined()
		tasktestutil.AssertContains(t, out, spec.Name)
		tasktestutil.AssertNotContains(t, strings.ToLower(out), "task not found")
		tasktestutil.AssertNotContains(t, strings.ToLower(out), "unknown task")
		tasktestutil.AssertNotContains(t, strings.ToLower(out), "no summary")
	})
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

func assertStubbedNpmTaskExits(t *testing.T, task string) {
	t.Helper()

	tasktestutil.AssertExitCode(t, runStubbedNpmTask(t, task), constNpmTestExitSuccess)
}

func runStubbedNpmTask(t *testing.T, task string, args ...string) tasktestutil.CommandResult {
	t.Helper()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip(constNpmTestUnixSkipMsg)
	}

	taskArgs := append([]string{constNpmTestYes, task}, args...)

	return tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{Root: tasktestutil.ModuleRoot(t), Env: npmStubEnv(t), Args: taskArgs},
	)
}

func runNpmTaskFor(t *testing.T, projectDir, taskName string) tasktestutil.CommandResult {
	t.Helper()

	taskfileDir := tasktestutil.ModuleRoot(t)

	return tasktestutil.RunTask(t, tasktestutil.TaskRun{
		Root: projectDir,
		Env:  npmStubEnv(t),
		Args: []string{
			constNpmTestTaskfileFlag,
			filepath.Join(taskfileDir, constNpmTestTaskfileYml),
			constNpmTestYes,
			taskName,
		},
	})
}

func writeProjectPackageJSON(t *testing.T, projectDir string) {
	t.Helper()

	err := os.WriteFile(
		filepath.Join(projectDir, constNpmTestPackageJSON),
		[]byte(`{"name":"test","version":"1.0.0"}`),
		constNpmTestFileMode0600,
	)
	if err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}
}

func assertResultErrContains(t *testing.T, result *tasktestutil.CommandResult, check *errCheck) {
	t.Helper()

	if result.Err == nil {
		t.Fatal(check.noErrMsg)
	}

	out := strings.ToLower(result.Combined())

	if !strings.Contains(out, check.substrA) && !strings.Contains(out, check.substrB) {
		t.Fatalf(check.msgFmt, result.Combined())
	}
}

func skipUnlessUnixShell(t *testing.T) {
	t.Helper()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip(constNpmTestUnixSkipMsg)
	}
}

// TestRealNpmFlowOnlyWhenExplicitlyEnabled
func skipUnlessRealNpmFlowEnabled(t *testing.T) {
	t.Helper()

	if os.Getenv("RUN_INSTALLER_TESTS") != "1" {
		t.Skip("set RUN_INSTALLER_TESTS=1 to run real npm install/build/test tests")
	}

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("real npm flow tests target Unix-like systems")
	}
}

// NpmStubEnv returns an isolated environment with stub nodejs, node, npm, and
// nix binaries so all preconditions pass without real installations.

// readmePublicTaskNames parses the npm README and returns sorted task names
// from the Public Tasks table.
func readmePublicTaskNames(content string) []string {
	return tasktestutil.ReadmePublicTaskNames(content)
}

func makeStubBinDir(t *testing.T, home string) string {
	t.Helper()

	binDir := filepath.Join(home, ".local", "bin")

	err := os.MkdirAll(binDir, constNpmTestFileMode0700)
	if err != nil {
		t.Fatalf("failed to create stub bin dir: %v", err)
	}

	return binDir
}

func writeNpmNodeStubs(t *testing.T, binDir string) {
	t.Helper()

	tasktestutil.WriteStub(
		t,
		binDir,
		"node",
		"#!/usr/bin/env bash\ncase \"$1\" in\n  --version) echo \"v20.11.0 stub\" ;;\n  *) exit 0 ;;\nesac\n",
	)
}

func writeNpmToolchainStubs(t *testing.T, binDir string) {
	t.Helper()

	writeNpmNodeStubs(t, binDir)
	tasktestutil.WriteStub(
		t,
		binDir,
		"npm",
		"#!/usr/bin/env bash\ncase \"$1\" in\n"+
			"  --version) echo \"10.9.0 stub\" ;;\n"+
			"  *) echo \"npm $* stub\"; exit 0 ;;\nesac\n",
	)
	tasktestutil.WriteStub(
		t,
		binDir,
		"nix",
		"#!/usr/bin/env bash\necho \"nix $* stub\"\n",
	)
}

func writeStubBashrc(t *testing.T, home string) {
	t.Helper()

	bashrc := filepath.Join(home, ".bashrc")

	err := os.WriteFile(
		bashrc,
		[]byte("export PATH=\"$HOME/.local/bin:$PATH\"\n"),
		constNpmTestFileMode0600,
	)
	if err != nil {
		t.Fatalf("failed to pre-populate shell profile: %v", err)
	}
}

func npmStubEnv(t *testing.T) []string {
	t.Helper()

	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")

	binDir := makeStubBinDir(t, home)
	writeNpmToolchainStubs(t, binDir)
	writeStubBashrc(t, home)

	path := tasktestutil.EnvValue(env, constNpmTestPath)

	return tasktestutil.SetEnv(env, constNpmTestPath, binDir+":"+path)
}
