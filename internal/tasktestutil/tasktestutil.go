// Package tasktestutil provides shared fixtures and assertions for Taskfile tests.
package tasktestutil

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	constTasktestutilDefault     = "default"
	constTasktestutilPublicTasks = "## Public Tasks"
	defaultTaskTimeout           = 2 * time.Minute
	taskWaitDelay                = 5 * time.Second
	readmeTableMatchCount        = 2
	privateFileMode              = 0o600
	stubExecutableMode           = 0o500
)

// PublicTaskSpec describes expectations for a single public task.
type PublicTaskSpec struct {
	Name                  string
	Args                  map[string]string
	MustDryRunWithArgs    bool
	MustDryRunWithoutArgs bool
	ExpectedDefaultTokens []string
	RequiresGroupOutput   bool
	RequiresPrompt        bool
	RequiresSummary       bool
}

// PublicTaskSpecOption customizes a PublicTaskSpec built by NewPublicTaskSpec.
type PublicTaskSpecOption func(*PublicTaskSpec)

// NewPublicTaskSpec returns a public task spec with explicit zero defaults.
func NewPublicTaskSpec(name string, options ...PublicTaskSpecOption) PublicTaskSpec {
	spec := PublicTaskSpec{
		Name:                  name,
		Args:                  nil,
		MustDryRunWithArgs:    false,
		MustDryRunWithoutArgs: false,
		ExpectedDefaultTokens: nil,
		RequiresGroupOutput:   false,
		RequiresPrompt:        false,
		RequiresSummary:       false,
	}

	for _, option := range options {
		option(&spec)
	}

	return spec
}

// WithArgs sets task variable assignments for a public task spec.
func WithArgs(args map[string]string) PublicTaskSpecOption {
	return func(spec *PublicTaskSpec) {
		spec.Args = args
	}
}

// WithDryRunArgs marks a public task as requiring dry-run execution with args.
func WithDryRunArgs() PublicTaskSpecOption {
	return func(spec *PublicTaskSpec) {
		spec.MustDryRunWithArgs = true
	}
}

// WithDryRunNoArgs marks a public task as requiring dry-run execution without args.
func WithDryRunNoArgs() PublicTaskSpecOption {
	return func(spec *PublicTaskSpec) {
		spec.MustDryRunWithoutArgs = true
	}
}

// WithExpectedDefaultTokens sets tokens expected from default task output.
func WithExpectedDefaultTokens(tokens ...string) PublicTaskSpecOption {
	return func(spec *PublicTaskSpec) {
		spec.ExpectedDefaultTokens = tokens
	}
}

// WithGroupOutput marks a public task as requiring GitHub group output.
func WithGroupOutput() PublicTaskSpecOption {
	return func(spec *PublicTaskSpec) {
		spec.RequiresGroupOutput = true
	}
}

// WithPrompt marks a public task as requiring an explicit prompt.
func WithPrompt() PublicTaskSpecOption {
	return func(spec *PublicTaskSpec) {
		spec.RequiresPrompt = true
	}
}

// WithSummary marks a public task as requiring task summary text.
func WithSummary() PublicTaskSpecOption {
	return func(spec *PublicTaskSpec) {
		spec.RequiresSummary = true
	}
}

// TaskNode wraps a YAML node with its task name for error messages.
type TaskNode struct {
	Name string
	Node *yaml.Node
}

// LoadedTaskfile holds the parsed content of a Taskfile.
type LoadedTaskfile struct {
	Path  string
	Root  TaskNode
	Tasks map[string]TaskNode
}

// CommandResult holds the output of a task invocation.
type CommandResult struct {
	Stdout string
	Stderr string
	Err    error
	Args   []string
}

// Combined returns stdout and stderr joined with a newline.
func (result CommandResult) Combined() string { return result.Stdout + "\n" + result.Stderr }

// SimpleTaskResult holds the output of a simple (non-isolated) task run.
type SimpleTaskResult struct {
	Output string
	Err    error
}

