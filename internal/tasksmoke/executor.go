// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/go-task/task/v3"
	"github.com/go-task/task/v3/taskfile/ast"
)

func assignCallVars(call *task.Call, vars map[string]string) {
	if len(vars) == emptyLength {
		return
	}

	call.Vars = ast.NewVars()
	setCallVars(call.Vars, vars)
}

func executeGoTask(request *runRequest) error {
	executor := newConfiguredExecutor(request)

	err := executor.Setup()
	if err != nil {
		return fmt.Errorf(errRunFormat, request.Dir, request.Name, err)
	}

	runErr := runExecutorTask(executor, request)
	if runErr != nil {
		return fmt.Errorf("run executor task: %w", runErr)
	}

	return nil
}

func executorOptions(request *runRequest) []task.ExecutorOption {
	var writer io.Writer = os.Stdout

	if request.Output != nil {
		writer = io.MultiWriter(os.Stdout, request.Output)
	}

	return []task.ExecutorOption{
		task.WithAssumeYes(false),
		task.WithColor(false),
		task.WithDir(request.Dir),
		task.WithEntrypoint(filepath.Join(request.Dir, taskfileName)),
		task.WithSilent(true),
		task.WithStderr(writer),
		task.WithStdout(writer),
		task.WithTimeout(request.Timeout),
	}
}

func newConfiguredExecutor(request *runRequest) *task.Executor {
	executor := task.NewExecutor(executorOptions(request)...)

	executor.UserWorkingDir = request.WorkDir

	return executor
}

func newTaskCall(name string, vars map[string]string) *task.Call {
	call := &task.Call{Task: name, Vars: nil, Silent: false, Indirect: false}

	assignCallVars(call, vars)

	return call
}

func runExecutorTask(executor *task.Executor, request *runRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), request.Timeout)

	defer cancel()

	err := executor.Run(ctx, newTaskCall(request.Name, request.Vars))
	if err != nil {
		return fmt.Errorf(errRunFormat, request.Dir, request.Name, err)
	}

	return nil
}

func setCallVars(vars *ast.Vars, values map[string]string) {
	keys := make([]string, emptyLength, len(values))

	for key := range values {
		keys = append(keys, key)
	}

	slices.Sort(keys)
	setSortedCallVars(vars, values, keys)
}

func setSortedCallVars(vars *ast.Vars, values map[string]string, keys []string) {
	for i := range keys {
		key := keys[i]
		vars.Set(key, ast.Var{Value: values[key]})
	}
}
