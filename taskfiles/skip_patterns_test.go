// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskfiles_test

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
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

	knipMergeCase struct {
		name     string
		file     string
		content  string
		overlay  string
		expected string
	}

	actionlintSkipFixture struct {
		taskfilePath string
		project      string
		logPath      string
		skippedPath  string
		cliGoodPath  string
		env          []string
	}

	cargoSkipFixture struct {
		taskfilePath string
		project      string
		logPath      string
		env          []string
	}

	govulncheckSkipFixture struct {
		taskfilePath string
		project      string
		logPath      string
		env          []string
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

	skipFixtureCommand struct {
		project      string
		env          []string
		taskfilePath string
		args         []string
	}

	actionlintFixtureDirs struct {
		binDirectory         string
		workflowDirectory    string
		generatedDirectory   string
		cliWorkflowDirectory string
	}

	actionlintFixturePaths struct {
		skippedPath string
		cliGoodPath string
	}

	cargoFixtureDirs struct {
		project                  string
		binDirectory             string
		goodSourceDirectory      string
		generatedDirectory       string
		generatedSourceDirectory string
	}

	govulncheckFixtureDirs struct {
		project            string
		binDirectory       string
		goodDirectory      string
		generatedDirectory string
	}

	fileWrite struct {
		content string
		perm    os.FileMode
	}
)

const (
	biomeLintSkipVar     = "BIOME_LINT_SKIP_PATTERN"
	biomeFmtSkipVar      = "BIOME_FMT_SKIP_PATTERN"
	depcheckLintSkipVar  = "DEPCHECK_LINT_SKIP_PATTERN"
	eslintLintSkipVar    = "ESLINT_LINT_SKIP_PATTERN"
	htmlhintLintSkipVar  = "HTMLHINT_LINT_SKIP_PATTERN"
	knipLintSkipVar      = "KNIP_LINT_SKIP_PATTERN"
	prettierFmtSkipVar   = "PRETTIER_FMT_SKIP_PATTERN"
	spectralLintSkipVar  = "SPECTRAL_LINT_SKIP_PATTERN"
	sqlfluffModule       = "sqlfluff"
	stylelintLintSkipVar = "STYLELINT_LINT_SKIP_PATTERN"
	taskfileFlag         = "--taskfile"
	skipTaskfileYML      = "Taskfile.yml"

	taskfilesDirName = "taskfiles"

	actionlintModule          = "actionlint"
	ansibleModule             = "ansible"
	bufModule                 = "buf"
	cargoModule               = "cargo"
	dotenvLinterModule        = "dotenv-linter"
	golangciLintModule        = "golangci-lint"
	govulncheckModule         = "govulncheck"
	hadolintModule            = "hadolint"
	jsonlintModule            = "jsonlint"
	protolintModule           = "protolint"
	shellcheckModule          = "shellcheck"
	biomeConfigSkipOverlay    = ".taskotter-biome-bun-skip.json"
	sqlfluffConfigSkipOverlay = ".taskotter-sqlfluff-skip.cfg"
	shfmtModule               = "shfmt"
	yamllintModule            = "yamllint"
	zizmorModule              = "zizmor"

	biomeBunModule         = "biome/bun"
	biomeNodeFnmNpmModule  = "biome/node/fnm/npm"
	biomeNodeNvmNpmModule  = "biome/node/nvm/npm"
	biomeNodeFnmPnpmModule = "biome/node/fnm/pnpm"
	biomeNodeNvmPnpmModule = "biome/node/nvm/pnpm"
	biomeNodeFnmYarnModule = "biome/node/fnm/yarn"
	biomeNodeNvmYarnModule = "biome/node/nvm/yarn"

	knipBunModule         = "knip/bun"
	knipNodeFnmNpmModule  = "knip/node/fnm/npm"
	knipNodeNvmNpmModule  = "knip/node/nvm/npm"
	knipNodeFnmPnpmModule = "knip/node/fnm/pnpm"
	knipNodeNvmPnpmModule = "knip/node/nvm/pnpm"
	knipNodeFnmYarnModule = "knip/node/fnm/yarn"
	knipNodeNvmYarnModule = "knip/node/nvm/yarn"

	biomeFamily     = "biome"
	depcheckFamily  = "depcheck"
	eslintFamily    = "eslint"
	htmlhintFamily  = "htmlhint"
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

	taskCommandName    = "task"
	silentFlag         = "--silent"
	configSkipTaskName = "config:skip"
	yesFlag            = "--yes"
	lintTaskName       = "lint"
	onelineFlag        = "-oneline"
	pathEnvVar         = "PATH"

	writeFileErrFormat = "write %s: %v"

	biomeLintGeneratedPattern      = "BIOME_LINT_SKIP_PATTERN=**/generated/**"
	biomeFmtVendorPattern          = "BIOME_FMT_SKIP_PATTERN=**/vendor/**"
	negatedGeneratedPattern        = `"!**/generated/**"`
	knipLintGeneratedPattern       = "KNIP_LINT_SKIP_PATTERN=**/generated/**"
	golangciLintGeneratedPattern   = "GOLANGCI_LINT_LINT_SKIP_PATTERN=**/generated/**"
	actionlintLintGeneratedPattern = "ACTIONLINT_LINT_SKIP_PATTERN=**/generated/**"

	staleOverlayRemovedCase = "no pattern removes a stale overlay"
	staleOverlayRemovedMsg  = "empty skip pattern did not remove the stale overlay"
	noProjectConfigCase     = "no project config"

	bunRuntime  = "bun"
	nodeRuntime = "node"

	generatedDirName    = "generated"
	generatedBadYmlPath = "generated/bad.yml"

	shortModeSkipMsg = "skipping task integration test in short mode"

	binDirName          = "bin"
	goodWorkflowFile    = "good workflow.yml"
	cliGoodWorkflowFile = "cli good.yml"

	perm0644 = 0o644
	perm0500 = 0o500

	goodDirName   = "good"
	srcDirName    = "src"
	cargoTomlFile = "Cargo.toml"
	libRsFile     = "lib.rs"

	githubDirName            = ".github"
	workflowsDirName         = "workflows"
	generatedPackageDirName  = "generated package"
	prepareOverlayTaskPrefix = "prepare-overlay:"
	cleanupTaskPrefix        = "cleanup:"
	filterShFile             = "filter.sh"
	filterSkipFilesPs1File   = "Filter-SkipFiles.ps1"
	prepareOverlayShFile     = "prepare-overlay.sh"
	prepareOverlayPs1File    = "Prepare-Overlay.ps1"
	knipConfigMjsFile        = "knip-config.mjs"

	constZero       = 0
	constOne        = 1
	constTwo        = 2
	constThree      = 3
	constFifteen    = 15
	constSixteen    = 16
	constSixtySeven = 67
	perm0600        = 0o600

	constDecimalBase   = 10
	constUint32BitSize = 32
	perm0700           = 0o700

	windowsSeparator = "\n"
	crlfSeparator    = "\r\n"
	staleFixtureBody = "stale\n"
)

