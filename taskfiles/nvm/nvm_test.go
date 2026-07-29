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
	"gopkg.in/yaml.v3"
)

const (
	nodeInstallTask   = "node:install"
	nodeVersion2400   = "24.0.0"
	nodeUninstallTask = "node:uninstall"
	versionArg        = "VERSION"
	listAllFlag       = "--list-all"
	windowsOS         = "windows"
)

func expectedPublicTasks() []tasktestutil.PublicTaskSpec {
	spec := tasktestutil.NewPublicTaskSpec
	versionArgs := map[string]string{versionArg: nodeVersion2400}

	return []tasktestutil.PublicTaskSpec{
		spec(
			"install",
			tasktestutil.WithDryRunArgs(),
			tasktestutil.WithGroupOutput(),
			tasktestutil.WithSummary(),
		),
		spec("install:undo", tasktestutil.WithDryRunArgs(), tasktestutil.WithGroupOutput(),
			tasktestutil.WithPrompt(), tasktestutil.WithSummary()),
		spec("ls", tasktestutil.WithDryRunArgs(), tasktestutil.WithSummary()),
		spec(nodeInstallTask, tasktestutil.WithArgs(versionArgs), tasktestutil.WithDryRunArgs(),
			tasktestutil.WithDryRunNoArgs(), tasktestutil.WithExpectedDefaultTokens("--lts"),
			tasktestutil.WithGroupOutput(), tasktestutil.WithSummary()),
		spec(nodeUninstallTask, tasktestutil.WithArgs(versionArgs), tasktestutil.WithDryRunArgs(),
			tasktestutil.WithGroupOutput(), tasktestutil.WithPrompt(), tasktestutil.WithSummary()),
		spec(
			"node:use",
			tasktestutil.WithArgs(versionArgs),
			tasktestutil.WithDryRunArgs(),
			tasktestutil.WithDryRunNoArgs(),
			tasktestutil.WithExpectedDefaultTokens("--lts"),
			tasktestutil.WithSummary(),
		),
		spec("node:version", tasktestutil.WithDryRunArgs(), tasktestutil.WithSummary()),
		spec("version", tasktestutil.WithDryRunArgs(), tasktestutil.WithSummary()),
	}
}

func isolatedEnv(t *testing.T) []string {
	t.Helper()
	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")

	return tasktestutil.SetEnv(env, "NVM_DIR", filepath.Join(home, ".nvm"))
}

func TestTaskBinaryIsAvailable(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)
	result := tasktestutil.RunTask(t, root, nil, "--version")
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertNotEmpty(t, result.Combined(), "task --version output is empty")
}

func TestTaskfileYamlIsCleanAndValid(t *testing.T) {
	t.Parallel()

	path := tasktestutil.ModuleTaskfilePath(t)
	content := tasktestutil.ReadFile(t, path)
	tasktestutil.AssertTextFileClean(t, path, content)

	var doc yaml.Node

	err := yaml.Unmarshal([]byte(content), &doc)
	if err != nil {
		t.Fatalf("Taskfile YAML is invalid: %v", err)
	}

	tasktestutil.AssertNoDuplicateMappingKeys(t, &doc, "Taskfile")
	tasktestutil.AssertNoYamlAliases(t, &doc, "Taskfile")

	root := tasktestutil.DocumentRoot(t, &doc)

	version := tasktestutil.ScalarField(root, "version")
	if version != "3" && !strings.HasPrefix(version, "3.") {
		t.Fatalf("Taskfile version must be 3 or 3.x, got %q", version)
	}

	tasks := tasktestutil.MappingField(root, "tasks")
	if tasks == nil || len(tasks.Content) == 0 {
		t.Fatal("Taskfile must contain non-empty tasks map")
	}
}

