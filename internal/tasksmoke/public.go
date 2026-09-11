// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"slices"
	"strings"

	"github.com/task-otter/store/internal/tasktest"
)

func declaredPublicTasks(taskfile *tasktest.Taskfile) []*taskSpec {
	tasks := make([]*taskSpec, emptyLength, len(taskfile.Tasks))

	for name := range taskfile.Tasks {
		tasks = appendPublicTask(tasks, name, taskfile.Tasks[name])
	}

	slices.SortFunc(tasks, compareTaskSpecs)

	return tasks
}

func appendPublicTask(tasks []*taskSpec, name string, task *tasktest.Task) []*taskSpec {
	if !isPublicTask(name, task) {
		return tasks
	}

	return append(tasks, newTaskSpec(name, task))
}

func compareTaskSpecs(left, right *taskSpec) int {
	return strings.Compare(left.Name, right.Name)
}

func isHiddenName(name string) bool {
	if name == defaultTaskName {
		return true
	}

	return strings.HasPrefix(name, privatePrefix)
}

func isPublicTask(name string, task *tasktest.Task) bool {
	if task == nil || isHiddenName(name) {
		return false
	}

	return !task.Internal
}

func newTaskSpec(name string, task *tasktest.Task) *taskSpec {
	return &taskSpec{
		Name:        name,
		Prompt:      task.Prompt,
		Requires:    requiredVarNames(task.Requires),
		Interactive: task.Interactive,
	}
}

func requiredVarNames(requires *tasktest.TaskRequires) []string {
	if requires == nil {
		return nil
	}

	return requires.Vars
}
