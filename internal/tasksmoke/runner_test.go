// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"path/filepath"
	"testing"
)

// TestEngineRunListsWithoutExecuting exercises EngineRunListsWithoutExecuting.
func TestEngineRunListsWithoutExecuting(t *testing.T) {
	t.Parallel()

	engine := testEngine(t)
	report, err := runSuite(engine, &runOptions{
		RepoRoot: testdataRepo(t),
		ListOnly: true,
		Module:   testEcho,
		Stdout:   nil,
		Stderr:   nil,
	})
	requireNoErr(t, err)
	requireSame(t, len(report.Results) > emptyLength, true)
	requireStatus(t, report.Results[0], statusPass)
}

// TestEngineRunSkipsDestructiveTasks exercises EngineRunSkipsDestructiveTasks.
func TestEngineRunSkipsDestructiveTasks(t *testing.T) {
	t.Parallel()

	engine := testEngine(t)
	report, err := runSuite(engine, &runOptions{
		RepoRoot: testdataRepo(t),
		ListOnly: true,
		Module:   testSkipme,
		Stdout:   nil,
		Stderr:   nil,
	})
	requireNoErr(t, err)
	requireSame(t, hasResultStatus(report, nameUninstall, statusSkip), true)
	requireSame(t, hasResultStatus(report, testFmtCheck, statusPass), true)
	requireSame(t, hasResultStatus(report, testGalaxyInstall, statusSkip), true)
}

// TestEngineRunRecordsExecutorFailure exercises EngineRunRecordsExecutorFailure.
func TestEngineRunRecordsExecutorFailure(t *testing.T) {
	t.Parallel()

	engine := testEngine(t)

	engine.runTask = func(*runRequest) error {
		return sentinelErr()
	}

	report, err := runSuite(engine, &runOptions{
		RepoRoot: testdataRepo(t),
		ListOnly: false,
		Module:   testEcho,
		Stdout:   nil,
		Stderr:   nil,
	})
	requireNoErr(t, err)
	requireSame(t, hasFailures(report.Results), true)
}

// TestEngineRunCopiesStubs exercises EngineRunCopiesStubs.
func TestEngineRunCopiesStubs(t *testing.T) {
	t.Parallel()

	workDir := captureEchoWorkDir(t)
	requireSame(t, pathExists(filepath.Join(workDir, testHelloFile)), true)
}

func captureEchoWorkDir(t *testing.T) string {
	t.Helper()

	engine := testEngine(t)
	dir := new(string)

	engine.runTask = bindWorkDir(dir)

	report, err := runIsolated(engine, echoModuleOpts(t))
	keepValue(report)
	requireNoErr(t, err)

	return *dir
}

func bindWorkDir(dir *string) runTaskFunc {
	return func(request *runRequest) error {
		*dir = request.WorkDir

		return nil
	}
}

func echoModuleOpts(t *testing.T) *runOptions {
	t.Helper()

	return &runOptions{Module: testEcho, RepoRoot: testdataRepo(t)}
}

// TestFilterModulesExactAndPrefix exercises FilterModulesExactAndPrefix.
func TestFilterModulesExactAndPrefix(t *testing.T) {
	t.Parallel()

	modules := []*module{
		{Name: testEslint},
		{Name: testEslintNPM},
		{Name: testYamllint},
	}
	selected := filterModules(modules, testEslint)

	requireSame(t, len(selected) == nameSplitParts, true)
	requireSame(t, len(filterModules(modules, emptyString)) == 3, true)
}

// TestFinishResultSetsFail exercises FinishResultSetsFail.
func TestFinishResultSetsFail(t *testing.T) {
	t.Parallel()

	result := finishResult(&taskResult{Status: statusPass}, sentinelErr())
	requireStatus(t, result, statusFail)
}

// TestRequestOutputNilBuffer exercises RequestOutputNilBuffer.
func TestRequestOutputNilBuffer(t *testing.T) {
	t.Parallel()

	requireEqual(t, requestOutput(&runRequest{Output: nil}), emptyString)
}

// TestSpecRequiresEmpty exercises SpecRequiresEmpty.
func TestSpecRequiresEmpty(t *testing.T) {
	t.Parallel()

	requireSame(t, specRequires(nil) == nil, true)
}

// TestPrepareWorkDirMkdirError exercises PrepareWorkDirMkdirError.
func TestPrepareWorkDirMkdirError(t *testing.T) {
	t.Parallel()

	engine := testEngine(t)

	engine.mkdirTemp = func() (string, error) {
		return emptyString, sentinelErr()
	}

	value, err := prepareWorkDir(engine, testdataRepo(t), testEcho)
	keepValue(value)
	requireErr(t, err)
}

// TestEngineRunIsolatedHomeError exercises EngineRunIsolatedHomeError.
func TestEngineRunIsolatedHomeError(t *testing.T) {
	t.Parallel()

	engine := testEngine(t)

	engine.mkdirTemp = func() (string, error) {
		return emptyString, sentinelErr()
	}

	value, err := runSuite(
		engine,
		&runOptions{RepoRoot: testdataRepo(t), ListOnly: false, Module: testEcho},
	)
	keepValue(value)
	requireErr(t, err)
}

func hasResultStatus(report *smokeReport, name, status string) bool {
	for i := range report.Results {
		if report.Results[i].Task == name && report.Results[i].Status == status {
			return true
		}
	}

	return false
}
