// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasksmoke

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/task-otter/store/internal/tasktest"
)

func appendMatching(selected []*module, module *module, filter string) []*module {
	if !moduleMatches(module.Name, filter) {
		return selected
	}

	return append(selected, module)
}

func attachConfigs(repoRoot string, modules []*module) error {
	for i := range modules {
		err := attachModuleConfig(repoRoot, modules[i])
		if err != nil {
			return fmt.Errorf("attach module config: %w", err)
		}
	}

	return nil
}

func attachModuleConfig(repoRoot string, module *module) error {
	config, err := moduleSmokeConfig(repoRoot, module.Name)
	if err != nil {
		return fmt.Errorf("load module smoke config: %w", err)
	}

	module.Config = config

	return nil
}

func collectResults(suite *suiteRun) (*smokeReport, error) {
	var results []*taskResult

	for i := range suite.modules {
		moduleResults, err := runModule(suite, suite.modules[i])
		if err != nil {
			return nil, fmt.Errorf("run module: %w", err)
		}

		results = append(results, moduleResults...)
	}

	return &smokeReport{Results: results}, nil
}

func executeOne(runner *engine, run *moduleRun, spec *taskSpec) *taskResult {
	started := runner.now()
	request := newRequest(runner, run, spec)
	err := runner.runTask(request)

	return finishResult(newExecutedResult(&executedResultInput{
		Finished: runner.now(),
		Module:   run.Module.Name,
		Name:     spec.Name,
		Request:  request,
		Started:  started,
	}), err)
}

func filterModules(modules []*module, filter string) []*module {
	if filter == emptyString {
		return modules
	}

	return matchingModules(modules, filter)
}

func finishResult(result *taskResult, err error) *taskResult {
	if err == nil {
		return result
	}

	result.Err = err
	result.Status = statusFail

	return result
}

func listModuleRun(opts *runOptions, home string, module *module) *moduleRun {
	return &moduleRun{Home: home, Module: module, Opts: opts, WorkDir: emptyString}
}

func matchingModules(modules []*module, filter string) []*module {
	var selected []*module

	for i := range modules {
		selected = appendMatching(selected, modules[i], filter)
	}

	return selected
}

func moduleMatches(name, filter string) bool {
	if name == filter {
		return true
	}

	return strings.HasPrefix(name, filter+pathSeparator)
}

func newExecutedResult(input *executedResultInput) *taskResult {
	return &taskResult{
		Duration: input.Finished.Sub(input.Started),
		Err:      nil,
		Module:   input.Module,
		Output:   requestOutput(input.Request),
		Status:   statusPass,
		Task:     input.Name,
	}
}

func newRequest(runner *engine, run *moduleRun, spec *taskSpec) *runRequest {
	return &runRequest{
		Dir:     run.Module.Dir,
		Home:    run.Home,
		Name:    spec.Name,
		Output:  new(bytes.Buffer),
		Timeout: runner.timeout,
		Vars:    smokeVars(run.Module.Config),
		WorkDir: run.WorkDir,
	}
}

func newSkipInput(module *module, spec *taskSpec) *skipInput {
	return &skipInput{
		Config: module.Config,
		Module: module.Name,
		Name:   spec.Name,
		Task:   specAsTask(spec),
	}
}

func bindSuite(runner *engine, opts *runOptions, home string) *suiteRun {
	return &suiteRun{engine: runner, home: home, modules: nil, opts: opts}
}

func plannedResult(module, name string) *taskResult {
	return &taskResult{
		Duration: emptyLength,
		Err:      nil,
		Module:   module,
		Output:   emptyString,
		Status:   statusPass,
		Task:     name,
	}
}

func prepareWorkDir(runner *engine, repoRoot, module string) (string, error) {
	workDir, err := runner.mkdirTemp()
	if err != nil {
		return emptyString, fmt.Errorf("create stub work dir: %w", err)
	}

	copyErr := copyStubTree(stubSource(repoRoot, module), workDir)
	if copyErr != nil {
		return emptyString, fmt.Errorf("copy smoke stubs: %w", copyErr)
	}

	return workDir, nil
}

func requestOutput(request *runRequest) string {
	if request.Output == nil {
		return emptyString
	}

	return request.Output.String()
}

func runSuite(runner *engine, opts *runOptions) (*smokeReport, error) {
	report, err := selectedRunner(opts)(runner, opts)
	if err != nil {
		return nil, fmt.Errorf(errWrapFormat, errSmokeEngine, err)
	}

	return report, nil
}

