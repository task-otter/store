// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package tasktestutil provides shared helpers for testing Taskfile modules:
// running the task binary, asserting on YAML Taskfile structure, and
// validating README/task documentation conventions across the repo.
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

	yaml "go.yaml.in/yaml/v3"
)

type (
	// PublicTaskSpec describes expectations for a single public task.
	PublicTaskSpec struct {
		Args                  map[string]string
		Name                  string
		ExpectedDefaultTokens []string
		MustDryRunWithArgs    bool
		MustDryRunWithoutArgs bool
		RequiresGroupOutput   bool
		RequiresPrompt        bool
		RequiresSummary       bool
	}

	// PublicTaskSpecOption customizes a PublicTaskSpec built by NewPublicTaskSpec.
	PublicTaskSpecOption func(*PublicTaskSpec)

	// CommandResult holds the output of a task invocation.
	CommandResult struct {
		Stdout string
		Stderr string
		Err    error
		Args   []string
	}

	// TestT is the subset of [testing.T] used by tasktestutil helpers.
	TestT interface {
		Helper()
		Fatal(args ...any)
		Fatalf(format string, args ...any)
		TempDir() string
	}
)

const (
	constTasktestutilDefault     = "default"
	constTasktestutilPublicTasks = "## Public Tasks"
	defaultTaskTimeout           = 2 * time.Minute
	taskWaitDelay                = 5 * time.Second
	readmeTableMatchCount        = 2
	privateFileMode              = 0o600
	stubExecutableMode           = 0o500
	trueString                   = "true"
	taskfileYML                  = "Taskfile.yml"
	taskfileYAML                 = "Taskfile.yaml"
	errTaskfileNotFound          = "could not find Taskfile.yml or Taskfile.yaml"
	taskBinaryName               = "task"
	errTaskTimeoutType           = "task timeout must be time.Duration, got %T"
	underscorePrefix             = "_"
	singleSpace                  = " "
	groupBeginTemplate           = "::group::{{.TASK}}"
	groupEndMarker               = "::endgroup::"
	seqIndexFormat               = "%s[%d]"
	newline                      = "\n"
	promptTextConfirm            = "confirm"
	promptTextRemove             = "remove"
	promptTextSure               = "sure"
	promptTextUninstall          = "uninstall"
	promptTextDelete             = "delete"
	promptTextContinue           = "continue"
	zeroIndex                    = 0
	valueOffset                  = 1
	emptyString                  = ""
	msgExpectedNotToExist        = "expected %s to not exist, but it does"
	placeholderTODO              = "TODO"
	placeholderFIXME             = "FIXME"
	placeholderChangeme          = "CHANGEME"
	placeholderCopyright         = "Copyright"
	placeholderLoremIpsum        = "LOREM IPSUM"
)

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

	for i := range options {
		option := options[i]
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

// Combined returns stdout and stderr joined with a newline.
func (result *CommandResult) Combined() string { return result.Stdout + newline + result.Stderr }

func workingDir() (string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return emptyString, fmt.Errorf("get working directory: %w", err)
	}

	return workingDirectory, nil
}

// MustTask returns the named task or fails the test if it is missing.
func MustTask(tester TestT, taskfile any, name string) TaskNode {
	tester.Helper()

	loaded := normalizeLoadedTaskfile(tester, taskfile)
	task, ok := loaded.Tasks[name]

	if !ok {
		tester.Fatalf("expected public task %q is missing", name)
	}

	return task
}

func normalizeLoadedTaskfile(tester TestT, taskfile any) *LoadedTaskfile {
	tester.Helper()

	switch loaded := taskfile.(type) {
	case *LoadedTaskfile:
		return loaded
	case LoadedTaskfile:
		return &loaded
	default:
		tester.Fatalf("taskfile must be LoadedTaskfile or *LoadedTaskfile, got %T", taskfile)

		return nil
	}
}

// BoolField returns true when the mapping field is the string "true" (case-insensitive).
func (node TaskNode) BoolField(name string) bool {
	field := node.Field(name)

	if field == nil {
		return false
	}

	return strings.EqualFold(field.Value, trueString)
}

// Field returns the YAML child node for the given mapping key.
func (node TaskNode) Field(name string) *yaml.Node {
	if node.Node == nil || node.Node.Kind != yaml.MappingNode {
		return nil
	}

	for i := zeroIndex; i < len(node.Node.Content); i += readmeTableMatchCount {
		if node.Node.Content[i].Value == name {
			return node.Node.Content[i+valueOffset]
		}
	}

	return nil
}

