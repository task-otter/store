// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"bytes"
	"io"
	"io/fs"
	"time"

	"github.com/task-otter/store/internal/tasktest"
)

type (
	engine = struct {
		applyHome applyHomeFunc
		mkdirTemp mkdirTempFunc
		now       nowFunc
		runTask   runTaskFunc
		timeout   time.Duration
	}

	suiteRun = struct {
		engine  *engine
		opts    *runOptions
		home    string
		modules []*module
	}

	module = struct {
		Config *smokeConfig
		Dir    string
		Name   string
		Tasks  []*taskSpec
	}

	taskSpec = struct {
		Prompt      any
		Name        string
		Requires    []string
		Interactive bool
	}

	smokeConfig = struct {
		Vars map[string]string `yaml:"vars"`
		Skip []string          `yaml:"skip"`
	}

	taskResult = struct {
		Err      error
		Module   string
		Output   string
		Status   string
		Task     string
		Duration time.Duration
	}

	smokeReport = struct {
		Results []*taskResult
	}

	runOptions = struct {
		Stderr   io.Writer
		Stdout   io.Writer
		Module   string
		RepoRoot string
		ListOnly bool
	}

	cliFlags = struct {
		Module string
		List   bool
	}

	cliRun = struct {
		Engine *engine
		Flags  *cliFlags
		Stderr io.Writer
		Stdout io.Writer
		Root   string
	}

	startedCLI = struct {
		Err    error
		Flags  *cliFlags
		Stderr io.Writer
		Stdout io.Writer
		Root   string
	}

	skipInput = struct {
		Config *smokeConfig
		Task   *tasktest.Task
		Module string
		Name   string
	}

	runRequest = struct {
		Output  *bytes.Buffer
		Vars    map[string]string
		Dir     string
		Home    string
		Name    string
		WorkDir string
		Timeout time.Duration
	}

	moduleRun = struct {
		Opts    *runOptions
		Home    string
		Module  *module
		WorkDir string
	}

	executedResultInput = struct {
		Finished time.Time
		Request  *runRequest
		Started  time.Time
		Module   string
		Name     string
	}

	folderScan = struct {
		root    string
		folders []string
	}

	walkVisit = struct {
		Entry fs.DirEntry
		Err   error
		Scan  *folderScan
		Path  string
	}

	reportSink = interface {
		Flush() error
		io.Writer
	}

	applyHomeFunc = func(string) error
	mkdirTempFunc = func() (string, error)
	nowFunc       = func() time.Time
	runTaskFunc   = func(*runRequest) error

	envPair = [nameSplitParts]string
)
