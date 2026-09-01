// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskintegration

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

type (
	fakeTest struct {
		fatalMessages chan string
	}
)

const (
	testModuleName        = "mod"
	publicTaskName        = "fmt:check"
	exportedName          = "ci:fix"
	variantTaskName       = "bun:ci"
	nixInclude            = "nix"
	nixTaskName           = "nix:help"
	orphanTaskName        = "orphan"
	stdoutText            = "listing"
	invalidJSON           = "not-json"
	invalidYAML           = "["
	invalidJSONWant       = "invalid JSON"
	parseWant             = "parse"
	missingListedWant     = "does not list"
	missingDeclaredWant   = "neither declares nor includes"
	missingExportedWant   = "does not expose it"
	missingVariantWant    = "does not expose variant task"
	installableVarName    = "ZIZMOR_NIX_INSTALLABLE"
	toolTaskName          = "install:tool"
	missingInstallWant    = "declares no public \"install\" task"
	unexportedInstallWant = "does not export the \"install\" task"
	nonZeroWant           = "exited non-zero"
	emptyOutputWant       = "produced no output"
	unknownSucceededWant  = "unexpectedly succeeded"
	summaryMismatchWant   = "does not name the task"
	modulePathWant        = "resolve the module folder"
	absoluteRepoRoot      = "/taskotter-repo"
	relativeModuleRoot    = "not-under-taskfiles"
	missingDefaultCase    = "missing_default"
	once                  = 1
	emptyContent          = ""
)

func (tester *fakeTest) Fatalf(format string, args ...any) {
	tester.reportFatal(fmt.Sprintf(format, args...))
}

func (*fakeTest) Helper() {}

func (*fakeTest) Skip(_ ...any) {}

func (*fakeTest) Skipf(_ string, _ ...any) {}

func (tester *fakeTest) reportFatal(message string) {
	tester.fatalMessages <- message

	runtime.Goexit()
}

// TestAssertDefaultTaskListsModuleSkipsMissingDefault skips when default is absent.
func TestAssertDefaultTaskListsModuleSkipsMissingDefault(t *testing.T) {
	t.Parallel()

	t.Run(missingDefaultCase, func(t *testing.T) {
		t.Parallel()
		assertDefaultTaskListsModule(t, moduleWithoutDefault())
	})
}

// TestIsIncludedAcceptsNamespacedName matches an include namespace prefix.
func TestIsIncludedAcceptsNamespacedName(t *testing.T) {
	t.Parallel()

	if !isIncluded([]string{nixInclude}, nixTaskName) {
		t.Fatalf("isIncluded(%s) = false", nixTaskName)
	}
}

// TestIsIncludedRejectsLocalName rejects a task that is not under an include.
func TestIsIncludedRejectsLocalName(t *testing.T) {
	t.Parallel()

	if isIncluded([]string{nixInclude}, publicTaskName) {
		t.Fatalf("isIncluded(%s) = true", publicTaskName)
	}
}

// TestLoadMetadataMissingFileReturnsEmpty treats a nested variant as empty metadata.
func TestLoadMetadataMissingFileReturnsEmpty(t *testing.T) {
	t.Parallel()

	metadata := loadMetadata(t, t.TempDir())

	if !metadataIsEmpty(metadata) {
		t.Fatalf("loadMetadata() = %+v", metadata)
	}
}

// TestModulePathFromRejectsUnrelatedPaths fatals when Rel cannot resolve the module.
func TestModulePathFromRejectsUnrelatedPaths(t *testing.T) {
	t.Parallel()

	expectFatal(t, modulePathWant, rejectUnrelatedModulePath)
}

// TestParseListedNamesRejectsInvalidJSON fatals on non-JSON list output.
func TestParseListedNamesRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	expectFatal(t, invalidJSONWant, parseInvalidListJSON)
}

// TestParseMetadataRejectsInvalidYAML fatals on broken metadata.yml contents.
func TestParseMetadataRejectsInvalidYAML(t *testing.T) {
	t.Parallel()

	expectFatal(t, parseWant, parseInvalidMetadata)
}

// TestRequireDeclaredOrIncludedAcceptsDeclared accepts a locally declared task.
func TestRequireDeclaredOrIncludedAcceptsDeclared(t *testing.T) {
	t.Parallel()

	requireDeclaredOrIncluded(t, declaredModule(), publicTaskName)
}

// TestRequireDeclaredOrIncludedRejectsUnknown fatals on a task with no source.
func TestRequireDeclaredOrIncludedRejectsUnknown(t *testing.T) {
	t.Parallel()

	expectFatal(t, missingDeclaredWant, rejectUndeclaredName)
}

