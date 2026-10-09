// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func testInteractiveTask() *tasktest.Task {
	task := new(tasktest.Task)

	task.Interactive = true

	return task
}

func testNamedModule(name string) *module {
	return &module{Config: nil, Dir: emptyString, Name: name, Tasks: nil}
}

func testNamedResult(module, name, status string) *taskResult {
	return &taskResult{
		Err:      nil,
		Module:   module,
		Output:   emptyString,
		Status:   status,
		Task:     name,
		Duration: emptyLength,
	}
}

func testNamedSkipInput(module, name string) *skipInput {
	return &skipInput{Config: nil, Task: nil, Module: module, Name: name}
}

func testPromptTask(prompt any) *tasktest.Task {
	task := new(tasktest.Task)

	task.Prompt = prompt

	return task
}

func testRequiredTask(requires *tasktest.TaskRequires) *tasktest.Task {
	task := new(tasktest.Task)

	task.Requires = requires

	return task
}

func testRunOptions(root, module string, list bool) *runOptions {
	return &runOptions{Stderr: nil, Stdout: nil, Module: module, RepoRoot: root, ListOnly: list}
}

func testNamedSpec(name string) *taskSpec {
	return &taskSpec{Prompt: nil, Name: name, Requires: nil, Interactive: false}
}

func testPingRequest(t *testing.T, name string) *runRequest {
	t.Helper()

	return &runRequest{
		Output: nil, Vars: nil, Dir: testdataAbs(t, testdataPingPath), Home: emptyString,
		Name: name, WorkDir: t.TempDir(), Timeout: defaultTimeout,
	}
}

func testReportRows() []*taskResult {
	return []*taskResult{testNamedResult(testEcho, testPing, statusPass)}
}
