// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package golangcilint_test

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

type (
	gclFixture struct {
		project  string
		bin      string
		log      string
		template string
		taskfile string
	}

	golangciLintConfig struct {
		name          string
		destination   string
		pluginVersion string
	}

	golangciLintRun struct {
		taskName string
		extraEnv []string
		args     []string
	}

	dependencyCheck struct {
		taskfile *tasktest.Taskfile
		taskName string
		expected []string
	}

	defaultVarCheck struct {
		name string
	}
)

const (
	constGolangciLintCi                 = "ci"
	constGolangciLintCiFix              = "ci:fix"
	constGolangciLintFmt                = "fmt"
	constGolangciLintFmtCheck           = "fmt:check"
	constGolangciLintLint               = "lint"
	constGolangciLintLintFix            = "lint:fix"
	constGolangciLintModule             = "golangci-lint"
	constStockCustomLog                 = "stock:custom -v"
	constCustomRunDefaultLog            = "custom:run ./..."
	constProjectCustomName              = "project-gcl"
	constProjectCustomDestination       = ".tools"
	constInitialPluginVersion           = "v1.0.0"
	constUpdatedPluginVersion           = "v1.0.1"
	constGolangciLintNixInstallable     = "GOLANGCI_LINT_NIX_INSTALLABLE"
	constGolangciLintLintSkipPatternVar = "GOLANGCI_LINT_LINT_SKIP_PATTERN"
	constEmptyValue                     = ""
	constTaskBaseArgCount               = 4
	constSecureFileMode                 = 0o600
	constPrivateDirectoryMode           = 0o750
	constExecutableFileMode             = 0o700
	constNewline                        = "\n"
	constZeroLen                        = 0
	stockGolangciLintScriptTemplate     = `#!/bin/sh
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
`
)

func publicTasks() []string {
	return []string{
		constGolangciLintCi,
		constGolangciLintCiFix,
		"config:skip",
		constGolangciLintFmt,
		constGolangciLintFmtCheck,
		constGolangciLintLint,
		constGolangciLintLintFix,
	}
}

func publicVars() []string {
	return []string{
		constGolangciLintNixInstallable,
		constGolangciLintLintSkipPatternVar,
		"GOLANGCI_LINT_INTERNAL_SKIP_CONFIG",
		"GOLANGCI_LINT_FMT_FORMATTER_FLAGS",
	}
}

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		constGolangciLintModule,
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

// TestGolangciLintCustomBuildLifecycle
func TestGolangciLintCustomBuildLifecycle(t *testing.T) {
	t.Parallel()
	skipWindows(t)

	fixture := newCustomGolangciLintFixture(t)

	assertInitialCustomBuild(t, &fixture)
	assertCachedCustomBuild(t, &fixture)
	assertUpdatedCustomBuild(t, &fixture)
	assertCustomRebuildAfterRemoval(t, &fixture)
}

func assertCustomBuildStockFallback(t *testing.T) {
	t.Parallel()

	fixture := newCustomGolangciLintFixture(t)
	fixture.run(t, constGolangciLintLint)
	fixture.assertLog(t, "stock:run ./...")
}

func assertCustomBuildDefaultOutput(t *testing.T) {
	t.Parallel()

	fixture := newCustomGolangciLintFixture(t)
	fixture.writeConfig(t, &golangciLintConfig{
		name:          constEmptyValue,
		destination:   constEmptyValue,
		pluginVersion: constInitialPluginVersion,
	})
	fixture.run(t, constGolangciLintLint)
	fixture.assertLog(t, constStockCustomLog, constCustomRunDefaultLog)
	assertDefaultCustomGolangciLintFile(t, &fixture)
}

func assertDefaultCustomGolangciLintFile(t *testing.T, fixture *gclFixture) {
	t.Helper()

	info, err := os.Stat(filepath.Join(fixture.project, "custom-gcl"))
	if err != nil {
		t.Fatalf("stat default custom golangci-lint: %v", err)
	}

	if info.IsDir() {
		t.Fatal("expected file at custom-gcl, found directory")
	}
}

func runFailingCustomBuild(t *testing.T, fixture *gclFixture) ([]byte, error) {
	t.Helper()

	fixture.writeConfig(t, projectCustomConfig(constInitialPluginVersion))

	output, err := fixture.runCommand(t.Context(), &golangciLintRun{
		taskName: constGolangciLintLint,
		extraEnv: []string{"GCL_CUSTOM_EXIT=9"},
		args:     nil,
	})
	if err != nil {
		return output, fmt.Errorf("run failing custom build: %w", err)
	}

	return output, nil
}

func assertCustomBuildFailure(t *testing.T) {
	t.Parallel()

	fixture := newCustomGolangciLintFixture(t)

	output, err := runFailingCustomBuild(t, &fixture)
	if err == nil {
		t.Fatalf("custom golangci-lint build succeeded unexpectedly\n%s", output)
	}

	fixture.assertLog(t, constStockCustomLog)
}

