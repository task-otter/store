// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskfiles_test

import (
	"fmt"
	"path/filepath"
	"regexp"
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
		t          *testing.T
		violations []varsOverridableViolation
	}

	overridableDefaultVarValueCase struct {
		value   any
		name    string
		varName string
		want    bool
	}
)

const (
	varsOverridableTestName = "TestTopLevelVarsOverridableDefault"
	defaultPipeMarker       = "| default"

	overridableCaseVarFoo     = "FOO"
	overridableCaseVarNixLoad = "NIX_LOAD"
)

// TestTopLevelVarsOverridableDefault requires every top-level Taskfile var to
// use an overridable default so parent includes/CLI can replace the value.
// Bare literals (e.g. VAR: "") lock Task's merge and block includes.*.vars.
func TestTopLevelVarsOverridableDefault(t *testing.T) {
	t.Parallel()

	runTopLevelVarsOverridableDefault(t)
}

func runTopLevelVarsOverridableDefault(t *testing.T) {
	t.Helper()

	taskfilesDir := filepath.Join(tasktest.RepoRoot(t), taskfilesDirName)
	collector := &varsOverridableCollector{t: t, violations: nil}
	walker := newTaskfileModuleWalk(&taskfileModuleWalkParams{
		t:         t,
		onModule:  collector.appendModuleViolations,
		dir:       taskfilesDir,
		errPrefix: "overridable vars module path",
	})

	err := filepath.WalkDir(taskfilesDir, walker.collect)
	if err != nil {
		t.Fatalf(walkTaskfilesErrFormat, err)
	}

	collector.failIfViolations()
}

// TestIsOverridableDefaultVarValue covers the template patterns that count as
// an overridable `| default` form versus bare literals that lock merge.
func TestIsOverridableDefaultVarValue(t *testing.T) {
	t.Parallel()

	runOverridableDefaultVarValueCases(t)
}

func runOverridableDefaultVarValueCases(t *testing.T) {
	t.Helper()

	cases := overridableDefaultVarValueCases()

	for i := range cases {
		assertOverridableDefaultVarValueCase(t, &cases[i])
	}
}

func assertOverridableDefaultVarValueCase(t *testing.T, testCase *overridableDefaultVarValueCase) {
	t.Helper()

	t.Run(testCase.name, func(t *testing.T) {
		t.Parallel()
		assertOverridableDefaultGot(t, testCase)
	})
}

func assertOverridableDefaultGot(t *testing.T, testCase *overridableDefaultVarValueCase) {
	t.Helper()

	got := isOverridableDefaultVarValue(testCase.varName, testCase.value)

	if got == testCase.want {
		return
	}

	t.Fatalf(
		"isOverridableDefaultVarValue(%q, %#v) = %v, want %v",
		testCase.varName,
		testCase.value,
		got,
		testCase.want,
	)
}

func overridableDefaultVarValueCases() []overridableDefaultVarValueCase {
	cases := overridableDefaultAcceptQuotedCases()

	cases = append(cases, overridableDefaultAcceptBacktickCases()...)

	return append(cases, overridableDefaultRejectCases()...)
}

func overridableDefaultAcceptQuotedCases() []overridableDefaultVarValueCase {
	cases := overridableDefaultAcceptQuotedEmptyCases()

	return append(cases, overridableDefaultAcceptQuotedExtraCases()...)
}

func overridableDefaultAcceptQuotedEmptyCases() []overridableDefaultVarValueCase {
	return []overridableDefaultVarValueCase{
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
	}
}

func overridableDefaultAcceptQuotedExtraCases() []overridableDefaultVarValueCase {
	return []overridableDefaultVarValueCase{
		{
			name:    "extra spaces around pipe",
			varName: overridableCaseVarFoo,
			value:   `{{.FOO  |  default "x"}}`,
			want:    true,
		},
	}
}

func overridableDefaultAcceptBacktickCases() []overridableDefaultVarValueCase {
	return []overridableDefaultVarValueCase{
		{
			name:    "backtick default",
			varName: overridableCaseVarNixLoad,
			value:   "{{.NIX_LOAD | default `echo hi`}}",
			want:    true,
		},
		{
			name:    "folded multiline",
			varName: overridableCaseVarNixLoad,
			value:   "{{.NIX_LOAD | default `line1\nline2`}}",
			want:    true,
		},
	}
}

func overridableDefaultRejectCases() []overridableDefaultVarValueCase {
	cases := append(
		overridableDefaultRejectBareCases(),
		overridableDefaultRejectTemplateCases()...,
	)

	return append(cases, overridableDefaultRejectTypeCases()...)
}

func overridableDefaultRejectBareCases() []overridableDefaultVarValueCase {
	return []overridableDefaultVarValueCase{
		{
			name:    "bare empty string",
			varName: overridableCaseVarFoo,
			value:   emptyString,
			want:    false,
		},
		{
			name:    "bare literal",
			varName: overridableCaseVarFoo,
			value:   "nixpkgs#go",
			want:    false,
		},
	}
}

func overridableDefaultRejectTemplateCases() []overridableDefaultVarValueCase {
	return []overridableDefaultVarValueCase{
		{
			name:    "template without default",
			varName: overridableCaseVarFoo,
			value:   "{{.FOO}}",
			want:    false,
		},
		{
			name:    "default for different var",
			varName: overridableCaseVarFoo,
			value:   `{{.BAR | default ""}}`,
			want:    false,
		},
	}
}

func overridableDefaultRejectTypeCases() []overridableDefaultVarValueCase {
	return []overridableDefaultVarValueCase{
		{
			name:    "non-string map value",
			varName: overridableCaseVarFoo,
			value:   map[string]any{"sh": "echo hi"},
			want:    false,
		},
		{
			name:    "nil value",
			varName: overridableCaseVarFoo,
			value:   nil,
			want:    false,
		},
	}
}

func (collector *varsOverridableCollector) appendModuleViolations(module string) {
	collector.t.Helper()

	if module == taskfilesRootModule {
		return
	}

	taskfile := tasktest.LoadTaskfile(collector.t, module)

	if len(taskfile.Vars) == constZero {
		return
	}

	collector.recordVars(module, taskfile.Vars)
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
		defaultPipeMarker,
		strings.Join(formatVarsOverridableViolations(collector.violations), windowsSeparator),
	)
}

func (collector *varsOverridableCollector) recordVars(module string, vars map[string]any) {
	collector.t.Helper()

	for name := range vars {
		value := vars[name]

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

func formatVarsOverridableViolations(violations []varsOverridableViolation) []string {
	sortByModuleThenName(violations, func(violation varsOverridableViolation) moduleNamePair {
		return moduleNamePair{module: violation.module, name: violation.name}
	})

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
