// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskintegration

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/task-otter/store/internal/tasktestutil"
)

type (
	// Result is the outcome of a single task CLI invocation.
	Result struct {
		// Stdout is what the task CLI wrote to standard output.
		Stdout string

		// Stderr is what the task CLI wrote to standard error.
		Stderr string

		// Args are the arguments the task CLI was invoked with.
		Args []string

		// Failed reports whether the task CLI exited with a non-zero status.
		Failed bool
	}

	taskListing struct {
		Tasks []listedTask `json:"tasks"`
	}

	listedTask struct {
		Name string `json:"name"`
	}
)

const (
	taskfileFlag   = "--taskfile"
	taskfileName   = "Taskfile.yml"
	listAllFlag    = "--list-all"
	jsonFlag       = "--json"
	summaryFlag    = "--summary"
	dryRunFlag     = "--dry"
	argSep         = " "
	lineSep        = "\n"
	prefixArgCount = 2
	taskTimeout    = 2 * time.Minute
)

// Combined returns the task CLI output with standard error appended to standard output.
func (result *Result) Combined() string { return result.Stdout + lineSep + result.Stderr }

// runTask invokes the task CLI against the folder's own Taskfile and returns its result.
func runTask(t *testing.T, module *Module, args ...string) Result {
	t.Helper()

	full := taskArgs(args)
	run := tasktestutil.TaskRun{Root: module.Dir, Env: module.Env, Args: full}
	result := tasktestutil.RunTaskTimeout(t, run, taskTimeout)

	return Result{
		Stdout: result.Stdout,
		Stderr: result.Stderr,
		Args:   full,
		Failed: result.Err != nil,
	}
}

func taskArgs(args []string) []string {
	full := make([]string, zeroLength, len(args)+prefixArgCount)

	full = append(full, taskfileFlag, taskfileName)

	return append(full, args...)
}

func listTaskNames(t *testing.T, module *Module) []string {
	t.Helper()

	result := runTask(t, module, listAllFlag, jsonFlag)
	requireSuccess(t, module, &result)

	return parseListedNames(t, module, &result)
}

func parseListedNames(t suiteT, module *Module, result *Result) []string {
	t.Helper()

	listing := new(taskListing)

	err := json.Unmarshal([]byte(result.Stdout), listing)
	if err != nil {
		t.Fatalf(
			"%s: %s produced invalid JSON: %v\n%s",
			module.Name,
			strings.Join(result.Args, argSep),
			err,
			result.Combined(),
		)
	}

	return listedNames(listing)
}

func listedNames(listing *taskListing) []string {
	names := make([]string, zeroLength, len(listing.Tasks))

	for i := range listing.Tasks {
		names = append(names, listing.Tasks[i].Name)
	}

	slices.Sort(names)

	return names
}

func requireSuccess(t suiteT, module *Module, result *Result) {
	t.Helper()

	if !result.Failed {
		return
	}

	t.Fatalf(
		"%s: %s exited non-zero\n%s",
		module.Name,
		strings.Join(result.Args, argSep),
		result.Combined(),
	)
}