func TestTaskCliCanLoadTaskfile(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)

	for _, args := range [][]string{
		{"--list"},
		{listAllFlag},
		{listAllFlag, "--sort", "alphanumeric"},
		{listAllFlag, "--json"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()

			result := tasktestutil.RunTask(t, root, isolatedEnv(t), args...)
			tasktestutil.AssertExitCode(t, result, 0)
			tasktestutil.AssertNotContains(
				t,
				strings.ToLower(result.Combined()),
				"taskfile does not exist",
			)
			tasktestutil.AssertNotContains(t, strings.ToLower(result.Combined()), "unknown")
		})
	}
}

func TestTaskListAllJsonIsValid(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)
	result := tasktestutil.RunTask(t, root, isolatedEnv(t), listAllFlag, "--json")
	tasktestutil.AssertExitCode(t, result, 0)

	err := tasktestutil.ValidateJSON(result.Stdout)
	if err != nil {
		t.Fatalf("task --list-all --json produced invalid JSON:\n%s\nerror: %v", result.Stdout, err)
	}
}

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

func TestEveryTaskIsEitherPublicOrInternal(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)
	for name, task := range taskfile.Tasks {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if strings.HasPrefix(name, "_") || task.BoolField("internal") {
				return
			}

			if task.StringField("desc") == "" {
				t.Fatalf(
					"task %q is not internal and has no desc. Either add desc/summary or mark it internal: true",
					name,
				)
			}
		})
	}
}

func TestPublicTasksHaveMetadata(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)

	for _, spec := range expectedPublicTasks() {
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()

			task := tasktestutil.MustTask(t, taskfile, spec.Name)
			if task.Node.Kind != yaml.MappingNode {
				t.Fatalf("public task %q must use full mapping syntax, not short syntax", spec.Name)
			}

			desc := task.StringField("desc")
			summary := task.StringField("summary")

			if strings.TrimSpace(desc) == "" {
				t.Fatalf("public task %q is missing desc", spec.Name)
			}

			if len(strings.TrimSpace(desc)) < 12 {
				t.Fatalf("public task %q desc is too short: %q", spec.Name, desc)
			}

			if spec.RequiresSummary && strings.TrimSpace(summary) == "" {
				t.Fatalf("public task %q is missing summary", spec.Name)
			}

			if spec.RequiresSummary && len(strings.TrimSpace(summary)) < 25 {
				t.Fatalf("public task %q summary is too short:\n%s", spec.Name, summary)
			}

			tasktestutil.AssertNoPlaceholderText(t, spec.Name, desc)
			tasktestutil.AssertNoPlaceholderText(t, spec.Name, summary)
		})
	}
}

func TestDestructivePublicTasksHavePrompt(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)

	for _, spec := range expectedPublicTasks() {
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

func TestInstallTasksUseGithubGroupOutput(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)

	for _, spec := range expectedPublicTasks() {
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()

			if !spec.RequiresGroupOutput {
				return
			}

			task := tasktestutil.MustTask(t, taskfile, spec.Name)

			outputNode := task.Field("output")
			if outputNode == nil {
				outputNode = taskfile.Root.Field("output")
			}

			tasktestutil.AssertGithubGroupOutput(t, spec.Name, outputNode)
		})
	}
}

func TestPublicTasksHaveCommands(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)

	for _, spec := range expectedPublicTasks() {
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()

			task := tasktestutil.MustTask(t, taskfile, spec.Name)
			if tasktestutil.IsEmptyNode(task.Field("cmds")) &&
				tasktestutil.IsEmptyNode(task.Field("deps")) {
				t.Fatalf("public task %q must have cmds or deps", spec.Name)
			}
		})
	}
}

func TestTaskSummariesWork(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)

	for _, spec := range expectedPublicTasks() {
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()

			if !spec.RequiresSummary {
				return
			}

			result := tasktestutil.RunTask(t, root, isolatedEnv(t), "--summary", spec.Name)
			tasktestutil.AssertExitCode(t, result, 0)
			out := result.Combined()
			tasktestutil.AssertContains(t, out, spec.Name)
			tasktestutil.AssertNotContains(t, strings.ToLower(out), "task not found")
			tasktestutil.AssertNotContains(t, strings.ToLower(out), "unknown task")
			tasktestutil.AssertNotContains(t, strings.ToLower(out), "no summary")
		})
	}
}