// TestRequireListedRejectsMissing fatals when a declared task is not listed.
func TestRequireListedRejectsMissing(t *testing.T) {
	t.Parallel()

	expectFatal(t, missingListedWant, rejectUnlistedName)
}

// TestRequireOutputRejectsEmpty fatals when the CLI prints nothing.
func TestRequireOutputRejectsEmpty(t *testing.T) {
	t.Parallel()

	expectFatal(t, emptyOutputWant, rejectEmptyOutput)
}

// TestRequireReachableRejectsMissing fatals when an exported task is invisible.
func TestRequireReachableRejectsMissing(t *testing.T) {
	t.Parallel()

	expectFatal(t, missingExportedWant, rejectUnreachableName)
}

// TestRequireSuccessRejectsFailure fatals when the CLI exits non-zero.
func TestRequireSuccessRejectsFailure(t *testing.T) {
	t.Parallel()

	expectFatal(t, nonZeroWant, rejectFailedResult)
}

// TestRequireSummaryNamesTaskRejectsMismatch fatals when --summary omits the task.
func TestRequireSummaryNamesTaskRejectsMismatch(t *testing.T) {
	t.Parallel()

	expectFatal(t, summaryMismatchWant, rejectMismatchedSummary)
}

// TestRequireUnknownRejectedRejectsSuccess fatals when an unknown task succeeds.
func TestRequireUnknownRejectedRejectsSuccess(t *testing.T) {
	t.Parallel()

	expectFatal(t, unknownSucceededWant, rejectSuccessfulUnknown)
}

// TestIsNixBackedAcceptsInstallerModule matches a module that includes nix and
// owns an installable.
func TestIsNixBackedAcceptsInstallerModule(t *testing.T) {
	t.Parallel()

	if !isNixBacked(nixBackedModule()) {
		t.Fatal("isNixBacked() = false for a module with a nix include and an installable")
	}
}

// TestIsNixBackedRejectsModuleWithoutInstallable rejects a module that includes
// nix but installs nothing of its own, such as a family root.
func TestIsNixBackedRejectsModuleWithoutInstallable(t *testing.T) {
	t.Parallel()

	module := nixBackedModule()

	module.Vars = nil

	if isNixBacked(module) {
		t.Fatal("isNixBacked() = true for a module that owns no installable")
	}
}

// TestIsNixBackedRejectsModuleWithoutNixInclude rejects docker and the other
// modules that own an installer outside the nix profile.
func TestIsNixBackedRejectsModuleWithoutNixInclude(t *testing.T) {
	t.Parallel()

	module := nixBackedModule()

	module.Includes = nil

	if isNixBacked(module) {
		t.Fatal("isNixBacked() = true for a module without a nix include")
	}
}

// TestRequireInstallerTaskAcceptsToolVariant accepts the install:tool name pnpm
// and yarn use, where install already means project dependencies.
func TestRequireInstallerTaskAcceptsToolVariant(t *testing.T) {
	t.Parallel()

	module := nixBackedModule()

	module.Declared = []string{toolTaskName}
	module.Exported = []string{toolTaskName}

	requireInstallerTask(t, module, installTaskName)
}

// TestRequireInstallerTaskRejectsMissingTask fatals when a nix-backed module
// declares no public install task.
func TestRequireInstallerTaskRejectsMissingTask(t *testing.T) {
	t.Parallel()

	expectFatal(t, missingInstallWant, rejectMissingInstallTask)
}

// TestRequireInstallerTaskRejectsUnexportedTask fatals when install exists but
// metadata.yml does not advertise it.
func TestRequireInstallerTaskRejectsUnexportedTask(t *testing.T) {
	t.Parallel()

	expectFatal(t, unexportedInstallWant, rejectUnexportedInstallTask)
}

// TestRequireVariantTaskRejectsMissing fatals when a variant task is not listed.
func TestRequireVariantTaskRejectsMissing(t *testing.T) {
	t.Parallel()

	expectFatal(t, missingVariantWant, rejectMissingVariant)
}

func assertFatalMessage(t *testing.T, want string, fatalMessages <-chan string) {
	t.Helper()

	fatalMessage, ok := receivedFatalMessage(fatalMessages)

	if !ok {
		t.Fatal("expected fatal call")
	}

	if !strings.Contains(fatalMessage, want) {
		t.Fatalf("fatal message %q does not contain %q", fatalMessage, want)
	}
}