func skipPatternModules() []skipPatternModule {
	return append(skipPatternFlatModules(), skipPatternVariantModules()...)
}

func skipPatternFlatModules() []skipPatternModule {
	return append(skipPatternFlatModulesA(), skipPatternFlatModulesB()...)
}

func skipPatternFlatModulesA() []skipPatternModule {
	return []skipPatternModule{
		{name: actionlintModule, vars: []string{"ACTIONLINT_LINT_SKIP_PATTERN"}},
		{name: ansibleModule, vars: []string{"ANSIBLE_LINT_SKIP_PATTERN"}},
		{name: bufModule, vars: []string{"BUF_LINT_SKIP_PATTERN", "BUF_FMT_SKIP_PATTERN"}},
		{name: cargoModule, vars: []string{"CARGO_LINT_SKIP_PATTERN", "CARGO_FMT_SKIP_PATTERN"}},
		{name: "djlint", vars: []string{"DJLINT_LINT_SKIP_PATTERN", "DJLINT_FMT_SKIP_PATTERN"}},
		{name: dotenvLinterModule, vars: []string{"DOTENV_LINTER_LINT_SKIP_PATTERN"}},
		{
			name: golangciLintModule,
			vars: []string{"GOLANGCI_LINT_LINT_SKIP_PATTERN", "GOLANGCI_LINT_FMT_SKIP_PATTERN"},
		},
		{name: govulncheckModule, vars: []string{"GOVULNCHECK_LINT_SKIP_PATTERN"}},
	}
}

func skipPatternFlatModulesB() []skipPatternModule {
	return []skipPatternModule{
		{name: hadolintModule, vars: []string{"HADOLINT_LINT_SKIP_PATTERN"}},
		{name: jsonlintModule, vars: []string{"JSONLINT_LINT_SKIP_PATTERN"}},
		{name: protolintModule, vars: []string{"PROTOLINT_LINT_SKIP_PATTERN"}},
		{name: "rumdl", vars: []string{"RUMDL_LINT_SKIP_PATTERN", "RUMDL_FMT_SKIP_PATTERN"}},
		{name: shellcheckModule, vars: []string{"SHELLCHECK_LINT_SKIP_PATTERN"}},
		{name: shfmtModule, vars: []string{"SHFMT_FMT_SKIP_PATTERN"}},
		{name: sqlfluffModule, vars: []string{"SQLFLUFF_LINT_SKIP_PATTERN"}},
		{name: yamllintModule, vars: []string{"YAMLLINT_LINT_SKIP_PATTERN"}},
		{name: zizmorModule, vars: []string{"ZIZMOR_LINT_SKIP_PATTERN"}},
	}
}

func skipPatternVariantModules() []skipPatternModule {
	return slices.Concat(
		skipPatternBiomeModules(),
		skipPatternDepcheckAndEslintModules(),
		skipPatternHtmlhintAndKnipModules(),
		skipPatternPrettierModules(),
		skipPatternSpectralAndStylelintModules(),
	)
}

func skipPatternBiomeModules() []skipPatternModule {
	return []skipPatternModule{
		{name: biomeBunModule, vars: []string{biomeLintSkipVar, biomeFmtSkipVar}},
		{name: biomeNodeFnmNpmModule, vars: []string{biomeLintSkipVar, biomeFmtSkipVar}},
		{name: biomeNodeNvmNpmModule, vars: []string{biomeLintSkipVar, biomeFmtSkipVar}},
		{name: biomeNodeFnmPnpmModule, vars: []string{biomeLintSkipVar, biomeFmtSkipVar}},
		{name: biomeNodeNvmPnpmModule, vars: []string{biomeLintSkipVar, biomeFmtSkipVar}},
		{name: biomeNodeFnmYarnModule, vars: []string{biomeLintSkipVar, biomeFmtSkipVar}},
		{name: biomeNodeNvmYarnModule, vars: []string{biomeLintSkipVar, biomeFmtSkipVar}},
	}
}

func skipPatternDepcheckAndEslintModules() []skipPatternModule {
	return []skipPatternModule{
		{name: "depcheck/bun", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/fnm/npm", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/nvm/npm", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/fnm/pnpm", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/nvm/pnpm", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/fnm/yarn", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/nvm/yarn", vars: []string{depcheckLintSkipVar}},
		{name: "eslint/bun", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/fnm/npm", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/nvm/npm", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/fnm/pnpm", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/nvm/pnpm", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/fnm/yarn", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/nvm/yarn", vars: []string{eslintLintSkipVar}},
	}
}

func skipPatternHtmlhintAndKnipModules() []skipPatternModule {
	return []skipPatternModule{
		{name: "htmlhint/node/fnm/npm", vars: []string{htmlhintLintSkipVar}},
		{name: "htmlhint/node/nvm/npm", vars: []string{htmlhintLintSkipVar}},
		{name: "htmlhint/node/fnm/pnpm", vars: []string{htmlhintLintSkipVar}},
		{name: "htmlhint/node/nvm/pnpm", vars: []string{htmlhintLintSkipVar}},
		{name: knipBunModule, vars: []string{knipLintSkipVar}},
		{name: knipNodeFnmNpmModule, vars: []string{knipLintSkipVar}},
		{name: knipNodeNvmNpmModule, vars: []string{knipLintSkipVar}},
		{name: knipNodeFnmPnpmModule, vars: []string{knipLintSkipVar}},
		{name: knipNodeNvmPnpmModule, vars: []string{knipLintSkipVar}},
		{name: knipNodeFnmYarnModule, vars: []string{knipLintSkipVar}},
		{name: knipNodeNvmYarnModule, vars: []string{knipLintSkipVar}},
	}
}

func skipPatternPrettierModules() []skipPatternModule {
	return []skipPatternModule{
		{name: "prettier/bun", vars: []string{prettierFmtSkipVar}},
		{name: "prettier/node/fnm/npm", vars: []string{prettierFmtSkipVar}},
		{name: "prettier/node/nvm/npm", vars: []string{prettierFmtSkipVar}},
		{name: "prettier/node/fnm/pnpm", vars: []string{prettierFmtSkipVar}},
		{name: "prettier/node/nvm/pnpm", vars: []string{prettierFmtSkipVar}},
		{name: "prettier/node/fnm/yarn", vars: []string{prettierFmtSkipVar}},
		{name: "prettier/node/nvm/yarn", vars: []string{prettierFmtSkipVar}},
	}
}