type testT interface {
	Helper()
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	TempDir() string
}

func workingDir() (string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	return workingDirectory, nil
}

// MustTask returns the named task or fails the test if it is missing.
func MustTask(tester testT, taskfile LoadedTaskfile, name string) TaskNode {
	tester.Helper()

	task, ok := taskfile.Tasks[name]
	if !ok {
		tester.Fatalf("expected public task %q is missing", name)
	}

	return task
}

// Field returns the YAML child node for the given mapping key.
func (n TaskNode) Field(name string) *yaml.Node {
	if n.Node == nil || n.Node.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(n.Node.Content); i += 2 {
		if n.Node.Content[i].Value == name {
			return n.Node.Content[i+1]
		}
	}

	return nil
}

// StringField returns the text value of a scalar mapping field.
func (n TaskNode) StringField(name string) string { return NodeText(n.Field(name)) }

// BoolField returns true when the mapping field is the string "true" (case-insensitive).
func (n TaskNode) BoolField(name string) bool {
	field := n.Field(name)
	if field == nil {
		return false
	}

	return strings.EqualFold(field.Value, "true")
}

// ModuleRoot walks up from the working directory to find the nearest ancestor
// that contains a Taskfile.yml or Taskfile.yaml. When tests run from a module
// directory (e.g. taskfiles/bun/) this returns that directory, not the repo root.
func ModuleRoot(tester testT) string {
	tester.Helper()

	workingDirectory, err := workingDir()
	if err != nil {
		tester.Fatalf("failed to get working directory: %v", err)
	}

	current := workingDirectory
	for {
		if FileExists(filepath.Join(current, "Taskfile.yml")) ||
			FileExists(filepath.Join(current, "Taskfile.yaml")) {
			return current
		}

		parent := filepath.Dir(current)
		if parent == current {
			tester.Fatal("could not find Taskfile.yml or Taskfile.yaml")
		}

		current = parent
	}
}

// ModuleTaskfilePath returns the path of the Taskfile.yml found by ModuleRoot.
func ModuleTaskfilePath(tester testT) string {
	tester.Helper()

	return moduleTaskfilePath(tester, ModuleRoot(tester))
}

func moduleTaskfilePath(tester testT, root string) string {
	tester.Helper()

	for _, name := range []string{"Taskfile.yml", "Taskfile.yaml"} {
		if path := filepath.Join(root, name); FileExists(path) {
			return path
		}
	}

	tester.Fatal("could not find Taskfile.yml or Taskfile.yaml")

	return ""
}

// ModuleReadmePath returns the README.md documenting the module. Flat modules
// keep their own README next to the Taskfile; nested family variants share one
// README at the family root (e.g. taskfiles/npm/fnm/ documents itself in
// taskfiles/npm/README.md), so ancestors are searched when the leaf has none.
func ModuleReadmePath(tester testT) string {
	tester.Helper()

	current := ModuleRoot(tester)
	for {
		if path := filepath.Join(current, "README.md"); FileExists(path) {
			return path
		}

		parent := filepath.Dir(current)
		if parent == current || filepath.Base(current) == "taskfiles" {
			tester.Fatal("could not find README.md for module")

			return ""
		}

		current = parent
	}
}

// LoadTaskfile parses the Taskfile in the module root and returns a LoadedTaskfile.
func LoadTaskfile(tester testT) LoadedTaskfile {
	tester.Helper()

	path := ModuleTaskfilePath(tester)
	content := ReadFile(tester, path)

	var doc yaml.Node

	err := yaml.Unmarshal([]byte(content), &doc)
	if err != nil {
		tester.Fatalf("failed to parse Taskfile: %v", err)
	}

	root := DocumentRoot(tester, &doc)

	tasksNode := MappingField(root, "tasks")
	if tasksNode == nil {
		tester.Fatal("Taskfile has no tasks map")
	}

	tasks := map[string]TaskNode{}

	for i := 0; i < len(tasksNode.Content); i += 2 {
		key := tasksNode.Content[i]
		tasks[key.Value] = TaskNode{Name: key.Value, Node: tasksNode.Content[i+1]}
	}

	return LoadedTaskfile{
		Path:  path,
		Root:  TaskNode{Name: "root", Node: root},
		Tasks: tasks,
	}
}

