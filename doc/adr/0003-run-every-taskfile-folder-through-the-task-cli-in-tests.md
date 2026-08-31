# 3. Run every Taskfile folder through the task CLI in tests

Date: 2026-08-26

## Status

Accepted

## Context

Until now every module was covered by a contract test that parses its
`Taskfile.yml`, `metadata.yml` and `README.md`. Those tests describe what a
module *declares*, and they never start the `task` binary.

That leaves a class of defects invisible. A Taskfile can parse as YAML and still
fail to load in the CLI: a broken `includes:` path, a template that does not
render, a task the CLI silently drops. A family root can advertise seven
variants in `metadata.yml` while its include tree wires only six. A `default`
task can be declared and still exit non-zero. All of these ship green today.

Running the tools themselves is not an option for a store of 88 modules: an
honest end-to-end test of `go:install` or `docker:build` downloads a toolchain,
needs privileges, and takes minutes. The gap worth closing is between "the YAML
parses" and "the tool runs", not the tool itself.

## Decision

Every folder under `taskfiles/` that holds a `Taskfile.yml` also holds a
`TestModuleIntegration` in its `_test.go` that runs the real `task` CLI
against that folder.
Shared fragments under `taskfiles/internal/` were excluded while that
directory existed; it is unused now, and the coverage test still skips it
if it reappears.

The per-folder test is one delegating call; all logic is shared in
`internal/taskintegration`:

```go
func TestModuleIntegration(t *testing.T) {
	t.Parallel()

	taskintegration.RunHere(t)
}
```

The shared suite runs these checks against the folder:

1. The Taskfile loads: `task --list-all --json` exits zero and returns JSON.
2. The CLI and the Taskfile agree in both directions — every declared public
   task is listed, and every listed task is declared locally or reached through
   a declared `includes:` namespace.
3. Every task `metadata.yml` exports is reachable, directly or under a namespace.
4. Every variant `metadata.yml` advertises exposes the full exported surface.
5. `task --summary` renders for each task the folder declares.
6. The `default` task runs and produces output.
7. An unknown task name is rejected.

Checks stay hermetic: every run uses an isolated `HOME`, and no check installs,
downloads, or mutates anything outside a temporary directory. `task --dry` is
opt-in per module through `taskintegration.RunSpec`, because most modules need a
tool the test environment does not install.

Enforcement: `TestEveryTaskfileFolderHasAnIntegrationTest` in
`taskfiles/integration_coverage_test.go`.

## Consequences

- Include-tree and variant-wiring bugs fail at test time instead of at a user's
  first `task {tool}:{variant}:install`.
- The full suite still runs in well under a minute, because the summary sweep is
  scoped to the tasks a folder declares; tasks pulled in through an include are
  covered by the test of the folder that declares them.
- A new module is not done until it has `TestModuleIntegration`; the coverage
  test names the missing folder.
- The suite proves a module's task graph resolves and its documentation-facing
  surface is real. It does not prove the underlying tool installs or runs — that
  remains the job of the opt-in installer flows in individual module tests.
