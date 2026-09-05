// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasktestutil_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/task-otter/store/internal/tasktestutil"
	yaml "go.yaml.in/yaml/v3"
)

type (
	fatalCall struct {
		message string
	}

	fakeTest struct {
		fatal    *fatalCall
		tempDirs []string
		nextDir  int
		stay     stayAfterFatal
	}

	stayAfterFatal bool

	fatalResult struct {
		fatal *fatalCall
		done  bool
	}

	groupOutput struct {
		errorOnly *string
		begin     string
		end       string
	}

	fatalExpectation struct {
		fake      *fakeTest
		fatalFunc func()
		want      string
	}

	filesystemAssertionFixture struct {
		dir      string
		file     string
		nonempty string
		empty    string
		missing  string
	}

	textFileAssertionCase struct {
		content string
		want    string
	}

	filesystemFailureCase struct {
		assert func(tasktestutil.TestT, string)
		path   string
		want   string
	}
)

const (
	publicTasksHeading                   = "## Public Tasks"
	alphaName                            = "alpha"
	zetaName                             = "zeta"
	defaultTaskName                      = "default"
	expectedFatalCall                    = "expected fatal call"
	dirModeAll                           = 0o700
	fileModeAll                          = 0o600
	taskBinaryName                       = "task"
	taskfileYML                          = "Taskfile.yml"
	taskfileYAML                         = "Taskfile.yaml"
	taskDescriptionTable                 = "| Task | Description |"
	tableDividerRow                      = "| --- | --- |"
	variablesHeading                     = "## Variables"
	descField                            = "desc"
	missingName                          = "missing"
	anythingField                        = "anything"
	descriptionValue                     = "description"
	twoAlias                             = "two"
	oneAlias                             = "one"
	xValue                               = "x"
	alphaBetaText                        = "alpha beta"
	spaceText                            = " "
	taskfileNotFoundMsg                  = "could not find Taskfile"
	rootName                             = "root"
	pathEnvVar                           = "PATH"
	bArg                                 = "B=2"
	simpleOutput                         = "simple"
	bashrcName                           = ".bashrc"
	trueValue                            = "true"
	envKeyA                              = "A"
	envValueNew                          = "new"
	envKeyB                              = "B"
	envValueGeneric                      = "value"
	internalTaskName                     = "internal"
	scalarName                           = "scalar"
	stubName                             = "stub"
	okName                               = "ok"
	betaName                             = "beta"
	emptySentinelMsg                     = "empty sentinel"
	entryName                            = "entry"
	groupBeginTemplate                   = "::group::{{.TASK}}"
	groupEndMarker                       = "::endgroup::"
	groupFieldName                       = "group"
	includeGroupConfig                   = "include group config"
	badValue                             = "bad"
	sameKeyName                          = "same"
	newline                              = "\n"
	zeroValue                            = 0
	oneValue                             = 1
	expectedTaskCount                    = 5
	exitCodeSeven                        = 7
	execStubMode                         = 0o500
	emptyStr                             = ""
	twoCount                             = 2
	stayAlive             stayAfterFatal = true
	readmeFileName                       = "README.md"
	dirExistsMsg                         = "but it does"
	noOutputConfigMsg                    = "no output config"
	taskfilesDirName                     = "taskfiles"
	childEntryName                       = "child"
	confirmPromptText                    = "please confirm"
	vaguePromptText                      = "maybe later"
	extraArgName                         = "extra"
	legacyTimeoutMsg                     = "legacy timeout call requires env and timeout"
	timeoutCountMsg                      = "exactly one timeout"
	taskfileTypeMsg                      = "LoadedTaskfile"
	resultTypeMsg                        = "CommandResult"
	stubPartsMsg                         = "name and body"
	positionalArgsMsg                    = "positional arguments"
	nonEmptyPromptMsg                    = "non-empty prompt"
	explicitPromptMsg                    = "not look explicit"
	readmeMissingMsg                     = "could not find README.md"
	durationTypeMsg                      = "time.Duration"
	validTaskfileTemplate                = `version: "3.5"
output:
  group:
    begin: "%s"
    end: "%s"
    error_only: false
tasks:
  default:
    desc: Show tasks
    cmds: [task --list]
  alpha:
    desc: Run the alpha fixture task
    aliases: [a]
    cmds:
      - cmd: echo alpha
  no-description:
    cmds: [echo hidden-by-contract]
  _private:
    desc: Private task
    cmds: [echo private]
  internal:
    desc: Internal task
    internal: true
    cmds: [echo internal]
`
	collectCommandStringsYAML = `cmds:
  - echo scalar
  - cmd: ./scripts/run.sh --flag
  - sh: echo shell
  - status:
      - cmd: echo status
  - preconditions:
      - sh: echo precondition
  - ignored: echo ignored
  - cmd:
      nested: ignored
`
)

var errSentinel = errors.New("sentinel")

func (fake *fakeTest) Fatal(args ...any) {
	fake.fatal = &fatalCall{message: fmt.Sprint(args...)}

	if !fake.stay {
		runtime.Goexit()
	}
}

func (fake *fakeTest) Fatalf(format string, args ...any) {
	fake.fatal = &fatalCall{message: fmt.Sprintf(format, args...)}

	if !fake.stay {
		runtime.Goexit()
	}
}

func (*fakeTest) Helper() {}

func (fake *fakeTest) TempDir() string {
	if fake.nextDir < len(fake.tempDirs) {
		dir := fake.tempDirs[fake.nextDir]
		fake.nextDir++

		return dir
	}

	dir, err := os.MkdirTemp(emptyStr, "tasktestutil-fake-")
	if err != nil {
		fake.Fatalf("create temp dir: %v", err)
	}

	fake.tempDirs = append(fake.tempDirs, dir)
	fake.nextDir++

	return dir
}

// TestTaskNodeAndYamlHelpers validates the behavior covered by this test case.
func TestTaskNodeAndYamlHelpers(t *testing.T) {
	t.Parallel()

	doc := parseYAML(t, `task:
  desc: "  description  "
  enabled: TRUE
  aliases: [one, two]
  nested:
    value: text
  sequence: [alpha, "", beta]
`)
	root := tasktestutil.DocumentRoot(t, doc)
	taskNode := tasktestutil.MappingField(root, taskBinaryName)
	task := tasktestutil.TaskNode{Name: taskBinaryName, Node: taskNode}

	assertTaskNodeHelpers(t, task)
	assertMappingHelpers(t, root, taskNode)
	assertTextAndEmptyHelpers(t, taskNode)
	assertDocumentRootFailures(t)
}

// TestModuleDiscoveryAndLoading validates the behavior covered by this test case.
func TestModuleDiscoveryAndLoading(t *testing.T) {
	t.Parallel()

	root := makeModule(t)
	nested := makeNestedModuleDir(t, root)
	taskfile := assertDiscoverAndLoad(t, root, nested)

	assertModuleDiscoveryFatals(t, &taskfile)
}

// TestModuleDiscoveryMissingTaskfile validates the behavior covered by this test case.
func TestModuleDiscoveryMissingTaskfile(t *testing.T) {
	inDir(t, t.TempDir(), func() {
		expectFatal(
			t,
			taskfileNotFoundMsg,
			func(fakeTester *fakeTest) { tasktestutil.ModuleRoot(fakeTester) },
		)
	})

	t.Parallel()
}

// TestLoadTaskfileFailures validates the behavior covered by this test case.
func TestLoadTaskfileFailures(t *testing.T) {
	root := makeModule(t)
	path := filepath.Join(root, taskfileYML)

	assertLoadTaskfileParseFailures(t, root, path)

	expectFatal(t, "failed to read", func(fakeTester *fakeTest) {
		tasktestutil.ReadFile(fakeTester, filepath.Join(root, missingName))
	})

	t.Parallel()
}