// StringField returns the text value of a scalar mapping field.
func (node TaskNode) StringField(name string) string { return NodeText(node.Field(name)) }

// ModuleRoot walks up from the working directory to find the nearest ancestor
// that contains a Taskfile.yml or Taskfile.yaml. When tests run from a module
// directory (e.g. taskfiles/bun/) this returns that directory, not the repo root.
func ModuleRoot(tester TestT) string {
	tester.Helper()

	workingDirectory, err := workingDir()
	if err != nil {
		tester.Fatalf("failed to get working directory: %v", err)
	}

	root, ok := findAncestorWithTaskfile(workingDirectory)

	if !ok {
		tester.Fatal(errTaskfileNotFound)
	}

	return root
}

func findAncestorWithTaskfile(start string) (string, bool) {
	current := start

	for {
		if hasTaskfile(current) {
			return current, true
		}

		parent := filepath.Dir(current)

		if parent == current {
			return emptyString, false
		}

		current = parent
	}
}

func hasTaskfile(dir string) bool {
	return FileExists(filepath.Join(dir, taskfileYML)) ||
		FileExists(filepath.Join(dir, taskfileYAML))
}

// ModuleTaskfilePath returns the path of the Taskfile.yml found by ModuleRoot.
func ModuleTaskfilePath(tester TestT) string {
	tester.Helper()

	return findModuleTaskfilePath(tester, ModuleRoot(tester))
}

func findModuleTaskfilePath(tester TestT, root string) string {
	tester.Helper()

	for i := range []string{taskfileYML, taskfileYAML} {
		name := []string{taskfileYML, taskfileYAML}[i]

		if path := filepath.Join(root, name); FileExists(path) {
			return path
		}
	}

	tester.Fatal(errTaskfileNotFound)

	return emptyString
}

// ModuleReadmePath returns the README.md documenting the module. Flat modules
// keep their own README next to the Taskfile; nested family variants share one
// README at the family root (e.g. taskfiles/npm/ documents itself in
// taskfiles/npm/README.md), so ancestors are searched when the leaf has none.
func ModuleReadmePath(tester TestT) string {
	tester.Helper()

	path, ok := findModuleReadme(ModuleRoot(tester))

	if !ok {
		tester.Fatal("could not find README.md for module")

		return emptyString
	}

	return path
}

func findModuleReadme(start string) (string, bool) {
	current := start

	for {
		if path := filepath.Join(current, "README.md"); FileExists(path) {
			return path, true
		}

		if reachedReadmeSearchBoundary(current) {
			return emptyString, false
		}

		current = filepath.Dir(current)
	}
}

func reachedReadmeSearchBoundary(dir string) bool {
	return filepath.Dir(dir) == dir || filepath.Base(dir) == "taskfiles"
}

// LoadTaskfile parses the Taskfile in the module root and returns a LoadedTaskfile.
func LoadTaskfile(tester TestT) LoadedTaskfile {
	tester.Helper()

	path := ModuleTaskfilePath(tester)
	root := parseTaskfileRoot(tester, path)
	tasksNode := MappingField(root, "tasks")

	if tasksNode == nil {
		tester.Fatal("Taskfile has no tasks map")
	}

	return LoadedTaskfile{
		Path:  path,
		Root:  TaskNode{Name: "root", Node: root},
		Tasks: buildTaskMap(tasksNode),
	}
}

func parseTaskfileRoot(tester TestT, path string) *yaml.Node {
	tester.Helper()

	content := ReadFile(tester, path)

	var doc yaml.Node

	err := yaml.Unmarshal([]byte(content), &doc)
	if err != nil {
		tester.Fatalf("failed to parse Taskfile: %v", err)
	}

	return DocumentRoot(tester, &doc)
}

func buildTaskMap(tasksNode *yaml.Node) map[string]TaskNode {
	tasks := map[string]TaskNode{}

	for i := zeroIndex; i < len(tasksNode.Content); i += readmeTableMatchCount {
		key := tasksNode.Content[i]

		tasks[key.Value] = TaskNode{Name: key.Value, Node: tasksNode.Content[i+valueOffset]}
	}

	return tasks
}

// HasAlias reports whether the task declares the given alias.
func HasAlias(task TaskNode, alias string) bool {
	aliases := task.Field("aliases")

	if aliases == nil || aliases.Kind != yaml.SequenceNode {
		return false
	}

	for i := range aliases.Content {
		item := aliases.Content[i]

		if item.Value == alias {
			return true
		}
	}

	return false
}

