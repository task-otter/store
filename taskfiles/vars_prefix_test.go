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
		t          *testing.T
		allowlists *varsPrefixAllowlists
		violations []varsPrefixViolation
	}

	taskfileModuleWalk struct {
		t            *testing.T
		onModule     func(string)
		taskfilesDir string
		errPrefix    string
	}

	taskfileModuleWalkParams = struct {
		t         *testing.T
		onModule  func(string)
		dir       string
		errPrefix string
	}

	walkedModuleParams = struct {
		entry        fs.DirEntry
		walkErr      error
		taskfilesDir string
		path         string
		errPrefix    string
	}

	moduleNamePair = struct {
		module string
		name   string
	}
)

const (
	varsPrefixTestName = "TestTopLevelVarsPrefix"
)

// TestTopLevelVarsPrefix enforces ADR 0002: top-level Taskfile vars must use
// the owning module prefix, or an allowlisted foreign/companion prefix.
func TestTopLevelVarsPrefix(t *testing.T) {
	t.Parallel()

	runTopLevelVarsPrefix(t)
}

func runTopLevelVarsPrefix(t *testing.T) {
	t.Helper()

	taskfilesDir := filepath.Join(tasktest.RepoRoot(t), taskfilesDirName)
	collector := newVarsPrefixCollector(t, taskfilesDir)
	walker := newTaskfileModuleWalk(&taskfileModuleWalkParams{
		t:         t,
		onModule:  collector.appendModuleViolations,
		dir:       taskfilesDir,
		errPrefix: "vars prefix module path",
	})

	err := filepath.WalkDir(taskfilesDir, walker.collect)
	if err != nil {
		t.Fatalf(walkTaskfilesErrFormat, err)
	}

	collector.failIfViolations()
}

func newVarsPrefixCollector(t *testing.T, taskfilesDir string) *varsPrefixCollector {
	t.Helper()

	allowlists := buildVarsPrefixAllowlists(t, taskfilesDir)

	return &varsPrefixCollector{
		t:          t,
		allowlists: &allowlists,
		violations: nil,
	}
}

func newTaskfileModuleWalk(params *taskfileModuleWalkParams) *taskfileModuleWalk {
	params.t.Helper()

	return &taskfileModuleWalk{
		t:            params.t,
		onModule:     params.onModule,
		taskfilesDir: params.dir,
		errPrefix:    params.errPrefix,
	}
}

func (collector *varsPrefixCollector) appendModuleViolations(module string) {
	collector.t.Helper()

	if module == taskfilesRootModule {
		return
	}

	taskfile := tasktest.LoadTaskfile(collector.t, module)

	if len(taskfile.Vars) == constZero {
		return
	}

	owned := ownedVarsPrefix(module)
	collector.recordVars(module, owned, taskfile.Vars)
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
		strings.Join(formatVarsPrefixViolations(collector.violations), windowsSeparator),
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

func (walk *taskfileModuleWalk) collect(path string, entry fs.DirEntry, walkErr error) error {
	walk.t.Helper()

	module, skip, err := resolveWalkedModule(&walkedModuleParams{
		entry:        entry,
		walkErr:      walkErr,
		taskfilesDir: walk.taskfilesDir,
		path:         path,
		errPrefix:    walk.errPrefix,
	})
	if err != nil {
		return fmt.Errorf("walk taskfile module: %w", err)
	}

	if skip {
		return nil
	}

	walk.onModule(module)

	return nil
}

func formatVarsPrefixViolations(violations []varsPrefixViolation) []string {
	sortByModuleThenName(violations, func(violation varsPrefixViolation) moduleNamePair {
		return moduleNamePair{module: violation.module, name: violation.name}
	})

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
		companionPrefixes: []string{"RUST_", "RUSTUP_", "PROTOC_", "NODE_", "WINDOWS_"},
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
		return emptyString, fmt.Errorf("taskfile path relative to taskfiles: %w", err)
	}

	return filepath.ToSlash(rel), nil
}

func resolveWalkedModule(params *walkedModuleParams) (module string, skip bool, err error) {
	if params.walkErr != nil {
		return emptyString, false, params.walkErr
	}

	if params.entry.IsDir() || params.entry.Name() != skipTaskfileYML {
		return emptyString, true, nil
	}

	module, err = modulePathForTaskfile(params.taskfilesDir, params.path)
	if err != nil {
		return emptyString, false, fmt.Errorf("%s: %w", params.errPrefix, err)
	}

	return module, false, nil
}

func compareModuleThenName(left, right *moduleNamePair) int {
	if left.module != right.module {
		return strings.Compare(left.module, right.module)
	}

	return strings.Compare(left.name, right.name)
}

func sortByModuleThenName[itemT any](items []itemT, pair func(itemT) moduleNamePair) {
	slices.SortFunc(items, func(left, right itemT) int {
		leftPair := pair(left)
		rightPair := pair(right)

		return compareModuleThenName(&leftPair, &rightPair)
	})
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
