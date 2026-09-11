// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"path/filepath"
	"testing"
)

// TestExecuteGoTaskRunsPing exercises ExecuteGoTaskRunsPing.
func TestExecuteGoTaskRunsPing(t *testing.T) {
	t.Parallel()

	dir := testdataAbs(t, testdataPingPath)
	err := executeGoTask(&runRequest{
		Dir:     dir,
		Name:    testPing,
		Timeout: defaultTimeout,
		WorkDir: t.TempDir(),
		Output:  nil,
		Home:    t.TempDir(),
		Vars:    map[string]string{testSample: "1"},
	})
	requireNoErr(t, err)
}

// TestExecuteGoTaskSetupError exercises ExecuteGoTaskSetupError.
func TestExecuteGoTaskSetupError(t *testing.T) {
	t.Parallel()

	err := executeGoTask(&runRequest{
		Dir:     t.TempDir(),
		Name:    testPing,
		Timeout: defaultTimeout,
		WorkDir: t.TempDir(),
	})
	requireErr(t, err)
}

// TestExecuteGoTaskUnknownTask exercises ExecuteGoTaskUnknownTask.
func TestExecuteGoTaskUnknownTask(t *testing.T) {
	t.Parallel()

	err := executeGoTask(&runRequest{
		Dir:     testdataAbs(t, testdataPingPath),
		Name:    testMissingTask,
		Timeout: defaultTimeout,
		WorkDir: t.TempDir(),
	})
	requireErr(t, err)
}

// TestExecuteGoTaskCapturesNilOutput exercises ExecuteGoTaskCapturesNilOutput.
func TestExecuteGoTaskCapturesNilOutput(t *testing.T) {
	t.Parallel()

	err := executeGoTask(&runRequest{
		Dir:     testdataAbs(t, testdataPingPath),
		Name:    testPing,
		Timeout: defaultTimeout,
		WorkDir: t.TempDir(),
	})
	requireNoErr(t, err)
}

// TestAssignCallVarsEmpty exercises AssignCallVarsEmpty.
func TestAssignCallVarsEmpty(t *testing.T) {
	t.Parallel()

	call := newTaskCall(testPing, nil)
	requireSame(t, call.Vars == nil, true)
}

// TestNewTaskCallSetsVars exercises NewTaskCallSetsVars.
func TestNewTaskCallSetsVars(t *testing.T) {
	t.Parallel()

	call := newTaskCall(testPing, map[string]string{testFOO: testBar})
	value, found := call.Vars.Get(testFOO)
	requireSame(t, found, true)

	text, isString := value.Value.(string)
	requireSame(t, isString, true)
	requireEqual(t, text, testBar)
}

// TestStubSourcePing exercises StubSourcePing.
func TestStubSourcePing(t *testing.T) {
	t.Parallel()

	got := stubSource(t.TempDir(), testPing)
	requireSame(t, filepath.Base(got) == testPing, true)
}
