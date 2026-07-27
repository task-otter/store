package tasktest

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"
)

// This file implements a golden-output harness for the `*_OVERRIDE` variable
// contract. Every module Taskfile resolves a base var against its `_OVERRIDE`
// twin; the goal is to capture the fully rendered command line that each public
// task emits, for every combination of (base set, override set), so that a
// refactor of the resolution templates can be proven byte-for-byte inert.
//
// Goldens are recorded per GOOS: modules use `platforms:` guards, and several
// compute values from `uname`, so output legitimately differs between darwin
// and linux. A platform without recorded goldens skips rather than fails.

// GoldenCase is one point in the base/override matrix.
type GoldenCase string

const (
	// CaseBare passes no variables; the module's own defaults apply.
	CaseBare GoldenCase = "bare"
	// CaseBase sets only the base variables.
	CaseBase GoldenCase = "base"
	// CaseOverride sets only the *_OVERRIDE variables.
	CaseOverride GoldenCase = "override"
	// CaseBoth sets base and override to *different* values. This is the case
	// that proves precedence: the override value must win.
	CaseBoth GoldenCase = "both"
)

// GoldenCases returns the matrix in a stable order.
func GoldenCases() []GoldenCase {
	return []GoldenCase{CaseBare, CaseBase, CaseOverride, CaseBoth}
}

var (
	overrideRefPattern = regexp.MustCompile(`\.([A-Z][A-Z0-9_]*)_OVERRIDE\b`)
	identPattern       = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
)

// GoldenModules returns every module path under taskfiles/ that declares at
// least one task, relative to the taskfiles/ directory (e.g. "eslint/bun").
// Family roots that only carry `includes:` are excluded because they have no
// commands of their own to render.
func GoldenModules(t testT) []string {
	t.Helper()

	root := filepath.Join(RepoRoot(t), "taskfiles")

	var modules []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "Taskfile.yml" {
			return nil
		}

		rel, relErr := filepath.Rel(root, filepath.Dir(path))
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}

		module := filepath.ToSlash(rel)
		if len(LoadTaskfile(t, module).Tasks) > 0 {
			modules = append(modules, module)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk taskfiles: %v", err)
	}

	sort.Strings(modules)
	return modules
}

// GoldenTasks returns the public tasks of a module in stable order.
func GoldenTasks(t testT, module string) []string {
	t.Helper()
	return publicTaskNames(LoadTaskfile(t, module))
}

// goldenVars derives the variable matrix for a module.
//
// The base set is the union of (a) every base name implied by a referenced
// `*_OVERRIDE` variable and (b) every declared var holding a simple scalar
// default. (b) matters because some modules pair a generic base name with a
// tool-prefixed override — eslint reads `.TARGETS` but `.ESLINT_TARGETS_OVERRIDE` —
// so neither source alone covers the contract.
//
// Vars holding script bodies (multi-line, templated, or long) are excluded:
// substituting a sentinel for `UV_LOAD` or `UNIX_GLOB_HELPERS` would produce
// pages of deterministic noise without testing any precedence rule.
func goldenVars(t testT, module string) (bases []string, overrides []string) {
	t.Helper()

	content, err := os.ReadFile(taskfilePath(t, module))
	if err != nil {
		t.Fatalf("read %s Taskfile: %v", module, err)
	}
	text := string(content)

	seen := map[string]bool{}
	for _, match := range overrideRefPattern.FindAllStringSubmatch(text, -1) {
		seen[match[1]] = true
	}

	for name, value := range LoadTaskfile(t, module).Vars {
		if !identPattern.MatchString(name) || !isSimpleScalar(value) {
			continue
		}
		seen[name] = true
	}

	for name := range seen {
		bases = append(bases, name)
		overrides = append(overrides, name+"_OVERRIDE")
	}

	sort.Strings(bases)
	sort.Strings(overrides)
	return bases, overrides
}