// TestCommandResultsAndRunners validates the behavior covered by this test case.
func TestCommandResultsAndRunners(t *testing.T) {
	root := t.TempDir()
	stub := writeExecutable(t, `printf 'stdout:%s' "$*"
printf 'stderr' >&2
if [ "${FAIL_TASK:-}" = yes ]; then exit 7; fi`)
	t.Setenv(pathEnvVar, filepath.Dir(stub)+string(os.PathListSeparator)+os.Getenv(pathEnvVar))

	assertHappyPathRun(t, root)
	assertFailedRun(t, root)
	assertTimedOutRun(t, root)
}

// TestDefaultTaskBinaryAndSimpleRunner validates the behavior covered by this test case.
//
//nolint:paralleltest // This test cannot call t.Parallel()
func TestDefaultTaskBinaryAndSimpleRunner(t *testing.T) {
	root := t.TempDir()

	setupDefaultTaskBinary(t, root)
	assertDefaultTaskBinaryResult(t, root)
	assertSimpleTaskRunnerResult(t, root)
}

// TestEnvironmentHelpers validates the behavior covered by this test case.
func TestEnvironmentHelpers(t *testing.T) {
	t.Parallel()

	env := tasktestutil.IsolatedEnv(t)
	home := tasktestutil.EnvValue(env, "HOME")

	assertIsolatedEnvValues(t, env, home)
	assertFileExistsHelpers(t, home)
	assertSetEnvHelpers(t)
	assertIsolatedEnvFailure(t)
}

// TestCollectionAndTextHelpers validates the behavior covered by this test case.
func TestCollectionAndTextHelpers(t *testing.T) {
	t.Parallel()

	assertCollectionHelpers(t)
	assertTextHelpers(t)
}

// TestFileHelpers validates the behavior covered by this test case.
func TestFileHelpers(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	tasktestutil.WriteStub(t, dir, stubName, "#!/bin/sh\necho stub\n")

	assertFileHelperReads(t, dir)
	assertFileHelperFailures(t, dir)
}

// TestCommandStringExtraction validates the behavior covered by this test case.
func TestCommandStringExtraction(t *testing.T) {
	t.Parallel()

	assertCollectCommandStrings(t)
	assertReferencedLocalShellScripts(t)
}

// TestAssertExitCode validates the behavior covered by this test case.
func TestAssertExitCode(t *testing.T) {
	t.Parallel()

	assertExitCodeSuccesses(t)
	assertExitCodeFailures(t)
}

// TestBasicAssertions validates the behavior covered by this test case.
func TestBasicAssertions(t *testing.T) {
	t.Parallel()

	tasktestutil.AssertContains(t, alphaBetaText, betaName)
	tasktestutil.AssertNotContains(t, alphaName, betaName)
	tasktestutil.AssertNotEmpty(t, " value ", "must not be empty")
	expectFatal(t, "expected output to contain", func(fakeTester *fakeTest) {
		tasktestutil.AssertContains(fakeTester, alphaName, betaName)
	})
	expectFatal(t, "expected output not to contain", func(fakeTester *fakeTest) {
		tasktestutil.AssertNotContains(fakeTester, alphaName, alphaName)
	})
	expectFatal(t, emptySentinelMsg, func(fakeTester *fakeTest) {
		tasktestutil.AssertNotEmpty(fakeTester, " \n", emptySentinelMsg)
	})
}

// TestFilesystemAssertions validates the behavior covered by this test case.
func TestFilesystemAssertions(t *testing.T) {
	t.Parallel()

	fixture := makeFilesystemAssertionFixture(t)

	assertFilesystemAssertionSuccesses(t, &fixture)
	assertFilesystemAssertionFailures(t, &fixture)
}

// TestGithubGroupAssertion validates the behavior covered by this test case.
func TestGithubGroupAssertion(t *testing.T) {
	t.Parallel()

	falseValue := "false"
	tasktestutil.AssertGithubGroupOutput(
		t,
		alphaName,
		groupOutputNode(t, groupBeginTemplate, groupEndMarker, &falseValue),
	)

	assertGithubGroupAssertionShapeFailures(t)
	assertGithubGroupAssertionValueFailures(t, falseValue)
	assertGithubGroupPostFatalReturn(t)
}

// TestTextFileAssertion validates the behavior covered by this test case.
func TestTextFileAssertion(t *testing.T) {
	t.Parallel()

	tasktestutil.AssertTextFileClean(t, "clean.yml", "key: value\n")

	tests := textFileAssertionCases()

	for i := range tests {
		testCase := tests[i]

		expectFatal(
			t,
			testCase.want,
			func(fakeTester *fakeTest) {
				tasktestutil.AssertTextFileClean(fakeTester, "bad.yml", testCase.content)
			},
		)
	}
}

// TestYamlStructureAssertions validates the behavior covered by this test case.
func TestYamlStructureAssertions(t *testing.T) {
	t.Parallel()

	assertYamlStructureCleanCases(t)
	assertYamlStructureFailureCases(t)
}

// TestPlaceholderJsonAndDangerousPatterns validates the behavior covered by this test case.
func TestPlaceholderJsonAndDangerousPatterns(t *testing.T) {
	t.Parallel()

	assertPlaceholderText(t)
	assertValidateJSON(t)
	assertDangerousCommandPatterns(t)
}

// TestAssertDestructivePrompt validates the behavior covered by this test case.
func TestAssertDestructivePrompt(t *testing.T) {
	t.Parallel()

	tasktestutil.AssertDestructivePrompt(t, alphaName, yamlScalar(confirmPromptText))
	assertDestructivePromptFailures(t)
}

// TestModuleReadmeSearch validates the behavior covered by this test case.
func TestModuleReadmeSearch(t *testing.T) {
	assertAncestorModuleReadme(t)
	assertMissingModuleReadme(t)
	assertTaskfilesReadmeBoundary(t)

	t.Parallel()
}

// TestRunNormalizationFatals validates the behavior covered by this test case.
func TestRunNormalizationFatals(t *testing.T) {
	t.Parallel()

	assertLegacyRunFatals(t)
	assertTimeoutRunFatals(t)
}

func expectFatal(t *testing.T, want string, fatalFunc func(*fakeTest)) {
	t.Helper()

	result := runFatalFunc(
		&fakeTest{fatal: nil, tempDirs: nil, nextDir: zeroValue, stay: false},
		func(fake *fakeTest) {
			fatalFunc(fake)
		},
	)

	assertFatalResult(t, want, result)
}

func expectFatalOn(t *testing.T, expectation *fatalExpectation) {
	t.Helper()

	result := runFatalFunc(expectation.fake, func(_ *fakeTest) {
		expectation.fatalFunc()
	})

	assertFatalResult(t, expectation.want, result)
}

func runFatalFunc(fake *fakeTest, callback func(*fakeTest)) fatalResult {
	done := make(chan fatalResult, oneValue)

	go func() {
		defer func() {
			done <- fatalResult{fatal: fake.fatal, done: true}
		}()

		callback(fake)
	}()

	return <-done
}

func assertFatalResult(t *testing.T, want string, result fatalResult) {
	t.Helper()

	if !result.done || result.fatal == nil {
		t.Fatal(expectedFatalCall)
	}

	if !strings.Contains(result.fatal.message, want) {
		t.Fatalf("fatal message %q does not contain %q", result.fatal.message, want)
	}
}

