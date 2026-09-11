// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"testing"
)

// TestSelectedSmokeTasksEmpty exercises SelectedSmokeTasksEmpty.
func TestSelectedSmokeTasksEmpty(t *testing.T) {
	t.Parallel()

	requireSelectedNames(t, namedSmokeModule())
}

// TestSelectedSmokeTasksFallsBackToFmtCheck exercises SelectedSmokeTasksFallsBackToFmtCheck.
func TestSelectedSmokeTasksFallsBackToFmtCheck(t *testing.T) {
	t.Parallel()

	requireSelectedNames(
		t,
		namedSmokeModule(testTypecheck, nameVersion, nameFmt, testFmtCheck, nameInstall),
		nameVersion,
		testFmtCheck,
	)
}

// TestSelectedSmokeTasksPrefersCI exercises SelectedSmokeTasksPrefersCI.
func TestSelectedSmokeTasksPrefersCI(t *testing.T) {
	t.Parallel()

	requireSelectedNames(
		t,
		namedSmokeModule(nameInstall, nameLint, nameVerify, nameVersion, nameCI),
		nameVersion,
		nameCI,
	)
}

// TestSelectedSmokeTasksPrefersVerify exercises SelectedSmokeTasksPrefersVerify.
func TestSelectedSmokeTasksPrefersVerify(t *testing.T) {
	t.Parallel()

	requireSelectedNames(
		t,
		namedSmokeModule(nameLint, nameVersion, nameVerify),
		nameVersion,
		nameVerify,
	)
}

// TestSelectedSmokeTasksSkippedHeuristicsUseWorkTask exercises SelectedSmokeTasksSkippedHeuristicsUseWorkTask.
func TestSelectedSmokeTasksSkippedHeuristicsUseWorkTask(t *testing.T) {
	t.Parallel()

	requireSelectedNames(
		t,
		skippedSmokeModule(
			[]string{nameCI, nameVerify, nameLint},
			nameVersion,
			nameCI,
			nameVerify,
			nameLint,
			testFmtCheck,
		),
		nameVersion,
		testFmtCheck,
	)
}

// TestSelectedSmokeTasksSkippedPrimaryFallsThrough exercises SelectedSmokeTasksSkippedPrimaryFallsThrough.
func TestSelectedSmokeTasksSkippedPrimaryFallsThrough(t *testing.T) {
	t.Parallel()

	requireSelectedNames(
		t,
		skippedSmokeModule([]string{nameCI}, nameVersion, nameCI, nameVerify, testFmtCheck),
		nameVersion,
		nameVerify,
	)
}

// TestSelectedSmokeTasksUsesVersionTool exercises SelectedSmokeTasksUsesVersionTool.
func TestSelectedSmokeTasksUsesVersionTool(t *testing.T) {
	t.Parallel()

	requireSelectedNames(
		t,
		namedSmokeModule(nameInstallTool, nameVersionTool, testFmtCheck),
		nameVersionTool,
		testFmtCheck,
	)
}

// TestSelectedSmokeTasksVersionOnly exercises SelectedSmokeTasksVersionOnly.
func TestSelectedSmokeTasksVersionOnly(t *testing.T) {
	t.Parallel()

	requireSelectedNames(
		t,
		namedSmokeModule(nameInstall, nameInstallTool, nameUninstall, nameUpgrade, nameVersion),
		nameVersion,
	)
}

func namedSmokeModule(names ...string) *module {
	return &module{Name: testGo, Tasks: namedTaskSpecs(names)}
}

func namedTaskSpecs(names []string) []*taskSpec {
	tasks := make([]*taskSpec, emptyLength, len(names))

	for i := range names {
		tasks = append(tasks, &taskSpec{Name: names[i]})
	}

	return tasks
}

func requireSelectedNames(t *testing.T, module *module, names ...string) {
	t.Helper()

	tasks := selectedSmokeTasks(module)

	requireSame(t, len(tasks), len(names))

	for i := range names {
		requireEqual(t, tasks[i].Name, names[i])
	}
}

func skippedSmokeModule(skip []string, names ...string) *module {
	module := namedSmokeModule(names...)

	module.Config = &smokeConfig{Skip: skip}

	return module
}