// HasAlias reports whether the task declares the given alias.
func HasAlias(task TaskNode, alias string) bool {
	aliases := task.Field("aliases")
	if aliases == nil || aliases.Kind != yaml.SequenceNode {
		return false
	}

	for _, item := range aliases.Content {
		if item.Value == alias {
			return true
		}
	}

	return false
}

// RunTask runs the task binary with the given args and returns the result.
func RunTask(tester testT, root string, env []string, args ...string) CommandResult {
	tester.Helper()

	return RunTaskTimeout(tester, root, env, defaultTaskTimeout, args...)
}

// RunTaskTimeout runs the task binary with a custom timeout.
func RunTaskTimeout(
	tester testT,
	root string,
	env []string,
	timeout time.Duration,
	args ...string,
) CommandResult {
	tester.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	commandContext := exec.CommandContext
	cmd := commandContext(ctx, "task", args...)
	cmd.Dir = root

	if env != nil {
		cmd.Env = env
	}

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.WaitDelay = taskWaitDelay

	err := cmd.Run()

	return CommandResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
		Args:   args,
	}
}

// RunSimpleTask runs task in the given directory and returns combined output.
// Use this for the simple pnpm/yarn-style tests that don'tester need separate stdout/stderr.
func RunSimpleTask(tester testT, dir string, env []string, args ...string) SimpleTaskResult {
	tester.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTaskTimeout)
	defer cancel()

	commandContext := exec.CommandContext
	cmd := commandContext(ctx, "task", args...)
	cmd.Dir = dir
	cmd.Env = env
	out, err := cmd.CombinedOutput()

	return SimpleTaskResult{Output: string(out), Err: err}
}

// IsolatedEnv returns a clean environment with a temporary HOME for tests that
// must not interact with the real user's shell profile or tool installations.
func IsolatedEnv(tester testT) []string {
	tester.Helper()

	home := tester.TempDir()
	profile := filepath.Join(home, ".bashrc")

	err := os.WriteFile(profile, []byte(""), privateFileMode)
	if err != nil {
		tester.Fatalf("failed to create fake shell profile: %v", err)
	}

	env := os.Environ()
	env = SetEnv(env, "HOME", home)
	env = SetEnv(env, "PROFILE", profile)
	env = SetEnv(env, "ZDOTDIR", home)
	env = SetEnv(env, "CI", "true")
	env = SetEnv(env, "TASK_COLOR", "0")
	env = SetEnv(env, "NO_COLOR", "1")
	env = SetEnv(env, "TASK_ASSUME_YES", "true")

	return env
}

// SetEnv sets or replaces a key=value pair in an env slice.
func SetEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, item := range env {
		if strings.HasPrefix(item, prefix) {
			env[i] = prefix + value

			return env
		}
	}

	return append(env, prefix+value)
}

// EnvValue returns the value for the given key from an env slice.
func EnvValue(env []string, key string) string {
	prefix := key + "="
	for _, item := range env {
		if after, ok := strings.CutPrefix(item, prefix); ok {
			return after
		}
	}

	return ""
}

// ExpectedPublicTaskNames returns sorted task names from a PublicTaskSpec slice.
func ExpectedPublicTaskNames(specs []PublicTaskSpec) []string {
	names := make([]string, 0, len(specs))
	for _, spec := range specs {
		names = append(names, spec.Name)
	}

	slices.Sort(names)

	return names
}

