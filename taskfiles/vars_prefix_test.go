// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskfiles_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

type (
	varsPrefixViolation struct {
		module string
		name   string
		want   string
	}

	varsPrefixAllowlists struct {
		foreignPrefixes   []string
		companionPrefixes []string
	}

	varsPrefixCollector struct {
		t            *testing.T
		taskfilesDir string
		allowlists   *varsPrefixAllowlists
		violations   []varsPrefixViolation
	}
)

const (
	varsPrefixTestName = "TestTopLevelVarsPrefix"
)

// TestTopLevelVarsPrefix enforces ADR 0002: top-level Taskfile vars must use
// the owning module prefix, or an allowlisted foreign/companion prefix.
func TestTopLevelVarsPrefix(t *testing.T) {
	t.Parallel()

	root := tasktest.RepoRoot(t)
	taskfilesDir := filepath.Join(root, taskfilesDirName)
	allowlists := buildVarsPrefixAllowlists(t, taskfilesDir)
	collector := varsPrefixCollector{
		t:            t,
		taskfilesDir: taskfilesDir,
		allowlists:   &allowlists,
		violations:   nil,
	}

	err := filepath.WalkDir(taskfilesDir, collector.collect)
	if err != nil {
		t.Fatalf(walkTaskfilesErrFormat, err)
	}

	collector.failIfViolations()
}

func (collector *varsPrefixCollector) appendModuleViolations(module string) {
	collector.t.Helper()

	if module == "." {
		return
	}

	taskfile := tasktest.LoadTaskfile(collector.t, module)

	if len(taskfile.Vars) == constZero {
		return
	}

	owned := ownedVarsPrefix(module)
	collector.recordVars(module, owned, taskfile.Vars)
}

func (collector *varsPrefixCollector) collect(path string, entry fs.DirEntry, walkErr error) error {
	collector.t.Helper()

	if walkErr != nil {
		return walkErr
	}

	if entry.IsDir() || entry.Name() != skipTaskfileYML {
		return nil
	}

	module, err := modulePathForTaskfile(collector.taskfilesDir, path)
	if err != nil {
		return fmt.Errorf("vars prefix module path: %w", err)
	}

	collector.appendModuleViolations(module)

	return nil
}

func (collector *varsPrefixCollector) failIfViolations() {
	collector.t.Helper()

	if len(collector.violations) == constZero {
		return
	}

	collector.t.Fatalf(
		"%s: %d top-level var(s) missing owned/foreign/companion prefix:\n%s",
		varsPrefixTestName,
		len(collector.violations),
		strings.Join(collector.violationLines(), windowsSeparator),
	)
}

func (collector *varsPrefixCollector) recordVars(module, owned string, vars map[string]any) {
	collector.t.Helper()

	for name := range vars {
		if varsPrefixAllowed(name, owned, collector.allowlists) {
			continue
		}

		collector.violations = append(collector.violations, varsPrefixViolation{
			module: module,
			name:   name,
			want:   owned,
		})
	}
}

func (collector *varsPrefixCollector) violationLines() []string {
	collector.t.Helper()

	slices.SortFunc(collector.violations, func(left, right varsPrefixViolation) int {
		if left.module != right.module {
			return strings.Compare(left.module, right.module)
		}

		return strings.Compare(left.name, right.name)
	})

	return formatVarsPrefixViolations(collector.violations)
}

func formatVarsPrefixViolations(violations []varsPrefixViolation) []string {
	lines := make([]string, constZero, len(violations))

	for i := range violations {
		violation := &violations[i]

		lines = append(lines, fmt.Sprintf(
			"%s: %s (want %s… or foreign/companion)",
			violation.module,
			violation.name,
			violation.want,
		))
	}

	return lines
}

func buildVarsPrefixAllowlists(t *testing.T, taskfilesDir string) varsPrefixAllowlists {
	t.Helper()

	return varsPrefixAllowlists{
		foreignPrefixes:   moduleForeignPrefixes(t, taskfilesDir),
		companionPrefixes: []string{"RUST_", "RUSTUP_", "PROTOC_", "YAMLFIX_", "NODE_", "WINDOWS_"},
	}
}

func moduleForeignPrefixes(t *testing.T, taskfilesDir string) []string {
	t.Helper()

	entries, err := os.ReadDir(taskfilesDir)
	if err != nil {
		t.Fatalf("read taskfiles dir: %v", err)
	}

	prefixes := dirEntriesToVarsPrefixes(entries)
	slices.Sort(prefixes)

	return prefixes
}

func dirEntriesToVarsPrefixes(entries []os.DirEntry) []string {
	prefixes := make([]string, constZero, len(entries))

	for i := range entries {
		entry := entries[i]

		if !entry.IsDir() {
			continue
		}

		prefixes = append(prefixes, dirNameToVarsPrefix(entry.Name()))
	}

	return prefixes
}

func modulePathForTaskfile(taskfilesDir, path string) (string, error) {
	rel, err := filepath.Rel(taskfilesDir, filepath.Dir(path))
	if err != nil {
		return "", fmt.Errorf("taskfile path relative to taskfiles: %w", err)
	}

	return filepath.ToSlash(rel), nil
}

func ownedVarsPrefix(module string) string {
	parts := strings.Split(module, pathSeparator)
	family := parts[constZero]

	if family == "internal" && len(parts) > constOne {
		family = parts[constOne]
	}

	return dirNameToVarsPrefix(family)
}

func dirNameToVarsPrefix(name string) string {
	return strings.ToUpper(strings.ReplaceAll(name, hyphenChar, underscoreChar)) + underscoreChar
}

func varsPrefixAllowed(name, owned string, allowlists *varsPrefixAllowlists) bool {
	if strings.HasPrefix(name, owned) {
		return true
	}

	if hasAnyPrefix(name, allowlists.companionPrefixes) {
		return true
	}

	return hasForeignPrefix(name, owned, allowlists.foreignPrefixes)
}

func hasAnyPrefix(name string, prefixes []string) bool {
	for i := range prefixes {
		if strings.HasPrefix(name, prefixes[i]) {
			return true
		}
	}

	return false
}

func hasForeignPrefix(name, owned string, prefixes []string) bool {
	for i := range prefixes {
		prefix := prefixes[i]

		if prefix == owned {
			continue
		}

		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	return false
}
