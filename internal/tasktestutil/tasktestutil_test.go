package tasktestutil

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
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
	dir, err := os.MkdirTemp("", "tasktestutil-fake-")
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

func chdir(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeExecutable(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "task")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func validTaskfile() string {
	return `version: "3.5"
output:
  group:
    begin: "::group::{{.TASK}}"
    end: "::endgroup::"
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
}

func makeModule(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "Taskfile.yml"), validTaskfile())
	writeFile(t, filepath.Join(root, "README.md"), "# Fixture\n\n## Public Tasks\n\n| Task | Description |\n| --- | --- |\n| `alpha` | Alpha |\n\n## Variables\n")
	return root
}

func parseYAML(t *testing.T, content string) *yaml.Node {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(content), &doc); err != nil {
		t.Fatalf("parse YAML: %v", err)
	}
	return &doc
}

func samePath(t *testing.T, left, right string) bool {
	t.Helper()
	leftInfo, leftErr := os.Stat(left)
	rightInfo, rightErr := os.Stat(right)
	return leftErr == nil && rightErr == nil && os.SameFile(leftInfo, rightInfo)
}

func TestTaskNodeAndYamlHelpers(t *testing.T) {
	doc := parseYAML(t, `task:
  desc: "  description  "
  enabled: TRUE
  aliases: [one, two]
  nested:
    value: text
  sequence: [alpha, "", beta]
`)
	root := DocumentRoot(t, doc)
	taskNode := MappingField(root, "task")
	task := TaskNode{Name: "task", Node: taskNode}

	if task.Field("desc") == nil || task.Field("missing") != nil {
		t.Fatal("TaskNode.Field lookup failed")
	}
	if (&TaskNode{}).Field("anything") != nil {
		t.Fatal("nil TaskNode returned a field")
	}
	if (TaskNode{Node: &yaml.Node{Kind: yaml.ScalarNode}}).Field("anything") != nil {
		t.Fatal("scalar TaskNode returned a field")
	}
	if got := task.StringField("desc"); got != "description" {
		t.Fatalf("StringField = %q", got)
	}
	if !task.BoolField("enabled") || task.BoolField("missing") || task.BoolField("desc") {
		t.Fatal("BoolField result mismatch")
	}
	if !HasAlias(task, "two") || HasAlias(task, "missing") {
		t.Fatal("HasAlias result mismatch")
	}
	if HasAlias(TaskNode{}, "one") || HasAlias(TaskNode{Node: &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{
		{Kind: yaml.ScalarNode, Value: "aliases"}, {Kind: yaml.ScalarNode, Value: "one"},
	}}}, "one") {
		t.Fatal("invalid aliases were accepted")
	}

	if MappingField(root, "missing") != nil || MappingField(root, "task") == nil {
		t.Fatal("MappingField mismatch")
	}
	if MappingField(root, "missing") != nil || MappingField(root, "task").Kind != yaml.MappingNode {
		t.Fatal("mapping kind mismatch")
	}
	if MappingField(root, "task").Kind == yaml.ScalarNode {
		t.Fatal("impossible mapping kind")
	}
	if ScalarField(taskNode, "desc") != "description" {
		t.Fatal("ScalarField mismatch")
	}
	if NodeMappingValue(nil, "x") != nil || NodeMappingValue(&yaml.Node{Kind: yaml.ScalarNode}, "x") != nil {
		t.Fatal("NodeMappingValue accepted invalid node")
	}
	if NodeMappingValue(taskNode, "missing") != nil || NodeMappingValue(taskNode, "desc") == nil {
		t.Fatal("NodeMappingValue lookup mismatch")
	}

	if NodeText(nil) != "" || NodeText(&yaml.Node{Kind: yaml.ScalarNode, Value: " x "}) != "x" {
		t.Fatal("NodeText scalar mismatch")
	}
	if got := NodeText(NodeMappingValue(taskNode, "sequence")); got != "alpha beta" {
		t.Fatalf("NodeText sequence = %q", got)
	}
	if !IsEmptyNode(nil) || !IsEmptyNode(&yaml.Node{Kind: yaml.ScalarNode, Value: " "}) ||
		IsEmptyNode(&yaml.Node{Kind: yaml.ScalarNode, Value: "x"}) ||
		!IsEmptyNode(&yaml.Node{Kind: yaml.SequenceNode}) ||
		IsEmptyNode(&yaml.Node{Kind: yaml.SequenceNode, Content: []*yaml.Node{{Kind: yaml.ScalarNode, Value: "x"}}}) {
		t.Fatal("IsEmptyNode mismatch")
	}

	expectFatal(t, "invalid YAML document", func(ft testT) {
		DocumentRoot(ft, &yaml.Node{Kind: yaml.ScalarNode})
	})
	expectFatal(t, "root must be a YAML mapping", func(ft testT) {
		DocumentRoot(ft, &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.SequenceNode}}})
	})
}

func TestModuleDiscoveryAndLoading(t *testing.T) {
	root := makeModule(t)
	nested := filepath.Join(root, "nested", "deeper")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	chdir(t, nested)

	if got := ModuleRoot(t); !samePath(t, got, root) {
		t.Fatalf("ModuleRoot = %s, want %s", got, root)
	}
	if got := ModuleTaskfilePath(t); !samePath(t, got, filepath.Join(root, "Taskfile.yml")) {
		t.Fatalf("ModuleTaskfilePath = %s", got)
	}

	tf := LoadTaskfile(t)
	if !samePath(t, tf.Path, filepath.Join(root, "Taskfile.yml")) || tf.Root.Name != "root" || len(tf.Tasks) != 5 {
		t.Fatalf("unexpected LoadedTaskfile: %#v", tf)
	}
	if MustTask(t, tf, "alpha").Name != "alpha" {
		t.Fatal("MustTask returned wrong task")
	}
	if got, want := PublicTaskNamesFromTaskfile(t, tf), []string{"alpha"}; !slices.Equal(got, want) {
		t.Fatalf("public tasks = %v, want %v", got, want)
	}

	if err := os.Rename(filepath.Join(root, "Taskfile.yml"), filepath.Join(root, "Taskfile.yaml")); err != nil {
		t.Fatal(err)
	}
	if got := ModuleTaskfilePath(t); !samePath(t, got, filepath.Join(root, "Taskfile.yaml")) {
		t.Fatalf("YAML path = %s", got)
	}

	expectFatal(t, "is missing", func(ft testT) { MustTask(ft, tf, "missing") })
	expectFatal(t, "could not find Taskfile", func(ft testT) { moduleTaskfilePath(ft, t.TempDir()) })
}

func TestModuleDiscoveryFailures(t *testing.T) {
	t.Run("getwd", func(t *testing.T) {
		old := getWorkingDir
		getWorkingDir = func() (string, error) { return "", errors.New("getwd sentinel") }
		t.Cleanup(func() { getWorkingDir = old })
		expectFatal(t, "failed to get working directory", func(ft testT) { ModuleRoot(ft) })
	})

	t.Run("no taskfile", func(t *testing.T) {
		chdir(t, t.TempDir())
		expectFatal(t, "could not find Taskfile", func(ft testT) { ModuleRoot(ft) })
	})
}

func TestLoadTaskfileFailures(t *testing.T) {
	root := makeModule(t)
	chdir(t, root)
	path := filepath.Join(root, "Taskfile.yml")

	writeFile(t, path, "version: [\n")
	expectFatal(t, "failed to parse Taskfile", func(ft testT) { LoadTaskfile(ft) })

	writeFile(t, path, "version: \"3\"\n")
	expectFatal(t, "has no tasks map", func(ft testT) { LoadTaskfile(ft) })

	expectFatal(t, "failed to read", func(ft testT) { ReadFile(ft, filepath.Join(root, "missing")) })
}

func TestCommandResultsAndRunners(t *testing.T) {
	root := t.TempDir()
	stub := writeExecutable(t, `printf 'stdout:%s' "$*"
printf 'stderr' >&2
if [ "${FAIL_TASK:-}" = yes ]; then exit 7; fi`)
	t.Setenv("TASK_BIN", stub)

	result := RunTask(t, root, nil, "alpha", "B=2")
	if result.Stdout != "stdout:alpha B=2" || result.Stderr != "stderr" || result.Err != nil ||
		!slices.Equal(result.Args, []string{"alpha", "B=2"}) {
		t.Fatalf("unexpected RunTask result: %#v", result)
	}
	if result.Combined() != "stdout:alpha B=2\nstderr" {
		t.Fatalf("Combined = %q", result.Combined())
	}

	env := SetEnv(os.Environ(), "FAIL_TASK", "yes")
	failed := RunTaskTimeout(t, root, env, time.Second, "alpha")
	if failed.Err == nil {
		t.Fatal("RunTaskTimeout succeeded unexpectedly")
	}

	sleeping := writeExecutable(t, "sleep 1")
	t.Setenv("TASK_BIN", sleeping)
	if timed := RunTaskTimeout(t, root, nil, 10*time.Millisecond, "alpha"); timed.Err == nil {
		t.Fatal("timed command succeeded")
	}
}

func TestDefaultTaskBinaryAndSimpleRunner(t *testing.T) {
	root := t.TempDir()
	bin := t.TempDir()
	stub := filepath.Join(bin, "task")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\nprintf simple\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TASK_BIN", "")

	// Generous: this asserts the happy path, and the suite runs tests in
	// parallel, so a short deadline turns CPU pressure into a flake.
	if result := RunTaskTimeout(t, root, os.Environ(), 30*time.Second, "alpha"); result.Stdout != "simple" || result.Err != nil {
		t.Fatalf("default task result: %#v", result)
	}
	if result := RunSimpleTask(t, root, os.Environ(), "alpha"); result.Output != "simple" || result.Err != nil {
		t.Fatalf("simple task result: %#v", result)
	}
}

func TestEnvironmentHelpers(t *testing.T) {
	env := IsolatedEnv(t)
	home := EnvValue(env, "HOME")
	if home == "" || EnvValue(env, "PROFILE") != filepath.Join(home, ".bashrc") ||
		EnvValue(env, "CI") != "true" || EnvValue(env, "MISSING") != "" {
		t.Fatalf("isolated env mismatch: %v", env)
	}
	if !FileExists(filepath.Join(home, ".bashrc")) || FileExists(filepath.Join(home, "missing")) {
		t.Fatal("FileExists mismatch")
	}

	values := []string{"A=old"}
	values = SetEnv(values, "A", "new")
	values = SetEnv(values, "B", "value")
	if EnvValue(values, "A") != "new" || EnvValue(values, "B") != "value" {
		t.Fatalf("SetEnv mismatch: %v", values)
	}

	homeFailure := t.TempDir()
	if err := os.Mkdir(filepath.Join(homeFailure, ".bashrc"), 0o755); err != nil {
		t.Fatal(err)
	}
	ft := &fakeTest{tempDirs: []string{homeFailure}}
	expectFatalWith(t, ft, "failed to create fake shell profile", func() { IsolatedEnv(ft) })
}

func TestCollectionAndTextHelpers(t *testing.T) {
	specs := []PublicTaskSpec{{Name: "zeta"}, {Name: "alpha"}}
	if got := ExpectedPublicTaskNames(specs); !slices.Equal(got, []string{"alpha", "zeta"}) {
		t.Fatalf("expected names = %v", got)
	}
	if TaskArgs(nil) != nil || TaskArgs(map[string]string{}) != nil {
		t.Fatal("empty TaskArgs must be nil")
	}
	if got := TaskArgs(map[string]string{"Z": "2", "A": "1"}); !slices.Equal(got, []string{"A=1", "Z=2"}) {
		t.Fatalf("TaskArgs = %v", got)
	}
	if FormatList([]string{"a", "b"}) != "- a\n- b" || FormatList(nil) != "- " {
		t.Fatal("FormatList mismatch")
	}

	tasks := map[string]any{
		"default":  map[string]any{},
		"_private": map[string]any{},
		"internal": map[string]any{"internal": true},
		"alpha":    map[string]any{"desc": "alpha"},
		"scalar":   "value",
	}
	if got := SimplePublicTaskNames(tasks); !slices.Equal(got, []string{"alpha", "scalar"}) {
		t.Fatalf("simple public names = %v", got)
	}

	readme := "# Module\n\n## Public Tasks\n\n| Task | Description |\n| --- | --- |\n| `zeta` | Z |\n| `alpha` | A |\n\n## Variables\n| Name | Value |\n"
	if got := ReadmePublicTaskNames(readme); !slices.Equal(got, []string{"alpha", "zeta"}) {
		t.Fatalf("README names = %v", got)
	}
	if got := ReadmePublicTaskNames("# No table\n"); len(got) != 0 {
		t.Fatalf("unexpected README names: %v", got)
	}
}

func TestFileHelpers(t *testing.T) {
	dir := t.TempDir()
	WriteStub(t, dir, "stub", "#!/bin/sh\necho stub\n")
	stub := filepath.Join(dir, "stub")
	if got := MustRead(t, stub); !strings.Contains(got, "echo stub") {
		t.Fatalf("MustRead = %q", got)
	}
	if got := ReadFile(t, stub); got == "" {
		t.Fatal("ReadFile returned empty content")
	}
	expectFatal(t, "write broken stub", func(ft testT) {
		WriteStub(ft, filepath.Join(dir, "missing"), "broken", "body")
	})
	expectFatal(t, "read", func(ft testT) { MustRead(ft, filepath.Join(dir, "missing")) })
}

func TestCommandStringExtraction(t *testing.T) {
	doc := parseYAML(t, `cmds:
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
`)
	root := DocumentRoot(t, doc)
	commands := CollectCommandStrings(NodeMappingValue(root, "cmds"))
	want := []string{"echo scalar", "./scripts/run.sh --flag", "echo shell", "echo status", "echo precondition"}
	if !slices.Equal(commands, want) {
		t.Fatalf("commands = %v, want %v", commands, want)
	}
	if CollectCommandStrings(nil) != nil || len(CollectCommandStrings(&yaml.Node{Kind: yaml.ScalarNode, Value: " "})) != 0 {
		t.Fatal("empty command extraction mismatch")
	}
	if got := ReferencedLocalShellScripts("./one.sh --flag && echo x\n ./two/path.sh "); !slices.Equal(got, []string{"./one.sh", "./two/path.sh"}) {
		t.Fatalf("script references = %v", got)
	}
	if got := ReferencedLocalShellScripts("scripts/no-prefix.sh"); len(got) != 0 {
		t.Fatalf("unexpected references = %v", got)
	}
}

func TestAssertExitCode(t *testing.T) {
	AssertExitCode(t, CommandResult{Args: []string{"ok"}}, 0)

	err := exec.Command("sh", "-c", "exit 7").Run()
	AssertExitCode(t, CommandResult{Err: err, Args: []string{"exit"}}, 7)

	expectFatal(t, "without exit code", func(ft testT) {
		AssertExitCode(ft, CommandResult{Err: errors.New("sentinel")}, 1)
	})
	expectFatal(t, "expected exit code", func(ft testT) {
		AssertExitCode(ft, CommandResult{Args: []string{"ok"}}, 2)
	})
}

func TestBasicAssertions(t *testing.T) {
	AssertContains(t, "alpha beta", "beta")
	AssertNotContains(t, "alpha", "beta")
	AssertNotEmpty(t, " value ", "must not be empty")
	expectFatal(t, "expected output to contain", func(ft testT) { AssertContains(ft, "alpha", "beta") })
	expectFatal(t, "expected output not to contain", func(ft testT) { AssertNotContains(ft, "alpha", "alpha") })
	expectFatal(t, "empty sentinel", func(ft testT) { AssertNotEmpty(ft, " \n", "empty sentinel") })
}

func TestFilesystemAssertions(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	writeFile(t, file, "content")
	nonempty := filepath.Join(dir, "nonempty")
	writeFile(t, filepath.Join(nonempty, "entry"), "entry")
	empty := filepath.Join(dir, "empty")
	if err := os.Mkdir(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "missing")

	AssertFileExists(t, file)
	AssertDirExists(t, dir)
	AssertDirNotExists(t, missing)
	AssertDirHasEntries(t, nonempty)

	expectFatal(t, "expected file", func(ft testT) { AssertFileExists(ft, missing) })
	expectFatal(t, "found directory", func(ft testT) { AssertFileExists(ft, dir) })
	expectFatal(t, "expected directory", func(ft testT) { AssertDirExists(ft, missing) })
	expectFatal(t, "found file", func(ft testT) { AssertDirExists(ft, file) })
	expectFatal(t, "but it does", func(ft testT) { AssertDirNotExists(ft, dir) })
	expectFatal(t, "failed to read directory", func(ft testT) { AssertDirHasEntries(ft, missing) })
	expectFatal(t, "at least one entry", func(ft testT) { AssertDirHasEntries(ft, empty) })
}

func groupOutputNode(t *testing.T, begin, end string, errorOnly *string) *yaml.Node {
	t.Helper()
	value := ""
	if errorOnly != nil {
		value = "\n    error_only: " + *errorOnly
	}
	doc := parseYAML(t, "output:\n  group:\n    begin: \""+begin+"\"\n    end: \""+end+"\""+value+"\n")
	return NodeMappingValue(DocumentRoot(t, doc), "output")
}

func TestGithubGroupAssertion(t *testing.T) {
	falseValue := "false"
	AssertGithubGroupOutput(t, "alpha", groupOutputNode(t, "::group::{{.TASK}}", "::endgroup::", &falseValue))

	expectFatal(t, "no output config", func(ft testT) { AssertGithubGroupOutput(ft, "alpha", nil) })
	expectFatal(t, "advanced object format", func(ft testT) {
		AssertGithubGroupOutput(ft, "alpha", &yaml.Node{Kind: yaml.ScalarNode, Value: "group"})
	})
	expectFatal(t, "include group config", func(ft testT) {
		AssertGithubGroupOutput(ft, "alpha", &yaml.Node{Kind: yaml.MappingNode})
	})
	expectFatal(t, "include group config", func(ft testT) {
		AssertGithubGroupOutput(ft, "alpha", &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "group"}, {Kind: yaml.ScalarNode, Value: "scalar"},
		}})
	})
	expectFatal(t, "output.group.begin", func(ft testT) {
		AssertGithubGroupOutput(ft, "alpha", groupOutputNode(t, "bad", "::endgroup::", &falseValue))
	})
	expectFatal(t, "output.group.end", func(ft testT) {
		AssertGithubGroupOutput(ft, "alpha", groupOutputNode(t, "::group::{{.TASK}}", "bad", &falseValue))
	})
	expectFatal(t, "explicitly set", func(ft testT) {
		AssertGithubGroupOutput(ft, "alpha", groupOutputNode(t, "::group::{{.TASK}}", "::endgroup::", nil))
	})
	trueValue := "true"
	expectFatal(t, "must be false", func(ft testT) {
		AssertGithubGroupOutput(ft, "alpha", groupOutputNode(t, "::group::{{.TASK}}", "::endgroup::", &trueValue))
	})
}

func TestTextFileAssertion(t *testing.T) {
	AssertTextFileClean(t, "clean.yml", "key: value\n")
	tests := []struct {
		content string
		want    string
	}{
		{content: "", want: "is empty"},
		{content: "key: value\r\n", want: "CRLF"},
		{content: "key:\tvalue\n", want: "contains tabs"},
		{content: "key: value", want: "end with a newline"},
		{content: "key: value \n", want: "trailing whitespace"},
	}
	for _, tt := range tests {
		expectFatal(t, tt.want, func(ft testT) { AssertTextFileClean(ft, "bad.yml", tt.content) })
	}
}

func TestYamlStructureAssertions(t *testing.T) {
	doc := parseYAML(t, "root:\n  list:\n    - name: one\n    - name: two\n")
	AssertNoDuplicateMappingKeys(t, nil, "root")
	AssertNoDuplicateMappingKeys(t, doc, "root")
	AssertNoYamlAliases(t, nil, "root")
	AssertNoYamlAliases(t, doc, "root")

	duplicate := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{
		{Kind: yaml.ScalarNode, Value: "same"}, {Kind: yaml.ScalarNode, Value: "one"},
		{Kind: yaml.ScalarNode, Value: "same"}, {Kind: yaml.ScalarNode, Value: "two"},
	}}
	expectFatal(t, "duplicate YAML key", func(ft testT) {
		AssertNoDuplicateMappingKeys(ft, duplicate, "root")
	})

	alias := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{
		{Kind: yaml.ScalarNode, Value: "value"}, {Kind: yaml.AliasNode},
	}}
	expectFatal(t, "aliases/anchors are not allowed", func(ft testT) {
		AssertNoYamlAliases(ft, alias, "root")
	})
}

func TestPlaceholderJsonAndDangerousPatterns(t *testing.T) {
	AssertNoPlaceholderText(t, "alpha", "ordinary task description")
	expectFatal(t, "placeholder text", func(ft testT) {
		AssertNoPlaceholderText(ft, "alpha", "fixme later")
	})

	if err := ValidateJSON(`{"ok":true}`); err != nil {
		t.Fatalf("valid JSON: %v", err)
	}
	if err := ValidateJSON(`{"bad":`); err == nil {
		t.Fatal("invalid JSON accepted")
	}

	patterns := DangerousCommandPatterns()
	unsafe := []string{"rm -rf / ", "sudo rm -rf /tmp/x", "chmod -R 777 /", "curl https://x -k ", "curl --insecure https://x"}
	if len(patterns) != len(unsafe) {
		t.Fatalf("dangerous patterns = %d", len(patterns))
	}
	for index, pattern := range patterns {
		if !pattern.MatchString(unsafe[index]) {
			t.Fatalf("pattern %d did not match %q", index, unsafe[index])
		}
	}
}