func skipPatternSpectralAndStylelintModules() []skipPatternModule {
	return []skipPatternModule{
		{name: "spectral/node/fnm/npm", vars: []string{spectralLintSkipVar}},
		{name: "spectral/node/nvm/npm", vars: []string{spectralLintSkipVar}},
		{name: "spectral/node/fnm/pnpm", vars: []string{spectralLintSkipVar}},
		{name: "spectral/node/nvm/pnpm", vars: []string{spectralLintSkipVar}},
		{name: "stylelint/bun", vars: []string{stylelintLintSkipVar}},
		{name: "stylelint/node/fnm/npm", vars: []string{stylelintLintSkipVar}},
		{name: "stylelint/node/nvm/npm", vars: []string{stylelintLintSkipVar}},
		{name: "stylelint/node/fnm/pnpm", vars: []string{stylelintLintSkipVar}},
		{name: "stylelint/node/nvm/pnpm", vars: []string{stylelintLintSkipVar}},
		{name: "stylelint/node/fnm/yarn", vars: []string{stylelintLintSkipVar}},
		{name: "stylelint/node/nvm/yarn", vars: []string{stylelintLintSkipVar}},
	}
}

// Modules that still delegate to the shared skipfiles helper. Biome and Knip
// dropped out when their config overlays moved into their own config:skip
// tasks; SQLFluff, golangci-lint, and govulncheck stay because they still use
// filter / go-packages.
func sharedSkipfilesConsumers() []string {
	return []string{
		actionlintModule,
		ansibleModule,
		bufModule,
		cargoModule,
		dotenvLinterModule,
		golangciLintModule,
		govulncheckModule,
		hadolintModule,
		jsonlintModule,
		protolintModule,
		shellcheckModule,
		shfmtModule,
		sqlfluffModule,
		yamllintModule,
		zizmorModule,
	}
}

// Modules that own their config overlay through a local config:skip task.
func configSkipModules() []string {
	return []string{
		biomeBunModule,
		biomeNodeFnmNpmModule,
		biomeNodeNvmNpmModule,
		biomeNodeFnmPnpmModule,
		biomeNodeNvmPnpmModule,
		biomeNodeFnmYarnModule,
		biomeNodeNvmYarnModule,
		golangciLintModule,
		knipBunModule,
		knipNodeFnmNpmModule,
		knipNodeNvmNpmModule,
		knipNodeFnmPnpmModule,
		knipNodeNvmPnpmModule,
		knipNodeFnmYarnModule,
		knipNodeNvmYarnModule,
		sqlfluffModule,
	}
}

// TestSkipPatternContract
func TestSkipPatternContract(t *testing.T) {
	t.Parallel()

	if len(skipPatternModules()) != constSixtySeven {
		t.Fatalf(
			"skip-pattern module count = %d, want %d",
			len(skipPatternModules()),
			constSixtySeven,
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
		biomeFamily:     {biomeLintSkipVar, biomeFmtSkipVar},
		depcheckFamily:  {depcheckLintSkipVar},
		eslintFamily:    {eslintLintSkipVar},
		htmlhintFamily:  {htmlhintLintSkipVar},
		knipFamily:      {knipLintSkipVar},
		prettierFamily:  {prettierFmtSkipVar},
		spectralFamily:  {spectralLintSkipVar},
		stylelintFamily: {stylelintLintSkipVar},
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

// runConfigSkip runs a module's config:skip task inside project, which is the
// USER_WORKING_DIR the task writes its overlay relative to.
func runConfigSkip(t *testing.T, project string, moduleAndVars ...string) {
	t.Helper()

	if len(moduleAndVars) == constZero {
		t.Fatal("module is required")
	}

	root := tasktest.RepoRoot(t)
	module := moduleAndVars[constZero]
	vars := moduleAndVars[constOne:]
	arguments := append([]string{
		silentFlag, taskfileFlag,
		filepath.Join(root, taskfilesDirName, module, skipTaskfileYML),
		configSkipTaskName,
	}, vars...)
	runCommand(t, project, arguments...)
}

func assertPathDoesNotExist(t *testing.T, path, message string) {
	t.Helper()

	info, err := os.Stat(path)

	if err == nil && info != nil {
		t.Fatal(message)
	}

	if !os.IsNotExist(err) {
		t.Fatal(message)
	}
}

func assertStaleOverlayRemoved(t *testing.T, overlay string, moduleAndVars ...string) {
	t.Helper()

	project := t.TempDir()
	writeFixture(t, project, overlay, staleFixtureBody)
	runConfigSkip(t, project, moduleAndVars...)
	assertPathDoesNotExist(t, filepath.Join(project, overlay), staleOverlayRemovedMsg)
}

func newSkipFixtureCommand(input *skipFixtureCommand) *exec.Cmd {
	commandContext := exec.CommandContext
	command := commandContext(
		context.Background(),
		taskCommandName,
		append([]string{taskfileFlag, input.taskfilePath}, input.args...)...)

	command.Dir = input.project
	command.Env = input.env

	return command
}

// writeFixture creates a file inside a fresh project directory.
func writeFixture(t *testing.T, project string, fixture ...string) {
	t.Helper()

	if len(fixture) != constTwo {
		t.Fatalf("fixture requires name and content, got %d values", len(fixture))
	}

	name := fixture[constZero]
	content := fixture[constOne]

	err := os.WriteFile(filepath.Join(project, name), []byte(content), perm0600)
	if err != nil {
		t.Fatalf(writeFileErrFormat, name, err)
	}
}

// assertOverlayContains reads an overlay written into project and checks tokens.
func assertOverlayContains(t *testing.T, project string, nameAndTokens ...string) {
	t.Helper()

	if len(nameAndTokens) == constZero {
		t.Fatal("overlay name is required")
	}

	name := nameAndTokens[constZero]
	tokens := nameAndTokens[constOne:]
	content := readFile(t, filepath.Join(project, name))

	for i := range tokens {
		token := tokens[i]

		if !strings.Contains(content, token) {
			t.Fatalf("%s does not contain %q:\n%s", name, token, content)
		}
	}
}

// TestBiomeConfigSkipTask
func testBiomeConfigSkipBothPatterns(t *testing.T) {
	t.Helper()

	t.Run("both patterns", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		runConfigSkip(t, project, biomeBunModule,
			biomeLintGeneratedPattern,
			biomeFmtVendorPattern)
		assertOverlayContains(
			t, project, biomeConfigSkipOverlay, negatedGeneratedPattern, `"!**/vendor/**"`,
		)
	})
}

func assertBiomeLintScopeOverlayHasNoFmtPattern(t *testing.T, project string) {
	t.Helper()

	content := readFile(t, filepath.Join(project, biomeConfigSkipOverlay))

	if strings.Contains(content, "vendor") {
		t.Fatalf("lint-scoped overlay leaked the fmt pattern:\n%s", content)
	}
}

func testBiomeConfigSkipLintScope(t *testing.T) {
	t.Helper()

	t.Run("lint scope excludes fmt pattern", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		runConfigSkip(t, project, biomeBunModule, "SKIP_SCOPE=lint",
			biomeLintGeneratedPattern,
			biomeFmtVendorPattern)
		assertOverlayContains(t, project, biomeConfigSkipOverlay, negatedGeneratedPattern)
		assertBiomeLintScopeOverlayHasNoFmtPattern(t, project)
	})
}

func testBiomeConfigSkipExtendsDiscovered(t *testing.T) {
	t.Helper()

	t.Run("extends a discovered config", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		writeFixture(t, project, "biome.json", `{"linter":{"enabled":true}}`)
		runConfigSkip(t, project, biomeBunModule, biomeLintGeneratedPattern)
		assertOverlayContains(t, project, biomeConfigSkipOverlay, `"extends":["biome.json"]`)
	})
}