func TestUndoPairsExist(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)
	for task, undo := range map[string]string{"install": "install:undo"} {
		if _, ok := taskfile.Tasks[task]; !ok {
			t.Fatalf("task %q is missing", task)
		}

		if _, ok := taskfile.Tasks[undo]; !ok {
			t.Fatalf("undo task %q for %q is missing", undo, task)
		}
	}

	for _, pair := range []struct{ task, undoTarget string }{
		{nodeInstallTask, nodeUninstallTask},
		{nodeUninstallTask, nodeInstallTask},
	} {
		if _, ok := taskfile.Tasks[pair.task]; !ok {
			t.Fatalf("task %q is missing", pair.task)
		}

		if _, ok := taskfile.Tasks[pair.undoTarget]; !ok {
			t.Fatalf("undo target %q is missing for task %q", pair.undoTarget, pair.task)
		}
	}
}

func TestReferencedScriptsExist(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)

	taskfile := tasktestutil.LoadTaskfile(t)
	for taskName, task := range taskfile.Tasks {
		for _, command := range tasktestutil.CollectCommandStrings(task.Node) {
			t.Run(taskName, func(t *testing.T) {
				t.Parallel()

				for _, scriptPath := range tasktestutil.ReferencedLocalShellScripts(command) {
					abs := filepath.Join(root, scriptPath)

					info, err := os.Stat(abs)
					if err != nil {
						t.Fatalf("task %q references missing script %q", taskName, scriptPath)
					}

					if info.IsDir() {
						t.Fatalf(
							"task %q references script path but it is a directory: %q",
							taskName,
							scriptPath,
						)
					}
				}
			})
		}
	}
}