// TestGolangciLintCustomBuildDefaultsAndFallback
func TestGolangciLintCustomBuildDefaultsAndFallback(t *testing.T) {
	t.Parallel()
	skipWindows(t)

	t.Run("stock fallback", assertCustomBuildStockFallback)
	t.Run("default output", assertCustomBuildDefaultOutput)
	t.Run("build failure", assertCustomBuildFailure)
}

func skipWindows(t *testing.T) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("shell fixture covers the Unix command path")
	}
}

func assertInitialCustomBuild(t *testing.T, fixture *gclFixture) {
	t.Helper()

	fixture.writeConfig(t, projectCustomConfig(constInitialPluginVersion))
	fixture.run(t, constGolangciLintLint)
	fixture.assertLog(t, constStockCustomLog, constCustomRunDefaultLog)
}

func assertCachedCustomBuild(t *testing.T, fixture *gclFixture) {
	t.Helper()

	fixture.clearLog(t)
	fixture.run(t, constGolangciLintLint)
	fixture.assertLog(t, constCustomRunDefaultLog)
}

func assertUpdatedCustomBuild(t *testing.T, fixture *gclFixture) {
	t.Helper()

	fixture.writeConfig(t, projectCustomConfig(constUpdatedPluginVersion))
	fixture.clearLog(t)
	fixture.run(t, constGolangciLintLint)
	fixture.assertLog(t, constStockCustomLog, constCustomRunDefaultLog)
}

func projectCustomConfig(pluginVersion string) *golangciLintConfig {
	return &golangciLintConfig{
		name:          constProjectCustomName,
		destination:   constProjectCustomDestination,
		pluginVersion: pluginVersion,
	}
}

func assertCustomRebuildAfterRemoval(t *testing.T, fixture *gclFixture) {
	t.Helper()

	err := os.Remove(
		filepath.Join(fixture.project, constProjectCustomDestination, constProjectCustomName),
	)
	if err != nil {
		t.Fatalf("remove custom golangci-lint: %v", err)
	}

	fixture.clearLog(t)
	fixture.run(t, constGolangciLintLintFix, "--", "./internal/...")
	fixture.assertLog(t,
		constStockCustomLog,
		"custom:run --fix ./internal/...",
	)
}

func golangciLintTaskfilePath(t *testing.T) string {
	t.Helper()

	return filepath.Join(
		tasktest.RepoRoot(t),
		"taskfiles",
		constGolangciLintModule,
		"Taskfile.yml",
	)
}

func newCustomGolangciLintFixture(t *testing.T) gclFixture {
	t.Helper()

	project, bin := newCustomGolangciLintFixtureDirs(t)

	logPath := filepath.Join(project, constGolangciLintModule+".log")
	template := filepath.Join(project, "custom-template")
	writeExecutable(t, template, customGolangciLintScript())
	writeExecutable(t, filepath.Join(bin, constGolangciLintModule), stockGolangciLintScript())

	return gclFixture{
		project:  project,
		bin:      bin,
		log:      logPath,
		template: template,
		taskfile: golangciLintTaskfilePath(t),
	}
}

func newCustomGolangciLintFixtureDirs(t *testing.T) (project, bin string) {
	t.Helper()

	project = t.TempDir()
	bin = filepath.Join(project, "bin")

	err := os.MkdirAll(bin, constPrivateDirectoryMode)
	if err != nil {
		t.Fatalf("create fixture bin: %v", err)
	}

	return project, bin
}

func customGolangciLintScript() string {
	return `#!/bin/sh
set -eu
printf 'custom:%s\n' "$*" >>"$GCL_LOG"
if [ "${1:-}" = "run" ]; then
  exit "${GCL_RUN_EXIT:-0}"
fi
`
}

func stockGolangciLintScript() string {
	return stockGolangciLintScriptTemplate
}

func (fixture *gclFixture) assertLog(t *testing.T, expected ...string) {
	t.Helper()

	content, err := os.ReadFile(fixture.log)
	if err != nil {
		t.Fatalf("read golangci-lint log: %v", err)
	}

	actual := strings.Split(strings.TrimSpace(string(content)), constNewline)

	if !slices.Equal(actual, expected) {
		t.Fatalf("golangci-lint log mismatch\nexpected: %v\nactual:   %v", expected, actual)
	}
}

func (fixture *gclFixture) clearLog(t *testing.T) {
	t.Helper()

	err := os.WriteFile(fixture.log, nil, constSecureFileMode)
	if err != nil {
		t.Fatalf("clear golangci-lint log: %v", err)
	}
}

func (fixture *gclFixture) command(ctx context.Context, run *golangciLintRun) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "task")

	cmd.Args = append(cmd.Args, fixture.taskArgs(run)...)

	cmd.Dir = fixture.project
	cmd.Env = append(os.Environ(),
		"CUSTOM_GCL_TEMPLATE="+fixture.template,
		"GCL_LOG="+fixture.log,
		"TASK_TEMP_DIR="+filepath.Join(fixture.project, ".task"),
	)
	cmd.Env = append(cmd.Env, run.extraEnv...)

	return cmd
}

