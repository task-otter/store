// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskfiles_test

import (
	"bytes"
	"context"
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
	depcheckLintSkipVar = "DEPCHECK_LINT_SKIP_PATTERN"
	eslintLintSkipVar   = "ESLINT_LINT_SKIP_PATTERN"
	htmlhintLintSkipVar = "HTMLHINT_LINT_SKIP_PATTERN"
	spectralLintSkipVar = "SPECTRAL_LINT_SKIP_PATTERN"
	sqlfluffModule      = "sqlfluff"
	taskfileFlag        = "--taskfile"
	skipTaskfileYML     = "Taskfile.yml"

	taskfilesDirName = "taskfiles"

	actionlintModule   = "actionlint"
	ansibleLintModule  = "ansible-lint"
	bufModule          = "buf"
	cargoModule        = "cargo"
	dotenvLinterModule = "dotenv-linter"
	golangciLintModule = "golangci-lint"
	govulncheckModule  = "govulncheck"
	hadolintModule     = "hadolint"
	jsonlintModule     = "jsonlint"
	protolintModule    = "protolint"
	shellcheckModule   = "shellcheck"
	shfmtModule        = "shfmt"
	yamlfixModule      = "yamlfix"
	yamllintModule     = "yamllint"
	zizmorModule       = "zizmor"

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
	yesFlag         = "--yes"
	lintTaskName    = "lint"
	ciTaskName      = "ci"
	onelineFlag     = "-oneline"
	pathEnvVar      = "PATH"

	bunRuntime  = "bun"
	nodeRuntime = "node"

	writeFileErrFormat = "write %s: %v"

	actionlintLintGeneratedPattern = "ACTIONLINT_LINT_SKIP_PATTERN=**/generated/**"

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

	constZero    = 0
	constOne     = 1
	constTwo     = 2
	constThree   = 3
	constFifteen = 15
	constTwenty  = 20
	perm0600     = 0o600

	underscoreChar = "_"
	hyphenChar     = "-"

	constDecimalBase   = 10
	constUint32BitSize = 32
	perm0700           = 0o700

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
	assertSharedSkipfilesConsumers(t, root)
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

// Modules that still delegate to the shared skipfiles helper.
func sharedSkipfilesConsumers() []string {
	return []string{
		actionlintModule,
		ansibleLintModule,
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

func (fixture *actionlintSkipFixture) assertAllSkipped(t *testing.T) {
	t.Helper()

	fixture.runTask(t, yesFlag, ciTaskName, "ACTIONLINT_LINT_SKIP_PATTERN=**")
	assertPathDoesNotExist(
		t,
		fixture.logPath,
		"actionlint ran even though every workflow was skipped",
	)
}

func (fixture *actionlintSkipFixture) assertCliTargets(t *testing.T) {
	t.Helper()

	output := fixture.runTask(t,
		yesFlag, ciTaskName, actionlintLintGeneratedPattern, "--",
		filepath.ToSlash(fixture.cliGoodPath), filepath.ToSlash(fixture.skippedPath), onelineFlag,
	)

	arguments := fixture.readLog(t, output)
	assertCliTargetArguments(t, arguments)
	fixture.removeLog(t, "CLI-target")
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

func (fixture *actionlintSkipFixture) assertDefaultDiscovery(t *testing.T) {
	t.Helper()

	output := fixture.runTask(t, yesFlag, ciTaskName, actionlintLintGeneratedPattern)
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

func writeCargoStub(t *testing.T, binDirectory string) {
	t.Helper()

	stub := `#!/usr/bin/env bash
printf '%s\n' "$@" >"$TASKOTTER_CARGO_LOG"
`

	mustWriteFile(t, filepath.Join(binDirectory, cargoModule), stub, perm0500)
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

	fixture.runTask(t, yesFlag, ciTaskName, "GOVULNCHECK_LINT_SKIP_PATTERN=**/*.go")
	assertPathDoesNotExist(
		t,
		fixture.logPath,
		"govulncheck ran even though every Go package was skipped",
	)
}

func (fixture *govulncheckSkipFixture) assertRetainedPackage(t *testing.T) {
	t.Helper()

	fixture.runTask(t, yesFlag, ciTaskName, "GOVULNCHECK_LINT_SKIP_PATTERN=generated/**")

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