func chdirOrFatal(t *testing.T, dir string) {
	t.Helper()

	err := syscall.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}
}

func inDir(t *testing.T, dir string, callback func()) {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	chdirOrFatal(t, dir)

	defer func() {
		err = syscall.Chdir(previous)
		if err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	callback()
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), dirModeAll)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path, []byte(content), fileModeAll)
	if err != nil {
		t.Fatal(err)
	}
}

func writeExecutable(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), taskBinaryName)
	writeExecutableFile(t, path, "#!/bin/sh\n"+body+"\n")

	return path
}

func writeExecutableFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.WriteFile(path, []byte(content), fileModeAll)
	if err != nil {
		t.Fatal(err)
	}

	err = syscall.Chmod(path, execStubMode)
	if err != nil {
		t.Fatal(err)
	}
}

func validTaskfile() string {
	return fmt.Sprintf(validTaskfileTemplate, groupBeginTemplate, groupEndMarker)
}

func makeModule(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, taskfileYML), validTaskfile())
	writeFile(t, filepath.Join(root, "README.md"), strings.Join([]string{
		"# Fixture",
		emptyStr,
		publicTasksHeading,
		emptyStr,
		taskDescriptionTable,
		tableDividerRow,
		"| `alpha` | Alpha |",
		emptyStr,
		variablesHeading,
		emptyStr,
	}, newline))

	return root
}

func parseYAML(t *testing.T, content string) *yaml.Node {
	t.Helper()

	var doc yaml.Node

	err := yaml.Unmarshal([]byte(content), &doc)
	if err != nil {
		t.Fatalf("parse YAML: %v", err)
	}

	return &doc
}

func yamlScalar(value string) *yaml.Node {
	node := yamlNode(yaml.ScalarNode, nil)

	node.Value = value

	return node
}

func yamlMapping(content ...*yaml.Node) *yaml.Node {
	return yamlNode(yaml.MappingNode, content)
}

func yamlSequence(content ...*yaml.Node) *yaml.Node {
	return yamlNode(yaml.SequenceNode, content)
}

func yamlDocument(content ...*yaml.Node) *yaml.Node {
	return yamlNode(yaml.DocumentNode, content)
}

func yamlAlias() *yaml.Node {
	return yamlNode(yaml.AliasNode, nil)
}

func yamlNode(kind yaml.Kind, content []*yaml.Node) *yaml.Node {
	return &yaml.Node{
		Kind:        kind,
		Content:     content,
		Style:       zeroValue,
		Tag:         emptyStr,
		Value:       emptyStr,
		Anchor:      emptyStr,
		Alias:       nil,
		HeadComment: emptyStr,
		LineComment: emptyStr,
		FootComment: emptyStr,
		Line:        zeroValue,
		Column:      zeroValue,
	}
}

func samePath(t *testing.T, left, right string) bool {
	t.Helper()

	leftInfo, leftErr := os.Stat(left)
	rightInfo, rightErr := os.Stat(right)

	return leftErr == nil && rightErr == nil && os.SameFile(leftInfo, rightInfo)
}

func assertTaskNodeHelpers(t *testing.T, task tasktestutil.TaskNode) {
	t.Helper()

	assertTaskFieldLookup(t, task)
	assertTaskScalarFields(t, task)
	assertAliasHelpers(t, task)
}

func assertTaskFieldLookup(t *testing.T, task tasktestutil.TaskNode) {
	t.Helper()

	if tasktestutil.Field(task, descField) == nil || tasktestutil.Field(task, missingName) != nil {
		t.Fatal("ttu.TaskNode.Field lookup failed")
	}

	if tasktestutil.Field(tasktestutil.TaskNode{Name: emptyStr, Node: nil}, anythingField) != nil {
		t.Fatal("nil ttu.TaskNode returned a field")
	}

	scalarField := tasktestutil.Field(
		tasktestutil.TaskNode{Node: yamlScalar(emptyStr), Name: emptyStr},
		anythingField,
	)

	if scalarField != nil {
		t.Fatal("scalar ttu.TaskNode returned a field")
	}
}

func assertTaskScalarFields(t *testing.T, task tasktestutil.TaskNode) {
	t.Helper()

	got := tasktestutil.StringField(task, descField)

	if got != descriptionValue {
		t.Fatalf("StringField = %q", got)
	}

	mismatch := !tasktestutil.BoolField(task, "enabled") ||
		tasktestutil.BoolField(task, missingName) ||
		tasktestutil.BoolField(task, descField)

	if mismatch {
		t.Fatal("BoolField result mismatch")
	}
}

func assertAliasHelpers(t *testing.T, task tasktestutil.TaskNode) {
	t.Helper()

	if !tasktestutil.HasAlias(task, twoAlias) || tasktestutil.HasAlias(task, missingName) {
		t.Fatal("ttu.HasAlias result mismatch")
	}

	nilNodeHasAlias := tasktestutil.HasAlias(
		tasktestutil.TaskNode{Name: emptyStr, Node: nil},
		oneAlias,
	)
	scalarAliasesNode := yamlMapping(yamlScalar("aliases"), yamlScalar(oneAlias))
	scalarHasAlias := tasktestutil.HasAlias(
		tasktestutil.TaskNode{Node: scalarAliasesNode, Name: emptyStr},
		oneAlias,
	)

	if nilNodeHasAlias || scalarHasAlias {
		t.Fatal("invalid aliases were accepted")
	}
}

func assertMappingHelpers(t *testing.T, root, taskNode *yaml.Node) {
	t.Helper()

	assertMappingFieldHelpers(t, root)
	assertNodeMappingValueHelpers(t, taskNode)
}

func assertMappingFieldHelpers(t *testing.T, root *yaml.Node) {
	t.Helper()

	if mappingFieldLookupMismatch(root) {
		t.Fatal("ttu.MappingField mismatch")
	}

	if mappingFieldKindMismatch(root) {
		t.Fatal("mapping kind mismatch")
	}

	if tasktestutil.MappingField(root, taskBinaryName).Kind == yaml.ScalarNode {
		t.Fatal("impossible mapping kind")
	}
}

func mappingFieldLookupMismatch(root *yaml.Node) bool {
	return tasktestutil.MappingField(root, missingName) != nil ||
		tasktestutil.MappingField(root, taskBinaryName) == nil
}

func mappingFieldKindMismatch(root *yaml.Node) bool {
	return tasktestutil.MappingField(root, missingName) != nil ||
		tasktestutil.MappingField(root, taskBinaryName).Kind != yaml.MappingNode
}

func assertNodeMappingValueHelpers(t *testing.T, taskNode *yaml.Node) {
	t.Helper()

	if tasktestutil.ScalarField(taskNode, descField) != descriptionValue {
		t.Fatal("ttu.ScalarField mismatch")
	}

	if nodeMappingValueRejectsInvalid() {
		t.Fatal("ttu.NodeMappingValue accepted invalid node")
	}

	if nodeMappingValueLookupMismatch(taskNode) {
		t.Fatal("ttu.NodeMappingValue lookup mismatch")
	}
}

func nodeMappingValueRejectsInvalid() bool {
	return tasktestutil.NodeMappingValue(nil, xValue) != nil ||
		tasktestutil.NodeMappingValue(yamlScalar(emptyStr), xValue) != nil
}

func nodeMappingValueLookupMismatch(taskNode *yaml.Node) bool {
	return tasktestutil.NodeMappingValue(taskNode, missingName) != nil ||
		tasktestutil.NodeMappingValue(taskNode, descField) == nil
}