// isSimpleScalar reports whether a declared var default is a plain short string
// rather than an embedded script or template.
func isSimpleScalar(value any) bool {
	text, ok := value.(string)
	if !ok {
		return false
	}
	if strings.Contains(text, "\n") || strings.Contains(text, "{{") {
		return false
	}
	return len(text) <= 40
}

// goldenValue picks the sentinel for a variable in a given role.
//
// Variables that the Taskfile compares against the literal "true" are treated
// as booleans, and are given values that make the override observable: the base
// disables the flag, the override enables it. For those, CaseBoth renders the
// flag if and only if the override correctly takes precedence.
func goldenValue(text, name string, override bool) string {
	if isBooleanVar(text, name) {
		if override {
			return "true"
		}
		return "false"
	}
	if override {
		return "OVR-" + name
	}
	return "BASE-" + name
}

func isBooleanVar(text, name string) bool {
	return strings.Contains(text, `eq .`+name+` "`) ||
		strings.Contains(text, `eq .`+name+`_OVERRIDE "`)
}

// GoldenRender runs every public task of a module under one matrix case and
// returns the normalized, concatenated dry-run output.
//
// A task that exits non-zero is recorded rather than failing the run: a failing
// precondition is itself behavior worth pinning, and it must stay identical
// across the refactor.
func GoldenRender(t testT, module string, kase GoldenCase) string {
	t.Helper()

	content, err := os.ReadFile(taskfilePath(t, module))
	if err != nil {
		t.Fatalf("read %s Taskfile: %v", module, err)
	}
	text := string(content)

	bases, overrides := goldenVars(t, module)

	var assignments []string
	if kase == CaseBase || kase == CaseBoth {
		for _, name := range bases {
			assignments = append(assignments, name+"="+goldenValue(text, name, false))
		}
	}
	if kase == CaseOverride || kase == CaseBoth {
		for _, name := range overrides {
			base := strings.TrimSuffix(name, "_OVERRIDE")
			assignments = append(assignments, name+"="+goldenValue(text, base, true))
		}
	}

	projectDir, env := setupDryRunEnv(t)
	home := dryRunGetEnv(env, "HOME")

	var out strings.Builder
	out.WriteString("# module: " + module + "\n")
	out.WriteString("# case:   " + string(kase) + "\n")
	if len(assignments) > 0 {
		out.WriteString("# vars:\n")
		for _, assignment := range assignments {
			out.WriteString("#   " + assignment + "\n")
		}
	}

	for _, task := range GoldenTasks(t, module) {
		if isUnpinnable(module, task) {
			out.WriteString("\n=== task: " + task + " ===\n")
			out.WriteString("--- unstable: output varies between runs; not pinned\n")
			continue
		}

		// --concurrency 1 serializes `deps:`. Without it, concurrent deps race
		// `run: once`, so a shared dependency is reported as "started" in one
		// run and "skipping execution" in the next.
		args := []string{
			"--taskfile", taskfilePath(t, module),
			"--dry", "--yes", "--verbose", "--concurrency", "1",
			task,
		}
		args = append(args, assignments...)

		rendered, exitErr, stable := stabilize(func() (string, string) {
			output, exitErr := runGolden(t, projectDir, env, args)
			lines := canonicalGoldenLines(normalizeGolden(output, projectDir, home, RepoRoot(t)))
			return strings.Join(lines, "\n"), exitErr
		})

		out.WriteString("\n=== task: " + task + " ===\n")
		if exitErr != "" {
			out.WriteString("--- exit: " + exitErr + "\n")
		}
		if !stable {
			// A handful of tasks fan out to concurrent `deps:` whose completion
			// order — and therefore whose rendered output — varies between runs.
			// Pinning an arbitrary branch would produce a flaky test, so the
			// instability is recorded instead of the output. These tasks are
			// genuinely not covered; see the harness notes in the README.
			out.WriteString("--- unstable: output varies between runs; not pinned\n")
			continue
		}
		out.WriteString(rendered + "\n")
	}

	return out.String()
}

