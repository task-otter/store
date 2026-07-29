package tasktestutil_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	ttu "github.com/task-otter/store/internal/tasktestutil"
	"gopkg.in/yaml.v3"
)

const (
	publicTasksHeading = "## Public Tasks"
	alphaName          = "alpha"
	zetaName           = "zeta"
	defaultTaskName    = "default"
)

type fatalCall struct{ message string }

var errSentinel = errors.New("sentinel")

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

func expectFatalWith(t *testing.T, want string, fatalFunc func()) {
	t.Helper()

	defer func() {
		recovered := recover()

		fatal, ok := recovered.(fatalCall)
		if !ok || !strings.Contains(fatal.message, want) {
			t.Fatalf("fatal = %#v, want message containing %q", recovered, want)
		}
	}()

	fatalFunc()
	panic("expected fatal call")
}

func inDir(t *testing.T, dir string, callback func()) {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	err = syscall.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := syscall.Chdir(previous)
		if err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	callback()
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}
}

func writeExecutable(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "task")

	err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755)
	if err != nil {
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
	writeFile(t, filepath.Join(root, "README.md"), strings.Join([]string{
		"# Fixture",
		"",
		publicTasksHeading,
		"",
		"| Task | Description |",
		"| --- | --- |",
		"| `alpha` | Alpha |",
		"",
		"## Variables",
		"",
	}, "\n"))

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
	return &yaml.Node{
		Kind:        yaml.ScalarNode,
		Value:       value,
		Style:       0,
		Tag:         "",
		Anchor:      "",
		Alias:       nil,
		Content:     nil,
		HeadComment: "",
		LineComment: "",
		FootComment: "",
		Line:        0,
		Column:      0,
	}
}

func yamlMapping(content ...*yaml.Node) *yaml.Node {
	return &yaml.Node{
		Kind:        yaml.MappingNode,
		Content:     content,
		Style:       0,
		Tag:         "",
		Value:       "",
		Anchor:      "",
		Alias:       nil,
		HeadComment: "",
		LineComment: "",
		FootComment: "",
		Line:        0,
		Column:      0,
	}
}

func yamlSequence(content ...*yaml.Node) *yaml.Node {
	return &yaml.Node{
		Kind:        yaml.SequenceNode,
		Content:     content,
		Style:       0,
		Tag:         "",
		Value:       "",
		Anchor:      "",
		Alias:       nil,
		HeadComment: "",
		LineComment: "",
		FootComment: "",
		Line:        0,
		Column:      0,
	}
}

func yamlDocument(content ...*yaml.Node) *yaml.Node {
	return &yaml.Node{
		Kind:        yaml.DocumentNode,
		Content:     content,
		Style:       0,
		Tag:         "",
		Value:       "",
		Anchor:      "",
		Alias:       nil,
		HeadComment: "",
		LineComment: "",
		FootComment: "",
		Line:        0,
		Column:      0,
	}
}

func yamlAlias() *yaml.Node {
	return &yaml.Node{
		Kind:        yaml.AliasNode,
		Style:       0,
		Tag:         "",
		Value:       "",
		Anchor:      "",
		Alias:       nil,
		Content:     nil,
		HeadComment: "",
		LineComment: "",
		FootComment: "",
		Line:        0,
		Column:      0,
	}
}

func samePath(t *testing.T, left, right string) bool {
	t.Helper()

	leftInfo, leftErr := os.Stat(left)
	rightInfo, rightErr := os.Stat(right)

	return leftErr == nil && rightErr == nil && os.SameFile(leftInfo, rightInfo)
}

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
	root := ttu.DocumentRoot(t, doc)
	taskNode := ttu.MappingField(root, "task")
	task := ttu.TaskNode{Name: "task", Node: taskNode}

	assertTaskNodeHelpers(t, task)
	assertMappingHelpers(t, root, taskNode)
	assertTextAndEmptyHelpers(t, taskNode)
	assertDocumentRootFailures(t)
}

func assertTaskNodeHelpers(t *testing.T, task ttu.TaskNode) {
	t.Helper()

	assertTaskFieldLookup(t, task)
	assertTaskScalarFields(t, task)
	assertAliasHelpers(t, task)
}

