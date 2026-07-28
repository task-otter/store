package go_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

var publicTasks = []string{
	"bench",
	"config:skip",
	"coverage",
	"fmt",
	"fmt:check",
	"fuzz",
	"golangci-lint:fmt",
	"golangci-lint:fmt:check",
	"golangci-lint:lint",
	"golangci-lint:lint:fix",
	"gosec:lint",
	"govulncheck:lint",
	"install",
	"install:golangci-lint",
	"install:gosec",
	"install:govulncheck",
	"install:undo",
	"lint",
	"lint:fix",
	"test",
	"upgrade",
	"verify",
	"version",
	"which",
}

var publicVars = []string{
	"GO_BIN_UNIX",
	"GO_CMD_UNIX",
	"GO_COVER_PROFILE",
	"GO_DOWNLOAD_BASE_URL",
	"GO_FMT_SKIP_PATTERN",
	"GO_FUZZTIME",
	"GO_LINT_SKIP_PATTERN",
	"GO_VERSION",
	"GO_ROOT_UNIX",
	"GO_VERSION_URL",
	"GOLANGCI_LINT_VERSION",
	"GOSEC_VERSION",
	"GLOBAL_GO_BIN",
	"GOVULNCHECK_VERSION",
	"INSTALL_DIR_UNIX",
}

func goAvailable() bool {
	_, err := exec.LookPath("go")
	return err == nil
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "go", publicTasks, publicVars)
}

func TestVersionDryRun(t *testing.T) {
	t.Parallel()

	if !goAvailable() {
		t.Skip("go is not installed")
	}

	tasktest.AssertDryRunContains(t, "go", []string{"version"}, "go version")
}

func TestLintDryRuns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		task  string
		token string
	}{
		{task: "golangci-lint:fmt:check", token: "--diff"},
		{task: "golangci-lint:lint", token: "golangci-lint"},
		{task: "golangci-lint:lint:fix", token: "--fix"},
		{task: "govulncheck:lint", token: "govulncheck"},
		{task: "gosec:lint", token: "gosec"},
	}

	for _, tt := range tests {
		t.Run(tt.task, func(t *testing.T) {
			tasktest.AssertDryRunContains(t, "go", []string{tt.task}, tt.token)
		})
	}
}

func TestLintFixDryRun(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "go", []string{"lint:fix"},
		"golangci-lint",
		"--fix",
		"fmt",
		"-E",
		"gofumpt",
		"goimports",
		"golines",
		"swaggo",
	)
}

func TestTestingTaskCommands(t *testing.T) {
	t.Parallel()

	tf := tasktest.LoadTaskfile(t, "go")

	tests := []struct {
		task   string
		tokens []string
	}{
		{task: "test", tokens: []string{"go test", "./..."}},
		{task: "bench", tokens: []string{"go test", "-bench", "-benchmem"}},
		{task: "fuzz", tokens: []string{"go test", "-fuzz", "-fuzztime"}},
		{task: "coverage", tokens: []string{"-coverprofile", "awk", "LC_ALL=C sort", "Sort-Object"}},
	}

	for _, tt := range tests {
		t.Run(tt.task, func(t *testing.T) {
			task, ok := tf.Tasks[tt.task]
			if !ok {
				t.Fatalf("go Taskfile missing task %q", tt.task)
			}

			cmds := fmt.Sprintf("%v", task.Cmds)
			for _, token := range tt.tokens {
				if !strings.Contains(cmds, token) {
					t.Fatalf("go task %q cmds missing %q: %s", tt.task, token, cmds)
				}
			}
		})
	}

	coverageCommands := fmt.Sprintf("%v", tf.Tasks["coverage"].Cmds)
	if strings.Contains(coverageCommands, "go tool cover") {
		t.Fatalf("go coverage task must not run the per-function cover report: %s", coverageCommands)
	}
}

func TestCoverageReportsStatementPackagesInAscendingOrder(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	writeCoverageFixture(t, projectDir, map[string]string{
		"go.mod": "module example.com/coveragefixture\n\ngo 1.22\n",
		"zero/zero.go": `package zero

func Value() int { return 0 }
`,
		"partial/partial.go": `package partial

func Covered() int { return 1 }
func Uncovered() int { return 2 }
`,
		"partial/partial_test.go": `package partial

import "testing"

func TestCovered(t *testing.T) { Covered() }
`,
		"fulla/fulla.go": `package fulla

func Value() int { return 1 }
`,
		"fulla/fulla_test.go": `package fulla

import "testing"

func TestValue(t *testing.T) { Value() }
`,
		"fullb/fullb.go": `package fullb

func Value() int { return 1 }
`,
		"fullb/fullb_test.go": `package fullb

import "testing"

func TestValue(t *testing.T) { Value() }
`,
		"nostmt/nostmt.go": `package nostmt

const Value = 1
`,
	})

	profile := filepath.Join(projectDir, "coverage.out")
	output, err := runCoverageTask(t, projectDir, profile)
	if err != nil {
		t.Fatalf("go coverage task failed: %v\n%s", err, output)
	}

	var rows []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "%  example.com/coveragefixture/") {
			rows = append(rows, line)
		}
	}
	want := []string{
		"0.0%  example.com/coveragefixture/zero",
		"50.0%  example.com/coveragefixture/partial",
		"100.0%  example.com/coveragefixture/fulla",
		"100.0%  example.com/coveragefixture/fullb",
	}
	if !slices.Equal(rows, want) {
		t.Fatalf("coverage rows mismatch\nwant: %v\ngot:  %v\noutput:\n%s", want, rows, output)
	}
	if strings.Contains(output, "nostmt") {
		t.Fatalf("coverage output contains package without statements:\n%s", output)
	}
	if strings.Contains(output, "total:") {
		t.Fatalf("coverage output contains an aggregate total:\n%s", output)
	}
	if info, statErr := os.Stat(profile); statErr != nil || info.Size() == 0 {
		t.Fatalf("coverage profile was not written: info=%v err=%v", info, statErr)
	}
}