func TestCommandsDoNotContainDangerousPatterns(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)
	for taskName, task := range taskfile.Tasks {
		for _, command := range tasktestutil.CollectCommandStrings(task.Node) {
			for _, pattern := range tasktestutil.DangerousCommandPatterns() {
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

func TestNoPlaceholderTextInTaskfile(t *testing.T) {
	t.Parallel()

	content := tasktestutil.ReadFile(t, tasktestutil.ModuleTaskfilePath(t))

	upper := strings.ToUpper(content)
	for _, p := range []string{"TODO", "FIXME", "CHANGEME", "REPLACE_ME", "YOUR VALUE HERE", "LOREM IPSUM"} {
		if strings.Contains(upper, p) {
			t.Fatalf("Taskfile contains placeholder text: %s", p)
		}
	}
}

func TestRealInstallerFlowOnlyWhenExplicitlyEnabled(t *testing.T) {
	t.Parallel()

	if os.Getenv("RUN_INSTALLER_TESTS") != "1" {
		t.Skip("set RUN_INSTALLER_TESTS=1 to run real install/uninstall tests")
	}

	if runtime.GOOS == windowsOS {
		t.Skip("real nvm shell installer tests are intended for Unix-like systems")
	}

	root := tasktestutil.ModuleRoot(t)
	env := isolatedEnv(t)

	nvmDir := tasktestutil.EnvValue(env, "NVM_DIR")
	if nvmDir == "" {
		t.Fatal("NVM_DIR was not set in isolated environment")
	}

	tasktestutil.AssertExitCode(
		t,
		tasktestutil.RunTaskTimeout(t, root, env, 10*time.Minute, "--yes", "install"),
		0,
	)

	_, err := os.Stat(nvmDir)
	if err != nil {
		t.Fatalf("expected NVM_DIR to exist after install: %s\nerror: %v", nvmDir, err)
	}

	tasktestutil.AssertExitCode(
		t,
		tasktestutil.RunTaskTimeout(t, root, env, 10*time.Minute, "version"),
		0,
	)
	tasktestutil.AssertExitCode(
		t,
		tasktestutil.RunTaskTimeout(t, root, env, 10*time.Minute, "--yes", "install:undo"),
		0,
	)

	_, err = os.Stat(nvmDir)
	if !os.IsNotExist(err) {
		t.Fatalf("expected NVM_DIR to be removed after install:undo: %s", nvmDir)
	}
}

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
	nvmDir := tasktestutil.EnvValue(env, "NVM_DIR")

	runNvmIntegrationSteps(t, root, env, nvmDir)
}

func runNvmIntegrationSteps(t *testing.T, root string, env []string, nvmDir string) {
	t.Helper()

	run := successfulIntegrationRun(root, env)

	runNvmInstallSteps(t, run, nvmDir)
	runNvmNodeSteps(t, run, nvmDir)
	runIntegrationStep(t, "install:undo — NVM directory is removed", func(t *testing.T) {
		t.Helper()
		run(t, "--yes", "install:undo")
		tasktestutil.AssertDirNotExists(t, nvmDir)
	})
}

func runNvmInstallSteps(
	t *testing.T,
	run func(t *testing.T, args ...string) tasktestutil.CommandResult,
	nvmDir string,
) {
	t.Helper()

	runIntegrationStep(t, "install — nvm.sh is present on disk", func(t *testing.T) {
		t.Helper()
		run(t, "--yes", "install")
		tasktestutil.AssertFileExists(t, filepath.Join(nvmDir, "nvm.sh"))
	})
	runIntegrationStep(t, "version — nvm version string is printed", func(t *testing.T) {
		t.Helper()
		result := run(t, "version")
		tasktestutil.AssertNotEmpty(t, result.Combined(), "version output is empty")
	})
	runIntegrationStep(
		t,
		"node:install — default LTS version directory is created",
		func(t *testing.T) {
			t.Helper()
			run(t, "--yes", nodeInstallTask)
			tasktestutil.AssertDirHasEntries(t, filepath.Join(nvmDir, "versions", "node"))
		},
	)
	runIntegrationStep(t, "ls — installed versions appear in output", func(t *testing.T) {
		t.Helper()
		result := run(t, "ls")
		tasktestutil.AssertNotEmpty(t, result.Combined(), "ls output is empty")
	})
}

func runNvmNodeSteps(
	t *testing.T,
	run func(t *testing.T, args ...string) tasktestutil.CommandResult,
	nvmDir string,
) {
	t.Helper()

	const secondary = "18.0.0"

	runIntegrationStep(
		t,
		"node:install VERSION=18.0.0 — specific version directory is created",
		func(t *testing.T) {
			t.Helper()
			run(t, "--yes", nodeInstallTask, "VERSION="+secondary)
			tasktestutil.AssertDirExists(
				t,
				filepath.Join(nvmDir, "versions", "node", "v"+secondary),
			)
		},
	)
	runIntegrationStep(
		t,
		"node:uninstall VERSION=18.0.0 — specific version directory is removed",
		func(t *testing.T) {
			t.Helper()
			run(t, "--yes", nodeUninstallTask, "VERSION="+secondary)
			tasktestutil.AssertDirNotExists(
				t,
				filepath.Join(nvmDir, "versions", "node", "v"+secondary),
			)
		},
	)
	runIntegrationStep(t, "node:use — LTS is activated without error", func(t *testing.T) {
		t.Helper()
		run(t, "--yes", "node:use")
	})
	runIntegrationStep(
		t,
		"node:version — active node and npm version strings are printed",
		func(t *testing.T) {
			t.Helper()
			result := run(t, "node:version")
			tasktestutil.AssertContains(t, result.Combined(), "v")
		},
	)
}

func runIntegrationStep(t *testing.T, name string, fn func(t *testing.T)) {
	t.Helper()
	t.Run(name, fn)

	if t.Failed() {
		t.FailNow()
	}
}

func successfulIntegrationRun(
	root string,
	env []string,
) func(t *testing.T, args ...string) tasktestutil.CommandResult {
	return func(t *testing.T, args ...string) tasktestutil.CommandResult {
		t.Helper()
		result := tasktestutil.RunTaskTimeout(t, root, env, 10*time.Minute, args...)
		tasktestutil.AssertExitCode(t, result, 0)

		return result
	}
}

func TestVersionTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip("stub nvm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), dryRunEnv(t), "--yes", "version")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestLsTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip("stub nvm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), dryRunEnv(t), "--yes", "ls")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestInstallIsIdempotentWithStubNvm(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip("stub nvm tests target Unix-like systems")
	}

	root := tasktestutil.ModuleRoot(t)
	env := dryRunEnv(t)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install"), 0)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install"), 0)
}