func (fixture *gclFixture) run(t *testing.T, taskName string, args ...string) {
	t.Helper()

	output, err := fixture.runCommand(t.Context(), &golangciLintRun{
		taskName: taskName,
		args:     args,
		extraEnv: nil,
	})
	if err != nil {
		t.Fatalf("run task %s: %v\n%s", taskName, err, output)
	}
}

func (fixture *gclFixture) runCommand(ctx context.Context, run *golangciLintRun) ([]byte, error) {
	cmd := fixture.command(ctx, run)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("run task %s: %w", run.taskName, err)
	}

	return output, nil
}

func (fixture *gclFixture) taskArgs(run *golangciLintRun) []string {
	taskArgs := make([]string, len(run.args), constTaskBaseArgCount+len(run.args))
	copy(taskArgs, run.args)

	return append([]string{
		"--taskfile",
		fixture.taskfile,
		run.taskName,
		"GO_GLOBAL_BIN=" + fixture.bin,
	}, taskArgs...)
}

func golangciLintConfigLines(config *golangciLintConfig) []string {
	lines := []string{"version: v2.12.2"}

	if config.name != constEmptyValue {
		lines = append(lines, "name: "+config.name)
	}

	if config.destination != constEmptyValue {
		lines = append(lines, "destination: "+config.destination)
	}

	return append(
		lines,
		"plugins:",
		"  - module: example.com/custom-linter",
		"    version: "+config.pluginVersion,
	)
}

func (fixture *gclFixture) writeConfig(t *testing.T, config *golangciLintConfig) {
	t.Helper()

	lines := golangciLintConfigLines(config)

	err := os.WriteFile(
		filepath.Join(fixture.project, ".custom-gcl.yml"),
		[]byte(strings.Join(lines, constNewline)+constNewline),
		constSecureFileMode,
	)
	if err != nil {
		t.Fatalf("write custom golangci-lint config: %v", err)
	}
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()

	err := os.WriteFile(path, []byte(content), constSecureFileMode)
	if err != nil {
		t.Fatalf("write executable %s: %v", path, err)
	}

	// #nosec G302 -- fixture commands must be executable.
	err = os.Chmod(path, constExecutableFileMode)
	if err != nil {
		t.Fatalf("make executable %s: %v", path, err)
	}
}

// TestGolangciLintVariableDefaultsEmpty
func TestGolangciLintVariableDefaultsEmpty(t *testing.T) {
	t.Parallel()

	for i := range defaultVarChecks() {
		check := defaultVarChecks()[i]
		t.Run(check.name, func(t *testing.T) {
			t.Parallel()
			assertDefaultVarEmpty(t, check)
		})
	}
}

func defaultVarChecks() []defaultVarCheck {
	return []defaultVarCheck{
		{name: constGolangciLintLintSkipPatternVar},
	}
}

func assertDefaultVarEmpty(t *testing.T, check defaultVarCheck) {
	t.Helper()

	taskfile := tasktest.LoadTaskfile(t, constGolangciLintModule)
	value, exists := taskfile.Vars[check.name]

	if !exists {
		t.Fatalf("%s must be defined", check.name)
	}

	if value != constEmptyValue {
		t.Fatalf("%s default = %#v, want empty", check.name, value)
	}
}

// TestDevelopmentToolDependencies
func TestDevelopmentToolDependencies(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, constGolangciLintModule)

	for taskName := range developmentToolDependencies() {
		expected := developmentToolDependencies()[taskName]
		assertTaskDependencies(
			t,
			&dependencyCheck{taskfile: taskfile, taskName: taskName, expected: expected},
		)
	}
}

func developmentToolDependencies() map[string][]string {
	return map[string][]string{
		constGolangciLintFmt:      {"_ensure"},
		constGolangciLintFmtCheck: {"_ensure"},
		constGolangciLintLint:     {"_ensure"},
		constGolangciLintLintFix:  {"_ensure"},
	}
}

func assertTaskDependencies(t *testing.T, check *dependencyCheck) {
	t.Helper()

	actual, ok := taskDependencyNames(check)

	if !ok {
		t.Fatalf(
			"%s deps have type %T, want []any",
			check.taskName,
			check.taskfile.Tasks[check.taskName].Deps,
		)
	}

	if !slices.Equal(actual, check.expected) {
		t.Fatalf(
			"%s deps mismatch\nexpected: %v\nactual:   %v",
			check.taskName,
			check.expected,
			actual,
		)
	}
}

func taskDependencyNames(check *dependencyCheck) ([]string, bool) {
	rawDeps, ok := check.taskfile.Tasks[check.taskName].Deps.([]any)

	if !ok {
		return nil, false
	}

	return dependencyNames(rawDeps)
}

func dependencyNames(rawDeps []any) ([]string, bool) {
	actual := make([]string, constZeroLen, len(rawDeps))

	for index := range rawDeps {
		rawDep := rawDeps[index]
		dep, ok := taskDependencyName(rawDep)

		if !ok {
			return nil, false
		}

		actual = append(actual, dep)
	}

	return actual, true
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
