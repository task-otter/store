// REPLACE_ME 2026
// SPDX-License-Identifier: Apache-2.0

package go_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

const (
	constGoTestGolangciLintFmtCheck = "golangci-lint:fmt:check"
	constGoTestGolangciLintLint     = "golangci-lint:lint"
	constGoTestGosecLint            = "gosec:lint"
	constGoTestGovulncheckLint      = "govulncheck:lint"
	constGoTestGolangciLintVersion  = "GOLANGCI_LINT_VERSION"
	constGoTestInstall              = "install"
	constGoTestInstallGoJunitReport = "install:go-junit-report"
	constGoTestInstallGolangciLint  = "install:golangci-lint"
	constGoTestInstallGosec         = "install:gosec"
	constGoTestInstallGovulncheck   = "install:govulncheck"
	constGoTestTest                 = "test"
)

func publicTasks() []string {
	return []string{
		"bench",
		"config:skip",
		"fmt",
		"fmt:check",
		"fuzz",
		"golangci-lint:fmt",
		constGoTestGolangciLintFmtCheck,
		constGoTestGolangciLintLint,
		"golangci-lint:lint:fix",
		constGoTestGosecLint,
		constGoTestGovulncheckLint,
		constGoTestInstall,
		constGoTestInstallGoJunitReport,
		constGoTestInstallGolangciLint,
		constGoTestInstallGosec,
		constGoTestInstallGovulncheck,
		"install:undo",
		"lint",
		"lint:fix",
		constGoTestTest,
		"upgrade",
		"verify",
		"version",
		"which",
	}
}

func publicVars() []string {
	return []string{
		"GO_BIN_UNIX",
		"GO_CMD_UNIX",
		"GO_COVER_PROFILE",
		"GO_DOWNLOAD_BASE_URL",
		"GO_FMT_SKIP_PATTERN",
		"GO_FUZZTIME",
		"GO_JUNIT_REPORT",
		"GO_LINT_SKIP_PATTERN",
		"GO_VERSION",
		"GO_ROOT_UNIX",
		"GO_VERSION_URL",
		constGoTestGolangciLintVersion,
		"GOSEC_VERSION",
		"GLOBAL_GO_BIN",
		"GOVULNCHECK_VERSION",
		"INSTALL_DIR_UNIX",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "go", publicTasks(), publicVars())
}

func TestTestingTaskCommands(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "go")

	tests := []struct {
		task   string
		tokens []string
	}{
		{
			task: constGoTestTest,
			tokens: []string{
				"go test -v",
				"-covermode atomic",
				"-coverprofile",
				"./...",
				"go-junit-report",
				"-set-exit-code",
				"-iocopy",
				"-out",
			},
		},
		{task: "bench", tokens: []string{"go test", "-bench", "-benchmem"}},
		{task: "fuzz", tokens: []string{"go test", "-fuzz", "-fuzztime"}},
	}

	for _, testCase := range tests {
		t.Run(testCase.task, func(t *testing.T) {
			t.Parallel()

			task, ok := taskfile.Tasks[testCase.task]

			if !ok {
				t.Fatalf("go Taskfile missing task %q", testCase.task)
			}

			cmds := fmt.Sprintf("%v", task.Cmds)

			for _, token := range testCase.tokens {
				if !strings.Contains(cmds, token) {
					t.Fatalf("go task %q cmds missing %q: %s", testCase.task, token, cmds)
				}
			}
		})
	}

	testVars := fmt.Sprintf("%v", taskfile.Tasks[constGoTestTest].Vars)

	for _, token := range []string{
		"GO_JUNIT_REPORT_OUT",
		"GO_JUNIT_REPORT",
		"junit.xml",
		"GO_COVER_OUT",
		"GO_COVER_PROFILE",
		"coverage.out",
	} {
		if !strings.Contains(testVars, token) {
			t.Fatalf("go test vars missing %q: %s", token, testVars)
		}
	}
}