func runCopiedModule(suite *suiteRun, module *module) ([]*taskResult, error) {
	workDir, err := prepareWorkDir(suite.engine, suite.opts.RepoRoot, module.Name)
	if err != nil {
		return nil, fmt.Errorf("prepare module workdir: %w", err)
	}

	return runModuleTasks(suite.engine, &moduleRun{
		Home:    suite.home,
		Module:  module,
		Opts:    suite.opts,
		WorkDir: workDir,
	}), nil
}

func runFiltered(suite *suiteRun) (*smokeReport, error) {
	suite.modules = filterModules(suite.modules, suite.opts.Module)

	report, err := runSelected(suite)
	if err != nil {
		return nil, fmt.Errorf("run selected modules: %w", err)
	}

	return report, nil
}

func runIsolated(runner *engine, opts *runOptions) (*smokeReport, error) {
	home, err := isolatedHomeDir(runner)
	if err != nil {
		return nil, fmt.Errorf("isolate smoke home: %w", err)
	}

	report, runErr := runWithHome(runner, opts, home)
	if runErr != nil {
		return nil, fmt.Errorf("run isolated smoke: %w", runErr)
	}

	return report, nil
}

func runList(runner *engine, opts *runOptions) (*smokeReport, error) {
	report, err := runWithHome(runner, opts, emptyString)
	if err != nil {
		return nil, fmt.Errorf("list smoke: %w", err)
	}

	return report, nil
}

func runModule(suite *suiteRun, module *module) ([]*taskResult, error) {
	if suite.opts.ListOnly {
		return runModuleTasks(suite.engine, listModuleRun(suite.opts, suite.home, module)), nil
	}

	results, err := runCopiedModule(suite, module)
	if err != nil {
		return nil, fmt.Errorf("run copied module: %w", err)
	}

	return results, nil
}

func runModuleTasks(runner *engine, run *moduleRun) []*taskResult {
	tasks := selectedSmokeTasks(run.Module)
	results := make([]*taskResult, emptyLength, len(tasks))

	for i := range tasks {
		results = append(results, runOne(runner, run, tasks[i]))
	}

	return results
}

func runOne(runner *engine, run *moduleRun, spec *taskSpec) *taskResult {
	reason := skipReason(newSkipInput(run.Module, spec))

	if reason != emptyString {
		return skippedResult(run.Module.Name, spec.Name, reason)
	}

	return runOrList(runner, run, spec)
}

func runOrList(runner *engine, run *moduleRun, spec *taskSpec) *taskResult {
	if run.Opts.ListOnly {
		return plannedResult(run.Module.Name, spec.Name)
	}

	return executeOne(runner, run, spec)
}

func runSelected(suite *suiteRun) (*smokeReport, error) {
	err := attachConfigs(suite.opts.RepoRoot, suite.modules)
	if err != nil {
		return nil, fmt.Errorf("attach smoke configs: %w", err)
	}

	report, collectErr := collectResults(suite)
	if collectErr != nil {
		return nil, fmt.Errorf("collect smoke results: %w", collectErr)
	}

	return report, nil
}

func runWithHome(runner *engine, opts *runOptions, home string) (*smokeReport, error) {
	modules, err := discoverModules(opts.RepoRoot)
	if err != nil {
		return nil, fmt.Errorf("discover smoke modules: %w", err)
	}

	suite := bindSuite(runner, opts, home)

	suite.modules = modules

	report, runErr := runFiltered(suite)
	if runErr != nil {
		return nil, fmt.Errorf("run filtered modules: %w", runErr)
	}

	return report, nil
}

func selectedRunner(opts *runOptions) func(*engine, *runOptions) (*smokeReport, error) {
	if opts.ListOnly {
		return runList
	}

	return runIsolated
}

func skippedResult(module, name, reason string) *taskResult {
	return &taskResult{
		Duration: emptyLength,
		Err:      nil,
		Module:   module,
		Output:   reason,
		Status:   statusSkip,
		Task:     name,
	}
}

func specAsTask(spec *taskSpec) *tasktest.Task {
	task := tasktest.Task{
		Interactive: spec.Interactive,
		Prompt:      spec.Prompt,
		Requires:    specRequires(spec.Requires),
	}

	return &task
}

func specRequires(names []string) *tasktest.TaskRequires {
	if len(names) == emptyLength {
		return nil
	}

	requires := tasktest.TaskRequires{Vars: names}

	return &requires
}