// PublicTaskNamesFromTaskfile returns sorted names of public tasks in the Taskfile.
func PublicTaskNamesFromTaskfile(tester testT, taskfile LoadedTaskfile) []string {
	tester.Helper()

	var names []string

	for name, task := range taskfile.Tasks {
		if name == constTasktestutilDefault || strings.HasPrefix(name, "_") ||
			task.BoolField("internal") {
			continue
		}

		if task.StringField("desc") != "" {
			names = append(names, name)
		}
	}

	slices.Sort(names)

	return names
}

// TaskArgs converts a map of task variable assignments to "KEY=VALUE" args.
func TaskArgs(args map[string]string) []string {
	if len(args) == 0 {
		return nil
	}

	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, fmt.Sprintf("%s=%s", key, args[key]))
	}

	return out
}

// FormatList formats a string slice as a bulleted list.
func FormatList(values []string) string { return "- " + strings.Join(values, "\n- ") }

// WriteStub writes a stub shell script to dir/name with the given body.
func WriteStub(tester testT, dir, name, body string) {
	tester.Helper()

	path := filepath.Join(dir, name)

	err := os.WriteFile(path, []byte(body), privateFileMode)
	if err != nil {
		tester.Fatalf("write %s stub: %v", name, err)
	}

	err = os.Chmod(path, stubExecutableMode)
	if err != nil {
		tester.Fatalf("mark %s stub executable: %v", name, err)
	}
}

// SimplePublicTaskNames extracts public task names from a decoded tasks map.
func SimplePublicTaskNames(tasks map[string]any) []string {
	var names []string

	for name, raw := range tasks {
		if name == constTasktestutilDefault || strings.HasPrefix(name, "_") {
			continue
		}

		if task, ok := raw.(map[string]any); ok && task["internal"] == true {
			continue
		}

		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

// ReadmePublicTaskNames parses a README and returns task names listed in the
// "## Public Tasks" table (backtick-quoted entries in the first column).
func ReadmePublicTaskNames(content string) []string {
	row := regexp.MustCompile(`^\|\s*` + "`" + `([^` + "`" + `]+)` + "`" + `\s*\|`)

	var names []string

	inTable := false

	for line := range strings.SplitSeq(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == constTasktestutilPublicTasks {
			inTable = true

			continue
		}

		if inTable && strings.HasPrefix(trimmed, "## ") {
			break
		}

		if inTable {
			if matches := row.FindStringSubmatch(trimmed); len(matches) == readmeTableMatchCount {
				names = append(names, matches[1])
			}
		}
	}

	slices.Sort(names)

	return names
}

// MustRead reads a file and fails the test on error.
func MustRead(tester testT, path string) string {
	tester.Helper()

	content, err := readFile(path)
	if err != nil {
		tester.Fatalf("read %s: %v", path, err)
	}

	return string(content)
}

// --- YAML helpers ---

// DocumentRoot returns the root mapping node of a YAML document node.
func DocumentRoot(tester testT, doc *yaml.Node) *yaml.Node {
	tester.Helper()

	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		tester.Fatal("invalid YAML document")
	}

	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		tester.Fatal("Taskfile root must be a YAML mapping")
	}

	return root
}

// MappingField returns the mapping-typed child of root named name, or nil.
func MappingField(root *yaml.Node, name string) *yaml.Node {
	node := NodeMappingValue(root, name)
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	return node
}

// ScalarField returns the text of the scalar child of root named name.
func ScalarField(root *yaml.Node, name string) string {
	return NodeText(NodeMappingValue(root, name))
}

// NodeMappingValue returns the value node for the given key in a mapping node.
func NodeMappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}

	return nil
}

// NodeText returns the trimmed text content of a YAML node.
func NodeText(node *yaml.Node) string {
	if node == nil {
		return ""
	}

	if node.Kind == yaml.ScalarNode {
		return strings.TrimSpace(node.Value)
	}

	var parts []string

	for _, child := range node.Content {
		if text := NodeText(child); text != "" {
			parts = append(parts, text)
		}
	}

	return strings.TrimSpace(strings.Join(parts, " "))
}