func TestGolangciLintInstallerUsesGoInstall(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "go")

	task, ok := taskfile.Tasks[constGoTestInstallGolangciLint]

	if !ok {
		t.Fatal("go Taskfile missing install:golangci-lint")
	}

	cmds := fmt.Sprintf("%v", task.Cmds)

	for _, token := range []string{
		"go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@",
		"default \"latest\" .GOLANGCI_LINT_VERSION",
		"New-Item -ItemType Directory -Force",
	} {
		if !strings.Contains(cmds, token) {
			t.Fatalf("install:golangci-lint cmds missing %q: %s", token, cmds)
		}
	}

	for _, token := range []string{
		"custom -v",
		".custom-gcl.yml",
		"github.com/gostafa/modularity",
		"github.com/gostafa/coverlint",
		"coverlint",
		"modularity",
	} {
		if strings.Contains(cmds, token) {
			t.Fatalf("install:golangci-lint cmds contain plugin token %q: %s", token, cmds)
		}
	}

	status := fmt.Sprintf("%v", task.Status)

	for _, token := range []string{
		"golangci-lint\" version --short",
		"GOLANGCI_LINT_VERSION",
	} {
		if !strings.Contains(status, token) {
			t.Fatalf("install:golangci-lint status missing %q: %s", token, status)
		}
	}

	for _, token := range []string{
		"golangci-lint.coverlint.version",
		"golangci-lint.modularity.version",
		"linters --config",
		"taskotter-golangci-status",
		"coverlint",
		"modularity",
	} {
		if strings.Contains(status, token) {
			t.Fatalf("install:golangci-lint status contains plugin token %q: %s", token, status)
		}
	}
}

func TestGolangciLintCustomBuildLifecycle(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell fixture covers the Unix command path")
	}

	fixture := newCustomGolangciLintFixture(t)
	fixture.writeConfig(t, "project-gcl", ".tools", "v1.0.0")

	fixture.run(t, "golangci-lint:lint")
	fixture.assertLog(t,
		"stock:custom -v",
		"custom:run ./...",
	)

	fixture.clearLog(t)
	fixture.run(t, "golangci-lint:lint")
	fixture.assertLog(t, "custom:run ./...")

	fixture.writeConfig(t, "project-gcl", ".tools", "v1.0.1")
	fixture.clearLog(t)
	fixture.run(t, "golangci-lint:lint")
	fixture.assertLog(t,
		"stock:custom -v",
		"custom:run ./...",
	)

	err := os.Remove(filepath.Join(fixture.project, ".tools", "project-gcl"))
	if err != nil {
		t.Fatalf("remove custom golangci-lint: %v", err)
	}

	fixture.clearLog(t)
	fixture.run(t, "golangci-lint:lint:fix", "--", "./internal/...")
	fixture.assertLog(t,
		"stock:custom -v",
		"custom:run --fix ./internal/...",
	)
}

func TestGolangciLintCustomBuildDefaultsAndFallback(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell fixture covers the Unix command path")
	}

	t.Run("stock fallback", func(t *testing.T) {
		t.Parallel()

		fixture := newCustomGolangciLintFixture(t)
		fixture.run(t, "golangci-lint:lint")
		fixture.assertLog(t, "stock:run ./...")
	})

	t.Run("default output", func(t *testing.T) {
		t.Parallel()

		fixture := newCustomGolangciLintFixture(t)
		fixture.writeConfig(t, "", "", "v1.0.0")
		fixture.run(t, "golangci-lint:lint")
		fixture.assertLog(t,
			"stock:custom -v",
			"custom:run ./...",
		)

		_, err := os.Stat(filepath.Join(fixture.project, "custom-gcl"))
		if err != nil {
			t.Fatalf("stat default custom golangci-lint: %v", err)
		}
	})

	t.Run("build failure", func(t *testing.T) {
		t.Parallel()

		fixture := newCustomGolangciLintFixture(t)
		fixture.writeConfig(t, "project-gcl", ".tools", "v1.0.0")

		output, err := fixture.runCommand(
			t.Context(),
			"golangci-lint:lint",
			[]string{"GCL_CUSTOM_EXIT=9"},
		)
		if err == nil {
			t.Fatalf("custom golangci-lint build succeeded unexpectedly\n%s", output)
		}

		fixture.assertLog(t, "stock:custom -v")
	})
}

type customGolangciLintFixture struct {
	project  string
	bin      string
	log      string
	template string
	taskfile string
}

