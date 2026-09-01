# 2. Prefix top-level Taskfile vars with the module name

Date: 2026-08-04

## Status

Accepted

## Context

TaskOtter modules are composed via Taskfile `includes:`. Top-level `vars:` from
included modules share one namespace in the composed graph. Bare knobs such as
`CACHE`, `IMAGE`, `BASE`, `COLLECTION`, or shared-looking names like
`EXTRA_ARGS` and `VERSION` collide when two families are included together, and
make it unclear which module owns a public override.

Per-task `vars:` stay local to that task and are out of scope.

## Decision

Top-level Taskfile vars must satisfy **one** of:

1. **Owned prefix:** `{TOOL}_…` where `TOOL` is the family root — the first path
   segment under `taskfiles/` (`go` → `GO_`, `prettier/node/npm` →
   `PRETTIER_`). For `taskfiles/internal/<name>/`, use `<NAME>_`.
2. **Foreign / dependency prefix:** starts with another module’s `{NAME}_`
   discovered from top-level directories under `taskfiles/` (for example `UV_`,
   `NPM_`, `CARGO_`, `GO_`), **or** a companion allowlist: `RUST_`, `RUSTUP_`,
   `PROTOC_`, `NODE_`, `WINDOWS_`.

Bare module knobs (`CACHE`, `IMAGE`, `PLAYBOOK`, `EXTRA_ARGS`, `VERSION`, …)
and go’s former `INSTALL_DIR_UNIX` / `GLOBAL_GO_BIN` names are not allowed;
rename them to the owned (or foreign) form. There is no `_OVERRIDE` suffix: a
prefixed top-level var is the single public name, overridden directly.

Enforcement: `TestTopLevelVarsPrefix` in `taskfiles/vars_prefix_test.go`.

## Consequences

- Public override knobs become longer but unambiguous when modules compose.
- Renames are breaking for callers who set the old bare names on the CLI or via
  `includes.vars`.
- Cross-module conventions use distinct owned names per family (for example
  `PRETTIER_EXTRA_ARGS` vs `DOCKER_EXTRA_ARGS`) instead of a shared bare API.
- Foreign prefixes let a module reference a dependency’s install paths (for
  example `GO_GLOBAL_BIN`) without inventing a duplicate owned name.
- Task-local knobs that are never declared at top-level (for example undeclared
  `EXTRA_ARGS` passed into `npm:add`) stay outside this rule.
- The parallel `{TOOL}_…_OVERRIDE` escape hatches were removed once every public
  knob lived at top-level: prefixed vars are already settable from the CLI, the
  environment, and `includes.vars`, so the second name was redundant. Modules
  that relied on it — `sqlfluff` (task-local `TARGETS` / `CONFIG` / `DIALECT` /
  `EXTRA_ARGS`, now hoisted to `SQLFLUFF_*`) and `proto` — were converted.
  Breaking for callers who set an `_OVERRIDE` name.
