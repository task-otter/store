# 2. Prefix top-level Taskfile vars with the module name

Date: 2026-08-04

## Status

Accepted

## Context

TaskOtter modules are composed via Taskfile `includes:`. Top-level `vars:` from
included modules share one namespace in the composed graph. Bare knobs such as
`CACHE`, `IMAGE`, `BASE`, or `COLLECTION` collide when two families are included
together, and make it unclear which module owns a public override.

Per-task `vars:` stay local to that task and are out of scope.

## Decision

Top-level Taskfile vars must satisfy **one** of:

1. **Owned prefix:** `{TOOL}_…` where `TOOL` is the family root — the first path
   segment under `taskfiles/` (`go` → `GO_`, `prettier/node/fnm/npm` →
   `PRETTIER_`). For `taskfiles/internal/<name>/`, use `<NAME>_`
   (`internal/skipfiles` → `SKIPFILES_`). Hyphens in the directory name become
   underscores (`bash-exec` → `BASH_EXEC_`).
2. **Shared API allowlist** (no module prefix): `EXTRA_ARGS`, `VERSION`,
   `TARGETS`, `CONFIG`, `FILE`, `ARGS`, `REQUIREMENTS`, `VENV`, `IGNORE_PATH`,
   `FORCE`.
3. **Foreign / dependency prefix:** starts with another module’s `{NAME}_`
   discovered from top-level directories under `taskfiles/` (for example `UV_`,
   `FNM_`, `CARGO_`, `GO_`), **or** a companion allowlist: `RUST_`, `RUSTUP_`,
   `PROTOC_`, `YAMLFIX_`, `NODE_`, `WINDOWS_`.

Bare module knobs (`CACHE`, `IMAGE`, `PLAYBOOK`, …) and go’s former
`INSTALL_DIR_UNIX` / `GLOBAL_GO_BIN` names are not allowed; rename them to the
owned (or foreign) form.

Enforcement: `TestTopLevelVarsPrefix` in `taskfiles/vars_prefix_test.go`.

## Consequences

- Public override knobs become longer but unambiguous when modules compose.
- Renames are breaking for callers who set the old bare names on the CLI or via
  `includes.vars`.
- Shared allowlisted names remain short for cross-module conventions.
- Foreign prefixes let a module reference a dependency’s install paths (for
  example `GO_GLOBAL_BIN`) without inventing a duplicate owned name.