// RunTask runs the task binary with the given args and returns the result.
func RunTask(tester TestT, run any, parts ...any) CommandResult {
	tester.Helper()

	return RunTaskTimeout(tester, normalizeTaskRun(tester, run, parts), defaultTaskTimeout)
}

// RunTaskTimeout runs the task binary with a custom timeout.
func RunTaskTimeout(tester TestT, run any, timeoutParts ...any) CommandResult {
	tester.Helper()

	normalizedRun, timeout := normalizeTaskRunTimeout(tester, run, timeoutParts)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()

	return executeTaskCommand(ctx, &normalizedRun)
}

func executeTaskCommand(ctx context.Context, run *TaskRun) CommandResult {
	cmd := buildTaskCmd(ctx, run)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.WaitDelay = taskWaitDelay

	err := cmd.Run()

	return CommandResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
		Args:   run.Args,
	}
}

func buildTaskCmd(ctx context.Context, run *TaskRun) *exec.Cmd {
	commandContext := exec.CommandContext
	cmd := commandContext(ctx, taskBinaryName, run.Args...)

	cmd.Dir = run.Root

	if run.Env != nil {
		cmd.Env = run.Env
	}

	return cmd
}

// RunSimpleTask runs task in the given directory and returns combined output
// in Stdout. Use this for the simple pnpm/yarn-style tests that don'tester
// need separate stdout/stderr.
func RunSimpleTask(tester TestT, run any, parts ...any) CommandResult {
	tester.Helper()

	normalizedRun := normalizeTaskRun(tester, run, parts)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTaskTimeout)

	defer cancel()

	cmd := buildTaskCmd(ctx, &normalizedRun)
	out, err := cmd.CombinedOutput()

	return CommandResult{
		Stdout: string(out),
		Stderr: emptyString,
		Err:    err,
		Args:   normalizedRun.Args,
	}
}

func requireNoExtraArgs(tester TestT, message string, parts []any) {
	tester.Helper()

	if len(parts) != zeroIndex {
		tester.Fatalf(message, parts)
	}
}

func requireStringValue(tester TestT, message string, value any) string {
	tester.Helper()

	str, ok := value.(string)

	if !ok {
		tester.Fatalf(message, value)
	}

	return str
}

func requireEnvParts(tester TestT, messages *envPartsMessages, parts []any) []string {
	tester.Helper()

	if len(parts) == zeroIndex {
		tester.Fatal(messages.required)
	}

	env, ok := parts[zeroIndex].([]string)

	if !ok && parts[zeroIndex] != nil {
		tester.Fatalf(messages.wrongType, parts[zeroIndex])
	}

	return env
}

func normalizeTaskRun(tester TestT, run any, parts []any) TaskRun {
	tester.Helper()

	if normalized, ok := run.(TaskRun); ok {
		requireNoExtraArgs(tester, "TaskRun does not accept positional arguments: %v", parts)

		return normalized
	}

	return legacyTaskRun(tester, &legacyTaskRunArgs{
		run:   run,
		parts: parts,
		messages: &legacyTaskRunMessages{
			rootType: "task run root must be string or TaskRun, got %T",
			env: envPartsMessages{
				required:  "task run environment is required",
				wrongType: "task run environment must be []string, got %T",
			},
		},
	})
}

func legacyTaskRun(tester TestT, args *legacyTaskRunArgs) TaskRun {
	tester.Helper()

	root := requireStringValue(tester, args.messages.rootType, args.run)
	env := requireEnvParts(tester, &args.messages.env, args.parts)

	return TaskRun{Root: root, Env: env, Args: stringArgs(tester, args.parts[valueOffset:])}
}

func normalizeTaskRunTimeout(tester TestT, run any, parts []any) (TaskRun, time.Duration) {
	tester.Helper()

	if _, ok := run.(TaskRun); ok {
		timeout := parseExplicitRunTimeout(tester, parts)

		return normalizeTaskRun(tester, run, nil), timeout
	}

	return parseLegacyRunTimeout(tester, run, parts)
}

func parseExplicitRunTimeout(tester TestT, parts []any) time.Duration {
	tester.Helper()

	if len(parts) != valueOffset {
		tester.Fatalf(
			"TaskRun timeout call requires exactly one timeout, got %d values",
			len(parts),
		)
	}

	timeout, ok := parts[zeroIndex].(time.Duration)

	if !ok {
		tester.Fatalf(errTaskTimeoutType, parts[zeroIndex])
	}

	return timeout
}

