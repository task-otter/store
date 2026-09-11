// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"path/filepath"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// TestIsPublicTaskNil exercises IsPublicTaskNil.
func TestIsPublicTaskNil(t *testing.T) {
	t.Parallel()

	requireSame(t, !isPublicTask(testCI, nil), true)
}

// TestIsHiddenName exercises IsHiddenName.
func TestIsHiddenName(t *testing.T) {
	t.Parallel()

	requireSame(t, isHiddenName(defaultTaskName), true)
	requireSame(t, isHiddenName(testHiddenLint), true)
	requireSame(t, !isHiddenName(testCI), true)
}

// TestPromptEmptyList exercises PromptEmptyList.
func TestPromptEmptyList(t *testing.T) {
	t.Parallel()

	requireSame(t, !hasPrompt([]any{}), true)
	requireSame(t, !hasPrompt([]string{}), true)
}

// TestSpecAsTaskCopiesRequires exercises SpecAsTaskCopiesRequires.
func TestSpecAsTaskCopiesRequires(t *testing.T) {
	t.Parallel()

	task := specAsTask(&taskSpec{Name: testInstallPkg, Requires: []string{testGOPKG}})
	requireSame(t, task.Requires != nil, true)
	requireEqual(t, task.Requires.Vars[0], testGOPKG)
}

// TestRequiredVarNamesNil exercises RequiredVarNamesNil.
func TestRequiredVarNamesNil(t *testing.T) {
	t.Parallel()

	requireSame(t, requiredVarNames(nil) == nil, true)
}

// TestCompareTaskSpecs exercises CompareTaskSpecs.
func TestCompareTaskSpecs(t *testing.T) {
	t.Parallel()

	requireSame(t, compareTaskSpecs(&taskSpec{Name: "a"}, &taskSpec{Name: "b"}) < 0, true)
}

// TestDeclaredPublicTasksFromParse exercises DeclaredPublicTasksFromParse.
func TestDeclaredPublicTasksFromParse(t *testing.T) {
	t.Parallel()

	taskfile, err := tasktest.ParseTaskfile(
		[]byte("version: '3'\ntasks:\n  ci:\n    cmds: [echo]\n"),
	)
	requireNoErr(t, err)

	tasks := declaredPublicTasks(taskfile)
	requireSame(t, len(tasks) == 1, true)
	requireEqual(t, tasks[0].Name, testCI)
}

// TestNewSkipInputUsesSpec exercises NewSkipInputUsesSpec.
func TestNewSkipInputUsesSpec(t *testing.T) {
	t.Parallel()

	input := newSkipInput(&module{Name: testGo}, &taskSpec{Name: nameFuzz})
	requireEqual(t, input.Name, nameFuzz)
	requireEqual(t, filepath.Base(input.Module), testGo)
}
