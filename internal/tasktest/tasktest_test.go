package tasktest

import (
	"fmt"
	"os"
	"path/filepath"

	"strings"
	"testing"
	"time"
)

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

func expectFatal(t *testing.T, want string, fn func(testT)) {
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
	fn(&fakeTest{})
	panic("expected fatal call")
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change directory to %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore directory: %v", err)
		}
	})
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func writeExecutable(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "task-stub")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}
	return path
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
	module := filepath.Join(root, "taskfiles", "fixture")
	writeFile(t, filepath.Join(module, "Taskfile.yml"), validTaskfile())
	writeFile(t, filepath.Join(module, "README.md"), validReadme())
	return root, module
}

func withTaskSettings(t *testing.T, binary string, timeout time.Duration) {
	t.Helper()
	oldBinary, oldTimeout := taskBinary, taskTimeout
	taskBinary, taskTimeout = binary, timeout
	t.Cleanup(func() {
		taskBinary, taskTimeout = oldBinary, oldTimeout
	})
}

func TestRepositoryAndTaskfilePaths(t *testing.T) {
	root, module := makeRepo(t)
	nested := filepath.Join(module, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	chdir(t, nested)

	if got := RepoRoot(t); !samePath(t, got, root) {
		t.Fatalf("RepoRoot = %s, want %s", got, root)
	}
	if got := moduleDir(t, "fixture"); !samePath(t, got, module) {
		t.Fatalf("moduleDir = %s, want %s", got, module)
	}
	if got := taskfilePath(t, "fixture"); !samePath(t, got, filepath.Join(module, "Taskfile.yml")) {
		t.Fatalf("taskfilePath = %s", got)
	}
}

func TestRepoRootFailures(t *testing.T) {
	t.Run("missing go.mod", func(t *testing.T) {
		chdir(t, t.TempDir())
		expectFatal(t, "could not find repository root", func(ft testT) { RepoRoot(ft) })
	})

	t.Run("getwd", func(t *testing.T) {
		old := getWorkingDir
		getWorkingDir = func() (string, error) { return "", fmt.Errorf("getwd sentinel") }
		t.Cleanup(func() { getWorkingDir = old })
		expectFatal(t, "get working directory", func(ft testT) { RepoRoot(ft) })
	})
}

func samePath(t *testing.T, left, right string) bool {
	t.Helper()
	leftInfo, leftErr := os.Stat(left)
	rightInfo, rightErr := os.Stat(right)
	return leftErr == nil && rightErr == nil && os.SameFile(leftInfo, rightInfo)
}

func TestLoadTaskfile(t *testing.T) {
	_, module := makeRepo(t)
	chdir(t, module)

	tf := LoadTaskfile(t, "fixture")
	if tf.Version != "3.5" || tf.Vars["FOO"] != "value" || len(tf.Tasks) != 3 {
		t.Fatalf("unexpected Taskfile: %#v", tf)
	}

	tests := []struct {
		name    string
		module  string
		content string
		want    string
	}{
		{name: "missing", module: "missing", want: "read missing Taskfile"},
		{name: "crlf", module: "fixture", content: "version: \"3\"\r\ntasks: {}\r\n", want: "LF line endings"},
		{name: "trailing whitespace", module: "fixture", content: "version: \"3\"\ntasks: {} \n", want: "trailing whitespace"},
		{name: "invalid yaml", module: "fixture", content: "version: [\n", want: "parse fixture Taskfile"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.content != "" {
				writeFile(t, filepath.Join(module, "Taskfile.yml"), tt.content)
			}
			expectFatal(t, tt.want, func(ft testT) { LoadTaskfile(ft, tt.module) })
			writeFile(t, filepath.Join(module, "Taskfile.yml"), validTaskfile())
		})
	}
}

func TestReadmeValidation(t *testing.T) {
	_, module := makeRepo(t)
	chdir(t, module)
	assertReadme(t, "fixture", []string{"build"})

	tests := []struct {
		name    string
		content *string
		want    string
	}{
		{name: "missing", content: nil, want: "must have README.md"},
		{name: "empty", content: ptr("\n"), want: "README.md is empty"},
		{name: "section", content: ptr("# Fixture\n"), want: "document public tasks"},
		{name: "task", content: ptr("# Fixture\n\n## Public Tasks\n"), want: "does not mention public task"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(module, "README.md")
			if tt.content == nil {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			} else {
				writeFile(t, path, *tt.content)
			}
			expectFatal(t, tt.want, func(ft testT) { assertReadme(ft, "fixture", []string{"build"}) })
			writeFile(t, path, validReadme())
		})
	}
}

func ptr(value string) *string { return &value }

func TestTaskfileValidation(t *testing.T) {
	_, module := makeRepo(t)
	chdir(t, module)
	assertTaskfile(t, "fixture", []string{"build"}, []string{"FOO"})

	tests := []struct {
		name    string
		content string
		tasks   []string
		vars    []string
		want    string
	}{
		{name: "version", content: strings.Replace(validTaskfile(), `version: "3.5"`, `version: "2"`, 1), tasks: []string{"build"}, vars: []string{"FOO"}, want: "version must be 3"},
		{name: "no tasks", content: "version: \"3\"\ntasks: {}\n", want: "must define tasks"},
		{name: "drift", content: validTaskfile(), tasks: []string{"other"}, vars: []string{"FOO"}, want: "public task drift"},
		{name: "description", content: strings.Replace(validTaskfile(), "Build the fixture project", "short", 1), tasks: []string{"build"}, vars: []string{"FOO"}, want: "desc is missing or too short"},
		{name: "commands", content: strings.Replace(validTaskfile(), "    cmds: [echo build]\n", "", 1), tasks: []string{"build"}, vars: []string{"FOO"}, want: "must define cmds or deps"},
		{name: "variable", content: validTaskfile(), tasks: []string{"build"}, vars: []string{"MISSING"}, want: "vars missing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeFile(t, filepath.Join(module, "Taskfile.yml"), tt.content)
			expectFatal(t, tt.want, func(ft testT) { assertTaskfile(ft, "fixture", tt.tasks, tt.vars) })
		})
	}
}

// deduplicated per family
