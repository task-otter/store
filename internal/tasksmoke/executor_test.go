// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

const (
	inheritedSmokeVarsTaskfile = `version: '3'
vars:
  TARGET: '{{.TARGET | default "wrong"}}'
  SPECIAL: '{{.SPECIAL | default "wrong"}}'
env:
  TARGET_VALUE: '{{.TARGET}}'
  SPECIAL_VALUE: '{{.SPECIAL}}'
tasks:
  parent:
    deps: [dependency]
    cmds:
      - task: child
  dependency:
    cmds:
      - test "$TARGET_VALUE" = src
      - printf '%s' "$SPECIAL_VALUE"
  child:
    cmds:
      - test "$TARGET_VALUE" = src
      - printf '%s' "$SPECIAL_VALUE"
`
	specialSmokeValue    = "spaces ; $(echo injected) & 'quoted' = value"
	runtimeSmokeTaskfile = `version: '3'
tasks:
  module:
    cmds:
      - test -f marker.txt
  workdir:
    dir: '{{.USER_WORKING_DIR}}'
    env:
      EXPECTED_DIR: '{{.EXPECTED_DIR}}'
    cmds:
      - test "$PWD" -ef "$EXPECTED_DIR"
  confirm:
    prompt: Confirm this action?
    cmds:
      - echo accepted
`
)

// TestExecuteTaskCLIInheritsSmokeVars checks overrides in nested tasks and dependencies.
func TestExecuteTaskCLIInheritsSmokeVars(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	requireNoErr(
		t,
		os.WriteFile(
			filepath.Join(dir, taskfileName),
			[]byte(inheritedSmokeVarsTaskfile),
			fileMode,
		),
	)

	output := new(bytes.Buffer)
	requireNoErr(t, executeTaskCLI(&runRequest{
		Home: emptyString,

		Dir: dir, Name: "parent", WorkDir: dir, Timeout: defaultTimeout,
		Output: output, Vars: map[string]string{"TARGET": "src", "SPECIAL": specialSmokeValue},
	}))
	requireEqual(t, output.String(), specialSmokeValue+specialSmokeValue)
}

// TestExecuteTaskCLIWorkingDirectories preserves module and caller directories.
func TestExecuteTaskCLIWorkingDirectories(t *testing.T) {
	t.Parallel()

	request := runtimeSmokeRequest(t, "module")
	requireNoErr(t, os.WriteFile(filepath.Join(request.Dir, "marker.txt"), nil, fileMode))
	requireNoErr(t, executeTaskCLI(request))

	request.Name = "workdir"
	request.Vars = map[string]string{"EXPECTED_DIR": request.WorkDir}

	requireNoErr(t, executeTaskCLI(request))
}

// TestExecuteTaskCLIDeniesPrompts overrides inherited automatic confirmation.
func TestExecuteTaskCLIDeniesPrompts(t *testing.T) {
	t.Setenv("TASK_ASSUME_YES", "true")

	request := runtimeSmokeRequest(t, "confirm")
	requireErr(t, executeTaskCLI(request))
	requireSame(t, bytes.Contains(request.Output.Bytes(), []byte("accepted")), false)
}

// TestExecuteTaskCLIRunsPing exercises the installed Task CLI.
func TestExecuteTaskCLIRunsPing(t *testing.T) {
	t.Parallel()

	dir := testdataAbs(t, testdataPingPath)
	err := executeTaskCLI(&runRequest{
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

// TestExecuteTaskCLIMissingTaskfile checks missing Taskfiles fail.
func TestExecuteTaskCLIMissingTaskfile(t *testing.T) {
	t.Parallel()

	err := executeTaskCLI(&runRequest{
		Output: nil,
		Vars:   nil,
		Home:   emptyString,

		Dir:     t.TempDir(),
		Name:    testPing,
		Timeout: defaultTimeout,
		WorkDir: t.TempDir(),
	})
	requireErr(t, err)
}

// TestExecuteTaskCLIUnknownTask checks unknown tasks fail.
func TestExecuteTaskCLIUnknownTask(t *testing.T) {
	t.Parallel()

	err := executeTaskCLI(testPingRequest(t, testMissingTask))
	requireErr(t, err)
}

// TestExecuteTaskCLICapturesNilOutput checks execution without report capture.
func TestExecuteTaskCLICapturesNilOutput(t *testing.T) {
	t.Parallel()

	err := executeTaskCLI(testPingRequest(t, testPing))
	requireNoErr(t, err)
}

// TestStubSourcePing exercises StubSourcePing.
func TestStubSourcePing(t *testing.T) {
	t.Parallel()

	got := stubSource(t.TempDir(), testPing)
	requireSame(t, filepath.Base(got) == testPing, true)
}

func runtimeSmokeRequest(t *testing.T, name string) *runRequest {
	t.Helper()

	dir := t.TempDir()
	requireNoErr(
		t,
		os.WriteFile(filepath.Join(dir, taskfileName), []byte(runtimeSmokeTaskfile), fileMode),
	)

	return &runRequest{
		Dir: dir, Name: name, WorkDir: t.TempDir(), Timeout: defaultTimeout,
		Output: new(bytes.Buffer), Vars: nil, Home: emptyString,
	}
}