func testBiomeConfigSkipStaleOverlayRemoved(t *testing.T) {
	t.Helper()

	t.Run(staleOverlayRemovedCase, func(t *testing.T) {
		t.Parallel()
		assertStaleOverlayRemoved(t, biomeConfigSkipOverlay, biomeBunModule)
	})
}

// TestBiomeConfigSkipTask validates the behavior covered by this test case.
func TestBiomeConfigSkipTask(t *testing.T) {
	t.Parallel()

	testBiomeConfigSkipBothPatterns(t)
	testBiomeConfigSkipLintScope(t)
	testBiomeConfigSkipExtendsDiscovered(t)
	testBiomeConfigSkipStaleOverlayRemoved(t)
}

// jsRuntimeVar pins the Knip overlay generator to a JS runtime that exists on
// this machine. The bun module defaults to bun, which CI does not install.
func jsRuntimeVar(t *testing.T) string {
	t.Helper()

	for i := range []string{bunRuntime, nodeRuntime} {
		runtimeName := []string{bunRuntime, nodeRuntime}[i]

		_, err := exec.LookPath(runtimeName)
		if err == nil {
			return "KNIP_INTERNAL_JS_RUNTIME=" + runtimeName
		}
	}

	t.Skip("neither bun nor node is installed")

	return ""
}

// TestKnipConfigSkipTask
func TestKnipConfigSkipTask(t *testing.T) {
	t.Parallel()

	const overlay = ".taskotter-knip-bun-skip.json"

	testKnipConfigSkipNoProjectConfig(t, overlay)
	testKnipConfigSkipMergesJSONC(t, overlay)
	testKnipConfigSkipMergesPackageJSON(t, overlay)
	testKnipConfigSkipRejectsDynamicJS(t)
	testKnipConfigSkipRemovesStaleOverlay(t, overlay)
}

func testKnipConfigSkipNoProjectConfig(t *testing.T, overlay string) {
	t.Helper()

	t.Run(noProjectConfigCase, func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		runConfigSkip(
			t,
			project,
			knipBunModule,
			knipLintGeneratedPattern,
			jsRuntimeVar(t),
		)
		assertOverlayContains(t, project, overlay, generatedGlob)
	})
}

func testKnipConfigSkipMergesJSONC(t *testing.T, overlay string) {
	t.Helper()

	assertKnipConfigSkipMerges(t, &knipMergeCase{
		name:     "merges jsonc",
		file:     "knip.jsonc",
		content:  "{\n  // keep this entry\n  \"entry\": [\"src/index.ts\"],\n}\n",
		overlay:  overlay,
		expected: "src/index.ts",
	})
}

func testKnipConfigSkipMergesPackageJSON(t *testing.T, overlay string) {
	t.Helper()

	assertKnipConfigSkipMerges(t, &knipMergeCase{
		name:     "merges the package.json knip section",
		file:     "package.json",
		content:  `{"name":"fixture","knip":{"ignore":["existing/**"]}}`,
		overlay:  overlay,
		expected: "existing/**",
	})
}

func assertKnipConfigSkipMerges(t *testing.T, testCase *knipMergeCase) {
	t.Helper()

	t.Run(testCase.name, func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		writeFixture(t, project, testCase.file, testCase.content)
		runConfigSkip(
			t,
			project,
			knipBunModule,
			knipLintGeneratedPattern,
			jsRuntimeVar(t),
		)
		assertOverlayContains(t, project, testCase.overlay, testCase.expected, generatedGlob)
	})
}

func runKnipDynamicJSRejectionCommand(t *testing.T, project string) ([]byte, error) {
	t.Helper()

	command := exec.CommandContext(t.Context(), taskCommandName)

	command.Args = append(command.Args,
		silentFlag,
		taskfileFlag,
		knipDynamicJSTaskfilePath(t),
		configSkipTaskName,
		knipLintGeneratedPattern,
	)

	command.Dir = project

	output, err := command.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("run knip dynamic js rejection command: %w", err)
	}

	return output, nil
}

func knipDynamicJSTaskfilePath(t *testing.T) string {
	t.Helper()

	return filepath.Join(
		tasktest.RepoRoot(t),
		taskfilesDirName,
		knipFamily,
		bunRuntime,
		skipTaskfileYML,
	)
}

func testKnipConfigSkipRejectsDynamicJS(t *testing.T) {
	t.Helper()

	t.Run("rejects a dynamic JS config", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		writeFixture(t, project, "knip.config.js", "export default {};\n")

		output, err := runKnipDynamicJSRejectionCommand(t, project)

		if err == nil || !strings.Contains(string(output), "dynamic JS/TS Knip config") {
			t.Fatalf("dynamic Knip config was not rejected clearly: err=%v\n%s", err, output)
		}
	})
}

func testKnipConfigSkipRemovesStaleOverlay(t *testing.T, overlay string) {
	t.Helper()

	t.Run(staleOverlayRemovedCase, func(t *testing.T) {
		t.Parallel()
		assertStaleOverlayRemoved(t, overlay, knipBunModule)
	})
}