func TestCoveragePreservesTestFailure(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	writeCoverageFixture(t, projectDir, map[string]string{
		"go.mod": "module example.com/coveragefailure\n\ngo 1.22\n",
		"failure/failure.go": `package failure

func Value() int { return 1 }
`,
		"failure/failure_test.go": `package failure

import "testing"

func TestFailure(t *testing.T) { t.Fatal("coverage failure sentinel") }
`,
	})

	output, err := runCoverageTask(t, projectDir, filepath.Join(projectDir, "coverage.out"))
	if err == nil {
		t.Fatalf("go coverage task succeeded despite a failing test:\n%s", output)
	}
	if !strings.Contains(output, "coverage failure sentinel") {
		t.Fatalf("go coverage task hid the test failure diagnostics:\n%s", output)
	}
}

func writeCoverageFixture(t *testing.T, root string, files map[string]string) {
	t.Helper()

	for name, contents := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create fixture directory for %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatalf("write fixture file %s: %v", name, err)
		}
	}
}

func runCoverageTask(t *testing.T, projectDir, profile string) (string, error) {
	t.Helper()

	taskfile := filepath.Join(tasktest.RepoRoot(t), "taskfiles", "go", "Taskfile.yml")
	command := exec.Command(
		"task",
		"--silent",
		"--taskfile", taskfile,
		"coverage",
		"GO_COVER_PROFILE="+profile,
	)
	command.Dir = projectDir
	output, err := command.CombinedOutput()
	return string(output), err
}

func TestFmtDryRuns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		task   string
		tokens []string
	}{
		{
			task: "golangci-lint:fmt",
			tokens: []string{
				"fmt",
				"-E",
				"gci",
				"gofmt",
				"gofumpt",
				"goimports",
				"golines",
				"swaggo",
			},
		},
		{
			task: "fmt",
			tokens: []string{
				"fmt",
				"-E",
				"gci",
				"gofmt",
				"gofumpt",
				"goimports",
				"golines",
				"swaggo",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.task, func(t *testing.T) {
			tasktest.AssertDryRunContains(t, "go", []string{tt.task}, tt.tokens...)
		})
	}
}

func TestFmtSkipPatternDefaultsEmpty(t *testing.T) {
	t.Parallel()

	tf := tasktest.LoadTaskfile(t, "go")

	value, exists := tf.Vars["GO_FMT_SKIP_PATTERN"]
	if !exists {
		t.Fatal("GO_FMT_SKIP_PATTERN must be defined")
	}
	if value != "" {
		t.Fatalf("GO_FMT_SKIP_PATTERN default = %#v, want empty", value)
	}
}

func TestLintSkipPatternDefaultsEmpty(t *testing.T) {
	t.Parallel()

	tf := tasktest.LoadTaskfile(t, "go")

	value, exists := tf.Vars["GO_LINT_SKIP_PATTERN"]
	if !exists {
		t.Fatal("GO_LINT_SKIP_PATTERN must be defined")
	}
	if value != "" {
		t.Fatalf("GO_LINT_SKIP_PATTERN default = %#v, want empty", value)
	}
}

func TestFmtSkipPatternDryRuns(t *testing.T) {
	t.Parallel()

	const pattern = "**/generated/**"

	tasktest.AssertDryRunContains(t, "go",
		[]string{"fmt", "GO_FMT_SKIP_PATTERN=" + pattern},
		"golangci-lint",
		pattern,
	)
	tasktest.AssertDryRunContains(t, "go",
		[]string{"fmt:check", "GO_FMT_SKIP_PATTERN=" + pattern},
		"golangci-lint",
		pattern,
		"--diff",
	)
}

func TestLintSkipPatternDryRuns(t *testing.T) {
	t.Parallel()

	const pattern = "**/generated/**"

	tasktest.AssertDryRunContains(t, "go",
		[]string{"lint", "GO_LINT_SKIP_PATTERN=" + pattern},
		"golangci-lint",
		"config:skip",
		pattern,
	)
	tasktest.AssertDryRunContains(t, "go",
		[]string{"lint:fix", "GO_LINT_SKIP_PATTERN=" + pattern},
		"golangci-lint",
		"config:skip",
		pattern,
		"--fix",
	)
}

