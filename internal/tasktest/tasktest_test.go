package tasktest_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	tasktest "github.com/task-otter/store/internal/tasktest"
)

const (
	missingModule = "missing"
	fixtureModule = "fixture"
	buildTask     = "build"
	fooVar        = "FOO"
)

type taskfileValidationCase struct {
	name    string
	content string
	tasks   []string
	vars    []string
	want    string
}

type fatalCall struct{ message string }

type fakeTest struct {
	tempDirs []string
	nextDir  int
}

func (*fakeTest) Helper() {}

func (*fakeTest) Fatal(args ...any) { panic(fatalCall{message: fmt.Sprint(args...)}) }

func (*fakeTest) Fatalf(format string, args ...any) {
	panic(fatalCall{message: fmt.Sprintf(format, args...)})
}

func (f *fakeTest) TempDir() string {
	if f.nextDir < len(f.tempDirs) {
		dir := f.tempDirs[f.nextDir]
		f.nextDir++

		return dir
	}

	dir, err := os.MkdirTemp("", "tasktest-fake-")
	if err != nil {
		panic(err)
	}

	f.tempDirs = append(f.tempDirs, dir)
	f.nextDir++

	return dir
}

func expectFatal(t *testing.T, want string, fatalFunc func(*fakeTest)) {
	t.Helper()

	defer func() {
		recovered := recover()

		fatal, ok := recovered.(fatalCall)
		if !ok {
			t.Fatalf("expected fatal call, recovered %#v", recovered)
		}

		if !strings.Contains(fatal.message, want) {
			t.Fatalf("fatal message %q does not contain %q", fatal.message, want)
		}
	}()

	fatalFunc(&fakeTest{tempDirs: nil, nextDir: 0})
	panic("expected fatal call")
}

