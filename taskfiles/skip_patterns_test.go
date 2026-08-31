// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskfiles_test

import (
	"bytes"
	"io/fs"
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
	skipPatternModule struct {
		name string
		vars []string
	}

	sharedSkipFileMatcherCase struct {
		name     string
		pattern  string
		paths    []string
		retained []string
	}

	skipPatternVariableCheck struct {
		variable        string
		taskfile        *tasktest.Taskfile
		taskfileContent string
		readmeContent   string
	}

	skipPatternVariablesCheck struct {
		taskfile        *tasktest.Taskfile
		taskfileContent string
		readmeContent   string
		vars            []string
	}

	skipPatternVariantParityCheck struct {
		root      string
		family    string
		variables []string
	}

	sharedSkipFileMatcher struct {
		root   string
		filter string
		test   sharedSkipFileMatcherCase
	}
)

const (
	depcheckLintSkipVar = "DEPCHECK_LINT_SKIP_PATTERN"
	eslintLintSkipVar   = "ESLINT_LINT_SKIP_PATTERN"
	htmlhintLintSkipVar = "HTMLHINT_LINT_SKIP_PATTERN"
	spectralLintSkipVar = "SPECTRAL_LINT_SKIP_PATTERN"
	taskfileFlag        = "--taskfile"
	skipTaskfileYML     = "Taskfile.yml"

	taskfilesDirName = "taskfiles"

	ansibleLintModule = "ansible-lint"
	yamlfixModule     = "yamlfix"

	depcheckFamily  = "depcheck"
	eslintFamily    = "eslint"
	htmlhintFamily  = "htmlhint"
	biomeFamily     = "biome"
	knipFamily      = "knip"
	prettierFamily  = "prettier"
	spectralFamily  = "spectral"
	stylelintFamily = "stylelint"

	internalDirName  = "internal"
	skipfilesDirName = "skipfiles"

	cmdMainGoPath    = "cmd/main.go"
	generatedGlob    = "**/generated/**"
	srcMainGoPath    = "src/main.go"
	srcToolsAGoPath  = "src/tools/a.go"
	srcMainGoWinPath = `src\main.go`
	globStar         = "*"

	nullSeparator = "\x00"

	taskCommandName = "task"
	silentFlag      = "--silent"

	bunRuntime  = "bun"
	nodeRuntime = "node"

	prepareOverlayTaskPrefix = "prepare-overlay:"
	cleanupTaskPrefix        = "cleanup:"

	constZero   = 0
	constOne    = 1
	constTwo    = 2
	constTwenty = 20

	underscoreChar = "_"
	hyphenChar     = "-"

	windowsSeparator = "\n"
	crlfSeparator    = "\r\n"
)

// TestSkipPatternContract
func TestSkipPatternContract(t *testing.T) {
	t.Parallel()

	if len(skipPatternModules()) != constTwenty {
		t.Fatalf(
			"skip-pattern module count = %d, want %d",
			len(skipPatternModules()),
			constTwenty,
		)
	}

	root := tasktest.RepoRoot(t)

	for i := range skipPatternModules() {
		module := skipPatternModules()[i]
		t.Run(module.name, func(t *testing.T) {
			t.Parallel()
			assertSkipPatternModule(t, root, &module)
		})
	}
}

// TestSkipPatternVariantParity
func TestSkipPatternVariantParity(t *testing.T) {
	t.Parallel()

	families := skipPatternVariantParityFamilies()

	root := tasktest.RepoRoot(t)

	for family := range families {
		variables := families[family]
		t.Run(family, func(t *testing.T) {
			t.Parallel()
			assertSkipPatternVariantParity(t, &skipPatternVariantParityCheck{
				root: root, family: family, variables: variables,
			})
		})
	}
}

// TestSharedSkipFileMatcher
func TestSharedSkipFileMatcher(t *testing.T) {
	t.Parallel()

	root := tasktest.RepoRoot(t)
	filter := sharedSkipFileMatcherFilter(root)

	for i := range sharedSkipFileMatcherCases() {
		test := sharedSkipFileMatcherCases()[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertSharedSkipFileMatcher(t, &sharedSkipFileMatcher{
				root:   root,
				filter: filter,
				test:   test,
			})
		})
	}
}