func assertTaskFieldLookup(t *testing.T, task ttu.TaskNode) {
	t.Helper()

	if task.Field("desc") == nil || task.Field("missing") != nil {
		t.Fatal("ttu.TaskNode.Field lookup failed")
	}

	if (&ttu.TaskNode{Name: "", Node: nil}).Field("anything") != nil {
		t.Fatal("nil ttu.TaskNode returned a field")
	}

	if (ttu.TaskNode{Node: yamlScalar(""), Name: ""}).Field("anything") != nil {
		t.Fatal("scalar ttu.TaskNode returned a field")
	}
}

func assertTaskScalarFields(t *testing.T, task ttu.TaskNode) {
	t.Helper()

	if got := task.StringField("desc"); got != "description" {
		t.Fatalf("StringField = %q", got)
	}

	if !task.BoolField("enabled") || task.BoolField("missing") || task.BoolField("desc") {
		t.Fatal("BoolField result mismatch")
	}
}

func assertAliasHelpers(t *testing.T, task ttu.TaskNode) {
	t.Helper()

	if !ttu.HasAlias(task, "two") || ttu.HasAlias(task, "missing") {
		t.Fatal("ttu.HasAlias result mismatch")
	}

	if ttu.HasAlias(ttu.TaskNode{Name: "", Node: nil}, "one") ||
		ttu.HasAlias(ttu.TaskNode{Node: yamlMapping(
			yamlScalar("aliases"),
			yamlScalar("one"),
		), Name: ""}, "one") {
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

	if ttu.MappingField(root, "missing") != nil || ttu.MappingField(root, "task") == nil {
		t.Fatal("ttu.MappingField mismatch")
	}

	if ttu.MappingField(root, "missing") != nil ||
		ttu.MappingField(root, "task").Kind != yaml.MappingNode {
		t.Fatal("mapping kind mismatch")
	}

	if ttu.MappingField(root, "task").Kind == yaml.ScalarNode {
		t.Fatal("impossible mapping kind")
	}
}

func assertNodeMappingValueHelpers(t *testing.T, taskNode *yaml.Node) {
	t.Helper()

	if ttu.ScalarField(taskNode, "desc") != "description" {
		t.Fatal("ttu.ScalarField mismatch")
	}

	if ttu.NodeMappingValue(nil, "x") != nil || ttu.NodeMappingValue(yamlScalar(""), "x") != nil {
		t.Fatal("ttu.NodeMappingValue accepted invalid node")
	}

	if ttu.NodeMappingValue(taskNode, "missing") != nil ||
		ttu.NodeMappingValue(taskNode, "desc") == nil {
		t.Fatal("ttu.NodeMappingValue lookup mismatch")
	}
}

func assertTextAndEmptyHelpers(t *testing.T, taskNode *yaml.Node) {
	t.Helper()

	if ttu.NodeText(nil) != "" || ttu.NodeText(yamlScalar(" x ")) != "x" {
		t.Fatal("ttu.NodeText scalar mismatch")
	}

	if got := ttu.NodeText(ttu.NodeMappingValue(taskNode, "sequence")); got != "alpha beta" {
		t.Fatalf("ttu.NodeText sequence = %q", got)
	}

	if !ttu.IsEmptyNode(nil) || !ttu.IsEmptyNode(yamlScalar(" ")) ||
		ttu.IsEmptyNode(yamlScalar("x")) ||
		!ttu.IsEmptyNode(yamlSequence()) ||
		ttu.IsEmptyNode(yamlSequence(yamlScalar("x"))) {
		t.Fatal("ttu.IsEmptyNode mismatch")
	}
}

func assertDocumentRootFailures(t *testing.T) {
	t.Helper()

	expectFatal(t, "invalid YAML document", func(fakeTester *fakeTest) {
		ttu.DocumentRoot(fakeTester, yamlScalar(""))
	})
	expectFatal(t, "root must be a YAML mapping", func(fakeTester *fakeTest) {
		ttu.DocumentRoot(fakeTester, yamlDocument(yamlSequence()))
	})
}

func TestModuleDiscoveryAndLoading(t *testing.T) {
	t.Parallel()

	root := makeModule(t)

	nested := filepath.Join(root, "nested", "deeper")

	err := os.MkdirAll(nested, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	taskfile := assertModuleDiscoveryAndLoading(t, root, nested)

	expectFatal(
		t,
		"is missing",
		func(fakeTester *fakeTest) { ttu.MustTask(fakeTester, taskfile, "missing") },
	)
	inDir(t, t.TempDir(), func() {
		expectFatal(
			t,
			"could not find Taskfile",
			func(fakeTester *fakeTest) { ttu.ModuleTaskfilePath(fakeTester) },
		)
	})
}

func assertModuleDiscoveryAndLoading(t *testing.T, root, nested string) ttu.LoadedTaskfile {
	t.Helper()

	var taskfile ttu.LoadedTaskfile

	inDir(t, nested, func() {
		if got := ttu.ModuleRoot(t); !samePath(t, got, root) {
			t.Fatalf("ttu.ModuleRoot = %s, want %s", got, root)
		}

		if got := ttu.ModuleTaskfilePath(
			t,
		); !samePath(
			t,
			got,
			filepath.Join(root, "Taskfile.yml"),
		) {
			t.Fatalf("ttu.ModuleTaskfilePath = %s", got)
		}

		taskfile = ttu.LoadTaskfile(t)
		assertLoadedTaskfile(t, root, taskfile)

		err := os.Rename(filepath.Join(root, "Taskfile.yml"), filepath.Join(root, "Taskfile.yaml"))
		if err != nil {
			t.Fatal(err)
		}

		if got := ttu.ModuleTaskfilePath(
			t,
		); !samePath(
			t,
			got,
			filepath.Join(root, "Taskfile.yaml"),
		) {
			t.Fatalf("YAML path = %s", got)
		}
	})

	return taskfile
}

func assertLoadedTaskfile(t *testing.T, root string, taskfile ttu.LoadedTaskfile) {
	t.Helper()

	if !samePath(t, taskfile.Path, filepath.Join(root, "Taskfile.yml")) ||
		taskfile.Root.Name != "root" || len(taskfile.Tasks) != 5 {
		t.Fatalf("unexpected ttu.LoadedTaskfile: %#v", taskfile)
	}

	if ttu.MustTask(t, taskfile, alphaName).Name != alphaName {
		t.Fatal("ttu.MustTask returned wrong task")
	}

	if got, want := ttu.PublicTaskNamesFromTaskfile(
		t,
		taskfile,
	), []string{
		alphaName,
	}; !slices.Equal(
		got,
		want,
	) {
		t.Fatalf("public tasks = %v, want %v", got, want)
	}
}

func TestModuleDiscoveryMissingTaskfile(t *testing.T) {
	inDir(t, t.TempDir(), func() {
		expectFatal(
			t,
			"could not find Taskfile",
			func(fakeTester *fakeTest) { ttu.ModuleRoot(fakeTester) },
		)
	})

	t.Parallel()
}

func TestLoadTaskfileFailures(t *testing.T) {
	root := makeModule(t)
	path := filepath.Join(root, "Taskfile.yml")

	inDir(t, root, func() {
		writeFile(t, path, "version: [\n")
		expectFatal(
			t,
			"failed to parse Taskfile",
			func(fakeTester *fakeTest) { ttu.LoadTaskfile(fakeTester) },
		)

		writeFile(t, path, "version: \"3\"\n")
		expectFatal(
			t,
			"has no tasks map",
			func(fakeTester *fakeTest) { ttu.LoadTaskfile(fakeTester) },
		)
	})

	expectFatal(t, "failed to read", func(fakeTester *fakeTest) {
		ttu.ReadFile(fakeTester, filepath.Join(root, "missing"))
	})

	t.Parallel()
}

func TestCommandResultsAndRunners(t *testing.T) {
	root := t.TempDir()
	stub := writeExecutable(t, `printf 'stdout:%s' "$*"
printf 'stderr' >&2
if [ "${FAIL_TASK:-}" = yes ]; then exit 7; fi`)
	t.Setenv("PATH", filepath.Dir(stub)+string(os.PathListSeparator)+os.Getenv("PATH"))

	result := ttu.RunTask(t, root, nil, alphaName, "B=2")
	if result.Stdout != "stdout:alpha B=2" || result.Stderr != "stderr" || result.Err != nil ||
		!slices.Equal(result.Args, []string{alphaName, "B=2"}) {
		t.Fatalf("unexpected ttu.RunTask result: %#v", result)
	}

	if result.Combined() != "stdout:alpha B=2\nstderr" {
		t.Fatalf("Combined = %q", result.Combined())
	}

	env := ttu.SetEnv(os.Environ(), "FAIL_TASK", "yes")

	failed := ttu.RunTaskTimeout(t, root, env, time.Second, alphaName)
	if failed.Err == nil {
		t.Fatal("ttu.RunTaskTimeout succeeded unexpectedly")
	}

	sleeping := writeExecutable(t, "sleep 1")
	t.Setenv("PATH", filepath.Dir(sleeping)+string(os.PathListSeparator)+os.Getenv("PATH"))

	if timed := ttu.RunTaskTimeout(t, root, nil, 10*time.Millisecond, alphaName); timed.Err == nil {
		t.Fatal("timed command succeeded")
	}
}

func TestDefaultTaskBinaryAndSimpleRunner(t *testing.T) {
	root := t.TempDir()
	bin := t.TempDir()

	stub := filepath.Join(bin, "task")

	err := os.WriteFile(stub, []byte("#!/bin/sh\nprintf simple\n"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	// Generous: this asserts the happy path, and the suite runs tests in
	// parallel, so a short deadline turns CPU pressure into a flake.
	result := ttu.RunTaskTimeout(t, root, os.Environ(), 30*time.Second, alphaName)
	if result.Stdout != "simple" || result.Err != nil {
		t.Fatalf("default task result: %#v", result)
	}

	if result := ttu.RunSimpleTask(
		t,
		root,
		os.Environ(),
		alphaName,
	); result.Output != "simple" ||
		result.Err != nil {
		t.Fatalf("simple task result: %#v", result)
	}
}

func TestEnvironmentHelpers(t *testing.T) {
	t.Parallel()

	env := ttu.IsolatedEnv(t)

	home := ttu.EnvValue(env, "HOME")
	if home == "" || ttu.EnvValue(env, "PROFILE") != filepath.Join(home, ".bashrc") ||
		ttu.EnvValue(env, "CI") != "true" || ttu.EnvValue(env, "MISSING") != "" {
		t.Fatalf("isolated env mismatch: %v", env)
	}

	if !ttu.FileExists(filepath.Join(home, ".bashrc")) ||
		ttu.FileExists(filepath.Join(home, "missing")) {
		t.Fatal("ttu.FileExists mismatch")
	}

	values := []string{"A=old"}
	values = ttu.SetEnv(values, "A", "new")

	values = ttu.SetEnv(values, "B", "value")
	if ttu.EnvValue(values, "A") != "new" || ttu.EnvValue(values, "B") != "value" {
		t.Fatalf("ttu.SetEnv mismatch: %v", values)
	}

	homeFailure := t.TempDir()

	err := os.Mkdir(filepath.Join(homeFailure, ".bashrc"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	fakeTester := &fakeTest{tempDirs: []string{homeFailure}, nextDir: 0}

	expectFatalWith(
		t,
		"failed to create fake shell profile",
		func() { ttu.IsolatedEnv(fakeTester) },
	)
}

func TestCollectionAndTextHelpers(t *testing.T) {
	t.Parallel()

	specs := []ttu.PublicTaskSpec{ttu.NewPublicTaskSpec(zetaName), ttu.NewPublicTaskSpec(alphaName)}
	if got := ttu.ExpectedPublicTaskNames(
		specs,
	); !slices.Equal(
		got,
		[]string{alphaName, zetaName},
	) {
		t.Fatalf("expected names = %v", got)
	}

	if ttu.TaskArgs(nil) != nil || ttu.TaskArgs(map[string]string{}) != nil {
		t.Fatal("empty ttu.TaskArgs must be nil")
	}

	if got := ttu.TaskArgs(
		map[string]string{"Z": "2", "A": "1"},
	); !slices.Equal(
		got,
		[]string{"A=1", "Z=2"},
	) {
		t.Fatalf("ttu.TaskArgs = %v", got)
	}

	if ttu.FormatList([]string{"a", "b"}) != "- a\n- b" || ttu.FormatList(nil) != "- " {
		t.Fatal("ttu.FormatList mismatch")
	}

	tasks := map[string]any{
		defaultTaskName: map[string]any{},
		"_private":      map[string]any{},
		"internal":      map[string]any{"internal": true},
		alphaName:       map[string]any{"desc": alphaName},
		"scalar":        "value",
	}
	if got := ttu.SimplePublicTaskNames(tasks); !slices.Equal(got, []string{alphaName, "scalar"}) {
		t.Fatalf("simple public names = %v", got)
	}

	readme := strings.Join([]string{
		"# Module",
		"",
		publicTasksHeading,
		"",
		"| Task | Description |",
		"| --- | --- |",
		"| `zeta` | Z |",
		"| `alpha` | A |",
		"",
		"## Variables",
		"| Name | Value |",
		"",
	}, "\n")
	if got := ttu.ReadmePublicTaskNames(readme); !slices.Equal(got, []string{alphaName, zetaName}) {
		t.Fatalf("README names = %v", got)
	}

	if got := ttu.ReadmePublicTaskNames("# No table\n"); len(got) != 0 {
		t.Fatalf("unexpected README names: %v", got)
	}
}

func TestFileHelpers(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	ttu.WriteStub(t, dir, "stub", "#!/bin/sh\necho stub\n")

	stub := filepath.Join(dir, "stub")
	if got := ttu.MustRead(t, stub); !strings.Contains(got, "echo stub") {
		t.Fatalf("ttu.MustRead = %q", got)
	}

	if got := ttu.ReadFile(t, stub); got == "" {
		t.Fatal("ttu.ReadFile returned empty content")
	}

	expectFatal(t, "write broken stub", func(fakeTester *fakeTest) {
		ttu.WriteStub(fakeTester, filepath.Join(dir, "missing"), "broken", "body")
	})
	expectFatal(
		t,
		"read",
		func(fakeTester *fakeTest) { ttu.MustRead(fakeTester, filepath.Join(dir, "missing")) },
	)
}

func TestCommandStringExtraction(t *testing.T) {
	t.Parallel()

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
	root := ttu.DocumentRoot(t, doc)
	commands := ttu.CollectCommandStrings(ttu.NodeMappingValue(root, "cmds"))

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

	if ttu.CollectCommandStrings(nil) != nil ||
		len(ttu.CollectCommandStrings(yamlScalar(" "))) != 0 {
		t.Fatal("empty command extraction mismatch")
	}

	got := ttu.ReferencedLocalShellScripts("./one.sh --flag && echo x\n ./two/path.sh ")
	if !slices.Equal(got, []string{"./one.sh", "./two/path.sh"}) {
		t.Fatalf("script references = %v", got)
	}

	if got := ttu.ReferencedLocalShellScripts("scripts/no-prefix.sh"); len(got) != 0 {
		t.Fatalf("unexpected references = %v", got)
	}
}

func TestAssertExitCode(t *testing.T) {
	t.Parallel()

	ttu.AssertExitCode(
		t,
		ttu.CommandResult{Args: []string{"ok"}, Stdout: "", Stderr: "", Err: nil},
		0,
	)

	err := exec.Command("sh", "-c", "exit 7").Run()
	ttu.AssertExitCode(t, ttu.CommandResult{
		Err:  err,
		Args: []string{"exit"}, Stdout: "", Stderr: "",
	}, 7)

	expectFatal(t, "without exit code", func(fakeTester *fakeTest) {
		ttu.AssertExitCode(fakeTester, ttu.CommandResult{
			Err: errSentinel, Stdout: "", Stderr: "",
			Args: nil,
		}, 1)
	})
	expectFatal(t, "expected exit code", func(fakeTester *fakeTest) {
		ttu.AssertExitCode(
			fakeTester,
			ttu.CommandResult{Args: []string{"ok"}, Stdout: "", Stderr: "", Err: nil},
			2,
		)
	})
}

func TestBasicAssertions(t *testing.T) {
	t.Parallel()

	ttu.AssertContains(t, "alpha beta", "beta")
	ttu.AssertNotContains(t, alphaName, "beta")
	ttu.AssertNotEmpty(t, " value ", "must not be empty")
	expectFatal(t, "expected output to contain", func(fakeTester *fakeTest) {
		ttu.AssertContains(fakeTester, alphaName, "beta")
	})
	expectFatal(t, "expected output not to contain", func(fakeTester *fakeTest) {
		ttu.AssertNotContains(fakeTester, alphaName, alphaName)
	})
	expectFatal(t, "empty sentinel", func(fakeTester *fakeTest) {
		ttu.AssertNotEmpty(fakeTester, " \n", "empty sentinel")
	})
}

func TestFilesystemAssertions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	writeFile(t, file, "content")

	nonempty := filepath.Join(dir, "nonempty")
	writeFile(t, filepath.Join(nonempty, "entry"), "entry")

	empty := filepath.Join(dir, "empty")

	err := os.Mkdir(empty, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	missing := filepath.Join(dir, "missing")

	ttu.AssertFileExists(t, file)
	ttu.AssertDirExists(t, dir)
	ttu.AssertDirNotExists(t, missing)
	ttu.AssertDirHasEntries(t, nonempty)

	expectFatal(
		t,
		"expected file",
		func(fakeTester *fakeTest) { ttu.AssertFileExists(fakeTester, missing) },
	)
	expectFatal(
		t,
		"found directory",
		func(fakeTester *fakeTest) { ttu.AssertFileExists(fakeTester, dir) },
	)
	expectFatal(
		t,
		"expected directory",
		func(fakeTester *fakeTest) { ttu.AssertDirExists(fakeTester, missing) },
	)
	expectFatal(
		t,
		"found file",
		func(fakeTester *fakeTest) { ttu.AssertDirExists(fakeTester, file) },
	)
	expectFatal(
		t,
		"but it does",
		func(fakeTester *fakeTest) { ttu.AssertDirNotExists(fakeTester, dir) },
	)
	expectFatal(
		t,
		"failed to read directory",
		func(fakeTester *fakeTest) { ttu.AssertDirHasEntries(fakeTester, missing) },
	)
	expectFatal(
		t,
		"at least one entry",
		func(fakeTester *fakeTest) { ttu.AssertDirHasEntries(fakeTester, empty) },
	)
}

func groupOutputNode(t *testing.T, begin, end string, errorOnly *string) *yaml.Node {
	t.Helper()

	value := ""
	if errorOnly != nil {
		value = "\n    error_only: " + *errorOnly
	}

	doc := parseYAML(
		t,
		"output:\n  group:\n    begin: \""+begin+"\"\n    end: \""+end+"\""+value+"\n",
	)

	return ttu.NodeMappingValue(ttu.DocumentRoot(t, doc), "output")
}

func TestGithubGroupAssertion(t *testing.T) {
	t.Parallel()

	falseValue := "false"
	ttu.AssertGithubGroupOutput(
		t,
		alphaName,
		groupOutputNode(t, "::group::{{.TASK}}", "::endgroup::", &falseValue),
	)

	expectFatal(t, "no output config", func(fakeTester *fakeTest) {
		ttu.AssertGithubGroupOutput(fakeTester, alphaName, nil)
	})
	expectFatal(t, "advanced object format", func(fakeTester *fakeTest) {
		ttu.AssertGithubGroupOutput(fakeTester, alphaName, yamlScalar("group"))
	})
	expectFatal(t, "include group config", func(fakeTester *fakeTest) {
		ttu.AssertGithubGroupOutput(fakeTester, alphaName, yamlMapping())
	})
	expectFatal(t, "include group config", func(fakeTester *fakeTest) {
		ttu.AssertGithubGroupOutput(fakeTester, alphaName, yamlMapping(
			yamlScalar("group"),
			yamlScalar("scalar"),
		))
	})
	expectFatal(t, "output.group.begin", func(fakeTester *fakeTest) {
		ttu.AssertGithubGroupOutput(
			fakeTester,
			alphaName,
			groupOutputNode(t, "bad", "::endgroup::", &falseValue),
		)
	})
	expectFatal(t, "output.group.end", func(fakeTester *fakeTest) {
		ttu.AssertGithubGroupOutput(
			fakeTester,
			alphaName,
			groupOutputNode(t, "::group::{{.TASK}}", "bad", &falseValue),
		)
	})
	expectFatal(t, "explicitly set", func(fakeTester *fakeTest) {
		ttu.AssertGithubGroupOutput(
			fakeTester,
			alphaName,
			groupOutputNode(t, "::group::{{.TASK}}", "::endgroup::", nil),
		)
	})

	trueValue := "true"

	expectFatal(t, "must be false", func(fakeTester *fakeTest) {
		ttu.AssertGithubGroupOutput(
			fakeTester,
			alphaName,
			groupOutputNode(t, "::group::{{.TASK}}", "::endgroup::", &trueValue),
		)
	})
}

func TestTextFileAssertion(t *testing.T) {
	t.Parallel()

	ttu.AssertTextFileClean(t, "clean.yml", "key: value\n")

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
		expectFatal(
			t,
			tt.want,
			func(fakeTester *fakeTest) { ttu.AssertTextFileClean(fakeTester, "bad.yml", tt.content) },
		)
	}
}

func TestYamlStructureAssertions(t *testing.T) {
	t.Parallel()

	doc := parseYAML(t, "root:\n  list:\n    - name: one\n    - name: two\n")
	ttu.AssertNoDuplicateMappingKeys(t, nil, "root")
	ttu.AssertNoDuplicateMappingKeys(t, doc, "root")
	ttu.AssertNoYamlAliases(t, nil, "root")
	ttu.AssertNoYamlAliases(t, doc, "root")

	duplicate := yamlMapping(
		yamlScalar("same"),
		yamlScalar("one"),
		yamlScalar("same"),
		yamlScalar("two"),
	)

	expectFatal(t, "duplicate YAML key", func(fakeTester *fakeTest) {
		ttu.AssertNoDuplicateMappingKeys(fakeTester, duplicate, "root")
	})

	alias := yamlMapping(
		yamlScalar("value"),
		yamlAlias(),
	)

	expectFatal(t, "aliases/anchors are not allowed", func(fakeTester *fakeTest) {
		ttu.AssertNoYamlAliases(fakeTester, alias, "root")
	})
}

func TestPlaceholderJsonAndDangerousPatterns(t *testing.T) {
	t.Parallel()

	ttu.AssertNoPlaceholderText(t, alphaName, "ordinary task description")
	expectFatal(t, "placeholder text", func(fakeTester *fakeTest) {
		ttu.AssertNoPlaceholderText(fakeTester, alphaName, "fixme later")
	})

	err := ttu.ValidateJSON(`{"ok":true}`)
	if err != nil {
		t.Fatalf("valid JSON: %v", err)
	}

	err = ttu.ValidateJSON(`{"bad":`)
	if err == nil {
		t.Fatal("invalid JSON accepted")
	}

	patterns := ttu.DangerousCommandPatterns()

	unsafe := []string{
		"rm -rf / ",
		"sudo rm -rf /tmp/x",
		"chmod -R 777 /",
		"curl https://x -k ",
		"curl --insecure https://x",
	}
	if len(patterns) != len(unsafe) {
		t.Fatalf("dangerous patterns = %d", len(patterns))
	}

	for index, pattern := range patterns {
		if !pattern.MatchString(unsafe[index]) {
			t.Fatalf("pattern %d did not match %q", index, unsafe[index])
		}
	}
}
