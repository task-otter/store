// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package npmnvm_test

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

const (
	constNpmTestPrettier = "prettier"
	constNpmTestListAll  = "--list-all"
	constNpmTestWindows  = "windows"
)

func expectedPublicTasks() []tasktestutil.PublicTaskSpec {
	spec := tasktestutil.NewPublicTaskSpec
	withArgs := tasktestutil.WithArgs
	dryGroupSummary := []tasktestutil.PublicTaskSpecOption{
		tasktestutil.WithDryRunArgs(),
		tasktestutil.WithGroupOutput(),
		tasktestutil.WithSummary(),
	}

	return []tasktestutil.PublicTaskSpec{
		spec("add", withArgs(map[string]string{"PACKAGES": constNpmTestPrettier}),
			tasktestutil.WithDryRunArgs(), tasktestutil.WithGroupOutput()),
		spec("audit", dryGroupSummary...),
		spec("audit:fix", dryGroupSummary...),
		spec("audit:json", dryGroupSummary...),
		spec("audit:report", dryGroupSummary...),
		spec("build", dryGroupSummary...),
		spec("cache:clean", dryGroupSummary...),
		spec("ci", dryGroupSummary...),
		spec("clean", append(dryGroupSummary, tasktestutil.WithPrompt())...),
		spec("clean:all", append(dryGroupSummary, tasktestutil.WithPrompt())...),
		spec("dev", dryGroupSummary...),
		spec("exec", withArgs(map[string]string{"BINARY": constNpmTestPrettier}),
			tasktestutil.WithDryRunArgs(), tasktestutil.WithGroupOutput()),
		spec("doctor", dryGroupSummary...),
		spec("format", dryGroupSummary...),
		spec("install", dryGroupSummary...),
		spec("install:undo", dryGroupSummary...),
		spec("lint", dryGroupSummary...),
		spec("manager:pin", append(
			[]tasktestutil.PublicTaskSpecOption{
				withArgs(map[string]string{"PACKAGE_MANAGER_VERSION": "latest"}),
			},
			dryGroupSummary...,
		)...),
		spec("manager:setup", dryGroupSummary...),
		spec("node:setup", dryGroupSummary...),
		spec("outdated", dryGroupSummary...),
		spec("outdated:strict", dryGroupSummary...),
		spec("remove", withArgs(map[string]string{"PACKAGES": constNpmTestPrettier}),
			tasktestutil.WithDryRunArgs(), tasktestutil.WithGroupOutput()),
		spec("run", append(
			[]tasktestutil.PublicTaskSpecOption{withArgs(map[string]string{"SCRIPT": "build"})},
			dryGroupSummary...,
		)...),
		spec("test", dryGroupSummary...),
		spec("typecheck", dryGroupSummary...),
		spec("update", dryGroupSummary...),
		spec("upgrade", dryGroupSummary...),
		spec("version", tasktestutil.WithDryRunArgs(), tasktestutil.WithSummary()),
	}
}

func TestTaskBinaryIsAvailable(t *testing.T) {
	t.Parallel()

	root := tasktestutil.ModuleRoot(t)
	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{Root: root, Env: nil, Args: []string{"--version"}},
	)
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
		{constNpmTestListAll},
		{constNpmTestListAll, "--sort", "alphanumeric"},
		{constNpmTestListAll, "--json"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()

			result := tasktestutil.RunTask(
				t,
				tasktestutil.TaskRun{Root: root, Env: tasktestutil.IsolatedEnv(t), Args: args},
			)
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
	result := tasktestutil.RunTask(
		t, tasktestutil.TaskRun{Root: root, Env: tasktestutil.IsolatedEnv(t), Args: []string{
			constNpmTestListAll,
			"--json",
		}})

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

			result := tasktestutil.RunTask(
				t,
				tasktestutil.TaskRun{Root: root, Env: tasktestutil.IsolatedEnv(t), Args: []string{
					"--summary",
					spec.Name,
				}},
			)

			tasktestutil.AssertExitCode(t, result, 0)

			out := result.Combined()
			tasktestutil.AssertContains(t, out, spec.Name)
			tasktestutil.AssertNotContains(t, strings.ToLower(out), "task not found")
			tasktestutil.AssertNotContains(t, strings.ToLower(out), "unknown task")
			tasktestutil.AssertNotContains(t, strings.ToLower(out), "no summary")
		})
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

	for _, p := range []string{"TODO", "FIXME", "CHANGEME", "Copyright", "YOUR VALUE HERE", "LOREM IPSUM"} {
		if strings.Contains(upper, p) {
			t.Fatalf("Taskfile contains placeholder text: %s", p)
		}
	}
}

