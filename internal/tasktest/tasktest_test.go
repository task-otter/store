// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasktest_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/task-otter/store/internal/tasktest"
)

type (
	taskfileValidationCase struct {
		name    string
		content string
		want    string
		tasks   []string
		vars    []string
	}

	loadTaskfileCase struct {
		name    string
		module  string
		content string
		want    string
	}

	readmeValidationCase struct {
		name    string
		content *string
		want    string
	}

	namedTestCase interface {
		testName() string
	}

	caseAssert[caseT namedTestCase] func(*testing.T, caseT)

	repoFixture struct {
		root   string
		module string
	}

	fakeTest struct {
		fatalMessages chan string
	}
)

const (
	missingModule      = "missing"
	fixtureModule      = "fixture"
	buildTask          = "build"
	fooVar             = "FOO"
	taskfileName       = "Taskfile.yml"
	readmeName         = "README.md"
	taskfileVersion    = "3.5"
	trailingWhitespace = "trailing whitespace"
	driftCaseName      = "drift"
	buildCommandLine   = "    cmds: [echo build]\n"
	emptyContent       = ""
	once               = 1
	dirMode            = 0o700
	fileMode           = 0o600
	chdirLockRetry     = time.Millisecond
	expectedTaskCount  = 3
)

func (tester *fakeTest) Fatal(args ...any) { tester.reportFatal(fmt.Sprint(args...)) }

func (tester *fakeTest) Fatalf(format string, args ...any) {
	tester.reportFatal(fmt.Sprintf(format, args...))
}

func (*fakeTest) Helper() {}

func (tester *fakeTest) TempDir() string {
	dir, err := os.MkdirTemp("", "tasktest-fake-")
	if err != nil {
		tester.Fatalf("create temporary directory: %v", err)

		return emptyContent
	}

	return dir
}

func (tester *fakeTest) reportFatal(message string) {
	tester.fatalMessages <- message

	runtime.Goexit()
}

// TestRepositoryAndTaskfilePaths verifies repo discovery from nested module paths.
// TestRepositoryAndTaskfilePaths
func TestRepositoryAndTaskfilePaths(t *testing.T) {
	fixture := makeRepo(t)
	assertRepositoryAndTaskfilePaths(t, &fixture)
	t.Parallel()
}

// TestRepoRootMissingGoMod verifies repo discovery fails outside a Go module.
// TestRepoRootMissingGoMod
func TestRepoRootMissingGoMod(t *testing.T) {
	inDir(t, t.TempDir(), func() {
		expectFatal(
			t,
			"could not find repository root",
			func(fakeTester *fakeTest) { tasktest.RepoRoot(fakeTester) },
		)
	})

	t.Parallel()
}

// TestLoadTaskfile verifies valid and invalid Taskfile loading behavior.
// TestLoadTaskfile
func TestLoadTaskfile(t *testing.T) {
	t.Parallel()
	assertValidLoadTaskfile(t)
	runCases(t, loadTaskfileCases(), assertInvalidLoadTaskfile)
}

// TestReadmeValidation verifies README validation for Taskfile modules.
// TestReadmeValidation
func TestReadmeValidation(t *testing.T) {
	t.Parallel()
	assertValidModule(t)
	runCases(t, readmeValidationCases(), assertInvalidReadme)
}

// TestTaskfileValidation verifies Taskfile content validation for modules.
// TestTaskfileValidation
func TestTaskfileValidation(t *testing.T) {
	t.Parallel()
	assertValidModule(t)
	runCases(t, taskfileValidationCases(), assertInvalidTaskfile)
}

func expectFatal(t *testing.T, want string, fatalFunc func(*fakeTest)) {
	t.Helper()

	fatalMessages := make(chan string, once)
	done := make(chan struct{})

	runFatalFunc(done, fatalMessages, fatalFunc)
	assertFatalMessage(t, want, fatalMessages)
}