func inDir(t *testing.T, dir string, callback func()) {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current directory: %v", err)
	}

	err = syscall.Chdir(dir)
	if err != nil {
		t.Fatalf("change directory to %s: %v", dir, err)
	}

	defer func() {
		err := syscall.Chdir(previous)
		if err != nil {
			t.Fatalf("restore directory: %v", err)
		}
	}()

	callback()
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0o700)
	if err != nil {
		t.Fatalf("create parent directory: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o600)
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

func makeRepo(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/fixture\n\ngo 1.22\n")
	module := filepath.Join(root, "taskfiles", fixtureModule)
	writeFile(t, filepath.Join(module, "Taskfile.yml"), validTaskfile())
	writeFile(t, filepath.Join(module, "README.md"), validReadme())

	return root, module
}

func TestRepositoryAndTaskfilePaths(t *testing.T) {
	root, module := makeRepo(t)

	nested := filepath.Join(module, "nested")

	err := os.MkdirAll(nested, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	inDir(t, nested, func() {
		if got := tasktest.RepoRoot(t); !samePath(t, got, root) {
			t.Fatalf("RepoRoot = %s, want %s", got, root)
		}

		taskfile := tasktest.LoadTaskfile(t, fixtureModule)
		if taskfile.Version != "3.5" {
			t.Fatalf("LoadTaskfile version = %s", taskfile.Version)
		}
	})

	t.Parallel()
}

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

func samePath(t *testing.T, left, right string) bool {
	t.Helper()

	leftInfo, leftErr := os.Stat(left)
	rightInfo, rightErr := os.Stat(right)

	return leftErr == nil && rightErr == nil && os.SameFile(leftInfo, rightInfo)
}

func TestLoadTaskfile(t *testing.T) {
	_, module := makeRepo(t)

	inDir(t, module, func() {
		taskfile := tasktest.LoadTaskfile(t, fixtureModule)
		if taskfile.Version != "3.5" || taskfile.Vars[fooVar] != "value" ||
			len(taskfile.Tasks) != 3 {
			t.Fatalf("unexpected Taskfile: %#v", taskfile)
		}
	})

	tests := []struct {
		name    string
		module  string
		content string
		want    string
	}{
		{name: missingModule, module: missingModule, content: "", want: "read missing Taskfile"},
		{
			name:    "crlf",
			module:  fixtureModule,
			content: "version: \"3\"\r\ntasks: {}\r\n",
			want:    "LF line endings",
		},
		{
			name:    "trailing whitespace",
			module:  fixtureModule,
			content: "version: \"3\"\ntasks: {} \n",
			want:    "trailing whitespace",
		},
		{
			name:    "invalid yaml",
			module:  fixtureModule,
			content: "version: [\n",
			want:    "parse fixture Taskfile",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, module := makeRepo(t)
			if testCase.content != "" {
				writeFile(t, filepath.Join(module, "Taskfile.yml"), testCase.content)
			}

			inDir(t, module, func() {
				expectFatal(
					t,
					testCase.want,
					func(fakeTester *fakeTest) { tasktest.LoadTaskfile(fakeTester, testCase.module) },
				)
			})

			t.Parallel()
		})
	}

	t.Parallel()
}

func TestReadmeValidation(t *testing.T) {
	_, module := makeRepo(t)
	inDir(t, module, func() {
		tasktest.AssertModule(t, fixtureModule, []string{buildTask}, []string{fooVar})
	})

	tests := []struct {
		name    string
		content *string
		want    string
	}{
		{name: missingModule, content: nil, want: "must have README.md"},
		{name: "empty", content: new("\n"), want: "README.md is empty"},
		{name: "section", content: new("# Fixture\n"), want: "document public tasks"},
		{
			name:    "task",
			content: new("# Fixture\n\n## Public Tasks\n"),
			want:    "does not mention public task",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, module := makeRepo(t)

			path := filepath.Join(module, "README.md")
			if testCase.content == nil {
				err := os.Remove(path)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				writeFile(t, path, *testCase.content)
			}

			inDir(t, module, func() {
				expectFatal(t, testCase.want, func(fakeTester *fakeTest) {
					tasktest.AssertModule(
						fakeTester,
						fixtureModule,
						[]string{buildTask},
						[]string{fooVar},
					)
				})
			})

			t.Parallel()
		})
	}

	t.Parallel()
}

func TestTaskfileValidation(t *testing.T) {
	t.Parallel()

	_, module := makeRepo(t)
	inDir(t, module, func() {
		tasktest.AssertModule(t, fixtureModule, []string{buildTask}, []string{fooVar})
	})

	for _, testCase := range taskfileValidationCases() {
		t.Run(testCase.name, func(t *testing.T) {
			_, module := makeRepo(t)
			prepareTaskfileValidationFixture(t, module, testCase)
			inDir(t, module, func() {
				expectFatal(t, testCase.want, func(fakeTester *fakeTest) {
					tasktest.AssertModule(fakeTester, fixtureModule, testCase.tasks, testCase.vars)
				})
			})

			t.Parallel()
		})
	}
}

func prepareTaskfileValidationFixture(
	t *testing.T,
	module string,
	testCase taskfileValidationCase,
) {
	t.Helper()

	writeFile(t, filepath.Join(module, "Taskfile.yml"), testCase.content)

	if testCase.name == "drift" {
		writeFile(
			t,
			filepath.Join(module, "README.md"),
			strings.Replace(validReadme(), "`build`", "`other`", 1),
		)
	}
}

func taskfileValidationCases() []taskfileValidationCase {
	return []taskfileValidationCase{
		{
			name:    "version",
			content: strings.Replace(validTaskfile(), `version: "3.5"`, `version: "2"`, 1),
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
		{
			name:    "drift",
			content: validTaskfile(),
			tasks:   []string{"other"},
			vars:    []string{fooVar},
			want:    "public task drift",
		},
		{
			name:    "description",
			content: strings.Replace(validTaskfile(), "Build the fixture project", "short", 1),
			tasks:   []string{buildTask},
			vars:    []string{fooVar},
			want:    "desc is missing or too short",
		},
		{
			name:    "commands",
			content: strings.Replace(validTaskfile(), "    cmds: [echo build]\n", "", 1),
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