func testSQLFluffConfigSkipMerges(t *testing.T) {
	t.Helper()

	t.Run("merges ignore_paths and normalizes separators", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		writeFixture(t, project, "source.cfg",
			"[sqlfluff]\ndialect = postgres\nignore_paths = build/**\n")
		runConfigSkip(t, project, sqlfluffModule,
			`SQLFLUFF_LINT_SKIP_PATTERN=**\generated\**`, "CONFIG_OVERRIDE=source.cfg")
		assertOverlayContains(t, project, sqlfluffConfigSkipOverlay,
			"dialect = postgres", "ignore_paths = build/**,**/generated/**")
	})
}

func testSQLFluffConfigSkipNoProjectConfig(t *testing.T) {
	t.Helper()

	t.Run(noProjectConfigCase, func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		runConfigSkip(t, project, sqlfluffModule, "SQLFLUFF_LINT_SKIP_PATTERN=**/generated/**")
		assertOverlayContains(
			t, project, sqlfluffConfigSkipOverlay, "[sqlfluff]", "ignore_paths = **/generated/**",
		)
	})
}

func testSQLFluffConfigSkipStaleOverlayRemoved(t *testing.T) {
	t.Helper()

	t.Run(staleOverlayRemovedCase, func(t *testing.T) {
		t.Parallel()
		assertStaleOverlayRemoved(t, sqlfluffConfigSkipOverlay, sqlfluffModule)
	})
}

// TestSQLFluffConfigSkipTask
func TestSQLFluffConfigSkipTask(t *testing.T) {
	t.Parallel()

	testSQLFluffConfigSkipMerges(t)
	testSQLFluffConfigSkipNoProjectConfig(t)
	testSQLFluffConfigSkipStaleOverlayRemoved(t)
}

// TestGolangciLintConfigSkipTask
func TestGolangciLintConfigSkipTask(t *testing.T) {
	t.Parallel()

	const overlay = ".golangci-taskotter-skip.yml"

	t.Run("translates the glob into an exclusion regex", func(t *testing.T) {
		t.Parallel()
		assertGolangciLintOverlayTranslatesGlob(t, overlay)
	})

	t.Run("rewrites rather than accumulating overlays", func(t *testing.T) {
		t.Parallel()
		assertGolangciLintOverlayRewrites(t, overlay)
	})

	t.Run(staleOverlayRemovedCase, func(t *testing.T) {
		t.Parallel()
		assertGolangciLintOverlayRemovesStale(t, overlay)
	})
}

func assertGolangciLintOverlayTranslatesGlob(t *testing.T, overlay string) {
	t.Helper()

	project := t.TempDir()
	runConfigSkip(t, project, golangciLintModule, golangciLintGeneratedPattern)
	assertOverlayContains(t, project, overlay,
		"linters:", "exclusions:", "paths:", `^(?:.*/)?generated/.*$`)
}

func assertGolangciLintOverlayRewrites(t *testing.T, overlay string) {
	t.Helper()

	project := t.TempDir()
	runConfigSkip(t, project, golangciLintModule, golangciLintGeneratedPattern)
	runConfigSkip(t, project, golangciLintModule, "GOLANGCI_LINT_LINT_SKIP_PATTERN=**/mocks/*.go")

	assertOverlayPatternReplaced(t, project, overlay)
	assertSingleOverlayFile(t, project)
}

func assertOverlayPatternReplaced(t *testing.T, project, overlay string) {
	t.Helper()

	content := readFile(t, filepath.Join(project, overlay))

	if strings.Contains(content, generatedDirName) {
		t.Fatalf("second run did not replace the first pattern:\n%s", content)
	}

	if !strings.Contains(content, `^(?:.*/)?mocks/[^/]*\.go$`) {
		t.Fatalf("second run did not write the new pattern:\n%s", content)
	}
}

func assertSingleOverlayFile(t *testing.T, project string) {
	t.Helper()

	entries, err := os.ReadDir(project)
	if err != nil {
		t.Fatalf("read project: %v", err)
	}

	if len(entries) != constOne {
		t.Fatalf("expected exactly one overlay file, found %d", len(entries))
	}
}

func assertGolangciLintOverlayRemovesStale(t *testing.T, overlay string) {
	t.Helper()

	assertStaleOverlayRemoved(t, overlay, golangciLintModule)
}

// TestSharedSkipfilesTaskfileContract
func TestSharedSkipfilesTaskfileContract(t *testing.T) {
	t.Parallel()

	root := tasktest.RepoRoot(t)
	helperDirectory := filepath.Join(root, taskfilesDirName, internalDirName, skipfilesDirName)
	assertSharedSkipfilesDirectory(t, helperDirectory)
	assertSharedSkipfilesConsumers(t, root)
	assertConfigSkipModules(t)
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

// assertSharedSkipfilesTaskfileTrimmed checks that the helper is down to
// filter and go-packages; overlay generation now lives in each tool's own
// config:skip task.
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

func assertSharedSkipfilesConsumers(t *testing.T, root string) {
	t.Helper()

	if len(sharedSkipfilesConsumers()) != constFifteen {
		t.Fatalf(
			"shared skipfiles consumer count = %d, want %d",
			len(sharedSkipfilesConsumers()),
			constFifteen,
		)
	}

	for i := range sharedSkipfilesConsumers() {
		module := sharedSkipfilesConsumers()[i]
		assertSharedSkipfilesConsumer(t, root, module)
	}
}

func assertNoRemovedSkipfilesHelpers(t *testing.T, module, content string) {
	t.Helper()

	removedHelpers := []string{
		filterShFile, filterSkipFilesPs1File, prepareOverlayShFile,
		prepareOverlayPs1File, knipConfigMjsFile,
	}

	for i := range removedHelpers {
		if strings.Contains(content, removedHelpers[i]) {
			t.Errorf("%s still references removed helper %s", module, removedHelpers[i])
		}
	}
}

func assertSharedSkipfilesConsumer(t *testing.T, root, module string) {
	t.Helper()

	content := readFile(t, filepath.Join(root, taskfilesDirName, module, skipTaskfileYML))

	missingSkipfiles := !strings.Contains(content, "internal/skipfiles/Taskfile.yml") ||
		!strings.Contains(content, "internal: true")

	if missingSkipfiles {
		t.Errorf("%s does not include the shared skipfiles Taskfile internally", module)
	}

	assertNoRemovedSkipfilesHelpers(t, module, content)
}

func assertConfigSkipModules(t *testing.T) {
	t.Helper()

	if len(configSkipModules()) != constSixteen {
		t.Fatalf("config:skip module count = %d, want %d", len(configSkipModules()), constSixteen)
	}

	for i := range configSkipModules() {
		module := configSkipModules()[i]
		assertConfigSkipModule(t, module)
	}
}

func assertConfigSkipModule(t *testing.T, module string) {
	t.Helper()

	taskfile := tasktest.LoadTaskfile(t, module)

	task, exists := taskfile.Tasks[configSkipTaskName]

	if !exists {
		t.Errorf("%s does not define a config:skip task", module)

		return
	}

	assertConfigSkipTaskShape(t, module, task)
}

func assertConfigSkipTaskShape(t *testing.T, module string, task *tasktest.Task) {
	t.Helper()

	if task.Internal {
		t.Errorf("%s config:skip is internal, want a public task", module)
	}

	// run: once would let one caller's overlay be reused by the next call in
	// the same run, which passes a different scope or pattern.
	if task.Run == "once" {
		t.Errorf("%s config:skip must not use run: once", module)
	}
}

// TestActionlintSkipPatternFiltersFiles
func TestActionlintSkipPatternFiltersFiles(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip(shortModeSkipMsg)
	}

	fixture := newActionlintSkipFixture(t)
	fixture.assertDefaultDiscovery(t)
	fixture.assertCliTargets(t)
	fixture.assertAllSkipped(t)
}