func assertTextAndEmptyHelpers(t *testing.T, taskNode *yaml.Node) {
	t.Helper()

	if nodeTextScalarMismatch() {
		t.Fatal("ttu.NodeText scalar mismatch")
	}

	if got := tasktestutil.NodeText(
		tasktestutil.NodeMappingValue(taskNode, "sequence"),
	); got != alphaBetaText {
		t.Fatalf("ttu.NodeText sequence = %q", got)
	}

	if isEmptyNodeMismatch() {
		t.Fatal("ttu.IsEmptyNode mismatch")
	}
}

func nodeTextScalarMismatch() bool {
	return tasktestutil.NodeText(nil) != emptyStr ||
		tasktestutil.NodeText(yamlScalar(" x ")) != xValue
}

func isEmptyNodeMismatch() bool {
	return !tasktestutil.IsEmptyNode(nil) || !tasktestutil.IsEmptyNode(yamlScalar(spaceText)) ||
		tasktestutil.IsEmptyNode(yamlScalar(xValue)) ||
		!tasktestutil.IsEmptyNode(yamlSequence()) ||
		tasktestutil.IsEmptyNode(yamlSequence(yamlScalar(xValue)))
}

func assertDocumentRootFailures(t *testing.T) {
	t.Helper()

	expectFatal(t, "invalid YAML document", func(fakeTester *fakeTest) {
		tasktestutil.DocumentRoot(fakeTester, yamlScalar(emptyStr))
	})
	expectFatal(t, "root must be a YAML mapping", func(fakeTester *fakeTest) {
		tasktestutil.DocumentRoot(fakeTester, yamlDocument(yamlSequence()))
	})
}

func makeNestedModuleDir(t *testing.T, root string) string {
	t.Helper()

	nested := filepath.Join(root, "nested", "deeper")

	err := os.MkdirAll(nested, dirModeAll)
	if err != nil {
		t.Fatal(err)
	}

	return nested
}

func assertModuleDiscoveryFatals(t *testing.T, taskfile *tasktestutil.LoadedTaskfile) {
	t.Helper()

	expectFatal(
		t,
		"is missing",
		func(fakeTester *fakeTest) { tasktestutil.MustTask(fakeTester, taskfile, missingName) },
	)
	expectFatal(
		t,
		taskfileTypeMsg,
		func(fakeTester *fakeTest) { tasktestutil.MustTask(fakeTester, zeroValue, alphaName) },
	)
	inDir(t, t.TempDir(), func() {
		expectFatal(
			t,
			taskfileNotFoundMsg,
			func(fakeTester *fakeTest) { tasktestutil.ModuleTaskfilePath(fakeTester) },
		)
	})
}

func assertDiscoverAndLoad(t *testing.T, root, nested string) tasktestutil.LoadedTaskfile {
	t.Helper()

	var taskfile tasktestutil.LoadedTaskfile

	inDir(t, nested, func() {
		taskfile = discoverAndLoadModule(t, root)
	})

	return taskfile
}

func discoverAndLoadModule(t *testing.T, root string) tasktestutil.LoadedTaskfile {
	t.Helper()

	assertModuleDiscoveryPaths(t, root)

	taskfile := tasktestutil.LoadTaskfile(t)
	assertLoadedTaskfile(t, root, &taskfile)
	renameModuleTaskfileToYAML(t, root)
	assertModuleTaskfilePathMatches(t, root, taskfileYAML)

	return taskfile
}

func assertModuleDiscoveryPaths(t *testing.T, root string) {
	t.Helper()

	assertModuleRootMatches(t, root)
	assertModuleTaskfilePathMatches(t, root, taskfileYML)
	assertModuleReadmePathMatches(t, root)
}

func renameModuleTaskfileToYAML(t *testing.T, root string) {
	t.Helper()

	err := os.Rename(filepath.Join(root, taskfileYML), filepath.Join(root, taskfileYAML))
	if err != nil {
		t.Fatal(err)
	}
}

func assertModuleRootMatches(t *testing.T, root string) {
	t.Helper()

	if got := tasktestutil.ModuleRoot(t); !samePath(t, got, root) {
		t.Fatalf("ttu.ModuleRoot = %s, want %s", got, root)
	}
}

func assertModuleTaskfilePathMatches(t *testing.T, root, name string) {
	t.Helper()

	if got := tasktestutil.ModuleTaskfilePath(t); !samePath(t, got, filepath.Join(root, name)) {
		t.Fatalf("ttu.ModuleTaskfilePath = %s", got)
	}
}

func assertLoadedTaskfile(t *testing.T, root string, taskfile *tasktestutil.LoadedTaskfile) {
	t.Helper()

	if loadedTaskfileMismatch(t, root, taskfile) {
		t.Fatalf("unexpected ttu.LoadedTaskfile: %#v", taskfile)
	}

	assertMustTaskNames(t, taskfile)
}

func assertMustTaskNames(t *testing.T, taskfile *tasktestutil.LoadedTaskfile) {
	t.Helper()

	if tasktestutil.MustTask(t, taskfile, alphaName).Name != alphaName {
		t.Fatal("ttu.MustTask returned wrong task")
	}

	if tasktestutil.MustTask(t, *taskfile, alphaName).Name != alphaName {
		t.Fatal("ttu.MustTask value returned wrong task")
	}

	got := tasktestutil.PublicTaskNamesFromTaskfile(t, *taskfile)
	want := []string{alphaName}

	if !slices.Equal(got, want) {
		t.Fatalf("public tasks = %v, want %v", got, want)
	}
}

func loadedTaskfileMismatch(t *testing.T, root string, taskfile *tasktestutil.LoadedTaskfile) bool {
	t.Helper()

	return !samePath(t, taskfile.Path, filepath.Join(root, taskfileYML)) ||
		taskfile.Root.Name != rootName || len(taskfile.Tasks) != expectedTaskCount
}

func assertLoadTaskfileParseFailures(t *testing.T, root, path string) {
	t.Helper()

	inDir(t, root, func() {
		writeFile(t, path, "version: [\n")
		expectFatal(
			t,
			"failed to parse Taskfile",
			func(fakeTester *fakeTest) { tasktestutil.LoadTaskfile(fakeTester) },
		)

		writeFile(t, path, "version: \"3\"\n")
		expectFatal(
			t,
			"has no tasks map",
			func(fakeTester *fakeTest) { tasktestutil.LoadTaskfile(fakeTester) },
		)
	})
}

func assertHappyPathRun(t *testing.T, root string) {
	t.Helper()

	result := tasktestutil.RunTask(
		t,
		tasktestutil.TaskRun{Root: root, Env: nil, Args: []string{alphaName, bArg}},
	)

	if happyPathResultMismatch(&result) {
		t.Fatalf("unexpected ttu.RunTask result: %#v", result)
	}

	if tasktestutil.Combined(&result) != "stdout:alpha "+bArg+"\nstderr" {
		t.Fatalf("Combined = %q", tasktestutil.Combined(&result))
	}
}

func happyPathResultMismatch(result *tasktestutil.CommandResult) bool {
	return result.Stdout != "stdout:alpha "+bArg || result.Stderr != "stderr" ||
		result.Err != nil ||
		!slices.Equal(result.Args, []string{alphaName, bArg})
}

func assertFailedRun(t *testing.T, root string) {
	t.Helper()

	env := tasktestutil.SetEnv(os.Environ(), "FAIL_TASK", "yes")

	failed := tasktestutil.RunTaskTimeout(
		t,
		tasktestutil.TaskRun{Root: root, Env: env, Args: []string{alphaName}},
		time.Second,
	)

	if failed.Err == nil {
		t.Fatal("ttu.RunTaskTimeout succeeded unexpectedly")
	}
}