func declaredModule() *Module {
	module := newTestModule()

	module.Declared = []string{publicTaskName}
	module.Includes = []string{nixInclude}

	return &module
}

func emptyOutputResult() *Result {
	result := newTestResult()

	result.Args = []string{defaultTaskName}

	return &result
}

func expectFatal(t *testing.T, want string, fatalFunc func(*fakeTest)) {
	t.Helper()

	fatalMessages := make(chan string, once)
	done := make(chan struct{})

	runFatalFunc(done, fatalMessages, fatalFunc)
	assertFatalMessage(t, want, fatalMessages)
}

func failedResult() *Result {
	result := newTestResult()

	result.Args = []string{defaultTaskName}
	result.Failed = true

	return &result
}

func invalidListResult() *Result {
	result := newTestResult()

	result.Stdout = invalidJSON
	result.Args = []string{listAllFlag, jsonFlag}

	return &result
}

func metadataIsEmpty(metadata *moduleMetadata) bool {
	return metadata != nil &&
		metadata.ExportedTasks == nil &&
		metadata.Variants == nil
}

func moduleWithoutDefault() *Module {
	module := newTestModule()

	module.Listed = []string{publicTaskName}

	return &module
}

func namedModule() *Module {
	module := newTestModule()

	return &module
}

func newTestModule() Module {
	return Module{
		Name:     testModuleName,
		Dir:      emptyContent,
		Env:      nil,
		Listed:   nil,
		Declared: nil,
		Exported: nil,
		Variants: nil,
		Includes: nil,
		Vars:     nil,
		DryRun:   nil,
	}
}

func newTestResult() Result {
	return Result{
		Stdout: emptyContent,
		Stderr: emptyContent,
		Args:   nil,
		Failed: false,
	}
}

func parseInvalidListJSON(tester *fakeTest) {
	parseListedNames(tester, namedModule(), invalidListResult())
}

func parseInvalidMetadata(tester *fakeTest) {
	parseMetadata(tester, metadataName, invalidYAML)
}

func receivedFatalMessage(fatalMessages <-chan string) (message string, ok bool) {
	select {
	case message = <-fatalMessages:
		return message, true
	default:
		return emptyContent, false
	}
}

func rejectEmptyOutput(tester *fakeTest) {
	requireOutput(tester, namedModule(), emptyOutputResult())
}

func rejectFailedResult(tester *fakeTest) {
	requireSuccess(tester, namedModule(), failedResult())
}

func rejectMismatchedSummary(tester *fakeTest) {
	requireSummaryNamesTask(tester, &summaryCheck{
		module: namedModule(),
		result: summaryResult(),
		name:   publicTaskName,
	})
}

func nixBackedModule() *Module {
	module := newTestModule()

	module.Includes = []string{nixInclude}
	module.Vars = []string{installableVarName}
	module.Declared = []string{installTaskName, versionTaskName}
	module.Exported = []string{installTaskName, versionTaskName}

	return &module
}

func rejectMissingInstallTask(tester *fakeTest) {
	module := nixBackedModule()

	module.Declared = nil

	requireInstallerTask(tester, module, installTaskName)
}

func rejectUnexportedInstallTask(tester *fakeTest) {
	module := nixBackedModule()

	module.Exported = nil

	requireInstallerTask(tester, module, installTaskName)
}

func rejectMissingVariant(tester *fakeTest) {
	requireVariantTask(tester, namedModule(), variantTaskName)
}

func rejectSuccessfulUnknown(tester *fakeTest) {
	requireUnknownRejected(tester, namedModule(), succeededResult())
}

func rejectUndeclaredName(tester *fakeTest) {
	requireDeclaredOrIncluded(tester, namedModule(), orphanTaskName)
}

func rejectUnlistedName(tester *fakeTest) {
	requireListed(tester, namedModule(), publicTaskName)
}

func rejectUnreachableName(tester *fakeTest) {
	requireReachable(tester, namedModule(), exportedName)
}

func rejectUnrelatedModulePath(tester *fakeTest) {
	modulePathFrom(tester, absoluteRepoRoot, relativeModuleRoot)
}

func runFatalFunc(done chan struct{}, fatalMessages chan string, fatalFunc func(*fakeTest)) {
	go func() {
		defer close(done)

		fatalFunc(&fakeTest{fatalMessages: fatalMessages})
	}()

	<-done
}

func succeededResult() *Result {
	result := newTestResult()

	result.Stdout = stdoutText
	result.Args = []string{unknownTaskName}

	return &result
}

func summaryResult() *Result {
	result := newTestResult()

	result.Stdout = stdoutText

	return &result
}