func actionlintSkipFixtureEnv(binDirectory, logPath string) []string {
	return append(os.Environ(),
		"PATH="+binDirectory+":"+os.Getenv(pathEnvVar),
		"TASKOTTER_ACTIONLINT_LOG="+logPath,
	)
}

func newActionlintSkipFixture(t *testing.T) actionlintSkipFixture {
	t.Helper()

	root := tasktest.RepoRoot(t)
	project := t.TempDir()
	dirs := newActionlintFixtureDirs(t, project)
	paths := writeActionlintFixtureFiles(t, &dirs)
	logPath := filepath.Join(project, "actionlint.args")

	writeActionlintStub(t, dirs.binDirectory)

	return actionlintSkipFixture{
		taskfilePath: filepath.Join(root, taskfilesDirName, actionlintModule, skipTaskfileYML),
		project:      project,
		logPath:      logPath,
		skippedPath:  paths.skippedPath,
		cliGoodPath:  paths.cliGoodPath,
		env:          actionlintSkipFixtureEnv(dirs.binDirectory, logPath),
	}
}

func newActionlintFixtureDirs(t *testing.T, project string) actionlintFixtureDirs {
	t.Helper()

	dirs := actionlintFixtureDirs{
		binDirectory:      filepath.Join(project, binDirName),
		workflowDirectory: filepath.Join(project, githubDirName, workflowsDirName),
		generatedDirectory: filepath.Join(
			project, githubDirName, workflowsDirName, generatedDirName,
		),
		cliWorkflowDirectory: filepath.Join(project, "custom workflows"),
	}

	directories := []string{dirs.binDirectory, dirs.generatedDirectory, dirs.cliWorkflowDirectory}

	for i := range directories {
		directory := directories[i]

		mustMkdirAll(t, directory)
	}

	return dirs
}

func writeActionlintFixtureFiles(t *testing.T, dirs *actionlintFixtureDirs) actionlintFixturePaths {
	t.Helper()

	goodPath := filepath.Join(dirs.workflowDirectory, goodWorkflowFile)
	skippedPath := filepath.Join(dirs.generatedDirectory, "bad.yml")
	cliGoodPath := filepath.Join(dirs.cliWorkflowDirectory, cliGoodWorkflowFile)

	paths := []string{goodPath, skippedPath, cliGoodPath}

	for i := range paths {
		path := paths[i]

		mustWriteFile(t, path, "name: test\n", perm0644)
	}

	return actionlintFixturePaths{skippedPath: skippedPath, cliGoodPath: cliGoodPath}
}

func writeActionlintStub(t *testing.T, binDirectory string) {
	t.Helper()

	stub := `#!/usr/bin/env bash
if [[ "${1-}" == "--version" ]]; then
  echo "1.7.12"
  exit 0
fi
printf '%s\n' "$@" >"$TASKOTTER_ACTIONLINT_LOG"
`

	mustWriteFile(t, filepath.Join(binDirectory, actionlintModule), stub, perm0500)
}

func (fixture *actionlintSkipFixture) assertAllSkipped(t *testing.T) {
	t.Helper()

	fixture.runTask(t, yesFlag, lintTaskName, "ACTIONLINT_LINT_SKIP_PATTERN=**")
	assertPathDoesNotExist(
		t,
		fixture.logPath,
		"actionlint ran even though every workflow was skipped",
	)
}

func (fixture *actionlintSkipFixture) assertCliTargets(t *testing.T) {
	t.Helper()

	output := fixture.runTask(t,
		yesFlag, lintTaskName, actionlintLintGeneratedPattern, "--",
		filepath.ToSlash(fixture.cliGoodPath), filepath.ToSlash(fixture.skippedPath), onelineFlag,
	)

	arguments := fixture.readLog(t, output)
	assertCliTargetArguments(t, arguments)
	fixture.removeLog(t, "CLI-target")
}

func assertCliTargetArguments(t *testing.T, arguments string) {
	t.Helper()

	missingCliTargets := !strings.Contains(arguments, cliGoodWorkflowFile) ||
		!strings.Contains(arguments, onelineFlag)

	if missingCliTargets {
		t.Fatalf("actionlint CLI targets or options were not retained:\n%s", arguments)
	}

	if strings.Contains(arguments, generatedBadYmlPath) {
		t.Fatalf("actionlint CLI target bypassed skip filtering:\n%s", arguments)
	}
}

func (fixture *actionlintSkipFixture) assertDefaultDiscovery(t *testing.T) {
	t.Helper()

	output := fixture.runTask(t, yesFlag, lintTaskName, actionlintLintGeneratedPattern)
	arguments := fixture.readLog(t, output)

	if !strings.Contains(arguments, goodWorkflowFile) {
		t.Fatalf("actionlint arguments do not contain retained file:\n%s", arguments)
	}

	if strings.Contains(arguments, generatedBadYmlPath) {
		t.Fatalf("actionlint arguments contain skipped file:\n%s", arguments)
	}

	fixture.removeLog(t, "first")
}

func (fixture *actionlintSkipFixture) readLog(t *testing.T, output []byte) string {
	t.Helper()

	argumentBytes, err := os.ReadFile(fixture.logPath)
	if err != nil {
		t.Fatalf("read actionlint log: %v\ntask output:\n%s", err, output)
	}

	return string(argumentBytes)
}

func (fixture *actionlintSkipFixture) removeLog(t *testing.T, name string) {
	t.Helper()

	err := os.Remove(fixture.logPath)
	if err != nil {
		t.Fatalf("remove %s actionlint log: %v", name, err)
	}
}