func assertTimedOutRun(t *testing.T, root string) {
	t.Helper()

	sleeping := writeExecutable(t, "sleep 1")
	t.Setenv(pathEnvVar, filepath.Dir(sleeping)+string(os.PathListSeparator)+os.Getenv(pathEnvVar))

	timed := tasktestutil.RunTaskTimeout(
		t, tasktestutil.TaskRun{Root: root, Env: nil, Args: []string{
			alphaName,
		}},

		10*time.Millisecond)

	if timed.Err == nil {
		t.Fatal("timed command succeeded")
	}
}

func setupDefaultTaskBinary(t *testing.T, _ string) {
	t.Helper()

	bin := t.TempDir()
	stub := filepath.Join(bin, taskBinaryName)

	writeExecutableFile(t, stub, "#!/bin/sh\nprintf "+simpleOutput+"\n")

	// exec.Command resolves the binary name via LookPath against the real
	// process PATH at call time, not against TaskRun.Env, so the stub must be
	// put on PATH with t.Setenv; that makes this test unable to run in parallel.
	t.Setenv(pathEnvVar, bin+string(os.PathListSeparator)+os.Getenv(pathEnvVar))
}

func assertDefaultTaskBinaryResult(t *testing.T, root string) {
	t.Helper()

	result := tasktestutil.RunTaskTimeout(
		t,
		tasktestutil.TaskRun{Root: root, Env: os.Environ(), Args: []string{alphaName}},
		30*time.Second,
	)

	if result.Stdout != simpleOutput || result.Err != nil {
		t.Fatalf("default task result: %#v", result)
	}
}

func assertSimpleTaskRunnerResult(t *testing.T, root string) {
	t.Helper()

	result := tasktestutil.RunSimpleTask(t, root, os.Environ(), alphaName)

	if result.Stdout != simpleOutput || result.Err != nil {
		t.Fatalf("simple task result: %#v", result)
	}

	assertLegacyTimeoutRunnerResult(t, root)
}

func assertIsolatedEnvValues(t *testing.T, env []string, home string) {
	t.Helper()

	mismatch := home == emptyStr ||
		tasktestutil.EnvValue(env, "PROFILE") != filepath.Join(home, bashrcName) ||
		tasktestutil.EnvValue(env, "CI") != trueValue ||
		tasktestutil.EnvValue(env, "MISSING") != emptyStr

	if mismatch {
		t.Fatalf("isolated env mismatch: %v", env)
	}
}

func assertFileExistsHelpers(t *testing.T, home string) {
	t.Helper()

	mismatch := !tasktestutil.FileExists(filepath.Join(home, bashrcName)) ||
		tasktestutil.FileExists(filepath.Join(home, missingName))

	if mismatch {
		t.Fatal("ttu.FileExists mismatch")
	}
}

func assertSetEnvHelpers(t *testing.T) {
	t.Helper()

	values := []string{"A=old"}

	values = tasktestutil.SetEnv(values, envKeyA, envValueNew)

	values = tasktestutil.SetEnv(values, envKeyB, envValueGeneric)

	mismatch := tasktestutil.EnvValue(values, envKeyA) != envValueNew ||
		tasktestutil.EnvValue(values, envKeyB) != envValueGeneric

	if mismatch {
		t.Fatalf("ttu.SetEnv mismatch: %v", values)
	}
}

func assertIsolatedEnvFailure(t *testing.T) {
	t.Helper()

	fakeTester := isolatedEnvFailureFake(t)

	expectFatalOn(t, &fatalExpectation{
		want:      "failed to create fake shell profile",
		fake:      fakeTester,
		fatalFunc: func() { tasktestutil.IsolatedEnv(fakeTester) },
	})
}

func isolatedEnvFailureFake(t *testing.T) *fakeTest {
	t.Helper()

	homeFailure := t.TempDir()

	err := os.Mkdir(filepath.Join(homeFailure, bashrcName), dirModeAll)
	if err != nil {
		t.Fatal(err)
	}

	return &fakeTest{
		fatal: nil, tempDirs: []string{homeFailure}, nextDir: zeroValue, stay: false,
	}
}

func assertCollectionHelpers(t *testing.T) {
	t.Helper()

	assertExpectedPublicTaskNames(t)
	assertPublicTaskSpecOptions(t)
	assertTaskArgsHelpers(t)
	assertFormatListHelpers(t)
}

func assertExpectedPublicTaskNames(t *testing.T) {
	t.Helper()

	specs := []tasktestutil.PublicTaskSpec{
		tasktestutil.NewPublicTaskSpec(zetaName),
		tasktestutil.NewPublicTaskSpec(alphaName),
	}

	got := tasktestutil.ExpectedPublicTaskNames(specs)
	want := []string{alphaName, zetaName}

	if !slices.Equal(got, want) {
		t.Fatalf("expected names = %v", got)
	}
}

func assertTaskArgsHelpers(t *testing.T) {
	t.Helper()

	if tasktestutil.TaskArgs(nil) != nil || tasktestutil.TaskArgs(map[string]string{}) != nil {
		t.Fatal("empty ttu.TaskArgs must be nil")
	}

	got := tasktestutil.TaskArgs(map[string]string{"Z": strconv.Itoa(twoCount), envKeyA: "1"})
	want := []string{"A=1", "Z=2"}

	if !slices.Equal(got, want) {
		t.Fatalf("ttu.TaskArgs = %v", got)
	}
}

func assertFormatListHelpers(t *testing.T) {
	t.Helper()

	mismatch := tasktestutil.FormatList([]string{"a", "b"}) != "- a\n- b" ||
		tasktestutil.FormatList(nil) != "- "

	if mismatch {
		t.Fatal("ttu.FormatList mismatch")
	}
}

func assertSimplePublicTaskNames(t *testing.T) {
	t.Helper()

	tasks := map[string]any{
		defaultTaskName:  map[string]any{},
		"_private":       map[string]any{},
		internalTaskName: map[string]any{internalTaskName: true},
		alphaName:        map[string]any{descField: alphaName},
		scalarName:       envValueGeneric,
	}

	got := tasktestutil.SimplePublicTaskNames(tasks)
	want := []string{alphaName, scalarName}

	if !slices.Equal(got, want) {
		t.Fatalf("simple public names = %v", got)
	}
}

func sampleReadmeText() string {
	return strings.Join([]string{
		"# Module",
		emptyStr,
		publicTasksHeading,
		emptyStr,
		taskDescriptionTable,
		tableDividerRow,
		"| `zeta` | Z |",
		"| `alpha` | A |",
		emptyStr,
		variablesHeading,
		"| Name | Value |",
		emptyStr,
	}, newline)
}

func assertReadmePublicTaskNames(t *testing.T) {
	t.Helper()

	got := tasktestutil.ReadmePublicTaskNames(sampleReadmeText())
	want := []string{alphaName, zetaName}

	if !slices.Equal(got, want) {
		t.Fatalf("README names = %v", got)
	}

	emptyGot := tasktestutil.ReadmePublicTaskNames("# No table\n")

	if len(emptyGot) != zeroValue {
		t.Fatalf("unexpected README names: %v", emptyGot)
	}
}

func assertTextHelpers(t *testing.T) {
	t.Helper()

	assertSimplePublicTaskNames(t)
	assertReadmePublicTaskNames(t)
}