func newCustomGolangciLintFixture(t *testing.T) customGolangciLintFixture {
	t.Helper()

	project := t.TempDir()

	bin := filepath.Join(project, "bin")

	err := os.MkdirAll(bin, 0o750)
	if err != nil {
		t.Fatalf("create fixture bin: %v", err)
	}

	logPath := filepath.Join(project, "golangci-lint.log")
	template := filepath.Join(project, "custom-template")
	writeExecutable(t, template, `#!/bin/sh
set -eu
printf 'custom:%s\n' "$*" >>"$GCL_LOG"
if [ "${1:-}" = "run" ]; then
  exit "${GCL_RUN_EXIT:-0}"
fi
`)

	writeExecutable(t, filepath.Join(bin, "golangci-lint"), `#!/bin/sh
set -eu
printf 'stock:%s\n' "$*" >>"$GCL_LOG"
if [ "${1:-}" = "custom" ]; then
  exit_code="${GCL_CUSTOM_EXIT:-0}"
  if [ "$exit_code" -ne 0 ]; then
    exit "$exit_code"
  fi
  name="$(awk '$1 == "name:" { print $2; exit }' .custom-gcl.yml)"
  destination="$(awk '$1 == "destination:" { print $2; exit }' .custom-gcl.yml)"
  name="${name:-custom-gcl}"
  destination="${destination:-.}"
  mkdir -p "$destination"
  cp "$CUSTOM_GCL_TEMPLATE" "$destination/$name"
  chmod +x "$destination/$name"
  exit 0
fi
if [ "${1:-}" = "run" ]; then
  exit "${GCL_RUN_EXIT:-0}"
fi
`)

	return customGolangciLintFixture{
		project:  project,
		bin:      bin,
		log:      logPath,
		template: template,
		taskfile: filepath.Join(tasktest.RepoRoot(t), "taskfiles", "go", "Taskfile.yml"),
	}
}

func (fixture customGolangciLintFixture) writeConfig(
	t *testing.T,
	name string,
	destination string,
	pluginVersion string,
) {

	t.Helper()

	var builder strings.Builder

	builder.WriteString("version: v2.12.2\n")

	if name != "" {
		fmt.Fprintf(&builder, "name: %s\n", name)
	}

	if destination != "" {
		fmt.Fprintf(&builder, "destination: %s\n", destination)
	}

	fmt.Fprintf(
		&builder,
		"plugins:\n  - module: example.com/custom-linter\n    version: %s\n",
		pluginVersion,
	)

	err := os.WriteFile(
		filepath.Join(fixture.project, ".custom-gcl.yml"),
		[]byte(builder.String()),
		0o600,
	)
	if err != nil {
		t.Fatalf("write custom golangci-lint config: %v", err)
	}
}

func (fixture customGolangciLintFixture) run(t *testing.T, taskName string, args ...string) {
	t.Helper()

	output, err := fixture.runCommand(t.Context(), taskName, nil, args...)
	if err != nil {
		t.Fatalf("run task %s: %v\n%s", taskName, err, output)
	}
}

func (fixture customGolangciLintFixture) runCommand(
	ctx context.Context,
	taskName string,
	extraEnv []string,
	args ...string,
) ([]byte, error) {

	taskArgs := make([]string, 0, 4+len(args))

	taskArgs = append(taskArgs,
		"--taskfile",
		fixture.taskfile,
		taskName,
		"GLOBAL_GO_BIN="+fixture.bin,
	)
	taskArgs = append(taskArgs, args...)

	// #nosec G204 -- every subprocess argument is assembled by this test fixture.
	cmd := exec.CommandContext(ctx, "task", taskArgs...)

	cmd.Dir = fixture.project
	cmd.Env = append(os.Environ(),
		"CUSTOM_GCL_TEMPLATE="+fixture.template,
		"GCL_LOG="+fixture.log,
		"TASK_TEMP_DIR="+filepath.Join(fixture.project, ".task"),
	)
	cmd.Env = append(cmd.Env, extraEnv...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("run task %s: %w", taskName, err)
	}

	return output, nil
}

func (fixture customGolangciLintFixture) clearLog(t *testing.T) {
	t.Helper()

	err := os.WriteFile(fixture.log, nil, 0o600)
	if err != nil {
		t.Fatalf("clear golangci-lint log: %v", err)
	}
}

func (fixture customGolangciLintFixture) assertLog(t *testing.T, expected ...string) {
	t.Helper()

	content, err := os.ReadFile(fixture.log)
	if err != nil {
		t.Fatalf("read golangci-lint log: %v", err)
	}

	actual := strings.Split(strings.TrimSpace(string(content)), "\n")

	if !slices.Equal(actual, expected) {
		t.Fatalf("golangci-lint log mismatch\nexpected: %v\nactual:   %v", expected, actual)
	}
}