func (fixture *actionlintSkipFixture) runTask(t *testing.T, args ...string) []byte {
	t.Helper()

	output, err := fixture.taskCommand(args...).CombinedOutput()
	if err != nil {
		t.Fatalf("run actionlint task: %v\n%s", err, output)
	}

	return output
}

func (fixture *actionlintSkipFixture) taskCommand(args ...string) *exec.Cmd {
	return newSkipFixtureCommand(&skipFixtureCommand{
		project:      fixture.project,
		env:          fixture.env,
		taskfilePath: fixture.taskfilePath,
		args:         args,
	})
}

// TestCargoSkipPatternExcludesWorkspacePackages
func TestCargoSkipPatternExcludesWorkspacePackages(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip(shortModeSkipMsg)
	}

	fixture := newCargoSkipFixture(t)
	fixture.assertRetainedPackage(t)
	fixture.assertAllSkipped(t)
}

func newCargoSkipFixture(t *testing.T) cargoSkipFixture {
	t.Helper()

	root := tasktest.RepoRoot(t)
	project := t.TempDir()
	dirs := newCargoFixtureDirs(t, project)
	writeCargoFixtureFiles(t, &dirs)

	logPath := filepath.Join(project, "cargo.args")

	writeCargoStub(t, dirs.binDirectory)

	return cargoSkipFixture{
		taskfilePath: filepath.Join(root, taskfilesDirName, cargoModule, skipTaskfileYML),
		project:      project,
		logPath:      logPath,
		env: append(os.Environ(),
			"PATH="+dirs.binDirectory+":"+os.Getenv(pathEnvVar),
			"TASKOTTER_CARGO_LOG="+logPath,
		),
	}
}

func mkdirAllCargoFixtureDirs(t *testing.T, dirs *cargoFixtureDirs) {
	t.Helper()

	directories := []string{
		dirs.binDirectory,
		dirs.goodSourceDirectory,
		dirs.generatedSourceDirectory,
	}

	for i := range directories {
		directory := directories[i]

		mustMkdirAll(t, directory)
	}
}

func newCargoFixtureDirs(t *testing.T, project string) cargoFixtureDirs {
	t.Helper()

	dirs := cargoFixtureDirs{
		project:                  project,
		binDirectory:             filepath.Join(project, binDirName),
		goodSourceDirectory:      filepath.Join(project, goodDirName, srcDirName),
		generatedDirectory:       filepath.Join(project, generatedPackageDirName),
		generatedSourceDirectory: filepath.Join(project, generatedPackageDirName, srcDirName),
	}

	mkdirAllCargoFixtureDirs(t, &dirs)

	return dirs
}

func writeCargoFixtureFiles(t *testing.T, dirs *cargoFixtureDirs) {
	t.Helper()

	workspaceManifest := "[workspace]\nmembers = [\"good\", \"generated package\"]\n"
	goodPackageManifest := "[package]\nname = \"good_package\"\nversion = \"0.1.0\"\n"
	generatedPackageManifest := "[package]\nname = \"generated_package\"\nversion = \"0.1.0\"\n"

	files := map[string]string{
		filepath.Join(dirs.project, cargoTomlFile):              workspaceManifest,
		filepath.Join(dirs.project, goodDirName, cargoTomlFile): goodPackageManifest,
		filepath.Join(dirs.generatedDirectory, cargoTomlFile):   generatedPackageManifest,
		filepath.Join(dirs.goodSourceDirectory, libRsFile):      "pub fn good() {}\n",
		filepath.Join(dirs.generatedSourceDirectory, libRsFile): "pub fn generated() {}\n",
	}

	for path := range files {
		content := files[path]

		mustWriteFile(t, path, content, perm0644)
	}
}

func writeCargoStub(t *testing.T, binDirectory string) {
	t.Helper()

	stub := `#!/usr/bin/env bash
printf '%s\n' "$@" >"$TASKOTTER_CARGO_LOG"
`

	mustWriteFile(t, filepath.Join(binDirectory, cargoModule), stub, perm0500)
}

func (fixture *cargoSkipFixture) assertAllSkipped(t *testing.T) {
	t.Helper()

	fixture.runTask(t, yesFlag, lintTaskName, "CARGO_LINT_SKIP_PATTERN=**/*.rs")
	assertPathDoesNotExist(
		t,
		fixture.logPath,
		"Cargo ran even though every workspace package was skipped",
	)
}

func (fixture *cargoSkipFixture) assertRetainedPackage(t *testing.T) {
	t.Helper()

	fixture.runTask(t, yesFlag, lintTaskName, "CARGO_LINT_SKIP_PATTERN=**/generated package/**")

	arguments := readFile(t, fixture.logPath)

	if !strings.Contains(arguments, "clippy") || !strings.Contains(arguments, "good_package") {
		t.Fatalf("Cargo did not lint retained package:\n%s", arguments)
	}

	if strings.Contains(arguments, "generated_package") {
		t.Fatalf("Cargo lint included skipped package:\n%s", arguments)
	}

	mustRemove(t, fixture.logPath)
}

func (fixture *cargoSkipFixture) runTask(t *testing.T, args ...string) {
	t.Helper()

	output, err := fixture.taskCommand(args...).CombinedOutput()
	if err != nil {
		t.Fatalf("run Cargo task: %v\n%s", err, output)
	}
}

func (fixture *cargoSkipFixture) taskCommand(args ...string) *exec.Cmd {
	return newSkipFixtureCommand(&skipFixtureCommand{
		project:      fixture.project,
		env:          fixture.env,
		taskfilePath: fixture.taskfilePath,
		args:         args,
	})
}

// TestGovulncheckSkipPatternExcludesPackages
func TestGovulncheckSkipPatternExcludesPackages(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip(shortModeSkipMsg)
	}

	fixture := newGovulncheckSkipFixture(t)
	fixture.assertRetainedPackage(t)
	fixture.assertAllSkipped(t)
}

func govulncheckSkipFixtureEnv(binDirectory, logPath string) []string {
	return append(os.Environ(),
		"GOBIN="+binDirectory,
		"TASKOTTER_GO_ANALYSIS_LOG="+logPath,
		"GOCACHE=/private/tmp/taskotter-gocache",
	)
}

func newGovulncheckSkipFixture(t *testing.T) govulncheckSkipFixture {
	t.Helper()

	root := tasktest.RepoRoot(t)
	project := t.TempDir()
	dirs := newGovulncheckFixtureDirs(t, project)
	writeGovulncheckFixtureFiles(t, &dirs)

	logPath := filepath.Join(project, "govulncheck.args")

	writeGovulncheckStub(t, dirs.binDirectory)

	return govulncheckSkipFixture{
		taskfilePath: filepath.Join(root, taskfilesDirName, govulncheckModule, skipTaskfileYML),
		project:      project,
		logPath:      logPath,
		env:          govulncheckSkipFixtureEnv(dirs.binDirectory, logPath),
	}
}