// IsEmptyNode reports whether a YAML node is nil or carries no content.
func IsEmptyNode(node *yaml.Node) bool {
	if node == nil {
		return true
	}

	if node.Kind == yaml.ScalarNode {
		return strings.TrimSpace(node.Value) == ""
	}

	return len(node.Content) == 0
}

// FileExists reports whether the given path exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

// ReadFile reads a file and fails the test on error.
func ReadFile(tester testT, path string) string {
	tester.Helper()

	content, err := readFile(path)
	if err != nil {
		tester.Fatalf("failed to read %s: %v", path, err)
	}

	return string(content)
}

// CollectCommandStrings extracts all command/sh scalar values from a task node.
func CollectCommandStrings(node *yaml.Node) []string {
	if node == nil {
		return nil
	}

	out := make([]string, 0, len(node.Content))

	switch node.Kind {
	case yaml.DocumentNode:
		out = append(out, collectSequenceCommandStrings(node)...)
	case yaml.ScalarNode:
		if strings.TrimSpace(node.Value) != "" {
			out = append(out, node.Value)
		}
	case yaml.SequenceNode:
		out = append(out, collectSequenceCommandStrings(node)...)
	case yaml.MappingNode:
		out = append(out, collectMappingCommandStrings(node)...)
	case yaml.AliasNode:
	}

	return out
}

func collectSequenceCommandStrings(node *yaml.Node) []string {
	out := make([]string, 0, len(node.Content))

	for _, child := range node.Content {
		out = append(out, CollectCommandStrings(child)...)
	}

	return out
}

func collectMappingCommandStrings(node *yaml.Node) []string {
	var out []string

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		value := node.Content[i+1]

		switch key.Value {
		case "cmd", "sh":
			out = append(out, scalarCommandString(value)...)
		case "cmds", "status", "preconditions":
			out = append(out, CollectCommandStrings(value)...)
		}
	}

	return out
}

func scalarCommandString(node *yaml.Node) []string {
	if node.Kind != yaml.ScalarNode {
		return nil
	}

	return []string{node.Value}
}

// ReferencedLocalShellScripts returns all ./path/to/script.sh references in a command string.
func ReferencedLocalShellScripts(command string) []string {
	re := regexp.MustCompile(`(?:^|\s)(\./[A-Za-z0-9_./-]+\.sh)(?:\s|$)`)
	matches := re.FindAllStringSubmatch(command, -1)

	var out []string

	for _, match := range matches {
		if len(match) > 1 {
			out = append(out, match[1])
		}
	}

	return out
}

// AssertExitCode fails the test if the command result exit code differs from expected.
func AssertExitCode(tester testT, result CommandResult, expected int) {
	tester.Helper()

	actual := 0

	if result.Err != nil {
		exitErr := new(exec.ExitError)

		ok := errors.As(result.Err, &exitErr)
		if !ok {
			tester.Fatalf(
				"command failed without exit code\nargs: %v\nerror: %v\nstdout:\n%s\nstderr:\n%s",
				result.Args, result.Err, result.Stdout, result.Stderr,
			)
		}

		actual = exitErr.ExitCode()
	}

	if actual != expected {
		tester.Fatalf(
			"expected exit code %d, got %d\nargs: %v\nerror: %v\nstdout:\n%s\nstderr:\n%s",
			expected, actual, result.Args, result.Err, result.Stdout, result.Stderr,
		)
	}
}

// AssertContains fails the test if value does not contain expected.
func AssertContains(tester testT, value, expected string) {
	tester.Helper()

	if !strings.Contains(value, expected) {
		tester.Fatalf("expected output to contain %q\n\nOutput:\n%s", expected, value)
	}
}

// AssertNotContains fails the test if value contains unexpected.
func AssertNotContains(tester testT, value, unexpected string) {
	tester.Helper()

	if strings.Contains(value, unexpected) {
		tester.Fatalf("expected output not to contain %q\n\nOutput:\n%s", unexpected, value)
	}
}