func TestDevelopmentToolDependencies(t *testing.T) {
	t.Parallel()

	tf := tasktest.LoadTaskfile(t, "go")

	installTasks := map[string][]string{
		"install:golangci-lint": {"install"},
		"install:govulncheck":   {"install"},
		"install:gosec":         {"install"},
	}
	lintTasks := map[string][]string{
		"fmt:check":               {"golangci-lint:fmt:check"},
		"golangci-lint:fmt":       {"install:golangci-lint"},
		"golangci-lint:fmt:check": {"install:golangci-lint"},
		"golangci-lint:lint":      {"install:golangci-lint"},
		"golangci-lint:lint:fix":  {"install:golangci-lint"},
		"govulncheck:lint":        {"install:govulncheck"},
		"gosec:lint":              {"install:gosec"},
		"lint": {
			"golangci-lint:lint",
			"golangci-lint:fmt:check",
			"govulncheck:lint",
			"gosec:lint",
		},
	}

	for taskName, expected := range installTasks {
		assertTaskDependencies(t, tf, taskName, expected)
	}
	for taskName, expected := range lintTasks {
		assertTaskDependencies(t, tf, taskName, expected)
	}
}

func TestDevelopmentToolInstallVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		task       string
		module     string
		versionVar string
	}{
		{
			task:       "install:golangci-lint",
			module:     "github.com/golangci/golangci-lint/v2/cmd/golangci-lint",
			versionVar: "GOLANGCI_LINT_VERSION",
		},
		{
			task:       "install:govulncheck",
			module:     "golang.org/x/vuln/cmd/govulncheck",
			versionVar: "GOVULNCHECK_VERSION",
		},
		{
			task:       "install:gosec",
			module:     "github.com/securego/gosec/v2/cmd/gosec",
			versionVar: "GOSEC_VERSION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.task+"/latest", func(t *testing.T) {
			tasktest.AssertDryRunContains(t, "go", []string{tt.task}, tt.module+"@latest")
		})
		t.Run(tt.task+"/explicit", func(t *testing.T) {
			tasktest.AssertDryRunContains(t, "go",
				[]string{tt.task, tt.versionVar + "=v0.0.0-test"},
				tt.module+"@v0.0.0-test",
			)
		})
	}
}

func TestVersionVariablesAreIndependentAndOptional(t *testing.T) {
	t.Parallel()

	tf := tasktest.LoadTaskfile(t, "go")

	if _, exists := tf.Vars["VERSION"]; exists {
		t.Fatal("shared VERSION variable must not be defined")
	}

	for _, name := range []string{
		"GO_VERSION",
		"GOLANGCI_LINT_VERSION",
		"GOVULNCHECK_VERSION",
		"GOSEC_VERSION",
	} {
		value, exists := tf.Vars[name]
		if !exists {
			t.Fatalf("%s must be defined", name)
		}
		if value != "" {
			t.Fatalf("%s default = %#v, want empty", name, value)
		}
	}
}

func TestGoInstallVersionDryRun(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin":
		tasktest.AssertDryRunContains(t, "go",
			[]string{"install", "GO_VERSION=go1.99.1"},
			"go1.99.1.darwin-",
			".pkg",
		)
	case "linux":
		tasktest.AssertDryRunContains(t, "go",
			[]string{"install", "GO_VERSION=go1.99.1"},
			"go1.99.1.linux-",
			".tar.gz",
		)
	default:
		t.Skip("explicit Go version dry-run is covered on macOS and Linux")
	}
}

func TestAggregateLintDryRun(t *testing.T) {
	t.Parallel()

	tasktest.AssertDryRunContains(t, "go", []string{"lint"},
		"golangci-lint",
		"--diff",
		"gci",
		"gofumpt",
		"goimports",
		"golines",
		"swaggo",
		"govulncheck",
		"gosec",
	)
}

func assertTaskDependencies(t *testing.T, tf tasktest.Taskfile, taskName string, expected []string) {
	t.Helper()

	rawDeps, ok := tf.Tasks[taskName].Deps.([]any)
	if !ok {
		t.Fatalf("%s deps have type %T, want []any", taskName, tf.Tasks[taskName].Deps)
	}

	actual := make([]string, len(rawDeps))
	for i, rawDep := range rawDeps {
		dep, ok := rawDep.(string)
		if !ok {
			t.Fatalf("%s dependency %d has type %T, want string", taskName, i, rawDep)
		}
		actual[i] = dep
	}

	if !slices.Equal(actual, expected) {
		t.Fatalf("%s deps mismatch\nexpected: %v\nactual:   %v", taskName, expected, actual)
	}
}

func TestUpgradeDryRun(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("brew"); err != nil {
			t.Skip("Homebrew is not installed")
		}
		tasktest.AssertDryRunContains(t, "go", []string{"upgrade"}, "brew upgrade go")
	case "linux":
		tasktest.AssertDryRunContains(t, "go", []string{"upgrade"},
			"https://go.dev/VERSION?m=text",
			"sudo tar",
		)
	default:
		t.Skip("upgrade dry-run is covered on macOS and Linux")
	}
}