// TestSharedSkipfilesTaskfileContract
func TestSharedSkipfilesTaskfileContract(t *testing.T) {
	t.Parallel()

	root := tasktest.RepoRoot(t)
	helperDirectory := filepath.Join(root, taskfilesDirName, internalDirName, skipfilesDirName)
	assertSharedSkipfilesDirectory(t, helperDirectory)
}

func skipPatternModules() []skipPatternModule {
	return append(skipPatternFlatModules(), skipPatternVariantModules()...)
}

func skipPatternFlatModules() []skipPatternModule {
	return append(skipPatternFlatModulesA(), skipPatternFlatModulesB()...)
}

func skipPatternFlatModulesA() []skipPatternModule {
	return []skipPatternModule{
		{name: ansibleLintModule, vars: []string{"ANSIBLE_LINT_LINT_SKIP_PATTERN"}},
		{name: "djlint", vars: []string{"DJLINT_LINT_SKIP_PATTERN", "DJLINT_FMT_SKIP_PATTERN"}},
	}
}

func skipPatternFlatModulesB() []skipPatternModule {
	return []skipPatternModule{
		{name: "rumdl", vars: []string{"RUMDL_LINT_SKIP_PATTERN", "RUMDL_FMT_SKIP_PATTERN"}},
		{name: yamlfixModule, vars: []string{"YAMLFIX_FMT_SKIP_PATTERN"}},
	}
}

func skipPatternVariantModules() []skipPatternModule {
	return slices.Concat(
		skipPatternDepcheckAndEslintModules(),
		skipPatternHtmlhintModules(),
		skipPatternSpectralModules(),
	)
}