// AssertNotEmpty fails the test if value is empty after trimming whitespace.
func AssertNotEmpty(tester testT, value, message string) {
	tester.Helper()

	if strings.TrimSpace(value) == "" {
		tester.Fatal(message)
	}
}

// AssertFileExists fails the test unless path exists and is a file.
func AssertFileExists(tester testT, path string) {
	tester.Helper()

	info, err := os.Stat(path)
	if err != nil {
		tester.Fatalf("expected file %s to exist: %v", path, err)
	}

	if info.IsDir() {
		tester.Fatalf("expected file but found directory at %s", path)
	}
}

// AssertDirExists fails the test unless path exists and is a directory.
func AssertDirExists(tester testT, path string) {
	tester.Helper()

	info, err := os.Stat(path)
	if err != nil {
		tester.Fatalf("expected directory %s to exist: %v", path, err)
	}

	if !info.IsDir() {
		tester.Fatalf("expected directory but found file at %s", path)
	}
}

// AssertDirNotExists fails the test unless path does not exist.
func AssertDirNotExists(tester testT, path string) {
	tester.Helper()

	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		tester.Fatalf("expected %s to not exist, but it does", path)
	}
}

// AssertDirHasEntries fails the test unless path is a non-empty directory.
func AssertDirHasEntries(tester testT, path string) {
	tester.Helper()

	entries, err := os.ReadDir(path)
	if err != nil {
		tester.Fatalf("failed to read directory %s: %v", path, err)
	}

	if len(entries) == 0 {
		tester.Fatalf("expected %s to contain at least one entry", path)
	}
}

// AssertGithubGroupOutput validates GitHub Actions grouped-output configuration.
func AssertGithubGroupOutput(tester testT, taskName string, outputNode *yaml.Node) {
	tester.Helper()

	if outputNode == nil {
		tester.Fatalf(
			"task %q requires output.group config but no output config was found",
			taskName,
		)

		return
	}

	if outputNode.Kind != yaml.MappingNode {
		tester.Fatalf("task %q output must use advanced object format, not scalar format", taskName)

		return
	}

	groupNode := NodeMappingValue(outputNode, "group")
	if groupNode == nil || groupNode.Kind != yaml.MappingNode {
		tester.Fatalf("task %q output must include group config", taskName)

		return
	}

	begin := NodeText(NodeMappingValue(groupNode, "begin"))
	end := NodeText(NodeMappingValue(groupNode, "end"))
	errorOnly := NodeMappingValue(groupNode, "error_only")

	if begin != "::group::{{.TASK}}" {
		tester.Fatalf(
			"task %q output.group.begin must be %q, got %q",
			taskName,
			"::group::{{.TASK}}",
			begin,
		)
	}

	if end != "::endgroup::" {
		tester.Fatalf("task %q output.group.end must be %q, got %q", taskName, "::endgroup::", end)
	}

	if errorOnly == nil {
		tester.Fatalf("task %q output.group.error_only must be explicitly set to false", taskName)

		return
	}

	if !strings.EqualFold(errorOnly.Value, "false") {
		tester.Fatalf(
			"task %q output.group.error_only must be false, got %q",
			taskName,
			errorOnly.Value,
		)
	}
}

// AssertDestructivePrompt validates that a destructive task has an explicit prompt.
func AssertDestructivePrompt(tester testT, taskName string, prompt *yaml.Node) {
	tester.Helper()

	if prompt == nil || NodeText(prompt) == "" {
		tester.Fatalf("destructive task %q must have a non-empty prompt", taskName)
	}

	if !explicitPromptText(NodeText(prompt)) {
		tester.Fatalf(
			"prompt for task %q does not look explicit enough:\n%s",
			taskName,
			NodeText(prompt),
		)
	}
}

func explicitPromptText(text string) bool {
	lower := strings.ToLower(text)

	for _, token := range []string{"sure", "confirm", "remove", "uninstall", "delete", "continue"} {
		if strings.Contains(lower, token) {
			return true
		}
	}

	return false
}