func parseLegacyRunTimeout(tester TestT, run any, parts []any) (TaskRun, time.Duration) {
	tester.Helper()

	if len(parts) < readmeTableMatchCount {
		tester.Fatal("legacy timeout call requires env and timeout")
	}

	timeout, ok := parts[valueOffset].(time.Duration)

	if !ok {
		tester.Fatalf(errTaskTimeoutType, parts[valueOffset])
	}

	normalizedRun := normalizeTaskRun(
		tester,
		run,
		append([]any{parts[zeroIndex]}, parts[readmeTableMatchCount:]...),
	)

	return normalizedRun, timeout
}

func stringArgs(tester TestT, values []any) []string {
	tester.Helper()

	args := make([]string, zeroIndex, len(values))

	for i := range values {
		value := values[i]
		arg, ok := value.(string)

		if !ok {
			tester.Fatalf("task argument must be string, got %T", value)
		}

		args = append(args, arg)
	}

	return args
}

// IsolatedEnv returns a clean environment with a temporary HOME for tests that
// must not interact with the real user's shell profile or tool installations.
func IsolatedEnv(tester TestT) []string {
	tester.Helper()

	home := tester.TempDir()
	profile := filepath.Join(home, ".bashrc")

	err := os.WriteFile(profile, []byte(emptyString), privateFileMode)
	if err != nil {
		tester.Fatalf("failed to create fake shell profile: %v", err)
	}

	return isolatedEnvVars(home, profile)
}

func isolatedEnvVars(home, profile string) []string {
	env := os.Environ()

	env = SetEnv(env, "HOME", home)
	env = SetEnv(env, "PROFILE", profile)
	env = SetEnv(env, "ZDOTDIR", home)
	env = SetEnv(env, "CI", trueString)
	env = SetEnv(env, "TASK_COLOR", "0")
	env = SetEnv(env, "NO_COLOR", "1")
	env = SetEnv(env, "TASK_ASSUME_YES", trueString)

	return env
}

