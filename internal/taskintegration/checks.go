// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskintegration

import (
	"slices"
	"strings"
	"testing"
)

type (
	// taskSweep applies one assertion to every task name in a list, or skips
	// the check when the module has no such tasks.
	taskSweep = struct {
		assert func(t *testing.T, module *Module, name string)
		module *Module
		reason string
		names  []string
	}

	summaryCheck = struct {
		module *Module
		result *Result
		name   string
	}
)

const (
	unknownTaskName = "taskotter-integration-missing-task"

	// installTaskName and versionTaskName are the public surface ADR 0004
	// requires of every Nix-backed module.
	installTaskName = "install"
	versionTaskName = "version"

	// toolSuffix names the variant pnpm and yarn use, where install and version
	// already mean project dependencies rather than the CLI itself.
	toolSuffix = ":tool"

	// nixNamespace is the include a Nix-backed module carries, and
	// installableSuffix names the installable it owns.
	nixNamespace      = "nix"
	installableSuffix = "_NIX_INSTALLABLE"
)

// assertDeclaredTasksAreListed proves the task CLI and the Taskfile agree in
// both directions: nothing the Taskfile publishes is invisible to the CLI, and
// nothing the CLI lists comes from outside the Taskfile or its includes.
func assertDeclaredTasksAreListed(t *testing.T, module *Module) {
	t.Helper()

	for i := range module.Declared {
		requireListed(t, module, module.Declared[i])
	}

	for i := range module.Listed {
		requireDeclaredOrIncluded(t, module, module.Listed[i])
	}
}

// assertMetadataTasksAreReachable proves every task metadata.yml advertises is
// actually reachable through the CLI, directly or under an include namespace.
func assertMetadataTasksAreReachable(t *testing.T, module *Module) {
	t.Helper()

	sweepTasks(t, &taskSweep{
		assert: requireReachableTask,
		module: module,
		names:  module.Exported,
		reason: module.Name + " has no " + metadataName,
	})
}

// assertVariantsExposeTasks proves a family root wires every variant it
// advertises, and that each variant namespace carries the full exported surface.
func assertVariantsExposeTasks(t *testing.T, module *Module) {
	t.Helper()

	sweepTasks(t, &taskSweep{
		assert: assertVariantExposesTasks,
		module: module,
		names:  module.Variants,
		reason: module.Name + " declares no variants",
	})
}

// assertTaskSummariesRender runs `task --summary` for every task the folder
// declares, which forces the CLI to resolve the task and render the
// documentation it carries. Tasks reached through an include belong to the
// integration test of the folder that declares them.
func assertTaskSummariesRender(t *testing.T, module *Module) {
	t.Helper()

	sweepTasks(t, &taskSweep{
		assert: assertSummaryRenders,
		module: module,
		names:  module.Declared,
		reason: module.Name + " declares no tasks of its own",
	})
}

// assertNixModulesExposeInstall proves a module that installs its tool from the
// Nix profile publishes the surface ADR 0004 requires: a public install task and
// a public version task, both advertised in metadata.yml. Modules that own no
// installable — the nix module itself, docker, and the family roots that only
// dispatch to variants — are skipped.
func assertNixModulesExposeInstall(t *testing.T, module *Module) {
	t.Helper()

	if !isNixBacked(module) {
		t.Skipf("%s installs no tool from the nix profile", module.Name)
	}

	requireInstallerTask(t, module, installTaskName)
	requireInstallerTask(t, module, versionTaskName)
}

// isNixBacked reports whether the module includes nix and owns an installable,
// which together mark it as installing its tool through nix:install:profile.
func isNixBacked(module *Module) bool {
	return slices.Contains(module.Includes, nixNamespace) && ownsInstallable(module.Vars)
}

func ownsInstallable(vars []string) bool {
	for i := range vars {
		if strings.HasSuffix(vars[i], installableSuffix) {
			return true
		}
	}

	return false
}

// requireInstallerTask proves the module both declares and exports name, or the
// ":tool" variant of it that pnpm and yarn use.
func requireInstallerTask(t suiteT, module *Module, name string) {
	t.Helper()

	requireInstallerDeclared(t, module, name)
	requireInstallerExported(t, module, name)
}

func requireInstallerDeclared(t suiteT, module *Module, name string) {
	t.Helper()

	if declaresEitherName(module.Declared, name) {
		return
	}

	t.Fatalf(
		"%s: %s installs from the nix profile but declares no public %q task\ndeclared: %v",
		module.Name,
		taskfileName,
		name,
		module.Declared,
	)
}

func requireInstallerExported(t suiteT, module *Module, name string) {
	t.Helper()

	if declaresEitherName(module.Exported, name) {
		return
	}

	t.Fatalf(
		"%s: %s does not export the %q task\nexported: %v",
		module.Name,
		metadataName,
		name,
		module.Exported,
	)
}

func declaresEitherName(names []string, name string) bool {
	return slices.Contains(names, name) || slices.Contains(names, name+toolSuffix)
}