func TestInstallUndoRemovesNvmDir(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip("stub nvm tests target Unix-like systems")
	}

	root := tasktestutil.ModuleRoot(t)
	env := dryRunEnv(t)
	nvmDir := tasktestutil.EnvValue(env, "NVM_DIR")
	tasktestutil.AssertDirExists(t, nvmDir)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install:undo"), 0)
	tasktestutil.AssertDirNotExists(t, nvmDir)
}

func TestNodeInstallWithVersionPrintsVersionInOutput(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip("stub nvm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.ModuleRoot(t),
		dryRunEnv(t),
		"--yes",
		nodeInstallTask,
		"VERSION=18.0.0",
	)
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertContains(t, result.Combined(), "18.0.0")
}

func TestNodeInstallDefaultVersionUsesLts(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip("stub nvm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.ModuleRoot(t),
		dryRunEnv(t),
		"--yes",
		nodeInstallTask,
	)
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertContains(t, result.Combined(), "--lts")
}

func TestNodeInstallSkipsAlreadyInstalledVersion(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip("stub nvm tests target Unix-like systems")
	}

	root := tasktestutil.ModuleRoot(t)
	env := dryRunEnv(t)
	nvmDir := tasktestutil.EnvValue(env, "NVM_DIR")

	versionDir := filepath.Join(nvmDir, "versions", "node", "v18.0.0")

	err := os.MkdirAll(versionDir, 0o755)
	if err != nil {
		t.Fatalf("failed to create stub version dir: %v", err)
	}

	result := tasktestutil.RunTask(t, root, env, "--yes", nodeInstallTask, "VERSION=18.0.0")
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertNotContains(t, result.Combined(), "Installing Node.js 18.0.0")
}

func TestNodeUninstallSkipsWhenVersionNotInstalled(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip("stub nvm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t, tasktestutil.ModuleRoot(t), dryRunEnv(t), "--yes", nodeUninstallTask, "VERSION=18.0.0",
	)
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertNotContains(t, result.Combined(), "Uninstalling Node.js 18.0.0")
}

func TestNodeUninstallWithInstalledVersionPrintsVersionInOutput(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == windowsOS {
		t.Skip("stub nvm tests target Unix-like systems")
	}

	root := tasktestutil.ModuleRoot(t)
	env := dryRunEnv(t)
	nvmDir := tasktestutil.EnvValue(env, "NVM_DIR")

	versionDir := filepath.Join(nvmDir, "versions", "node", "v18.0.0")

	err := os.MkdirAll(versionDir, 0o755)
	if err != nil {
		t.Fatalf("failed to create stub version dir: %v", err)
	}

	result := tasktestutil.RunTask(t, root, env, "--yes", nodeUninstallTask, "VERSION=18.0.0")
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertContains(t, result.Combined(), "18.0.0")
}

// dryRunEnv returns an isolated environment with a stub nvm.sh so that nvm
// preconditions resolve without a real nvm installation.
func dryRunEnv(t *testing.T) []string {
	t.Helper()

	env := isolatedEnv(t)
	nvmDir := tasktestutil.EnvValue(env, "NVM_DIR")

	err := os.MkdirAll(nvmDir, 0o755)
	if err != nil {
		t.Fatalf("failed to create fake NVM dir: %v", err)
	}

	stub := "nvm() { case \"$1\" in --version) echo 0.40.1 ;; version|current) echo stub ;; *) return 0 ;; esac; }\n"

	err = os.WriteFile(filepath.Join(nvmDir, "nvm.sh"), []byte(stub), 0o644)
	if err != nil {
		t.Fatalf("failed to create fake nvm.sh: %v", err)
	}

	return env
}