// SetEnv sets or replaces a key=value pair in an env slice.
func SetEnv(env []string, key, value string) []string {
	prefix := key + "="

	for i := range env {
		item := env[i]

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

	for i := range env {
		item := env[i]

		if after, ok := strings.CutPrefix(item, prefix); ok {
			return after
		}
	}

	return emptyString
}

// ExpectedPublicTaskNames returns sorted task names from a PublicTaskSpec slice.
func ExpectedPublicTaskNames(specs []PublicTaskSpec) []string {
	names := make([]string, zeroIndex, len(specs))

	for i := range specs {
		spec := specs[i]

		names = append(names, spec.Name)
	}

	slices.Sort(names)

	return names
}

// PublicTaskNamesFromTaskfile returns sorted names of public tasks in the Taskfile.
func PublicTaskNamesFromTaskfile(tester TestT, taskfile any) []string {
	tester.Helper()

	loaded := normalizeLoadedTaskfile(tester, taskfile)

	var names []string

	for name := range loaded.Tasks {
		if isPublicTaskfileTask(name, loaded.Tasks[name]) {
			names = append(names, name)
		}
	}

	slices.Sort(names)

	return names
}

func isPublicTaskfileTask(name string, task TaskNode) bool {
	isDefaultOrPrivate := name == constTasktestutilDefault ||
		strings.HasPrefix(name, underscorePrefix)

	if isDefaultOrPrivate || task.BoolField("internal") {
		return false
	}

	return task.StringField("desc") != emptyString
}

// TaskArgs converts a map of task variable assignments to "KEY=VALUE" args.
func TaskArgs(args map[string]string) []string {
	if len(args) == zeroIndex {
		return nil
	}

	keys := sortedMapKeys(args)
	slices.Sort(keys)

	out := make([]string, zeroIndex, len(keys))

	for i := range keys {
		key := keys[i]

		out = append(out, fmt.Sprintf("%s=%s", key, args[key]))
	}

	return out
}

func sortedMapKeys(args map[string]string) []string {
	keys := make([]string, zeroIndex, len(args))

	for key := range args {
		keys = append(keys, key)
	}

	return keys
}

// FormatList formats a string slice as a bulleted list.
func FormatList(values []string) string { return "- " + strings.Join(values, "\n- ") }

// WriteStub writes a stub shell script to dir/name with the given body.
func WriteStub(tester TestT, dirOrStub any, parts ...string) {
	tester.Helper()

	normalizedStub := normalizeStub(tester, dirOrStub, parts)

	path := filepath.Join(normalizedStub.Dir, normalizedStub.Name)

	err := os.WriteFile(path, []byte(normalizedStub.Body), privateFileMode)
	if err != nil {
		tester.Fatalf("write %s stub: %v", normalizedStub.Name, err)
	}

	err = os.Chmod(path, stubExecutableMode)
	if err != nil {
		tester.Fatalf("mark %s stub executable: %v", normalizedStub.Name, err)
	}
}

func normalizeStub(tester TestT, dirOrStub any, parts []string) stub {
	tester.Helper()

	normalized, ok := dirOrStub.(stub)

	if !ok {
		return newStubFromParts(tester, dirOrStub, parts)
	}

	if len(parts) != zeroIndex {
		tester.Fatalf("stub does not accept positional arguments: %v", parts)
	}

	return normalized
}

func newStubFromParts(tester TestT, dirOrStub any, parts []string) stub {
	tester.Helper()

	dir := requireStringValue(tester, "stub dir must be string, got %T", dirOrStub)

	if len(parts) != readmeTableMatchCount {
		tester.Fatalf("stub requires name and body, got %d values", len(parts))
	}

	return stub{Dir: dir, Name: parts[zeroIndex], Body: parts[valueOffset]}
}

// SimplePublicTaskNames extracts public task names from a decoded tasks map.
func SimplePublicTaskNames(tasks map[string]any) []string {
	var names []string

	for name := range tasks {
		raw := tasks[name]

		if isSimplePublicTask(name, raw) {
			names = append(names, name)
		}
	}

	slices.Sort(names)

	return names
}

func isSimplePublicTask(name string, raw any) bool {
	if isDefaultOrPrivateTaskName(name) {
		return false
	}

	return !isInternalSimpleTask(raw)
}

func isDefaultOrPrivateTaskName(name string) bool {
	return name == constTasktestutilDefault || strings.HasPrefix(name, underscorePrefix)
}

func isInternalSimpleTask(raw any) bool {
	task, ok := raw.(map[string]any)

	if !ok {
		return false
	}

	internal, internalOK := task["internal"].(bool)

	return internalOK && internal
}

// ReadmePublicTaskNames parses a README and returns task names listed in the
// "## Public Tasks" table (backtick-quoted entries in the first column).
func ReadmePublicTaskNames(content string) []string {
	row := regexp.MustCompile(`^\|\s*` + "`" + `([^` + "`" + `]+)` + "`" + `\s*\|`)

	names := collectReadmeTaskNames(content, row)

	slices.Sort(names)

	return names
}

func collectReadmeTaskNames(content string, row *regexp.Regexp) []string {
	scanner := readmeTableScanner{row: row, names: nil, inTable: false}

	for line := range strings.SplitSeq(content, newline) {
		if !scanner.consume(strings.TrimSpace(line)) {
			break
		}
	}

	return scanner.names
}

func (scanner *readmeTableScanner) consume(trimmed string) bool {
	if trimmed == constTasktestutilPublicTasks {
		scanner.inTable = true

		return true
	}

	if !scanner.inTable {
		return true
	}

	if strings.HasPrefix(trimmed, "## ") {
		return false
	}

	if name, ok := matchReadmeTaskRow(scanner.row, trimmed); ok {
		scanner.names = append(scanner.names, name)
	}

	return true
}

func matchReadmeTaskRow(row *regexp.Regexp, line string) (string, bool) {
	matches := row.FindStringSubmatch(line)

	if len(matches) != readmeTableMatchCount {
		return emptyString, false
	}

	return matches[valueOffset], true
}

// MustRead reads a file and fails the test on error.
func MustRead(tester TestT, path string) string {
	tester.Helper()

	content, err := readFileBytes(path)
	if err != nil {
		tester.Fatalf("read %s: %v", path, err)
	}

	return string(content)
}

// --- YAML helpers ---.

// DocumentRoot returns the root mapping node of a YAML document node.
func DocumentRoot(tester TestT, doc *yaml.Node) *yaml.Node {
	tester.Helper()

	if doc.Kind != yaml.DocumentNode || len(doc.Content) == zeroIndex {
		tester.Fatal("invalid YAML document")
	}

	root := doc.Content[zeroIndex]

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

	for i := zeroIndex; i < len(mapping.Content); i += readmeTableMatchCount {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+valueOffset]
		}
	}

	return nil
}

// NodeText returns the trimmed text content of a YAML node.
func NodeText(node *yaml.Node) string {
	if node == nil {
		return emptyString
	}

	if node.Kind == yaml.ScalarNode {
		return strings.TrimSpace(node.Value)
	}

	var parts []string

	for i := range node.Content {
		child := node.Content[i]

		if text := NodeText(child); text != emptyString {
			parts = append(parts, text)
		}
	}

	return strings.TrimSpace(strings.Join(parts, singleSpace))
}