func assertFileHelperReads(t *testing.T, dir string) {
	t.Helper()

	stub := filepath.Join(dir, stubName)

	if got := tasktestutil.MustRead(t, stub); !strings.Contains(got, "echo stub") {
		t.Fatalf("ttu.MustRead = %q", got)
	}

	if got := tasktestutil.ReadFile(t, stub); got == emptyStr {
		t.Fatal("ttu.ReadFile returned empty content")
	}
}

func assertFileHelperFailures(t *testing.T, dir string) {
	t.Helper()

	assertWriteStubFatals(t, dir)
	expectFatal(
		t,
		"read",
		func(fakeTester *fakeTest) {
			tasktestutil.MustRead(fakeTester, filepath.Join(dir, missingName))
		},
	)
}

func assertWriteStubFatals(t *testing.T, dir string) {
	t.Helper()

	expectFatal(t, "write broken stub", func(fakeTester *fakeTest) {
		tasktestutil.WriteStub(fakeTester, filepath.Join(dir, missingName), "broken", "body")
	})
	expectFatal(t, stubPartsMsg, func(fakeTester *fakeTest) {
		tasktestutil.WriteStub(fakeTester, dir)
	})
	expectFatal(t, "stub dir must be string", func(fakeTester *fakeTest) {
		tasktestutil.WriteStub(fakeTester, zeroValue, stubName, stubName)
	})
}

func assertCollectCommandStringsMatch(t *testing.T) {
	t.Helper()

	doc := parseYAML(t, collectCommandStringsYAML)
	root := tasktestutil.DocumentRoot(t, doc)
	commands := tasktestutil.CollectCommandStrings(tasktestutil.NodeMappingValue(root, "cmds"))

	want := []string{
		"echo scalar",
		"./scripts/run.sh --flag",
		"echo shell",
		"echo status",
		"echo precondition",
	}

	if !slices.Equal(commands, want) {
		t.Fatalf("commands = %v, want %v", commands, want)
	}
}

func assertCollectCommandStrings(t *testing.T) {
	t.Helper()

	assertCollectCommandStringsMatch(t)

	emptyExtraction := tasktestutil.CollectCommandStrings(nil) != nil ||
		len(tasktestutil.CollectCommandStrings(yamlScalar(spaceText))) != zeroValue ||
		tasktestutil.CollectCommandStrings(yamlAlias()) != nil

	if emptyExtraction {
		t.Fatal("empty command extraction mismatch")
	}
}

func assertReferencedLocalShellScripts(t *testing.T) {
	t.Helper()

	got := tasktestutil.ReferencedLocalShellScripts("./one.sh --flag && echo x\n ./two/path.sh ")

	if !slices.Equal(got, []string{"./one.sh", "./two/path.sh"}) {
		t.Fatalf("script references = %v", got)
	}

	got = tasktestutil.ReferencedLocalShellScripts("scripts/no-prefix.sh")

	if len(got) != zeroValue {
		t.Fatalf("unexpected references = %v", got)
	}
}

func assertExitCodeSuccesses(t *testing.T) {
	t.Helper()

	okResult := tasktestutil.CommandResult{
		Args: []string{okName}, Stdout: emptyStr, Stderr: emptyStr, Err: nil,
	}
	tasktestutil.AssertExitCode(t, okResult, zeroValue)
	tasktestutil.AssertExitCode(t, &okResult, zeroValue)

	err := exec.CommandContext(t.Context(), "sh", "-c", "exit 7").Run()
	tasktestutil.AssertExitCode(t, tasktestutil.CommandResult{
		Err:  err,
		Args: []string{"exit"}, Stdout: emptyStr, Stderr: emptyStr,
	}, exitCodeSeven)
}

func assertExitCodeFailures(t *testing.T) {
	t.Helper()

	expectFatal(t, "without exit code", func(fakeTester *fakeTest) {
		tasktestutil.AssertExitCode(fakeTester, tasktestutil.CommandResult{
			Err: errSentinel, Stdout: emptyStr, Stderr: emptyStr,
			Args: nil,
		}, oneValue)
	})
	expectFatal(t, "expected exit code", func(fakeTester *fakeTest) {
		tasktestutil.AssertExitCode(
			fakeTester,
			tasktestutil.CommandResult{
				Args: []string{okName}, Stdout: emptyStr, Stderr: emptyStr, Err: nil,
			},
			twoCount,
		)
	})
	expectFatal(t, resultTypeMsg, func(fakeTester *fakeTest) {
		tasktestutil.AssertExitCode(fakeTester, zeroValue, zeroValue)
	})
}

func makeEmptyAssertionDir(t *testing.T, dir string) string {
	t.Helper()

	empty := filepath.Join(dir, "empty")

	err := os.Mkdir(empty, dirModeAll)
	if err != nil {
		t.Fatal(err)
	}

	return empty
}

func makeFilesystemAssertionFixture(t *testing.T) filesystemAssertionFixture {
	t.Helper()

	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	writeFile(t, file, "content")

	nonempty := filepath.Join(dir, "nonempty")
	writeFile(t, filepath.Join(nonempty, entryName), entryName)

	empty := makeEmptyAssertionDir(t, dir)
	missing := filepath.Join(dir, missingName)

	return filesystemAssertionFixture{
		dir: dir, file: file, nonempty: nonempty, empty: empty, missing: missing,
	}
}

func assertFilesystemAssertionSuccesses(t *testing.T, fixture *filesystemAssertionFixture) {
	t.Helper()

	tasktestutil.AssertFileExists(t, fixture.file)
	tasktestutil.AssertDirExists(t, fixture.dir)
	tasktestutil.AssertDirNotExists(t, fixture.missing)
	tasktestutil.AssertDirHasEntries(t, fixture.nonempty)
}

func filesystemFailureCases(fixture *filesystemAssertionFixture) []filesystemFailureCase {
	return []filesystemFailureCase{
		{assert: tasktestutil.AssertFileExists, path: fixture.missing, want: "expected file"},
		{assert: tasktestutil.AssertFileExists, path: fixture.dir, want: "found directory"},
		{assert: tasktestutil.AssertDirExists, path: fixture.missing, want: "expected directory"},
		{assert: tasktestutil.AssertDirExists, path: fixture.file, want: "found file"},
		{assert: tasktestutil.AssertDirNotExists, path: fixture.dir, want: dirExistsMsg},
		{
			assert: tasktestutil.AssertDirNotExists,
			path:   filepath.Join(fixture.file, childEntryName),
			want:   dirExistsMsg,
		},
		{
			assert: tasktestutil.AssertDirHasEntries,
			path:   fixture.missing,
			want:   "failed to read directory",
		},
		{assert: tasktestutil.AssertDirHasEntries, path: fixture.empty, want: "at least one entry"},
	}
}

func assertFilesystemAssertionFailures(t *testing.T, fixture *filesystemAssertionFixture) {
	t.Helper()

	cases := filesystemFailureCases(fixture)

	for i := range cases {
		testCase := cases[i]

		expectFatal(
			t,
			testCase.want,
			func(fakeTester *fakeTest) { testCase.assert(fakeTester, testCase.path) },
		)
	}
}

func groupOutputNode(t *testing.T, groupValue any, parts ...any) *yaml.Node {
	t.Helper()

	group := normalizeGroupOutput(t, groupValue, parts)

	value := emptyStr

	if group.errorOnly != nil {
		value = "\n    error_only: " + *group.errorOnly
	}

	doc := parseYAML(
		t,
		"output:\n  group:\n    begin: \""+group.begin+"\"\n    end: \""+group.end+"\""+value+"\n",
	)

	return tasktestutil.NodeMappingValue(tasktestutil.DocumentRoot(t, doc), "output")
}