func TestRunTaskRequiresScriptVariable(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmNvmStubEnv(t),
			Args: []string{"--yes", "run"},
		},
	)

	if result.Err == nil {
		t.Fatal("expected task run to fail without SCRIPT variable but it succeeded")
	}
}

func TestVersionTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmNvmStubEnv(t),
			Args: []string{
				"--yes",
				"version",
			},
		},
	)

	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertContains(t, result.Combined(), "stub")
}

func TestInstallTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmNvmStubEnv(t),
			Args: []string{
				"--yes",
				"install",
			},
		},
	)

	tasktestutil.AssertExitCode(t, result, 0)
}

func TestCiTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmNvmStubEnv(t),
			Args: []string{"--yes", "ci"},
		},
	)
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestBuildTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmNvmStubEnv(t),
			Args: []string{
				"--yes",
				"build",
			},
		},
	)

	tasktestutil.AssertExitCode(t, result, 0)
}

func TestRunTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmNvmStubEnv(t),
			Args: []string{
				"--yes",
				"run",
				"SCRIPT=build",
			},
		},
	)

	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertContains(t, result.Combined(), "build")
}

func TestCleanTaskSkipsWhenNodeModulesAbsent(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmNvmStubEnv(t),
			Args: []string{
				"--yes",
				"clean",
			},
		},
	)

	tasktestutil.AssertExitCode(t, result, 0)
}

func TestOutdatedTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmNvmStubEnv(t),
			Args: []string{
				"--yes",
				"outdated",
			},
		},
	)

	tasktestutil.AssertExitCode(t, result, 0)
}

func TestOutdatedStrictTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmNvmStubEnv(t),
			Args: []string{
				"--yes",
				"outdated:strict",
			},
		},
	)

	tasktestutil.AssertExitCode(t, result, 0)
}

func TestAuditReportTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmNvmStubEnv(t),
			Args: []string{
				"--yes",
				"audit:report",
			},
		},
	)

	tasktestutil.AssertExitCode(t, result, 0)
}

func TestRunTaskForwardsCliArgs(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.ModuleRoot(t),
		npmNvmStubEnv(t),
		"--yes",
		"run",
		"SCRIPT=test",
		"--",
		"--watch",
	)
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestRunTaskCliArgsWiredInYaml(t *testing.T) {
	t.Parallel()

	taskfile := tasktestutil.LoadTaskfile(t)
	task := tasktestutil.MustTask(t, taskfile, "_run:unix")

	cmds := task.Field("cmds")

	if cmds == nil {
		t.Fatal("_run:unix task has no cmds")
	}

	if !strings.Contains(tasktestutil.NodeText(cmds), "CLI_ARGS") {
		t.Fatal(
			"_run:unix cmds do not reference CLI_ARGS; extra arguments after -- will not be forwarded to npm run",
		)
	}
}

func TestDevTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmNvmStubEnv(t),
			Args: []string{"--yes", "dev"},
		},
	)
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestInstallFailsOutsideProjectRoot(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	taskfileDir := tasktestutil.ModuleRoot(t)
	projectDir := t.TempDir()

	result := tasktestutil.RunTask(t, tasktestutil.TaskRun{
		Root: projectDir,
		Env:  npmNvmStubEnv(t),
		Args: []string{
			"--taskfile",
			filepath.Join(taskfileDir, "Taskfile.yml"),
			"--yes",
			"install",
		},
	})

	if result.Err == nil {
		t.Fatal("expected task install to fail outside a project root but it succeeded")
	}

	if !strings.Contains(strings.ToLower(result.Combined()), "package.json") {
		t.Fatalf("expected error mentioning package.json, got:\n%s", result.Combined())
	}
}

func TestCiFailsWithoutLockfile(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	taskfileDir := tasktestutil.ModuleRoot(t)

	projectDir := t.TempDir()

	err := os.WriteFile(
		filepath.Join(projectDir, "package.json"),
		[]byte(`{"name":"test","version":"1.0.0"}`),
		0o600,
	)
	if err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	result := tasktestutil.RunTask(t, tasktestutil.TaskRun{
		Root: projectDir,
		Env:  npmNvmStubEnv(t),
		Args: []string{"--taskfile", filepath.Join(taskfileDir, "Taskfile.yml"), "--yes", "ci"},
	})

	if result.Err == nil {
		t.Fatal("expected task ci to fail without package-lock.json but it succeeded")
	}

	out := strings.ToLower(result.Combined())

	if !strings.Contains(out, "package-lock.json") && !strings.Contains(out, "lockfile") {
		t.Fatalf("expected error mentioning lockfile, got:\n%s", result.Combined())
	}
}