func skipPatternDepcheckAndEslintModules() []skipPatternModule {
	return []skipPatternModule{
		{name: "depcheck/bun", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/npm", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/pnpm", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/yarn", vars: []string{depcheckLintSkipVar}},
		{name: "eslint/bun", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/npm", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/pnpm", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/yarn", vars: []string{eslintLintSkipVar}},
	}
}

func skipPatternHtmlhintModules() []skipPatternModule {
	return skipPatternSingleVarModules([]string{
		"htmlhint/bun",
		"htmlhint/node/npm",
		"htmlhint/node/pnpm",
		"htmlhint/node/yarn",
	}, htmlhintLintSkipVar)
}

func skipPatternSingleVarModules(names []string, skipVar string) []skipPatternModule {
	modules := make([]skipPatternModule, constZero, len(names))

	for index := range names {
		modules = append(modules, skipPatternModule{
			name: names[index],
			vars: []string{skipVar},
		})
	}

	return modules
}

func skipPatternSpectralModules() []skipPatternModule {
	return skipPatternSingleVarModules(
		[]string{
			"spectral/bun",
			"spectral/node/npm",
			"spectral/node/pnpm",
			"spectral/node/yarn",
		},
		spectralLintSkipVar,
	)
}

// skipPatternDocModule resolves the module whose README documents a skip
// pattern variable. Tool families keep a single README at the tool root;
// flat modules keep their own.
func skipPatternDocModule(name string) string {
	if before, _, ok := strings.Cut(name, "/"); ok {
		return before
	}

	return name
}

func assertSkipPatternModule(t *testing.T, root string, module *skipPatternModule) {
	t.Helper()

	taskfile := tasktest.LoadTaskfile(t, module.name)
	taskfileContent := readFile(
		t,
		filepath.Join(root, taskfilesDirName, module.name, skipTaskfileYML),
	)
	readmeContent := readFile(
		t,
		filepath.Join(root, taskfilesDirName, skipPatternDocModule(module.name), "README.md"),
	)

	assertSkipPatternVariables(t, &skipPatternVariablesCheck{
		vars:            module.vars,
		taskfile:        taskfile,
		taskfileContent: taskfileContent,
		readmeContent:   readmeContent,
	})
}

func assertSkipPatternVariables(t *testing.T, check *skipPatternVariablesCheck) {
	t.Helper()

	for i := range check.vars {
		variable := check.vars[i]
		assertSkipPatternVariable(t, &skipPatternVariableCheck{
			variable:        variable,
			taskfile:        check.taskfile,
			taskfileContent: check.taskfileContent,
			readmeContent:   check.readmeContent,
		})
	}
}

func assertSkipPatternVariable(t *testing.T, check *skipPatternVariableCheck) {
	t.Helper()

	value, exists := check.taskfile.Vars[check.variable]

	switch {
	case !exists:
		t.Errorf("%s is not defined", check.variable)
	case value != "":
		t.Errorf("%s default = %#v, want empty", check.variable, value)
	default:
	}

	assertSkipPatternVariableUsage(t, check)
}

func assertSkipPatternVariableUsage(t *testing.T, check *skipPatternVariableCheck) {
	t.Helper()

	if strings.Count(check.taskfileContent, check.variable) < constTwo {
		t.Errorf("%s is declared but not used by a task", check.variable)
	}

	if !strings.Contains(check.readmeContent, "`"+check.variable+"`") {
		t.Errorf("README does not document %s", check.variable)
	}
}

func skipPatternVariantParityFamilies() map[string][]string {
	return map[string][]string{
		depcheckFamily: {depcheckLintSkipVar},
		eslintFamily:   {eslintLintSkipVar},
		htmlhintFamily: {htmlhintLintSkipVar},
		spectralFamily: {spectralLintSkipVar},
	}
}

func assertSkipPatternVariantParity(t *testing.T, check *skipPatternVariantParityCheck) {
	t.Helper()

	paths := variantLeaves(t, check.root, check.family)

	if len(paths) < constTwo {
		t.Fatalf("found %d variants, want at least 2", len(paths))
	}

	for i := range check.variables {
		variable := check.variables[i]
		assertVariantUsageParity(t, paths, variable)
	}
}

func assertVariantUsageParity(t *testing.T, paths []string, variable string) {
	t.Helper()

	want := strings.Count(readFile(t, paths[constZero]), variable)

	for i := range paths[constOne:] {
		path := paths[constOne:][i]

		if got := strings.Count(readFile(t, path), variable); got != want {
			t.Errorf(
				"%s uses %s %d times, want %d",
				filepath.Base(filepath.Dir(path)),
				variable,
				got,
				want,
			)
		}
	}
}

func sharedSkipFileMatcherFilter(root string) string {
	return filepath.Join(
		root,
		taskfilesDirName,
		internalDirName,
		skipfilesDirName,
		skipTaskfileYML,
	)
}

func sharedSkipFileMatcherCasesA() []sharedSkipFileMatcherCase {
	return []sharedSkipFileMatcherCase{
		{
			name:     "single star stays in segment",
			pattern:  "*.go",
			paths:    []string{"main.go", cmdMainGoPath},
			retained: []string{cmdMainGoPath},
		},
		{
			name:     "double star crosses directories",
			pattern:  generatedGlob,
			paths:    []string{"generated/a.go", "src/generated/a.go", srcMainGoPath},
			retained: []string{srcMainGoPath},
		},
	}
}

func sharedSkipFileMatcherCasesB() []sharedSkipFileMatcherCase {
	return []sharedSkipFileMatcherCase{
		{
			name:     "question mark and spaces",
			pattern:  "src/?ock/*.go",
			paths:    []string{"src/mock/a.go", "src/lock/file with space.go", srcToolsAGoPath},
			retained: []string{srcToolsAGoPath},
		},
		{
			name:     "windows separators normalize",
			pattern:  generatedGlob,
			paths:    []string{`src\generated\a.go`, srcMainGoWinPath},
			retained: []string{srcMainGoWinPath},
		},
	}
}

func sharedSkipFileMatcherCases() []sharedSkipFileMatcherCase {
	return append(sharedSkipFileMatcherCasesA(), sharedSkipFileMatcherCasesB()...)
}

func assertSharedSkipFileMatcher(t *testing.T, matcher *sharedSkipFileMatcher) {
	t.Helper()

	separator := filterSeparator()
	output := runFilterCommand(t, matcher, separator)
	actual := parseFilterOutput(output, separator)

	if strings.Join(actual, nullSeparator) != strings.Join(matcher.test.retained, nullSeparator) {
		t.Fatalf("retained paths = %q, want %q", actual, matcher.test.retained)
	}
}

func filterSeparator() string {
	if runtime.GOOS == "windows" {
		return windowsSeparator
	}

	return nullSeparator
}

func runFilterCommand(t *testing.T, matcher *sharedSkipFileMatcher, separator string) []byte {
	t.Helper()

	input := []byte(strings.Join(matcher.test.paths, separator) + separator)
	commandContext := exec.CommandContext
	command := commandContext(
		t.Context(),
		taskCommandName, silentFlag, taskfileFlag, matcher.filter, "filter",
		"SKIPFILES_PATTERN="+matcher.test.pattern,
	)

	command.Dir = matcher.root
	command.Stdin = bytes.NewReader(input)

	output, err := command.Output()
	if err != nil {
		t.Fatalf("run filter: %v", err)
	}

	return output
}

func parseFilterOutput(output []byte, separator string) []string {
	if len(output) == constZero {
		return nil
	}

	outputText := strings.ReplaceAll(string(output), crlfSeparator, windowsSeparator)

	return strings.Split(strings.TrimSuffix(outputText, separator), separator)
}

func assertSharedSkipfilesDirectory(t *testing.T, helperDirectory string) {
	t.Helper()

	entries := readSkipfilesDirEntries(t, helperDirectory)
	assertOnlySharedTaskfile(t, entries)
	assertSharedSkipfilesTaskfileTrimmed(t, helperDirectory)
}

func readSkipfilesDirEntries(t *testing.T, helperDirectory string) []os.DirEntry {
	t.Helper()

	entries, err := os.ReadDir(helperDirectory)
	if err != nil {
		t.Fatalf("read shared skipfiles directory: %v", err)
	}

	return entries
}

func assertOnlySharedTaskfile(t *testing.T, entries []os.DirEntry) {
	t.Helper()

	if len(entries) == constOne && entries[constZero].Name() == skipTaskfileYML {
		return
	}

	names := make([]string, constZero, len(entries))

	for i := range entries {
		entry := entries[i]

		names = append(names, entry.Name())
	}

	t.Fatalf("shared skipfiles directory contains %v, want only Taskfile.yml", names)
}

func assertSharedSkipfilesTaskfileTrimmed(t *testing.T, helperDirectory string) {
	t.Helper()

	content := readFile(t, filepath.Join(helperDirectory, skipTaskfileYML))

	for i := range []string{prepareOverlayTaskPrefix, cleanupTaskPrefix} {
		removed := []string{prepareOverlayTaskPrefix, cleanupTaskPrefix}[i]

		if strings.Contains(content, removed) {
			t.Errorf("shared skipfiles Taskfile still defines %s", removed)
		}
	}
}

// variantLeaves returns every concrete leaf Taskfile of a nested tool family
// (the bun leaf plus each node/<pm> leaf), excluding aggregators.
func variantLeafPatterns(root, family string) []string {
	return []string{
		filepath.Join(root, taskfilesDirName, family, bunRuntime, skipTaskfileYML),
		filepath.Join(
			root,
			taskfilesDirName,
			family,
			nodeRuntime,
			globStar,
			skipTaskfileYML,
		),
	}
}

func variantLeaves(t *testing.T, root, family string) []string {
	t.Helper()

	var paths []string

	patterns := variantLeafPatterns(root, family)

	for i := range patterns {
		matches, err := filepath.Glob(patterns[i])
		if err != nil {
			t.Fatalf("glob variant leaves: %v", err)
		}

		paths = append(paths, matches...)
	}

	return paths
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return string(content)
}
