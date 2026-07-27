package tasktest

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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

func TestCollectionAndEnvironmentHelpers(t *testing.T) {
	env := []string{"A=old", "PATH=/bin"}
	env = dryRunSetEnv(env, "A", "new")
	env = dryRunSetEnv(env, "B", "value")
	if got := dryRunGetEnv(env, "A"); got != "new" {
		t.Fatalf("A = %q", got)
	}
	if got := dryRunGetEnv(env, "B"); got != "value" {
		t.Fatalf("B = %q", got)
	}
	if got := dryRunGetEnv(env, "MISSING"); got != "" {
		t.Fatalf("missing env = %q", got)
	}

	tf := Taskfile{Tasks: map[string]Task{
		"default":  {Desc: "default"},
		"internal": {Internal: true},
		"zeta":     {Desc: "zeta"},
		"alpha":    {Desc: "alpha"},
	}}
	if got, want := publicTaskNames(tf), []string{"alpha", "zeta"}; !slices.Equal(got, want) {
		t.Fatalf("public tasks = %v, want %v", got, want)
	}

	original := []string{"z", "a"}
	if got := sortedCopy(original); !slices.Equal(got, []string{"a", "z"}) {
		t.Fatalf("sorted copy = %v", got)
	}
	if !slices.Equal(original, []string{"z", "a"}) {
		t.Fatalf("sortedCopy modified input: %v", original)
	}
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
	if tf.Version != "3" || tf.Vars["FOO"] != "value" || len(tf.Tasks) != 3 {
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

func TestDryRunEnvironment(t *testing.T) {
	project, env := setupDryRunEnv(t)
	for _, path := range []string{
		filepath.Join(project, ".stub-bin", "node"),
		filepath.Join(project, "package.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s: %v", path, err)
		}
	}
	for key, want := range map[string]string{"CI": "true", "NO_COLOR": "1", "TASK_ASSUME_YES": "true"} {
		if got := dryRunGetEnv(env, key); got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestDryRunEnvironmentFailures(t *testing.T) {
	base := t.TempDir()
	newDirs := func(name string) (string, string) {
		home := filepath.Join(base, name, "home")
		project := filepath.Join(base, name, "project")
		if err := os.MkdirAll(home, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(project, 0o755); err != nil {
			t.Fatal(err)
		}
		return home, project
	}

	t.Run("mkdir", func(t *testing.T) {
		home, project := newDirs("mkdir")
		if err := os.Remove(project); err != nil {
			t.Fatal(err)
		}
		writeFile(t, project, "file")
		ft := &fakeTest{tempDirs: []string{home, project}}
		expectFatalWith(t, ft, "create stub dir", func() { setupDryRunEnv(ft) })
	})

	tests := []struct {
		name  string
		block func(home, project string) string
		want  string
	}{
		{name: "stub", want: "write stub fnm", block: func(_, project string) string { return filepath.Join(project, ".stub-bin", "fnm") }},
		{name: "bun", want: "write bun file stub", block: func(home, _ string) string { return filepath.Join(home, ".bun", "bin", "bun") }},
		{name: "fnm", want: "write fnm file stub", block: func(home, _ string) string { return filepath.Join(home, ".local", "share", "fnm", "fnm") }},
		{name: "nvm", want: "write nvm.sh stub", block: func(home, _ string) string { return filepath.Join(home, ".nvm", "nvm.sh") }},
		{name: "package", want: "write package.json", block: func(_, project string) string { return filepath.Join(project, "package.json") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home, project := newDirs(tt.name)
			blocked := tt.block(home, project)
			if err := os.MkdirAll(blocked, 0o755); err != nil {
				t.Fatal(err)
			}
			ft := &fakeTest{tempDirs: []string{home, project}}
			expectFatalWith(t, ft, tt.want, func() { setupDryRunEnv(ft) })
		})
	}
}

func expectFatalWith(t *testing.T, ft *fakeTest, want string, fn func()) {
	t.Helper()
	defer func() {
		recovered := recover()
		fatal, ok := recovered.(fatalCall)
		if !ok || !strings.Contains(fatal.message, want) {
			t.Fatalf("fatal = %#v, want message containing %q", recovered, want)
		}
	}()
	fn()
	panic("expected fatal call")
}

func TestTaskCommandsAndAssertions(t *testing.T) {
	_, module := makeRepo(t)
	chdir(t, module)

	success := writeExecutable(t, `case "$*" in
  *--json*) printf '%s\n' '{}' ;;
  *) printf '%s\n' 'token tool download build' ;;
esac`)
	withTaskSettings(t, success, 10*time.Second)

	if output := RootDryRun(t, "build"); !strings.Contains(output, "token") {
		t.Fatalf("RootDryRun output = %q", output)
	}
	if output := DryRun(t, "fixture", "build"); !strings.Contains(output, "token") {
		t.Fatalf("DryRun output = %q", output)
	}
	AssertDryRunContains(t, "fixture", []string{"build"}, "token", "tool")
	AssertInstallDryRun(t, "fixture", "tool", "download")
	if output := runTask(t, "build"); !strings.Contains(output, "token") {
		t.Fatalf("runTask output = %q", output)
	}
	if output, err := runTaskOutput(t, "build"); err != nil || !strings.Contains(output, "token") {
		t.Fatalf("runTaskOutput = %q, %v", output, err)
	}
	assertTaskCliCanLoad(t, "fixture")
	AssertModule(t, "fixture", []string{"build"}, []string{"FOO"})

	expectFatal(t, "missing \"absent\"", func(ft testT) {
		AssertDryRunContains(ft, "fixture", []string{"build"}, "absent")
	})
	expectFatal(t, "install missing", func(ft testT) {
		AssertInstallDryRun(ft, "fixture", "tool", "absent")
	})
}

func TestInstallDryRunBranches(t *testing.T) {
	_, module := makeRepo(t)
	chdir(t, module)

	withTaskSettings(t, writeExecutable(t, "printf '%s\\n' 'tool up to date'"), 10*time.Second)
	AssertInstallDryRun(t, "fixture", "tool", "unused")

	withTaskSettings(t, writeExecutable(t, "printf '%s\\n' 'up to date'"), 10*time.Second)
	expectFatal(t, "skipped but missing", func(ft testT) {
		AssertInstallDryRun(ft, "fixture", "tool", "unused")
	})
}

func TestTaskCommandFailures(t *testing.T) {
	_, module := makeRepo(t)
	chdir(t, module)

	t.Run("dry run error", func(t *testing.T) {
		withTaskSettings(t, writeExecutable(t, "echo failure; exit 3"), 10*time.Second)
		expectFatal(t, "task command failed", func(ft testT) { DryRun(ft, "fixture", "build") })
	})

	t.Run("dry run timeout", func(t *testing.T) {
		withTaskSettings(t, writeExecutable(t, "sleep 1"), 10*time.Millisecond)
		expectFatal(t, "task command timed out", func(ft testT) { DryRun(ft, "fixture", "build") })
	})

	t.Run("run error", func(t *testing.T) {
		withTaskSettings(t, writeExecutable(t, "echo failure; exit 4"), 10*time.Second)
		if _, err := runTaskOutput(t, "build"); err == nil {
			t.Fatal("runTaskOutput succeeded")
		}
		expectFatal(t, "task command failed", func(ft testT) { runTask(ft, "build") })
	})

	t.Run("run timeout", func(t *testing.T) {
		withTaskSettings(t, writeExecutable(t, "sleep 1"), 10*time.Millisecond)
		expectFatal(t, "task command timed out", func(ft testT) { runTaskOutput(ft, "build") })
	})

	t.Run("invalid list json", func(t *testing.T) {
		withTaskSettings(t, writeExecutable(t, "echo not-json"), 10*time.Second)
		expectFatal(t, "invalid JSON", func(ft testT) { assertTaskCliCanLoad(ft, "fixture") })
	})
}
