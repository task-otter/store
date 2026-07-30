// REPLACE_ME 2026
// SPDX-License-Identifier: Apache-2.0

package bun_test

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/task-otter/store/internal/tasktestutil"
	yaml "gopkg.in/yaml.v3"
)

const (
	constBunTestPrettier = "prettier"
	constBunTestListAll  = "--list-all"
	constBunTestWindows  = "windows"
)

func expectedPublicTasks() []tasktestutil.PublicTaskSpec {
	spec := tasktestutil.NewPublicTaskSpec

	return []tasktestutil.PublicTaskSpec{
		spec("add", tasktestutil.WithArgs(map[string]string{"PACKAGES": constBunTestPrettier}),
			tasktestutil.WithDryRunArgs(), tasktestutil.WithGroupOutput()),
		spec("exec", tasktestutil.WithArgs(map[string]string{"BINARY": constBunTestPrettier}),
			tasktestutil.WithDryRunArgs(), tasktestutil.WithGroupOutput()),
		spec("remove", tasktestutil.WithArgs(map[string]string{"PACKAGES": constBunTestPrettier}),
			tasktestutil.WithDryRunArgs(), tasktestutil.WithGroupOutput()),
		spec(
			"install",
			tasktestutil.WithDryRunArgs(),
			tasktestutil.WithGroupOutput(),
			tasktestutil.WithSummary(),
		),
		spec("install:undo", tasktestutil.WithDryRunArgs(), tasktestutil.WithGroupOutput(),
			tasktestutil.WithPrompt(), tasktestutil.WithSummary()),
		spec(
			"upgrade",
			tasktestutil.WithDryRunArgs(),
			tasktestutil.WithGroupOutput(),
			tasktestutil.WithSummary(),
		),
		spec(
			"upgrade:canary",
			tasktestutil.WithDryRunArgs(),
			tasktestutil.WithGroupOutput(),
			tasktestutil.WithSummary(),
		),
		spec(
			"upgrade:stable",
			tasktestutil.WithDryRunArgs(),
			tasktestutil.WithGroupOutput(),
			tasktestutil.WithSummary(),
		),
		spec("version", tasktestutil.WithDryRunArgs(), tasktestutil.WithSummary()),
	}
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
		{constBunTestListAll},
		{constBunTestListAll, "--sort", "alphanumeric"},
		{constBunTestListAll, "--json"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()

			result := tasktestutil.RunTask(t, root, tasktestutil.IsolatedEnv(t), args...)
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
		t,
		root,
		tasktestutil.IsolatedEnv(t),
		constBunTestListAll,
		"--json",
	)
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

			result := tasktestutil.RunTask(
				t,
				root,
				tasktestutil.IsolatedEnv(t),
				"--summary",
				spec.Name,
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

	dangerousPatterns := tasktestutil.DangerousCommandPatterns()

	taskfile := tasktestutil.LoadTaskfile(t)

	for taskName, task := range taskfile.Tasks {
		for _, command := range tasktestutil.CollectCommandStrings(task.Node) {
			for _, pattern := range dangerousPatterns {
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

func TestVersionTaskExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constBunTestWindows {
		t.Skip("stub bun tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), bunStubEnv(t), "--yes", "version")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestVersionTaskPrintsBunVersion(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constBunTestWindows {
		t.Skip("stub bun tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), bunStubEnv(t), "--yes", "version")
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertContains(t, result.Combined(), "1.")
}

func TestInstallIsIdempotentWithStubBun(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constBunTestWindows {
		t.Skip("stub bun tests target Unix-like systems")
	}

	root := tasktestutil.ModuleRoot(t)
	env := bunStubEnv(t)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install"), 0)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install"), 0)
}

func TestInstallSkipsWhenBunIsAlreadyPresent(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constBunTestWindows {
		t.Skip("stub bun tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), bunStubEnv(t), "--yes", "install")
	tasktestutil.AssertExitCode(t, result, 0)
	tasktestutil.AssertNotContains(t, result.Combined(), "Installing Bun")
}

func TestInstallUndoRemovesBunDir(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constBunTestWindows {
		t.Skip("stub bun tests target Unix-like systems")
	}

	root := tasktestutil.ModuleRoot(t)
	env := bunStubEnv(t)
	bunDir := filepath.Join(tasktestutil.EnvValue(env, "HOME"), ".bun")
	tasktestutil.AssertDirExists(t, bunDir)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install:undo"), 0)
	tasktestutil.AssertDirNotExists(t, bunDir)
}

func TestInstallUndoIsIdempotent(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constBunTestWindows {
		t.Skip("stub bun tests target Unix-like systems")
	}

	root := tasktestutil.ModuleRoot(t)
	env := bunStubEnv(t)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install:undo"), 0)
	tasktestutil.AssertExitCode(t, tasktestutil.RunTask(t, root, env, "--yes", "install:undo"), 0)
}

func TestUpgradeExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constBunTestWindows {
		t.Skip("stub bun tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(t, tasktestutil.ModuleRoot(t), bunStubEnv(t), "--yes", "upgrade")
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestUpgradeCanaryExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constBunTestWindows {
		t.Skip("stub bun tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.ModuleRoot(t),
		bunStubEnv(t),
		"--yes",
		"upgrade:canary",
	)
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestUpgradeStableExitsSuccessfully(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == constBunTestWindows {
		t.Skip("stub bun tests target Unix-like systems")
	}

	result := tasktestutil.RunTask(
		t,
		tasktestutil.ModuleRoot(t),
		bunStubEnv(t),
		"--yes",
		"upgrade:stable",
	)
	tasktestutil.AssertExitCode(t, result, 0)
}

func TestRealInstallerFlowOnlyWhenExplicitlyEnabled(t *testing.T) {
	t.Parallel()

	if os.Getenv("RUN_INSTALLER_TESTS") != "1" {
		t.Skip("set RUN_INSTALLER_TESTS=1 to run real install/uninstall tests")
	}

	if runtime.GOOS == constBunTestWindows {
		t.Skip("real bun installer tests are intended for Unix-like systems")
	}

	root := tasktestutil.ModuleRoot(t)
	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")
	bunBin := filepath.Join(home, ".bun", "bin", "bun")

	tasktestutil.AssertExitCode(
		t,
		tasktestutil.RunTaskTimeout(t, root, env, 10*time.Minute, "--yes", "install"),
		0,
	)
	tasktestutil.AssertFileExists(t, bunBin)
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

	_, err := os.Stat(filepath.Join(home, ".bun"))

	if !os.IsNotExist(err) {
		t.Fatalf("expected .bun directory to be removed after install:undo: %s", home)
	}
}

func TestAllPublicTasksIntegration(t *testing.T) {
	t.Parallel()

	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("set RUN_INTEGRATION_TESTS=1 to run integration tests (downloads and installs Bun)")
	}

	if runtime.GOOS == constBunTestWindows {
		t.Skip("integration tests target Unix-like systems")
	}

	root := tasktestutil.ModuleRoot(t)
	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")
	bunBin := filepath.Join(home, ".bun", "bin", "bun")

	runBunIntegrationSteps(t, root, env, home, bunBin)
}

func runBunIntegrationSteps(t *testing.T, root string, env []string, home, bunBin string) {
	t.Helper()

	step := func(name string, fn func(t *testing.T)) {
		t.Helper()
		t.Run(name, fn)

		if t.Failed() {
			t.FailNow()
		}
	}
	run := func(t *testing.T, args ...string) tasktestutil.CommandResult {
		t.Helper()

		result := tasktestutil.RunTaskTimeout(t, root, env, 10*time.Minute, args...)
		tasktestutil.AssertExitCode(t, result, 0)

		return result
	}

	step("install — bun binary is present on disk", func(t *testing.T) {
		t.Helper()
		run(t, "--yes", "install")
		tasktestutil.AssertFileExists(t, bunBin)
	})
	step("version — bun version string is printed", func(t *testing.T) {
		t.Helper()

		result := run(t, "version")
		tasktestutil.AssertNotEmpty(t, result.Combined(), "version output is empty")
	})
	step("upgrade — bun upgrades without error", func(t *testing.T) {
		t.Helper()
		run(t, "--yes", "upgrade")
		tasktestutil.AssertFileExists(t, bunBin)
	})
	step("upgrade:canary — bun switches to canary without error", func(t *testing.T) {
		t.Helper()
		run(t, "--yes", "upgrade:canary")
		tasktestutil.AssertFileExists(t, bunBin)
	})
	step("upgrade:stable — bun switches back to stable without error", func(t *testing.T) {
		t.Helper()
		run(t, "--yes", "upgrade:stable")
		tasktestutil.AssertFileExists(t, bunBin)
	})
	step("install:undo — .bun directory is removed", func(t *testing.T) {
		t.Helper()
		run(t, "--yes", "install:undo")
		tasktestutil.AssertDirNotExists(t, filepath.Join(home, ".bun"))
	})
}

// BunStubEnv returns an isolated environment with a stub bun binary that
// satisfies precondition checks without performing real operations.

func bunStubEnv(t *testing.T) []string {
	t.Helper()

	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")

	bunBinDir := filepath.Join(home, ".bun", "bin")

	err := os.MkdirAll(bunBinDir, 0o700)
	if err != nil {
		t.Fatalf("failed to create stub bun dir: %v", err)
	}

	stub := "#!/usr/bin/env bash\n" +
		"case \"$1\" in\n" +
		"  --version) echo \"1.2.3\" ;;\n" +
		"  --revision) echo \"abc1234\" ;;\n" +
		"  upgrade) echo \"Bun is already at the latest version\" ;;\n" +
		"  *) exit 0 ;;\n" +
		"esac\n"

	bunPath := filepath.Join(bunBinDir, "bun")

	err = os.WriteFile(bunPath, []byte(stub), 0o600)
	if err != nil {
		t.Fatalf("failed to create stub bun binary: %v", err)
	}

	err = syscall.Chmod(bunPath, 0o500)
	if err != nil {
		t.Fatalf("make stub bun executable: %v", err)
	}

	path := tasktestutil.EnvValue(env, "PATH")

	return tasktestutil.SetEnv(env, "PATH", bunBinDir+":"+path)
}