func runFatalFunc(done chan struct{}, fatalMessages chan string, fatalFunc func(*fakeTest)) {
	go func() {
		defer close(done)

		fatalFunc(&fakeTest{fatalMessages: fatalMessages})
	}()

	<-done
}

func assertFatalMessage(t *testing.T, want string, fatalMessages <-chan string) {
	t.Helper()

	fatalMessage, ok := receivedFatalMessage(fatalMessages)

	if !ok {
		t.Fatal("expected fatal call")
	}

	if !strings.Contains(fatalMessage, want) {
		t.Fatalf("fatal message %q does not contain %q", fatalMessage, want)
	}
}

func receivedFatalMessage(fatalMessages <-chan string) (message string, ok bool) {
	select {
	case message = <-fatalMessages:
		return message, true
	default:
		return emptyContent, false
	}
}

func inDir(t *testing.T, dir string, callback func()) {
	t.Helper()

	unlock := lockWorkingDirectory(t)

	defer unlock()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current directory: %v", err)
	}

	changeDir(t, dir)

	defer restoreDir(t, previous)

	callback()
}

func lockWorkingDirectory(t *testing.T) func() {
	t.Helper()

	lockPath := workingDirectoryLockPath(t)

	for {
		if lockAcquired(t, lockPath) {
			return func() { removePath(t, lockPath) }
		}

		time.Sleep(chdirLockRetry)
	}
}

func workingDirectoryLockPath(t *testing.T) string {
	t.Helper()

	return filepath.Join(
		filepath.Dir(filepath.Dir(t.TempDir())),
		fmt.Sprintf("tasktest-chdir-%d.lock", os.Getpid()),
	)
}

func lockAcquired(t *testing.T, lockPath string) bool {
	t.Helper()

	err := os.Mkdir(lockPath, dirMode)
	if err == nil {
		return true
	}

	if !os.IsExist(err) {
		t.Fatalf("lock working directory: %v", err)
	}

	return false
}

func changeDir(t *testing.T, dir string) {
	t.Helper()

	err := syscall.Chdir(dir)
	if err != nil {
		t.Fatalf("change directory to %s: %v", dir, err)
	}
}