func runGolden(t testT, dir string, env []string, args []string) (output string, exitErr string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, taskBinary, args...)
	cmd.Dir = dir
	cmd.Env = env

	// Capture the streams separately. Interleaving them into one pipe lets the
	// write boundaries of stdout and stderr fall differently between runs,
	// which changes where physical lines break.
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("task command timed out: task %s", strings.Join(args, " "))
	}
	if err != nil {
		exitErr = err.Error()
	}

	return stdout.String() + stderr.String(), exitErr
}

// normalizeGolden strips per-run absolute paths so goldens compare across runs.
// Trailing whitespace is deliberately preserved: the whitespace emitted around
// each rendered flag is precisely what the refactor must not disturb.
func normalizeGolden(output, projectDir, home, repoRoot string) string {
	replacements := []struct{ from, to string }{
		{projectDir, "$PROJECT"},
		{home, "$HOME"},
		{repoRoot, "$REPO"},
	}
	// Longest path first, so a nested path is not partially rewritten.
	slices.SortFunc(replacements, func(a, b struct{ from, to string }) int {
		return len(b.from) - len(a.from)
	})

	for _, replacement := range replacements {
		if replacement.from == "" {
			continue
		}
		output = strings.ReplaceAll(output, replacement.from, replacement.to)
	}
	return output
}

// unpinnableTasks lists module tasks whose dry-run output cannot be pinned.
//
// `go:lint` fans out to four concurrent `deps:`; Task renders them in whatever
// order they acquire a worker, so the captured output legitimately differs
// between runs even at --concurrency 1. Detecting this at capture time is not
// enough, because the instability is itself intermittent — two runs sometimes
// agree by chance — which would make the marker flaky. Declaring it keeps both
// capture and comparison deterministic.
//
// These tasks are NOT covered by the harness. A refactor touching them must be
// verified by reading the diff of the rendered Taskfile by hand.
var unpinnableTasks = map[string][]string{
	"go": {"lint"},
}

func isUnpinnable(module, task string) bool {
	return slices.Contains(unpinnableTasks[module], task)
}

// stabilizeAttempts is how many renders are taken while looking for two that
// agree. Three is enough to clear incidental scheduler jitter without masking a
// task that is genuinely nondeterministic.
const stabilizeAttempts = 3

// stabilize renders repeatedly until two consecutive results match, and reports
// whether it succeeded. Requiring agreement is what keeps the golden honest: a
// task whose output genuinely varies is flagged rather than silently pinned to
// whichever branch happened to run first.
func stabilize(render func() (output string, exitErr string)) (output string, exitErr string, stable bool) {
	previous, previousExit := render()
	for range stabilizeAttempts - 1 {
		current, currentExit := render()
		if current == previous && currentExit == previousExit {
			return current, currentExit, true
		}
		previous, previousExit = current, currentExit
	}
	return previous, previousExit, false
}

// canonicalGoldenLines sorts a task's output lines.
//
// Task writes progress to stderr and rendered commands to stdout, so even a
// serialized run interleaves the two streams unpredictably. Sorting makes the
// capture reproducible while preserving every line verbatim — including the
// trailing whitespace this harness exists to protect. Execution *order* is
// deliberately not pinned; it is already nondeterministic today. What is pinned
// is the exact set of commands each task renders.
func canonicalGoldenLines(output string) []string {
	// Task emits some progress messages joined by a literal backslash-n rather
	// than a real newline. How many end up on one physical line depends on
	// write chunking, so expand them before splitting.
	output = strings.ReplaceAll(output, `\n`, "\n")

	lines := strings.Split(output, "\n")

	// Drop only the empty element produced by a trailing newline; blank lines
	// inside the output are real content and are kept.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	sort.Strings(lines)
	return lines
}

// GoldenPath returns the on-disk location of a module/case golden.
func GoldenPath(t testT, module string, kase GoldenCase) string {
	t.Helper()
	return filepath.Join(
		RepoRoot(t), "internal", "tasktest", "testdata", "golden",
		runtime.GOOS,
		filepath.FromSlash(module),
		string(kase)+".txt",
	)
}