// IsEmptyNode reports whether a YAML node is nil or carries no content.
func IsEmptyNode(node *yaml.Node) bool {
	if node == nil {
		return true
	}

	if node.Kind == yaml.ScalarNode {
		return strings.TrimSpace(node.Value) == emptyString
	}

	return len(node.Content) == zeroIndex
}

// FileExists reports whether the given path exists.
func FileExists(path string) bool {
	info, err := os.Stat(path)

	if info == nil {
		return false
	}

	return err == nil
}

// ReadFile reads a file and fails the test on error.
func ReadFile(tester TestT, path string) string {
	tester.Helper()

	content, err := readFileBytes(path)
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

	return collectByKind(node)
}

func collectByKind(node *yaml.Node) []string {
	commandCollectors := map[yaml.Kind]func(*yaml.Node) []string{
		yaml.DocumentNode: collectSequenceCommandStrings,
		yaml.MappingNode:  collectMappingCommandStrings,
		yaml.ScalarNode:   collectNonEmptyScalar,
		yaml.SequenceNode: collectSequenceCommandStrings,
		yaml.AliasNode:    nil,
	}
	collector, ok := commandCollectors[node.Kind]

	if !ok || collector == nil {
		return nil
	}

	return collector(node)
}

func collectNonEmptyScalar(node *yaml.Node) []string {
	if strings.TrimSpace(node.Value) == emptyString {
		return nil
	}

	return []string{node.Value}
}

func collectSequenceCommandStrings(node *yaml.Node) []string {
	out := make([]string, zeroIndex, len(node.Content))

	for i := range node.Content {
		child := node.Content[i]

		out = append(out, CollectCommandStrings(child)...)
	}

	return out
}

func collectMappingCommandStrings(node *yaml.Node) []string {
	out := make([]string, zeroIndex, len(node.Content)/readmeTableMatchCount)

	for i := zeroIndex; i < len(node.Content); i += readmeTableMatchCount {
		out = append(out, commandStringsForMappingPair(node, i)...)
	}

	return out
}

func commandStringsForMappingPair(node *yaml.Node, i int) []string {
	key := node.Content[i]
	value := node.Content[i+valueOffset]

	switch key.Value {
	case "cmd", "sh":
		return scalarCommandString(value)
	case "cmds", "status", "preconditions":
		return CollectCommandStrings(value)
	default:
		return nil
	}
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

	for i := range matches {
		match := matches[i]

		if len(match) > valueOffset {
			out = append(out, match[valueOffset])
		}
	}

	return out
}

// AssertExitCode fails the test if the command result exit code differs from expected.
func AssertExitCode(tester TestT, result any, expected int) {
	tester.Helper()

	commandResult := normalizeCommandResult(tester, result)
	actual := commandExitCode(tester, commandResult)

	if actual != expected {
		failExitCodeMismatch(tester, &exitCodeMismatch{
			result:   commandResult,
			expected: expected,
			actual:   actual,
		})
	}
}

func commandExitCode(tester TestT, result *CommandResult) int {
	tester.Helper()

	if result.Err == nil {
		return zeroIndex
	}

	exitErr := new(exec.ExitError)

	if !errors.As(result.Err, &exitErr) {
		tester.Fatalf(
			"command failed without exit code\nargs: %v\nerror: %v\nstdout:\n%s\nstderr:\n%s",
			result.Args,
			result.Err,
			result.Stdout,
			result.Stderr,
		)
	}

	return exitErr.ExitCode()
}

func failExitCodeMismatch(tester TestT, mismatch *exitCodeMismatch) {
	tester.Helper()

	tester.Fatalf(
		"expected exit code %d, got %d\nargs: %v\nerror: %v\nstdout:\n%s\nstderr:\n%s",
		mismatch.expected,
		mismatch.actual,
		mismatch.result.Args,
		mismatch.result.Err,
		mismatch.result.Stdout,
		mismatch.result.Stderr,
	)
}

func normalizeCommandResult(tester TestT, result any) *CommandResult {
	tester.Helper()

	switch commandResult := result.(type) {
	case *CommandResult:
		return commandResult
	case CommandResult:
		return &commandResult
	default:
		tester.Fatalf("result must be CommandResult or *CommandResult, got %T", result)

		return nil
	}
}

// AssertContains fails the test if value does not contain expected.
func AssertContains(tester TestT, value, expected string) {
	tester.Helper()

	if !strings.Contains(value, expected) {
		tester.Fatalf("expected output to contain %q\n\nOutput:\n%s", expected, value)
	}
}

