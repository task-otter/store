// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package taskintegration exercises Taskfile modules through the real task CLI.
//
// The helpers in internal/tasktest read a Taskfile and check what it declares.
// This package runs task against the same folder and checks what it does: the
// Taskfile loads, the tasks the Taskfile declares are the tasks the CLI
// exposes, metadata.yml describes the surface the CLI actually reaches, every
// listed task renders a summary, the default task lists the module, and an
// unknown task is rejected.
//
// Every run happens in an isolated HOME so a task can neither read nor write
// the developer's shell profile, and no check installs or downloads anything.
package taskintegration

import (
	"testing"
)

type (
	// Spec describes how the shared suite exercises one Taskfile folder.
	Spec struct {
		// Module is the folder path relative to taskfiles/, such as "go"
		// or "biome/node/pnpm".
		Module string

		// DryRunTasks are public tasks that must also survive `task --dry`.
		// Leave it empty for tasks that reach the network or need a tool
		// that the test environment does not install.
		DryRunTasks []string
	}

	// suiteCheck is one named assertion of the shared integration suite.
	suiteCheck struct {
		assert func(t *testing.T, module *Module)
		name   string
	}
)

// Run executes the default integration suite against the named Taskfile folder.
func Run(t *testing.T, module string) {
	t.Helper()

	RunSpec(t, &Spec{Module: module, DryRunTasks: nil})
}

// RunHere executes the default integration suite against the Taskfile folder
// that holds the calling test, so a module test never repeats its own path.
func RunHere(t *testing.T) {
	t.Helper()

	Run(t, currentModule(t))
}

// RunSpec executes the integration suite described by spec.
func RunSpec(t *testing.T, spec *Spec) {
	t.Helper()

	module := newModule(t, spec)
	checks := suiteChecks()

	for i := range checks {
		runSuiteCheck(t, module, &checks[i])
	}
}

func runSuiteCheck(t *testing.T, module *Module, check *suiteCheck) {
	t.Helper()

	t.Run(check.name, func(t *testing.T) {
		t.Parallel()
		check.assert(t, module)
	})
}

func suiteChecks() []suiteCheck {
	return []suiteCheck{
		{name: "declared_tasks_are_listed", assert: assertDeclaredTasksAreListed},
		{name: "metadata_tasks_are_reachable", assert: assertMetadataTasksAreReachable},
		{name: "metadata_variants_expose_tasks", assert: assertVariantsExposeTasks},
		{name: "task_summaries_render", assert: assertTaskSummariesRender},
		{name: "default_task_lists_module", assert: assertDefaultTaskListsModule},
		{name: "unknown_task_is_rejected", assert: assertUnknownTaskIsRejected},
		{name: "dry_run_tasks_succeed", assert: assertDryRunTasksSucceed},
	}
}
