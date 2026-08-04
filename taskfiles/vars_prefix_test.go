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

const (
	varsPrefixTestName = "TestTopLevelVarsPrefix"
)

type (
	varsPrefixViolation struct {
		module string
		name   string
		want   string
	}

	varsPrefixAllowlists struct {
		shared            map[string]struct{}
		foreignPrefixes   []string
		companionPrefixes []string
	}
)

func TestTopLevelVarsPrefix(t *testing.T) {
	t.Parallel()

	root := tasktest.RepoRoot(t)
	taskfilesDir := filepath.Join(root, taskfilesDirName)
	allowlists := buildVarsPrefixAllowlists(t, taskfilesDir)

	var violations []varsPrefixViolation

	err := filepath.WalkDir(taskfilesDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() || d.Name() != skipTaskfileYML {
			return nil
		}

		module, err := modulePathForTaskfile(taskfilesDir, path)
		if err != nil {
			return err
		}

		if module == "." {
			return nil
		}

		taskfile := tasktest.LoadTaskfile(t, module)
		if len(taskfile.Vars) == constZero {
			return nil
		}

		owned := ownedVarsPrefix(module)
		for name := range taskfile.Vars {
			if varsPrefixAllowed(name, owned, allowlists) {
				continue
			}

			violations = append(violations, varsPrefixViolation{
				module: module,
				name:   name,
				want:   owned,
			})
		}

		return nil
	})
	if err != nil {
		t.Fatalf(walkTaskfilesErrFormat, err)
	}

	if len(violations) == constZero {
		return
	}

	slices.SortFunc(violations, func(left, right varsPrefixViolation) int {
		if left.module != right.module {
			return strings.Compare(left.module, right.module)
		}

		return strings.Compare(left.name, right.name)
	})

	lines := make([]string, constZero, len(violations))
	for i := range violations {
		violation := violations[i]
		lines = append(lines, fmt.Sprintf(
			"%s: %s (want %s… or allowlisted)",
			violation.module,
			violation.name,
			violation.want,
		))
	}

	t.Fatalf(
		"%s: %d top-level var(s) missing owned/shared/foreign prefix:\n%s",
		varsPrefixTestName,
		len(violations),
		strings.Join(lines, "\n"),
	)
}

func buildVarsPrefixAllowlists(t *testing.T, taskfilesDir string) varsPrefixAllowlists {
	t.Helper()

	return varsPrefixAllowlists{
		shared: map[string]struct{}{
			"EXTRA_ARGS":   {},
			"VERSION":      {},
			"TARGETS":      {},
			"CONFIG":       {},
			"FILE":         {},
			"ARGS":         {},
			"REQUIREMENTS": {},
			"VENV":         {},
			"IGNORE_PATH":  {},
			"FORCE":        {},
		},
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

	prefixes := make([]string, constZero, len(entries))
	for i := range entries {
		entry := entries[i]
		if !entry.IsDir() {
			continue
		}

		prefixes = append(prefixes, dirNameToVarsPrefix(entry.Name()))
	}

	slices.Sort(prefixes)

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
	if family == "internal" && len(parts) > 1 {
		family = parts[1]
	}

	return dirNameToVarsPrefix(family)
}

func dirNameToVarsPrefix(name string) string {
	return strings.ToUpper(strings.ReplaceAll(name, "-", "_")) + "_"
}

func varsPrefixAllowed(name, owned string, allowlists varsPrefixAllowlists) bool {
	if strings.HasPrefix(name, owned) {
		return true
	}

	if _, ok := allowlists.shared[name]; ok {
		return true
	}

	for i := range allowlists.companionPrefixes {
		if strings.HasPrefix(name, allowlists.companionPrefixes[i]) {
			return true
		}
	}

	for i := range allowlists.foreignPrefixes {
		prefix := allowlists.foreignPrefixes[i]
		if prefix == owned {
			continue
		}

		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	return false
}