func TestRunTaskRejectsUnsafeScript(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("stub npm tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{
			Root: tasktestutil.ModuleRoot(t),
			Env:  npmNvmStubEnv(t),
			Args: []string{
				"--yes",
				"run",
				"SCRIPT=dev; rm -rf /",
			},
		},
	)

	if result.Err == nil {
		t.Fatal("expected task run to reject unsafe SCRIPT but it succeeded")
	}

	out := strings.ToLower(result.Combined())

	if !strings.Contains(out, "invalid") && !strings.Contains(out, "script") {
		t.Fatalf("expected error about invalid SCRIPT characters, got:\n%s", result.Combined())
	}
}

func TestRealNpmFlowOnlyWhenExplicitlyEnabled(t *testing.T) {
	t.Parallel()

	if os.Getenv("RUN_INSTALLER_TESTS") != "1" {
		t.Skip("set RUN_INSTALLER_TESTS=1 to run real npm install/build/test tests")
	}

	if runtime.GOOS == constNpmTestWindows {
		t.Skip("real npm flow tests target Unix-like systems")
	}

	root := tasktestutil.ModuleRoot(t)
	env := tasktestutil.IsolatedEnv(t)
	result := tasktestutil.RunTaskTimeout(
		t,
		tasktestutil.TaskRun{Root: root, Env: env, Args: []string{"--yes", "version"}},
		10*time.Minute,
	)
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertNotEmpty(t, result.Combined(), "version output is empty")
}

// NpmNvmStubEnv returns an isolated environment with stub fnm, node, npm, and
// corepack binaries so all preconditions pass without real installations.

// readmePublicTaskNames parses the npm README and returns sorted task names
// from the Public Tasks table.
func readmePublicTaskNames(content string) []string {
	return tasktestutil.ReadmePublicTaskNames(content)
}

func npmNvmStubEnv(t *testing.T) []string {
	t.Helper()

	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")

	binDir := filepath.Join(home, ".local", "bin")

	err := os.MkdirAll(binDir, 0o700)
	if err != nil {
		t.Fatalf("failed to create stub bin dir: %v", err)
	}

	nvmDir := filepath.Join(home, ".nvm")

	err = os.MkdirAll(nvmDir, 0o700)
	if err != nil {
		t.Fatalf("failed to create nvm dir: %v", err)
	}

	err = os.WriteFile(filepath.Join(nvmDir, "nvm.sh"), []byte("# nvm stub\n"), 0o600)
	if err != nil {
		t.Fatalf("failed to create nvm.sh stub: %v", err)
	}

	tasktestutil.WriteStub(t, tasktestutil.Stub{
		Dir:  binDir,
		Name: "nvm",
		Body: "#!/usr/bin/env bash\ncase \"$1\" in\n" +
			"  --version) echo \"nvm 0.40.1 stub\" ;;\n" +
			"  use) echo \"Using Node.js stub\" ;;\n" +
			"  *) exit 0 ;;\nesac\n",
	})
	tasktestutil.WriteStub(
		t,
		binDir,
		"node",
		"#!/usr/bin/env bash\ncase \"$1\" in\n  --version) echo \"v20.11.0 stub\" ;;\n  *) exit 0 ;;\nesac\n",
	)
	tasktestutil.WriteStub(t, tasktestutil.Stub{
		Dir:  binDir,
		Name: "npm",
		Body: "#!/usr/bin/env bash\ncase \"$1\" in\n" +
			"  --version) echo \"10.9.0 stub\" ;;\n" +
			"  *) echo \"npm $* stub\"; exit 0 ;;\nesac\n",
	})
	tasktestutil.WriteStub(
		t,
		binDir,
		"corepack",
		"#!/usr/bin/env bash\ncase \"$1\" in --version) echo \"0.34.0\" ;; *) echo \"corepack $* stub\" ;; esac\n",
	)

	bashrc := filepath.Join(home, ".bashrc")

	err = os.WriteFile(bashrc, []byte("export PATH=\"$HOME/.local/bin:$PATH\"\n"), 0o600)
	if err != nil {
		t.Fatalf("failed to pre-populate shell profile: %v", err)
	}

	path := tasktestutil.EnvValue(env, "PATH")

	env = tasktestutil.SetEnv(env, "PATH", binDir+":"+path)
	env = tasktestutil.SetEnv(env, "NVM_DIR", nvmDir)

	return env
}