// assertDryRunTasksSucceed runs the tasks a module opted into under `task --dry`,
// which resolves variables, preconditions and status checks without side effects.
func assertDryRunTasksSucceed(t *testing.T, module *Module) {
	t.Helper()

	sweepTasks(t, &taskSweep{
		assert: assertDryRunSucceeds,
		module: module,
		names:  module.DryRun,
		reason: module.Name + " declares no dry-run tasks",
	})
}

// assertDefaultTaskListsModule runs the module's default task, the one task
// every module can run without installing anything.
func assertDefaultTaskListsModule(t *testing.T, module *Module) {
	t.Helper()

	if !slices.Contains(module.Listed, defaultTaskName) {
		t.Skipf("%s declares no %s task", module.Name, defaultTaskName)
	}

	result := runTask(t, module, defaultTaskName)

	requireSuccess(t, module, &result)
	requireOutput(t, module, &result)
}

// assertUnknownTaskIsRejected proves the module fails loudly on a task it does
// not define, rather than succeeding silently.
func assertUnknownTaskIsRejected(t *testing.T, module *Module) {
	t.Helper()

	result := runTask(t, module, unknownTaskName)

	requireUnknownRejected(t, module, &result)
}

func requireUnknownRejected(t suiteT, module *Module, result *Result) {
	t.Helper()

	if result.Failed {
		return
	}

	t.Fatalf(
		"%s: task %s unexpectedly succeeded:\n%s",
		module.Name,
		unknownTaskName,
		result.Combined(),
	)
}

func sweepTasks(t *testing.T, sweep *taskSweep) {
	t.Helper()

	if len(sweep.names) == zeroLength {
		t.Skip(sweep.reason)
	}

	for i := range sweep.names {
		sweep.assert(t, sweep.module, sweep.names[i])
	}
}

func requireListed(t suiteT, module *Module, name string) {
	t.Helper()

	if slices.Contains(module.Listed, name) {
		return
	}

	t.Fatalf(
		"%s: %s declares public task %q but the task CLI does not list it\nlisted: %v",
		module.Name,
		taskfileName,
		name,
		module.Listed,
	)
}

func requireDeclaredOrIncluded(t suiteT, module *Module, name string) {
	t.Helper()

	skippable := name == defaultTaskName ||
		strings.HasPrefix(name, privatePrefix) ||
		isIncluded(module.Includes, name)

	if skippable || slices.Contains(module.Declared, name) {
		return
	}

	t.Fatalf(
		"%s: the task CLI lists %q but %s neither declares nor includes it",
		module.Name,
		name,
		taskfileName,
	)
}

func isIncluded(includes []string, name string) bool {
	for i := range includes {
		if strings.HasPrefix(name, includes[i]+namespaceSep) {
			return true
		}
	}

	return false
}

func requireReachableTask(t *testing.T, module *Module, exported string) {
	t.Helper()

	requireReachable(t, module, exported)
}

func requireReachable(t suiteT, module *Module, exported string) {
	t.Helper()

	if isReachable(module.Listed, exported) {
		return
	}

	t.Fatalf(
		"%s: %s exports %q but the task CLI does not expose it\nlisted: %v",
		module.Name,
		metadataName,
		exported,
		module.Listed,
	)
}

func isReachable(listed []string, exported string) bool {
	suffix := namespaceSep + exported

	for i := range listed {
		if listed[i] == exported || strings.HasSuffix(listed[i], suffix) {
			return true
		}
	}

	return false
}

func assertVariantExposesTasks(t *testing.T, module *Module, variant string) {
	t.Helper()

	namespace := strings.ReplaceAll(variant, pathSep, namespaceSep)

	for i := range module.Exported {
		requireVariantTask(t, module, namespace+namespaceSep+module.Exported[i])
	}
}

func requireVariantTask(t suiteT, module *Module, name string) {
	t.Helper()

	if slices.Contains(module.Listed, name) {
		return
	}

	t.Fatalf("%s: the task CLI does not expose variant task %q", module.Name, name)
}

func assertSummaryRenders(t *testing.T, module *Module, name string) {
	t.Helper()

	result := runTask(t, module, summaryFlag, name)

	requireSuccess(t, module, &result)
	requireSummaryNamesTask(t, &summaryCheck{module: module, result: &result, name: name})
}

func requireSummaryNamesTask(t suiteT, check *summaryCheck) {
	t.Helper()

	if strings.Contains(check.result.Stdout, check.name) {
		return
	}

	t.Fatalf(
		"%s: task %s %s does not name the task:\n%s",
		check.module.Name,
		summaryFlag,
		check.name,
		check.result.Stdout,
	)
}

func assertDryRunSucceeds(t *testing.T, module *Module, name string) {
	t.Helper()

	result := runTask(t, module, dryRunFlag, name)

	requireSuccess(t, module, &result)
}

func requireOutput(t suiteT, module *Module, result *Result) {
	t.Helper()

	if strings.TrimSpace(result.Combined()) != emptyString {
		return
	}

	t.Fatalf("%s: %s produced no output", module.Name, strings.Join(result.Args, argSep))
}