// AssertTextFileClean validates portable text-file formatting rules.
func AssertTextFileClean(tester testT, path, content string) {
	tester.Helper()

	if content == "" {
		tester.Fatalf("%s is empty", path)
	}

	if strings.Contains(content, "\r\n") {
		tester.Fatalf("%s uses CRLF line endings; use LF only", path)
	}

	if strings.Contains(content, "\t") {
		tester.Fatalf("%s contains tabs; use spaces in YAML", path)
	}

	if !strings.HasSuffix(content, "\n") {
		tester.Fatalf("%s must end with a newline", path)
	}

	for i, line := range strings.Split(content, "\n") {
		if strings.TrimRight(line, " ") != line {
			tester.Fatalf("%s has trailing whitespace at line %d", path, i+1)
		}
	}
}

// AssertNoDuplicateMappingKeys fails the test if a YAML tree contains duplicate keys.
func AssertNoDuplicateMappingKeys(tester testT, node *yaml.Node, path string) {
	tester.Helper()

	if node == nil {
		return
	}

	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		AssertNoDuplicateMappingKeys(tester, node.Content[0], path)

		return
	}

	if node.Kind == yaml.MappingNode {
		seen := map[string]bool{}

		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]

			if seen[key.Value] {
				tester.Fatalf("duplicate YAML key at %s.%s", path, key.Value)
			}

			seen[key.Value] = true
			AssertNoDuplicateMappingKeys(tester, value, path+"."+key.Value)
		}
	}

	if node.Kind == yaml.SequenceNode {
		for i, child := range node.Content {
			AssertNoDuplicateMappingKeys(tester, child, fmt.Sprintf("%s[%d]", path, i))
		}
	}
}

// AssertNoYamlAliases fails the test if a YAML tree contains anchors or aliases.
func AssertNoYamlAliases(tester testT, node *yaml.Node, path string) {
	tester.Helper()

	if node == nil {
		return
	}

	if node.Kind == yaml.AliasNode {
		tester.Fatalf("YAML aliases/anchors are not allowed for clean Taskfile config at %s", path)
	}

	for i, child := range node.Content {
		AssertNoYamlAliases(tester, child, fmt.Sprintf("%s[%d]", path, i))
	}
}

// AssertNoPlaceholderText fails the test if value contains common placeholder text.
func AssertNoPlaceholderText(tester testT, taskName, value string) {
	tester.Helper()

	upper := strings.ToUpper(value)
	for _, placeholder := range []string{"TODO", "FIXME", "CHANGEME", "REPLACE_ME", "LOREM IPSUM"} {
		if strings.Contains(upper, placeholder) {
			tester.Fatalf("task %q contains placeholder text %q", taskName, placeholder)
		}
	}
}

// ValidateJSON returns an error if s is not valid JSON.
func ValidateJSON(s string) error {
	var payload any

	err := json.Unmarshal([]byte(s), &payload)
	if err != nil {
		return fmt.Errorf("validate JSON: %w", err)
	}

	return nil
}

// DangerousCommandPatterns returns regexps that match unsafe shell command patterns.
func DangerousCommandPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		regexp.MustCompile(`(?m)\brm\s+-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*\s+/(?:\s|$)`),
		regexp.MustCompile(`(?m)\bsudo\s+rm\s+-[a-zA-Z]*r[a-zA-Z]*f`),
		regexp.MustCompile(`(?m)\bchmod\s+-R\s+777\s+/`),
		regexp.MustCompile(`(?m)\bcurl\b.*\s-k(?:\s|$)`),
		regexp.MustCompile(`(?m)\bcurl\b.*--insecure`),
	}
}

func readFile(path string) ([]byte, error) {
	clean := filepath.Clean(path)

	content, err := fs.ReadFile(os.DirFS(filepath.Dir(clean)), filepath.Base(clean))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", clean, err)
	}

	return content, nil
}