func writeExecutable(t *testing.T, path string, content string) {
	t.Helper()

	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("write executable %s: %v", path, err)
	}

	err = os.Chmod(path, 0o700) // #nosec G302 -- fixture commands must be executable.
	if err != nil {
		t.Fatalf("make executable %s: %v", path, err)
	}
}

func TestFmtSkipPatternDefaultsEmpty(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "go")

	value, exists := taskfile.Vars["GO_FMT_SKIP_PATTERN"]

	if !exists {
		t.Fatal("GO_FMT_SKIP_PATTERN must be defined")
	}

	if value != "" {
		t.Fatalf("GO_FMT_SKIP_PATTERN default = %#v, want empty", value)
	}
}

func TestLintSkipPatternDefaultsEmpty(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "go")

	value, exists := taskfile.Vars["GO_LINT_SKIP_PATTERN"]

	if !exists {
		t.Fatal("GO_LINT_SKIP_PATTERN must be defined")
	}

	if value != "" {
		t.Fatalf("GO_LINT_SKIP_PATTERN default = %#v, want empty", value)
	}
}

func TestDevelopmentToolDependencies(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "go")

	installTasks := map[string][]string{
		constGoTestInstallGolangciLint:  {constGoTestInstall},
		constGoTestInstallGoJunitReport: {constGoTestInstall},
		constGoTestInstallGovulncheck:   {constGoTestInstall},
		constGoTestInstallGosec:         {constGoTestInstall},
	}
	lintTasks := map[string][]string{
		"fmt:check":                     {constGoTestGolangciLintFmtCheck},
		"golangci-lint:fmt":             {constGoTestInstallGolangciLint},
		constGoTestGolangciLintFmtCheck: {constGoTestInstallGolangciLint},
		constGoTestGolangciLintLint:     {constGoTestInstallGolangciLint},
		"golangci-lint:lint:fix":        {constGoTestInstallGolangciLint},
		constGoTestGovulncheckLint:      {constGoTestInstallGovulncheck},
		constGoTestGosecLint:            {constGoTestInstallGosec},
		"lint": {
			constGoTestGolangciLintLint,
			constGoTestGolangciLintFmtCheck,
			constGoTestGovulncheckLint,
			constGoTestGosecLint,
		},
	}
	testTasks := map[string][]string{
		constGoTestTest: {constGoTestInstallGoJunitReport},
	}

	for taskName, expected := range installTasks {
		assertTaskDependencies(t, taskfile, taskName, expected)
	}

	for taskName, expected := range lintTasks {
		assertTaskDependencies(t, taskfile, taskName, expected)
	}

	for taskName, expected := range testTasks {
		assertTaskDependencies(t, taskfile, taskName, expected)
	}
}

func TestVersionVariablesAreIndependentAndOptional(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "go")

	if _, exists := taskfile.Vars["VERSION"]; exists {
		t.Fatal("shared VERSION variable must not be defined")
	}

	for _, name := range []string{
		"GO_VERSION",
		"GOLANGCI_LINT_VERSION",
		"GOVULNCHECK_VERSION",
		"GOSEC_VERSION",
	} {
		value, exists := taskfile.Vars[name]

		if !exists {
			t.Fatalf("%s must be defined", name)
		}

		if value != "" {
			t.Fatalf("%s default = %#v, want empty", name, value)
		}
	}
}

func assertTaskDependencies(
	t *testing.T,
	taskfile tasktest.Taskfile,
	taskName string,
	expected []string,
) {

	t.Helper()

	rawDeps, ok := taskfile.Tasks[taskName].Deps.([]any)

	if !ok {
		t.Fatalf("%s deps have type %T, want []any", taskName, taskfile.Tasks[taskName].Deps)
	}

	actual := make([]string, len(rawDeps))

	for index, rawDep := range rawDeps {
		dep, ok := taskDependencyName(rawDep)

		if !ok {
			t.Fatalf("%s dependency %d has unsupported value %v", taskName, index, rawDep)
		}

		actual[index] = dep
	}

	if !slices.Equal(actual, expected) {
		t.Fatalf("%s deps mismatch\nexpected: %v\nactual:   %v", taskName, expected, actual)
	}
}

func taskDependencyName(rawDep any) (string, bool) {
	switch dep := rawDep.(type) {
	case string:
		return dep, true
	case map[string]any:
		name, ok := dep["task"].(string)

		return name, ok
	default:
		return "", false
	}
}
