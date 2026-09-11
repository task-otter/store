// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"slices"
)

func appendRunnable(tasks []*taskSpec, module *module, spec *taskSpec) []*taskSpec {
	if skipReason(newSkipInput(module, spec)) != emptyString {
		return tasks
	}

	return append(tasks, spec)
}

func appendSelected(tasks []*taskSpec, task *taskSpec) []*taskSpec {
	if task == nil {
		return tasks
	}

	return append(tasks, task)
}

func excludedWorkNames() []string {
	return []string{
		nameCacheClean, nameConfigInit, nameHelp, nameInstall, nameInstallTool,
		nameInstallUndo, nameUninstall, nameUpgrade, nameVersion, nameVersionTool,
		nameWhich,
	}
}

func fallbackWorkTask(tasks []*taskSpec) *taskSpec {
	for i := range tasks {
		if !isExcludedWorkName(tasks[i].Name) {
			return tasks[i]
		}
	}

	return nil
}

func findNamedTask(tasks []*taskSpec, name string) *taskSpec {
	for i := range tasks {
		if tasks[i].Name == name {
			return tasks[i]
		}
	}

	return nil
}

func firstNamedTask(tasks []*taskSpec, names []string) *taskSpec {
	for i := range names {
		found := findNamedTask(tasks, names[i])

		if found != nil {
			return found
		}
	}

	return nil
}

func isExcludedWorkName(name string) bool {
	return slices.Contains(excludedWorkNames(), name)
}

func primarySmokeNames() []string {
	return []string{nameCI, nameVerify, nameLint}
}

func primarySmokeTask(tasks []*taskSpec) *taskSpec {
	found := firstNamedTask(tasks, primarySmokeNames())

	if found != nil {
		return found
	}

	return fallbackWorkTask(tasks)
}

func runnableTasks(module *module) []*taskSpec {
	tasks := make([]*taskSpec, emptyLength, len(module.Tasks))

	for i := range module.Tasks {
		tasks = appendRunnable(tasks, module, module.Tasks[i])
	}

	return tasks
}

func selectedSmokeTasks(module *module) []*taskSpec {
	runnable := sortedRunnableTasks(module)
	selected := make([]*taskSpec, emptyLength, nameSplitParts)

	selected = appendSelected(selected, versionSmokeTask(runnable))

	return appendSelected(selected, primarySmokeTask(runnable))
}

func sortedRunnableTasks(module *module) []*taskSpec {
	tasks := runnableTasks(module)

	slices.SortFunc(tasks, compareTaskSpecs)

	return tasks
}

func versionSmokeNames() []string {
	return []string{nameVersion, nameVersionTool}
}

func versionSmokeTask(tasks []*taskSpec) *taskSpec {
	return firstNamedTask(tasks, versionSmokeNames())
}