func normalizeGroupOutput(t *testing.T, groupValue any, parts []any) groupOutput {
	t.Helper()

	group, ok := groupValue.(groupOutput)

	if !ok {
		begin := requireGroupBegin(t, groupValue)
		end, errorOnly := requireGroupEndAndErrorOnly(t, parts)

		return groupOutput{begin: begin, end: end, errorOnly: errorOnly}
	}

	if len(parts) != zeroValue {
		t.Fatalf("groupOutput does not accept positional arguments: %v", parts)
	}

	return group
}

func requireGroupBegin(t *testing.T, groupValue any) string {
	t.Helper()

	begin, ok := groupValue.(string)

	if !ok {
		t.Fatalf("group begin must be string or groupOutput, got %T", groupValue)
	}

	return begin
}

func requireGroupEndAndErrorOnly(t *testing.T, parts []any) (end string, errorOnly *string) {
	t.Helper()

	if len(parts) != twoCount {
		t.Fatalf("group output requires end and error_only values; got %d values", len(parts))
	}

	return requireGroupEnd(t, parts), requireGroupErrorOnly(t, parts)
}

func requireGroupEnd(t *testing.T, parts []any) string {
	t.Helper()

	end, ok := parts[zeroValue].(string)

	if !ok {
		t.Fatalf("group end must be string, got %T", parts[zeroValue])
	}

	return end
}

func requireGroupErrorOnly(t *testing.T, parts []any) *string {
	t.Helper()

	errorOnly, ok := parts[oneValue].(*string)

	if parts[oneValue] != nil && !ok {
		t.Fatalf("group error_only must be *string, got %T", parts[oneValue])
	}

	return errorOnly
}

func assertGithubGroupAssertionShapeFailures(t *testing.T) {
	t.Helper()

	expectFatal(t, noOutputConfigMsg, func(fakeTester *fakeTest) {
		tasktestutil.AssertGithubGroupOutput(fakeTester, alphaName, nil)
	})
	expectFatal(t, "advanced object format", func(fakeTester *fakeTest) {
		tasktestutil.AssertGithubGroupOutput(fakeTester, alphaName, yamlScalar(groupFieldName))
	})
	expectFatal(t, includeGroupConfig, func(fakeTester *fakeTest) {
		tasktestutil.AssertGithubGroupOutput(fakeTester, alphaName, yamlMapping())
	})
	expectFatal(t, includeGroupConfig, func(fakeTester *fakeTest) {
		tasktestutil.AssertGithubGroupOutput(fakeTester, alphaName, yamlMapping(
			yamlScalar(groupFieldName),
			yamlScalar(scalarName),
		))
	})
}

func assertGithubGroupBeginEndFailures(t *testing.T, falseValue string) {
	t.Helper()

	expectFatal(t, "output.group.begin", func(fakeTester *fakeTest) {
		tasktestutil.AssertGithubGroupOutput(
			fakeTester,
			alphaName,
			groupOutputNode(t, badValue, groupEndMarker, &falseValue),
		)
	})
	expectFatal(t, "output.group.end", func(fakeTester *fakeTest) {
		tasktestutil.AssertGithubGroupOutput(
			fakeTester,
			alphaName,
			groupOutputNode(t, groupBeginTemplate, badValue, &falseValue),
		)
	})
}

func assertGithubGroupErrorOnlyFailures(t *testing.T) {
	t.Helper()

	expectFatal(t, "explicitly set", func(fakeTester *fakeTest) {
		tasktestutil.AssertGithubGroupOutput(
			fakeTester,
			alphaName,
			groupOutputNode(t, groupBeginTemplate, groupEndMarker, nil),
		)
	})

	trueGroupValue := trueValue

	expectFatal(t, "must be false", func(fakeTester *fakeTest) {
		tasktestutil.AssertGithubGroupOutput(
			fakeTester,
			alphaName,
			groupOutputNode(t, groupBeginTemplate, groupEndMarker, &trueGroupValue),
		)
	})
}

func assertGithubGroupAssertionValueFailures(t *testing.T, falseValue string) {
	t.Helper()

	assertGithubGroupBeginEndFailures(t, falseValue)
	assertGithubGroupErrorOnlyFailures(t)
}

func textFileAssertionCases() []textFileAssertionCase {
	return []textFileAssertionCase{
		{content: emptyStr, want: "is empty"},
		{content: "key: value\r\n", want: "CRLF"},
		{content: "key:\tvalue\n", want: "contains tabs"},
		{content: "key: value", want: "end with a newline"},
		{content: "key: value \n", want: "trailing whitespace"},
	}
}

func assertYamlStructureCleanCases(t *testing.T) {
	t.Helper()

	doc := parseYAML(t, "root:\n  list:\n    - name: one\n    - name: two\n")
	tasktestutil.AssertNoDuplicateMappingKeys(t, nil, rootName)
	tasktestutil.AssertNoDuplicateMappingKeys(t, doc, rootName)
	tasktestutil.AssertNoYamlAliases(t, nil, rootName)
	tasktestutil.AssertNoYamlAliases(t, doc, rootName)
}

func assertDuplicateMappingKeyFailure(t *testing.T) {
	t.Helper()

	duplicate := yamlMapping(
		yamlScalar(sameKeyName),
		yamlScalar(oneAlias),
		yamlScalar(sameKeyName),
		yamlScalar(twoAlias),
	)

	expectFatal(t, "duplicate YAML key", func(fakeTester *fakeTest) {
		tasktestutil.AssertNoDuplicateMappingKeys(fakeTester, duplicate, rootName)
	})
}

func assertYamlAliasFailure(t *testing.T) {
	t.Helper()

	alias := yamlMapping(
		yamlScalar(envValueGeneric),
		yamlAlias(),
	)

	expectFatal(t, "aliases/anchors are not allowed", func(fakeTester *fakeTest) {
		tasktestutil.AssertNoYamlAliases(fakeTester, alias, rootName)
	})
}

func assertYamlStructureFailureCases(t *testing.T) {
	t.Helper()

	assertDuplicateMappingKeyFailure(t)
	assertYamlAliasFailure(t)
}

func assertPlaceholderText(t *testing.T) {
	t.Helper()

	tasktestutil.AssertNoPlaceholderText(t, alphaName, "ordinary task description")
	expectFatal(t, "placeholder text", func(fakeTester *fakeTest) {
		tasktestutil.AssertNoPlaceholderText(fakeTester, alphaName, "fixme later")
	})
}

func assertValidateJSON(t *testing.T) {
	t.Helper()

	err := tasktestutil.ValidateJSON(`{"ok":true}`)
	if err != nil {
		t.Fatalf("valid JSON: %v", err)
	}

	err = tasktestutil.ValidateJSON(`{"bad":`)
	if err == nil {
		t.Fatal("invalid JSON accepted")
	}
}

func unsafeCommandSamples() []string {
	return []string{
		"rm -rf / ",
		"sudo rm -rf /tmp/x",
		"chmod -R 777 /",
		"curl https://x -k ",
		"curl --insecure https://x",
	}
}

func assertDangerousCommandPatterns(t *testing.T) {
	t.Helper()

	patterns := tasktestutil.DangerousCommandPatterns()
	unsafe := unsafeCommandSamples()

	if len(patterns) != len(unsafe) {
		t.Fatalf("dangerous patterns = %d", len(patterns))
	}

	for index := range patterns {
		pattern := patterns[index]

		if !pattern.MatchString(unsafe[index]) {
			t.Fatalf("pattern %d did not match %q", index, unsafe[index])
		}
	}
}