func newGovulncheckFixtureDirs(t *testing.T, project string) govulncheckFixtureDirs {
	t.Helper()

	dirs := govulncheckFixtureDirs{
		project:            project,
		binDirectory:       filepath.Join(project, binDirName),
		goodDirectory:      filepath.Join(project, goodDirName),
		generatedDirectory: filepath.Join(project, generatedDirName),
	}

	directories := []string{dirs.binDirectory, dirs.goodDirectory, dirs.generatedDirectory}

	for i := range directories {
		directory := directories[i]

		mustMkdirAll(t, directory)
	}

	return dirs
}

func writeGovulncheckFixtureFiles(t *testing.T, dirs *govulncheckFixtureDirs) {
	t.Helper()

	files := map[string]string{
		filepath.Join(dirs.project, "go.mod"):            "module example.com/skipfixture\n\ngo 1.25\n",
		filepath.Join(dirs.goodDirectory, "good.go"):     "package good\n",
		filepath.Join(dirs.generatedDirectory, "bad.go"): "package generated\n",
	}

	for path := range files {
		content := files[path]

		mustWriteFile(t, path, content, perm0644)
	}
}

func writeGovulncheckStub(t *testing.T, binDirectory string) {
	t.Helper()

	stub := `#!/usr/bin/env bash
printf '%s\n' "$@" >"$TASKOTTER_GO_ANALYSIS_LOG"
`

	mustWriteFile(t, filepath.Join(binDirectory, govulncheckModule), stub, perm0500)
}

func (fixture *govulncheckSkipFixture) assertAllSkipped(t *testing.T) {
	t.Helper()

	fixture.runTask(t, yesFlag, lintTaskName, "GOVULNCHECK_LINT_SKIP_PATTERN=**/*.go")
	assertPathDoesNotExist(
		t,
		fixture.logPath,
		"govulncheck ran even though every Go package was skipped",
	)
}

func (fixture *govulncheckSkipFixture) assertRetainedPackage(t *testing.T) {
	t.Helper()

	fixture.runTask(t, yesFlag, lintTaskName, "GOVULNCHECK_LINT_SKIP_PATTERN=generated/**")

	arguments := readFile(t, fixture.logPath)

	if !strings.Contains(arguments, "example.com/skipfixture/good") {
		t.Fatalf("govulncheck did not receive retained package:\n%s", arguments)
	}

	if strings.Contains(arguments, "example.com/skipfixture/generated") {
		t.Fatalf("govulncheck received skipped package:\n%s", arguments)
	}

	mustRemove(t, fixture.logPath)
}

func (fixture *govulncheckSkipFixture) runTask(t *testing.T, args ...string) {
	t.Helper()

	output, err := fixture.taskCommand(args...).CombinedOutput()
	if err != nil {
		t.Fatalf("run govulncheck task: %v\n%s", err, output)
	}
}

func (fixture *govulncheckSkipFixture) taskCommand(args ...string) *exec.Cmd {
	return newSkipFixtureCommand(&skipFixtureCommand{
		project:      fixture.project,
		env:          fixture.env,
		taskfilePath: fixture.taskfilePath,
		args:         args,
	})
}

// variantLeaves returns every concrete leaf Taskfile of a nested tool family
// (the bun leaf plus each node/<vm>/<pm> leaf), excluding aggregators.
func variantLeafPatterns(root, family string) []string {
	return []string{
		filepath.Join(root, taskfilesDirName, family, bunRuntime, skipTaskfileYML),
		filepath.Join(
			root,
			taskfilesDirName,
			family,
			nodeRuntime,
			globStar,
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

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()

	err := os.MkdirAll(path, perm0700)
	if err != nil {
		t.Fatalf("create directory %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, values ...any) {
	t.Helper()

	path, file := normalizeFileWrite(t, values)

	err := os.WriteFile(path, []byte(file.content), file.perm)
	if err != nil {
		t.Fatalf(writeFileErrFormat, path, err)
	}
}

func normalizeFileWrite(t *testing.T, values []any) (string, fileWrite) {
	t.Helper()

	if len(values) != constTwo && len(values) != constThree {
		t.Fatalf("mustWriteFile requires path plus file data, got %d values", len(values))
	}

	path := normalizeFileWritePath(t, values)

	if len(values) == constTwo {
		return path, normalizeFileWriteData(t, values)
	}

	return path, normalizeFileWriteContent(t, values)
}

func normalizeFileWritePath(t *testing.T, values []any) string {
	t.Helper()

	path, ok := values[constZero].(string)

	if !ok {
		t.Fatalf("file path must be string, got %T", values[constZero])
	}

	return path
}

func normalizeFileWriteData(t *testing.T, values []any) fileWrite {
	t.Helper()

	file, ok := values[constOne].(fileWrite)

	if !ok {
		t.Fatalf("file data must be fileWrite, got %T", values[constOne])
	}

	return file
}

func normalizeFileWriteContent(t *testing.T, values []any) fileWrite {
	t.Helper()

	content, ok := values[constOne].(string)

	if !ok {
		t.Fatalf("file content must be string, got %T", values[constOne])
	}

	return fileWrite{content: content, perm: filePerm(t, values[constTwo])}
}

func filePerm(t *testing.T, value any) os.FileMode {
	t.Helper()

	switch perm := value.(type) {
	case os.FileMode:
		return perm
	case int:
		return filePermFromInt(t, perm)
	default:
		t.Fatalf("file permission must be os.FileMode, got %T", value)
	}

	return constZero
}

func filePermFromInt(t *testing.T, perm int) os.FileMode {
	t.Helper()

	if perm < constZero || perm > math.MaxUint32 {
		t.Fatalf("file permission out of range: %d", perm)
	}

	mode, err := strconv.ParseUint(strconv.Itoa(perm), constDecimalBase, constUint32BitSize)
	if err != nil {
		t.Fatalf("parse file permission: %v", err)
	}

	return os.FileMode(mode)
}

func mustRemove(t *testing.T, path string) {
	t.Helper()

	err := os.Remove(path)
	if err != nil {
		t.Fatalf("remove %s: %v", path, err)
	}
}

func runCommand(t *testing.T, directory string, arguments ...string) {
	t.Helper()

	command := exec.CommandContext(t.Context(), taskCommandName)

	command.Args = append(command.Args, arguments...)

	command.Dir = directory

	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run task: %v\n%s", err, output)
	}
}
