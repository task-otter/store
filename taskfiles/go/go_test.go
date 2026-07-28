package go_test

import (
	"fmt"
	"os/exec"

	"slices"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"bench",
	"config:skip",
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
	"install:go-junit-report",
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
	"GO_JUNIT_REPORT",
	"GO_LINT_SKIP_PATTERN",
	"GO_VERSION",
	"GO_ROOT_UNIX",
	"GO_VERSION_URL",
	"GOLANGCI_LINT_VERSION",
	"GOLANGCI_LINT_MODULARITY_VERSION",
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

func TestTestingTaskCommands(t *testing.T) {
	t.Parallel()

	tf := tasktest.LoadTaskfile(t, "go")

	tests := []struct {
		task   string
		tokens []string
	}{
		{
			task: "test",
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

	testVars := fmt.Sprintf("%v", tf.Tasks["test"].Vars)
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

func TestGolangciLintInstallerBuildsModularityPlugin(t *testing.T) {
	t.Parallel()

	tf := tasktest.LoadTaskfile(t, "go")
	task, ok := tf.Tasks["install:golangci-lint"]
	if !ok {
		t.Fatal("go Taskfile missing install:golangci-lint")
	}

	cmds := fmt.Sprintf("%v", task.Cmds)
	for _, token := range []string{
		"go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@",
		"custom -v",
		".custom-gcl.yml",
		"name: golangci-lint",
		"destination:",
		"github.com/gostafa/modularity",
		"github.com/gostafa/modularity/plugin",
		"GOLANGCI_LINT_MODULARITY_VERSION",
		"golangci-lint.modularity.version",
		"Go 1.26.5 or newer",
	} {
		if !strings.Contains(cmds, token) {
			t.Fatalf("install:golangci-lint cmds missing %q: %s", token, cmds)
		}
	}

	status := fmt.Sprintf("%v", task.Status)
	for _, token := range []string{
		"golangci-lint\" version --short",
		"golangci-lint.modularity.version",
		"--issues-exit-code=0",
		"taskotter-modularity-status",
		"modularity",
		"GOLANGCI_LINT_MODULARITY_VERSION",
	} {
		if !strings.Contains(status, token) {
			t.Fatalf("install:golangci-lint status missing %q: %s", token, status)
		}
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

func TestDevelopmentToolDependencies(t *testing.T) {
	t.Parallel()

	tf := tasktest.LoadTaskfile(t, "go")

	installTasks := map[string][]string{
		"install:golangci-lint":   {"install"},
		"install:go-junit-report": {"install"},
		"install:govulncheck":     {"install"},
		"install:gosec":           {"install"},
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
	testTasks := map[string][]string{
		"test": {"install:go-junit-report"},
	}

	for taskName, expected := range installTasks {
		assertTaskDependencies(t, tf, taskName, expected)
	}
	for taskName, expected := range lintTasks {
		assertTaskDependencies(t, tf, taskName, expected)
	}
	for taskName, expected := range testTasks {
		assertTaskDependencies(t, tf, taskName, expected)
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

	value, exists := tf.Vars["GOLANGCI_LINT_MODULARITY_VERSION"]
	if !exists {
		t.Fatal("GOLANGCI_LINT_MODULARITY_VERSION must be defined")
	}
	if value != "v0.0.1" {
		t.Fatalf("GOLANGCI_LINT_MODULARITY_VERSION default = %#v, want v0.0.1", value)
	}
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
