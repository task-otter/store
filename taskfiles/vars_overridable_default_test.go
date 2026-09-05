// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskfiles_test

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

type (
	varsOverridableViolation struct {
		module string
		name   string
		value  string
	}

	varsOverridableCollector struct {
		t            *testing.T
		taskfilesDir string
		violations   []varsOverridableViolation
	}
)

const (
	varsOverridableTestName = "TestTopLevelVarsOverridableDefault"
	defaultPipeToken        = "| default"
)

// TestTopLevelVarsOverridableDefault requires every top-level Taskfile var to
// use an overridable default so parent includes/CLI can replace the value.
// Bare literals (e.g. VAR: "") lock Task's merge and block includes.*.vars.
func TestTopLevelVarsOverridableDefault(t *testing.T) {
	t.Parallel()

	root := tasktest.RepoRoot(t)
	taskfilesDir := filepath.Join(root, taskfilesDirName)
	collector := varsOverridableCollector{
		t:            t,
		taskfilesDir: taskfilesDir,
		violations:   nil,
	}

	err := filepath.WalkDir(taskfilesDir, collector.collect)
	if err != nil {
		t.Fatalf(walkTaskfilesErrFormat, err)
	}

	collector.failIfViolations()
}

func TestIsOverridableDefaultVarValue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		varName string
		value   any
		want    bool
	}{
		{
			name:    "quoted default empty",
			varName: "GO_FUZZTIME",
			value:   `{{.GO_FUZZTIME | default ""}}`,
			want:    true,
		},
		{
			name:    "quoted default non-empty",
			varName: "GIT_BASE",
			value:   `{{.GIT_BASE | default "main"}}`,
			want:    true,
		},
		{
			name:    "backtick default",
			varName: "NIX_LOAD",
			value:   "{{.NIX_LOAD | default `echo hi`}}",
			want:    true,
		},
		{
			name:    "folded multiline",
			varName: "NIX_LOAD",
			value:   "{{.NIX_LOAD | default `line1\nline2`}}",
			want:    true,
		},
		{
			name:    "extra spaces around pipe",
			varName: "FOO",
			value:   `{{.FOO  |  default "x"}}`,
			want:    true,
		},
		{
			name:    "bare empty string",
			varName: "FOO",
			value:   "",
			want:    false,
		},
		{
			name:    "bare literal",
			varName: "FOO",
			value:   "nixpkgs#go",
			want:    false,
		},
		{
			name:    "template without default",
			varName: "FOO",
			value:   "{{.FOO}}",
			want:    false,
		},
		{
			name:    "default for different var",
			varName: "FOO",
			value:   `{{.BAR | default ""}}`,
			want:    false,
		},
		{
			name:    "non-string map value",
			varName: "FOO",
			value:   map[string]any{"sh": "echo hi"},
			want:    false,
		},
		{
			name:    "nil value",
			varName: "FOO",
			value:   nil,
			want:    false,
		},
	}

	for i := range cases {
		testCase := cases[i]
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := isOverridableDefaultVarValue(testCase.varName, testCase.value)
			if got != testCase.want {
				t.Fatalf(
					"isOverridableDefaultVarValue(%q, %#v) = %v, want %v",
					testCase.varName,
					testCase.value,
					got,
					testCase.want,
				)
			}
		})
	}
}

func (collector *varsOverridableCollector) collect(path string, entry fs.DirEntry, walkErr error) error {
	collector.t.Helper()

	if walkErr != nil {
		return walkErr
	}

	if entry.IsDir() || entry.Name() != skipTaskfileYML {
		return nil
	}

	module, err := modulePathForTaskfile(collector.taskfilesDir, path)
	if err != nil {
		return fmt.Errorf("overridable vars module path: %w", err)
	}

	collector.appendModuleViolations(module)

	return nil
}

func (collector *varsOverridableCollector) appendModuleViolations(module string) {
	collector.t.Helper()

	if module == "." {
		return
	}

	taskfile := tasktest.LoadTaskfile(collector.t, module)

	if len(taskfile.Vars) == constZero {
		return
	}

	collector.recordVars(module, taskfile.Vars)
}

func (collector *varsOverridableCollector) recordVars(module string, vars map[string]any) {
	collector.t.Helper()

	for name, value := range vars {
		if isOverridableDefaultVarValue(name, value) {
			continue
		}

		collector.violations = append(collector.violations, varsOverridableViolation{
			module: module,
			name:   name,
			value:  formatVarValue(value),
		})
	}
}

func (collector *varsOverridableCollector) failIfViolations() {
	collector.t.Helper()

	if len(collector.violations) == constZero {
		return
	}

	collector.t.Fatalf(
		"%s: %d top-level var(s) missing overridable %s form:\n%s",
		varsOverridableTestName,
		len(collector.violations),
		defaultPipeToken,
		strings.Join(collector.violationLines(), windowsSeparator),
	)
}

func (collector *varsOverridableCollector) violationLines() []string {
	collector.t.Helper()

	slices.SortFunc(collector.violations, func(left, right varsOverridableViolation) int {
		if left.module != right.module {
			return strings.Compare(left.module, right.module)
		}

		return strings.Compare(left.name, right.name)
	})

	return formatVarsOverridableViolations(collector.violations)
}

func formatVarsOverridableViolations(violations []varsOverridableViolation) []string {
	lines := make([]string, constZero, len(violations))

	for i := range violations {
		violation := &violations[i]

		lines = append(lines, fmt.Sprintf(
			`%s: %s = %s (want '{{.%s | default ...}}')`,
			violation.module,
			violation.name,
			violation.value,
			violation.name,
		))
	}

	return lines
}

func isOverridableDefaultVarValue(name string, value any) bool {
	text, ok := value.(string)
	if !ok {
		return false
	}

	pattern := regexp.MustCompile(
		`\{\{\s*\.` + regexp.QuoteMeta(name) + `\s*\|\s*default\b`,
	)

	return pattern.MatchString(text)
}

func formatVarValue(value any) string {
	if value == nil {
		return "<nil>"
	}

	text, ok := value.(string)
	if ok {
		return strconv.Quote(text)
	}

	return fmt.Sprintf("%T(%v)", value, value)
}