func assertPublicTaskSpecOptions(t *testing.T) {
	t.Helper()

	spec := tasktestutil.NewPublicTaskSpec(alphaName, publicTaskSpecOptions()...)

	if publicTaskSpecMismatch(&spec) {
		t.Fatalf("public task spec mismatch: %#v", spec)
	}
}

func publicTaskSpecOptions() []tasktestutil.PublicTaskSpecOption {
	return []tasktestutil.PublicTaskSpecOption{
		tasktestutil.WithArgs(map[string]string{envKeyA: "1"}),
		tasktestutil.WithDryRunArgs(),
		tasktestutil.WithDryRunNoArgs(),
		tasktestutil.WithExpectedDefaultTokens(okName),
		tasktestutil.WithGroupOutput(),
		tasktestutil.WithPrompt(),
		tasktestutil.WithSummary(),
	}
}

func publicTaskSpecMismatch(spec *tasktestutil.PublicTaskSpec) bool {
	return publicTaskSpecDryRunMismatch(spec) || publicTaskSpecRequireMismatch(spec) ||
		publicTaskSpecValueMismatch(spec)
}

func publicTaskSpecDryRunMismatch(spec *tasktestutil.PublicTaskSpec) bool {
	return !spec.MustDryRunWithArgs || !spec.MustDryRunWithoutArgs
}

func publicTaskSpecRequireMismatch(spec *tasktestutil.PublicTaskSpec) bool {
	return !spec.RequiresGroupOutput || !spec.RequiresPrompt || !spec.RequiresSummary
}

func publicTaskSpecValueMismatch(spec *tasktestutil.PublicTaskSpec) bool {
	return spec.Name != alphaName || spec.Args[envKeyA] != "1" ||
		len(spec.ExpectedDefaultTokens) != oneValue ||
		spec.ExpectedDefaultTokens[zeroValue] != okName
}

func assertModuleReadmePathMatches(t *testing.T, root string) {
	t.Helper()

	want := filepath.Join(root, readmeFileName)

	if got := tasktestutil.ModuleReadmePath(t); !samePath(t, got, want) {
		t.Fatalf("ttu.ModuleReadmePath = %s", got)
	}
}

func assertLegacyTimeoutRunnerResult(t *testing.T, root string) {
	t.Helper()

	result := tasktestutil.RunTaskTimeout(t, root, os.Environ(), 30*time.Second, alphaName)

	if result.Stdout != simpleOutput || result.Err != nil {
		t.Fatalf("legacy timeout result: %#v", result)
	}
}

func assertGithubGroupPostFatalReturn(t *testing.T) {
	t.Helper()

	expectFatalStay(t, noOutputConfigMsg, func(fakeTester *fakeTest) {
		tasktestutil.AssertGithubGroupOutput(fakeTester, alphaName, nil)
	})
}

func expectFatalStay(t *testing.T, want string, fatalFunc func(*fakeTest)) {
	t.Helper()

	fake := &fakeTest{fatal: nil, tempDirs: nil, nextDir: zeroValue, stay: stayAlive}

	fatalFunc(fake)
	assertFatalResult(t, want, fatalResult{fatal: fake.fatal, done: true})
}

func assertDestructivePromptFailures(t *testing.T) {
	t.Helper()

	expectFatal(t, nonEmptyPromptMsg, func(fakeTester *fakeTest) {
		tasktestutil.AssertDestructivePrompt(fakeTester, alphaName, nil)
	})
	expectFatal(t, explicitPromptMsg, func(fakeTester *fakeTest) {
		tasktestutil.AssertDestructivePrompt(fakeTester, alphaName, yamlScalar(vaguePromptText))
	})
}

func assertAncestorModuleReadme(t *testing.T) {
	t.Helper()

	family, variant := makeFamilyVariantModule(t)
	writeFile(t, filepath.Join(family, readmeFileName), "# Family\n")

	inDir(t, variant, func() {
		assertModuleReadmePathMatches(t, family)
	})
}

func assertMissingModuleReadme(t *testing.T) {
	t.Helper()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, taskfileYML), validTaskfile())

	inDir(t, root, func() {
		expectFatalStay(t, readmeMissingMsg, func(fakeTester *fakeTest) {
			assertEmptyModuleReadmePath(t, fakeTester)
		})
	})
}

func assertTaskfilesReadmeBoundary(t *testing.T) {
	t.Helper()

	root, module := makeTaskfilesModule(t)
	writeFile(t, filepath.Join(root, readmeFileName), "# Root\n")

	inDir(t, module, func() {
		expectFatal(t, readmeMissingMsg, func(fakeTester *fakeTest) {
			tasktestutil.ModuleReadmePath(fakeTester)
		})
	})
}

func assertEmptyModuleReadmePath(t *testing.T, fakeTester *fakeTest) {
	t.Helper()

	if got := tasktestutil.ModuleReadmePath(fakeTester); got != emptyStr {
		t.Fatalf("ModuleReadmePath = %q", got)
	}
}

func makeFamilyVariantModule(t *testing.T) (family, variant string) {
	t.Helper()

	family = filepath.Join(t.TempDir(), taskfilesDirName, "family")
	variant = filepath.Join(family, "variant")
	writeFile(t, filepath.Join(variant, taskfileYML), validTaskfile())

	return family, variant
}

func makeTaskfilesModule(t *testing.T) (root, module string) {
	t.Helper()

	root = t.TempDir()
	module = filepath.Join(root, taskfilesDirName, "module")
	writeFile(t, filepath.Join(module, taskfileYML), validTaskfile())

	return root, module
}

func assertLegacyRunFatals(t *testing.T) {
	t.Helper()

	root := t.TempDir()
	env := []string{}

	expectFatal(t, positionalArgsMsg, func(fakeTester *fakeTest) {
		tasktestutil.RunTask(fakeTester, tasktestutil.TaskRun{
			Root: root, Env: env, Args: nil,
		}, extraArgName)
	})
	assertLegacyRunValueFatals(t, root, env)
}

func assertLegacyRunValueFatals(t *testing.T, root string, env []string) {
	t.Helper()

	expectFatal(t, "must be string or TaskRun", func(fakeTester *fakeTest) {
		tasktestutil.RunTask(fakeTester, zeroValue, env)
	})
	expectFatal(t, "environment is required", func(fakeTester *fakeTest) {
		tasktestutil.RunTask(fakeTester, root)
	})
	expectFatal(t, "environment must be []string", func(fakeTester *fakeTest) {
		tasktestutil.RunTask(fakeTester, root, zeroValue)
	})
	expectFatal(t, "argument must be string", func(fakeTester *fakeTest) {
		tasktestutil.RunTask(fakeTester, root, env, zeroValue)
	})
}

func assertTimeoutRunFatals(t *testing.T) {
	t.Helper()

	run := tasktestutil.TaskRun{Root: t.TempDir(), Env: nil, Args: nil}

	expectFatal(t, timeoutCountMsg, func(fakeTester *fakeTest) {
		tasktestutil.RunTaskTimeout(fakeTester, run)
	})
	expectFatal(t, durationTypeMsg, func(fakeTester *fakeTest) {
		tasktestutil.RunTaskTimeout(fakeTester, run, extraArgName)
	})
	assertLegacyTimeoutFatals(t, run.Root)
}

func assertLegacyTimeoutFatals(t *testing.T, root string) {
	t.Helper()

	expectFatal(t, legacyTimeoutMsg, func(fakeTester *fakeTest) {
		tasktestutil.RunTaskTimeout(fakeTester, root)
	})
	expectFatal(t, durationTypeMsg, func(fakeTester *fakeTest) {
		tasktestutil.RunTaskTimeout(fakeTester, root, []string{}, extraArgName)
	})
}