func restoreDir(t *testing.T, previous string) {
	t.Helper()

	err := syscall.Chdir(previous)
	if err != nil {
		t.Fatalf("restore directory: %v", err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), dirMode)
	if err != nil {
		t.Fatalf("create parent directory: %v", err)
	}

	err = os.WriteFile(path, []byte(content), fileMode)
	if err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func validTaskfile() string {
	return `version: "3.5"
vars:
  FOO: value
tasks:
  default:
    desc: Show available tasks
    cmds: [task --list]
  build:
    desc: Build the fixture project
    cmds: [echo build]
  hidden:
    internal: true
    cmds: [echo hidden]
`
}

func validReadme() string {
	return "# Fixture\n\n## Public Tasks\n\n| Task | Description |\n| --- | --- |\n| `build` | Build fixture |\n"
}

func makeRepo(t *testing.T) repoFixture {
	t.Helper()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/fixture\n\ngo 1.22\n")

	module := filepath.Join(root, "taskfiles", fixtureModule)
	writeFile(t, filepath.Join(module, taskfileName), validTaskfile())
	writeFile(t, filepath.Join(module, readmeName), validReadme())

	return repoFixture{root: root, module: module}
}

func assertRepositoryAndTaskfilePaths(t *testing.T, fixture *repoFixture) {
	t.Helper()

	nested := filepath.Join(fixture.module, "nested")

	err := os.MkdirAll(nested, dirMode)
	if err != nil {
		t.Fatal(err)
	}

	inDir(t, nested, func() {
		if got := tasktest.RepoRoot(t); !samePath(t, got, fixture.root) {
			t.Fatalf("RepoRoot = %s, want %s", got, fixture.root)
		}

		taskfile := tasktest.LoadTaskfile(t, fixtureModule)

		if taskfile.Version != taskfileVersion {
			t.Fatalf("LoadTaskfile version = %s", taskfile.Version)
		}
	})
}

func (testCase *loadTaskfileCase) testName() string {
	return testCase.name
}

func samePath(t *testing.T, left, right string) bool {
	t.Helper()

	leftInfo, leftErr := os.Stat(left)
	rightInfo, rightErr := os.Stat(right)

	return leftErr == nil && rightErr == nil && os.SameFile(leftInfo, rightInfo)
}

func runCases[caseT namedTestCase](t *testing.T, cases []caseT, assert caseAssert[caseT]) {
	t.Helper()

	for index := range cases {
		testCase := cases[index]
		t.Run(testCase.testName(), func(t *testing.T) {
			t.Parallel()
			assert(t, testCase)
		})
	}
}

func assertValidLoadTaskfile(t *testing.T) {
	t.Helper()

	fixture := makeRepo(t)

	inDir(t, fixture.module, func() {
		taskfile := tasktest.LoadTaskfile(t, fixtureModule)

		if !loadedTaskfileMatchesFixture(taskfile) {
			t.Fatalf("unexpected Taskfile: %#v", taskfile)
		}
	})
}

func loadedTaskfileMatchesFixture(taskfile *tasktest.Taskfile) bool {
	return taskfile.Version == taskfileVersion &&
		taskfile.Vars[fooVar] == "value" &&
		len(taskfile.Tasks) == expectedTaskCount
}

func assertInvalidLoadTaskfile(t *testing.T, testCase *loadTaskfileCase) {
	t.Helper()

	fixture := makeRepo(t)

	if testCase.content != emptyContent {
		writeFile(t, filepath.Join(fixture.module, taskfileName), testCase.content)
	}

	inDir(t, fixture.module, func() {
		expectFatal(
			t,
			testCase.want,
			func(fakeTester *fakeTest) { tasktest.LoadTaskfile(fakeTester, testCase.module) },
		)
	})
}

func loadTaskfileCases() []*loadTaskfileCase {
	testCases := []*loadTaskfileCase{
		{
			name:    missingModule,
			module:  missingModule,
			content: emptyContent,
			want:    "read missing Taskfile",
		},
	}

	return append(testCases, invalidLoadTaskfileFormattingCases()...)
}

func invalidLoadTaskfileFormattingCases() []*loadTaskfileCase {
	return []*loadTaskfileCase{
		{
			name:    trailingWhitespace,
			module:  fixtureModule,
			content: "version: \"3\"\ntasks: {} \n",
			want:    trailingWhitespace,
		},
		{
			name:    "invalid yaml",
			module:  fixtureModule,
			content: "version: [\n",
			want:    "parse fixture Taskfile",
		},
		{
			name:    "crlf",
			module:  fixtureModule,
			content: "version: \"3\"\r\ntasks: {}\r\n",
			want:    "LF line endings",
		},
	}
}

func assertValidModule(t *testing.T) {
	t.Helper()

	fixture := makeRepo(t)
	expected := moduleExpectations([]string{buildTask}, []string{fooVar})

	inDir(t, fixture.module, func() {
		tasktest.AssertModule(t, fixtureModule, expected)
	})
}

func assertInvalidReadme(t *testing.T, testCase *readmeValidationCase) {
	t.Helper()

	fixture := makeRepo(t)
	prepareReadmeValidationFixture(t, fixture.module, testCase)

	expected := moduleExpectations([]string{buildTask}, []string{fooVar})

	inDir(t, fixture.module, func() {
		expectModuleFatal(t, testCase.want, expected)
	})
}

func prepareReadmeValidationFixture(t *testing.T, module string, testCase *readmeValidationCase) {
	t.Helper()

	path := filepath.Join(module, readmeName)

	if testCase.content == nil {
		removeFile(t, path)

		return
	}

	writeFile(t, path, *testCase.content)
}

func (testCase *readmeValidationCase) testName() string {
	return testCase.name
}

func readmeValidationCases() []*readmeValidationCase {
	return []*readmeValidationCase{
		{name: missingModule, content: nil, want: "must have README.md"},
		{name: "empty", content: new("\n"), want: "README.md is empty"},
		{name: "section", content: new("# Fixture\n"), want: "document public tasks"},
		{
			name:    "task",
			content: new("# Fixture\n\n## Public Tasks\n"),
			want:    "does not mention public task",
		},
	}
}

func removeFile(t *testing.T, path string) {
	t.Helper()

	removePath(t, path)
}

func removePath(t *testing.T, path string) {
	t.Helper()

	err := os.Remove(path)
	if err != nil {
		t.Fatal(err)
	}
}

func moduleExpectations(tasks, vars []string) *tasktest.ModuleExpectations {
	return &tasktest.ModuleExpectations{Tasks: tasks, Vars: vars}
}

func expectModuleFatal(t *testing.T, want string, expected *tasktest.ModuleExpectations) {
	t.Helper()

	expectFatal(t, want, func(fakeTester *fakeTest) {
		tasktest.AssertModule(fakeTester, fixtureModule, expected)
	})
}

func assertInvalidTaskfile(t *testing.T, testCase *taskfileValidationCase) {
	t.Helper()

	fixture := makeRepo(t)
	prepareTaskfileFixture(t, fixture.module, testCase)
	inDir(t, fixture.module, func() {
		expectModuleFatal(t, testCase.want, moduleExpectations(testCase.tasks, testCase.vars))
	})
}

func (testCase *taskfileValidationCase) testName() string {
	return testCase.name
}

func prepareTaskfileFixture(t *testing.T, module string, testCase *taskfileValidationCase) {
	t.Helper()

	writeFile(t, filepath.Join(module, taskfileName), testCase.content)

	if testCase.name == driftCaseName {
		writeFile(
			t,
			filepath.Join(module, readmeName),
			strings.Replace(validReadme(), "`build`", "`other`", once),
		)
	}
}

func taskfileValidationCases() []*taskfileValidationCase {
	testCases := taskfileValidationBaseCases()

	return append(testCases, taskfileValidationExpectationCases()...)
}

func taskfileValidationBaseCases() []*taskfileValidationCase {
	return []*taskfileValidationCase{
		{
			name:    "version",
			content: strings.Replace(validTaskfile(), `version: "3.5"`, `version: "2"`, once),
			tasks:   []string{buildTask},
			vars:    []string{fooVar},
			want:    "version must be 3",
		},
		{
			name:    "no tasks",
			content: "version: \"3\"\ntasks: {}\n",
			tasks:   nil,
			vars:    nil,
			want:    "must define tasks",
		},
	}
}

func taskfileValidationExpectationCases() []*taskfileValidationCase {
	testCases := taskfileValidationReadmeCases()

	return append(testCases, taskfileValidationCommandCases()...)
}

func taskfileValidationReadmeCases() []*taskfileValidationCase {
	return []*taskfileValidationCase{
		{
			name:    driftCaseName,
			content: validTaskfile(),
			tasks:   []string{"other"},
			vars:    []string{fooVar},
			want:    "public task drift",
		},
		{
			name:    "description",
			content: strings.Replace(validTaskfile(), "Build the fixture project", "short", once),
			tasks:   []string{buildTask},
			vars:    []string{fooVar},
			want:    "desc is missing or too short",
		},
	}
}

func taskfileValidationCommandCases() []*taskfileValidationCase {
	return []*taskfileValidationCase{
		{
			name:    "commands",
			content: strings.Replace(validTaskfile(), buildCommandLine, emptyContent, once),
			tasks:   []string{buildTask},
			vars:    []string{fooVar},
			want:    "must define cmds or deps",
		},
		{
			name:    "variable",
			content: validTaskfile(),
			tasks:   []string{buildTask},
			vars:    []string{"MISSING"},
			want:    "vars missing",
		},
	}
}
