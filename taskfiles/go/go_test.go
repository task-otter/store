package go_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

const (
	constGoTestGolangciLintFmtCheck          = "golangci-lint:fmt:check"
	constGoTestGolangciLintLint              = "golangci-lint:lint"
	constGoTestGosecLint                     = "gosec:lint"
	constGoTestGovulncheckLint               = "govulncheck:lint"
	constGoTestInstall                       = "install"
	constGoTestInstallGoJunitReport          = "install:go-junit-report"
	constGoTestInstallGolangciLint           = "install:golangci-lint"
	constGoTestInstallGosec                  = "install:gosec"
	constGoTestInstallGovulncheck            = "install:govulncheck"
	constGoTestTest                          = "test"
	constGoTestGOLANGCILINTCOVERLINTVERSION  = "GOLANGCI_LINT_COVERLINT_VERSION"
	constGoTestGOLANGCILINTMODULARITYVERSION = "GOLANGCI_LINT_MODULARITY_VERSION"
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
		constGoTestGOLANGCILINTCOVERLINTVERSION,
		"GOLANGCI_LINT_VERSION",
		constGoTestGOLANGCILINTMODULARITYVERSION,
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

func TestGolangciLintInstallerBuildsCustomPlugins(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "go")

	task, ok := taskfile.Tasks[constGoTestInstallGolangciLint]
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
		"github.com/gostafa/coverlint",
		constGoTestGOLANGCILINTCOVERLINTVERSION,
		constGoTestGOLANGCILINTMODULARITYVERSION,
		"golangci-lint.coverlint.version",
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
		"golangci-lint.coverlint.version",
		"golangci-lint.modularity.version",
		"linters --config",
		"taskotter-golangci-status",
		"coverlint",
		"modularity",
		constGoTestGOLANGCILINTCOVERLINTVERSION,
		constGoTestGOLANGCILINTMODULARITYVERSION,
	} {
		if !strings.Contains(status, token) {
			t.Fatalf("install:golangci-lint status missing %q: %s", token, status)
		}
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

	for _, name := range []string{
		constGoTestGOLANGCILINTCOVERLINTVERSION,
		constGoTestGOLANGCILINTMODULARITYVERSION,
	} {
		value, exists := taskfile.Vars[name]
		if !exists {
			t.Fatalf("%s must be defined", name)
		}

		if value != "v0.0.1" {
			t.Fatalf("%s default = %#v, want v0.0.1", name, value)
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
