// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/task-otter/store/internal/taskcli"
)

func configureSmokeOutput(invocation *taskcli.Request, request *runRequest) {
	var writer io.Writer = os.Stdout

	if request.Output != nil {
		writer = io.MultiWriter(os.Stdout, request.Output)
	}

	invocation.Stdout = writer
	invocation.Stderr = writer
}

func executeTaskCLI(request *runRequest) error {
	invocation := smokeInvocation(request)

	runErr := taskcli.Run(context.Background(), invocation)
	if runErr != nil {
		return fmt.Errorf(errRunFormat, request.Dir, request.Name, runErr)
	}

	return nil
}

func smokeInvocation(request *runRequest) *taskcli.Request {
	invocation := &taskcli.Request{
		Args: nil, Stdout: nil, Stderr: nil,
		Dir: request.WorkDir, Timeout: request.Timeout,
		Env: append(os.Environ(), "TASK_ASSUME_YES=false"),
	}

	invocation.Args = smokeTaskArgs(request)
	configureSmokeOutput(invocation, request)

	return invocation
}

func smokeTaskArgs(request *runRequest) []string {
	return append([]string{
		"--taskfile", filepath.Join(request.Dir, taskfileName),
		"--silent", "--color=false", "--yes=false", request.Name,
	}, taskcli.VarsArgs(request.Vars)...)
}