// AssertNotContains fails the test if value contains unexpected.
func AssertNotContains(tester TestT, value, unexpected string) {
	tester.Helper()

	if strings.Contains(value, unexpected) {
		tester.Fatalf("expected output not to contain %q\n\nOutput:\n%s", unexpected, value)
	}
}

// AssertNotEmpty fails the test if value is empty after trimming whitespace.
func AssertNotEmpty(tester TestT, value, message string) {
	tester.Helper()

	if strings.TrimSpace(value) == emptyString {
		tester.Fatal(message)
	}
}

// AssertFileExists fails the test unless path exists and is a file.
func AssertFileExists(tester TestT, path string) {
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
func AssertDirExists(tester TestT, path string) {
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
func AssertDirNotExists(tester TestT, path string) {
	tester.Helper()

	info, err := os.Stat(path)

	if info != nil && err == nil {
		tester.Fatalf(msgExpectedNotToExist, path)
	}

	if !os.IsNotExist(err) {
		tester.Fatalf(msgExpectedNotToExist, path)
	}
}

// AssertDirHasEntries fails the test unless path is a non-empty directory.
func AssertDirHasEntries(tester TestT, path string) {
	tester.Helper()

	entries, err := os.ReadDir(path)
	if err != nil {
		tester.Fatalf("failed to read directory %s: %v", path, err)
	}

	if len(entries) == zeroIndex {
		tester.Fatalf("expected %s to contain at least one entry", path)
	}
}

// AssertGithubGroupOutput validates GitHub Actions grouped-output configuration.
func AssertGithubGroupOutput(tester TestT, taskName string, outputNode *yaml.Node) {
	tester.Helper()

	groupNode := requireGroupNode(tester, taskName, outputNode)

	if groupNode == nil {
		return
	}

	assertGroupBeginEnd(tester, taskName, groupNode)
	assertGroupErrorOnly(tester, taskName, groupNode)
}

func requireGroupNode(tester TestT, taskName string, outputNode *yaml.Node) *yaml.Node {
	tester.Helper()

	mapping := requireGroupOutputMapping(tester, taskName, outputNode)

	if mapping == nil {
		return nil
	}

	groupNode := NodeMappingValue(mapping, "group")

	if groupNode == nil || groupNode.Kind != yaml.MappingNode {
		tester.Fatalf("task %q output must include group config", taskName)

		return nil
	}

	return groupNode
}

func requireGroupOutputMapping(tester TestT, taskName string, outputNode *yaml.Node) *yaml.Node {
	tester.Helper()

	if outputNode == nil {
		tester.Fatalf(
			"task %q requires output.group config but no output config was found",
			taskName,
		)

		return nil
	}

	if outputNode.Kind != yaml.MappingNode {
		tester.Fatalf("task %q output must use advanced object format, not scalar format", taskName)

		return nil
	}

	return outputNode
}

func assertGroupBeginEnd(tester TestT, taskName string, groupNode *yaml.Node) {
	tester.Helper()

	begin := NodeText(NodeMappingValue(groupNode, "begin"))
	end := NodeText(NodeMappingValue(groupNode, "end"))

	if begin != groupBeginTemplate {
		tester.Fatalf(
			"task %q output.group.begin must be %q, got %q",
			taskName,
			groupBeginTemplate,
			begin,
		)
	}

	if end != groupEndMarker {
		tester.Fatalf("task %q output.group.end must be %q, got %q", taskName, groupEndMarker, end)
	}
}

func assertGroupErrorOnly(tester TestT, taskName string, groupNode *yaml.Node) {
	tester.Helper()

	errorOnly := NodeMappingValue(groupNode, "error_only")

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
func AssertDestructivePrompt(tester TestT, taskName string, prompt *yaml.Node) {
	tester.Helper()

	if prompt == nil || NodeText(prompt) == emptyString {
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

func explicitPromptTokens() []string {
	return []string{
		promptTextSure,
		promptTextConfirm,
		promptTextRemove,
		promptTextUninstall,
		promptTextDelete,
		promptTextContinue,
	}
}

func explicitPromptText(text string) bool {
	lower := strings.ToLower(text)

	tokens := explicitPromptTokens()

	for i := range tokens {
		if strings.Contains(lower, tokens[i]) {
			return true
		}
	}

	return false
}

// AssertTextFileClean validates portable text-file formatting rules.
func AssertTextFileClean(tester TestT, path, content string) {
	tester.Helper()

	assertTextFileNotEmpty(tester, path, content)
	assertTextFileLineEndings(tester, path, content)
	assertTextFileNoTabs(tester, path, content)
	assertTextFileTrailingNewline(tester, path, content)
	assertTextFileNoTrailingWhitespace(tester, path, content)
}

func assertTextFileNotEmpty(tester TestT, path, content string) {
	tester.Helper()

	if content == emptyString {
		tester.Fatalf("%s is empty", path)
	}
}

func assertTextFileLineEndings(tester TestT, path, content string) {
	tester.Helper()

	if strings.Contains(content, "\r\n") {
		tester.Fatalf("%s uses CRLF line endings; use LF only", path)
	}
}

func assertTextFileNoTabs(tester TestT, path, content string) {
	tester.Helper()

	if strings.Contains(content, "\t") {
		tester.Fatalf("%s contains tabs; use spaces in YAML", path)
	}
}

func assertTextFileTrailingNewline(tester TestT, path, content string) {
	tester.Helper()

	if !strings.HasSuffix(content, newline) {
		tester.Fatalf("%s must end with a newline", path)
	}
}

func assertTextFileNoTrailingWhitespace(tester TestT, path, content string) {
	tester.Helper()

	for i := range strings.Split(content, newline) {
		line := strings.Split(content, newline)[i]

		if strings.TrimRight(line, singleSpace) != line {
			tester.Fatalf("%s has trailing whitespace at line %d", path, i+valueOffset)
		}
	}
}

// AssertNoDuplicateMappingKeys fails the test if a YAML tree contains duplicate keys.
func AssertNoDuplicateMappingKeys(tester TestT, node *yaml.Node, path string) {
	tester.Helper()

	if node == nil {
		return
	}

	if isNonEmptyDocumentNode(node) {
		AssertNoDuplicateMappingKeys(tester, node.Content[zeroIndex], path)

		return
	}

	assertNoDuplicateKeysByKind(tester, node, path)
}

func assertNoDuplicateKeysByKind(tester TestT, node *yaml.Node, path string) {
	tester.Helper()

	if node.Kind == yaml.MappingNode {
		assertNoDuplicateKeysInMapping(tester, node, path)

		return
	}

	if node.Kind == yaml.SequenceNode {
		assertNoDuplicateKeysInSequence(tester, node, path)
	}
}

func isNonEmptyDocumentNode(node *yaml.Node) bool {
	return node.Kind == yaml.DocumentNode && len(node.Content) > zeroIndex
}

func assertNoDuplicateKeysInMapping(tester TestT, node *yaml.Node, path string) {
	tester.Helper()

	seen := map[string]bool{}

	for i := zeroIndex; i < len(node.Content); i += readmeTableMatchCount {
		key := node.Content[i]
		value := node.Content[i+valueOffset]

		if seen[key.Value] {
			tester.Fatalf("duplicate YAML key at %s.%s", path, key.Value)
		}

		seen[key.Value] = true
		AssertNoDuplicateMappingKeys(tester, value, path+"."+key.Value)
	}
}

func assertNoDuplicateKeysInSequence(tester TestT, node *yaml.Node, path string) {
	tester.Helper()

	for i := range node.Content {
		child := node.Content[i]
		AssertNoDuplicateMappingKeys(tester, child, fmt.Sprintf(seqIndexFormat, path, i))
	}
}

// AssertNoYamlAliases fails the test if a YAML tree contains anchors or aliases.
func AssertNoYamlAliases(tester TestT, node *yaml.Node, path string) {
	tester.Helper()

	if node == nil {
		return
	}

	if node.Kind == yaml.AliasNode {
		tester.Fatalf("YAML aliases/anchors are not allowed for clean Taskfile config at %s", path)
	}

	for i := range node.Content {
		child := node.Content[i]
		AssertNoYamlAliases(tester, child, fmt.Sprintf(seqIndexFormat, path, i))
	}
}

// AssertNoPlaceholderText fails the test if value contains common placeholder text.
func AssertNoPlaceholderText(tester TestT, taskName, value string) {
	tester.Helper()

	upper := strings.ToUpper(value)

	for i := range []string{
		placeholderTODO, placeholderFIXME, placeholderChangeme, placeholderCopyright, placeholderLoremIpsum,
	} {
		placeholder := []string{
			placeholderTODO, placeholderFIXME, placeholderChangeme, placeholderCopyright, placeholderLoremIpsum,
		}[i]

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

func readFileBytes(path string) ([]byte, error) {
	clean := filepath.Clean(path)

	content, err := fs.ReadFile(os.DirFS(filepath.Dir(clean)), filepath.Base(clean))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", clean, err)
	}

	return content, nil
}
