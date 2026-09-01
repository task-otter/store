# 4. Install CLI tools via nix:install:profile

Date: 2026-08-31

## Status

Accepted

## Context

CLI and system modules in this store each shipped a public installer surface:
`install`, `install:undo`, `upgrade`, and `version`, plus platform internals
(`_install:*` / `_upgrade:*` for brew, apt, scoop, official scripts, and
language-specific tool installers). That duplicated package-manager logic,
drifted across families, and leaked `{TOOL}_VERSION` as the pin knob.

The nix module already exposes a persistent user-profile install
(`nix:install:profile`) that auto-installs Nix, enables `nix-command` and
`flakes`, and adds a flake installable to `~/.nix-profile`. Callers that only
need a CLI on PATH can depend on that task instead of owning an installer.

Two families are not a Nix profile install:

- **JS local-devDep tools** (biome, eslint, prettier, and their runtime
  variants) install the linter as a project `devDependency`. Their `install` /
  `upgrade` / `{TOOL}_VERSION` stay.
- **Project package managers** (npm, yarn, pnpm) run project install
  (`npm install`, and the yarn/pnpm equivalents). Node.js, yarn, and pnpm CLIs
  are installed via the Nix user profile (`nodejs`, `yarn`, `pnpm` modules).

One tool cannot be replaced by `nix profile add` without changing meaning:

- **docker:** the current install is Docker Desktop / get.docker.com (the
  daemon). `nixpkgs#docker` is the CLI/engine only.

ADR 0002 forbids a bare top-level `NIX_INSTALLABLE` on a consuming module: the
nix include already owns that name, and a second top-level copy would collide
in the composed graph.

## Decision

CLI and system modules install their tool through `nix:install:profile`. They
publish exactly two installer tasks — `install` and `version` — and they delete
`install:undo`, `upgrade`, and the `_install:*` / `_upgrade:*` platform
internals.

`install` is the module's single caller of `nix:install:profile`. It is guarded
by a `status:` check so an already-present tool is a no-op, which keeps it safe
as a dependency of every work task. `version` depends on `install` and prints
the tool's own version. Work tasks auto-install by depending on `install`
rather than on `nix:install:profile` directly, so the installable is named in
one place per module.

Each module includes nix and owns a prefixed installable (not
`{TOOL}_VERSION`):

```yaml
includes:
  nix:
    taskfile: ../nix/Taskfile.yml

vars:
  ACTIONLINT_NIX_INSTALLABLE: nixpkgs#actionlint

tasks:
  install:
    desc: Install actionlint via the Nix profile
    status:
      - command -v actionlint >/dev/null 2>&1
    cmds:
      - task: nix:install:profile
        vars:
          NIX_INSTALLABLE: "{{.ACTIONLINT_NIX_INSTALLABLE}}"

  version:
    desc: Show the active actionlint version
    deps:
      - task: install
    cmds:
      - bash -lc '{{.NIX_LOAD}}; actionlint -version'

  ci:
    deps:
      - task: install
    cmds:
      - bash -lc '{{.NIX_LOAD}}; actionlint ...'
```

A module that depends on another module's tool depends on that module's
`install` task — `git` on `gh:install`, `vault` on `jq:install` — never on a
namespaced `nix:install:profile`.

`NIX_LOAD` comes from the nix include. Unix commands that invoke the installed
CLI wrap with `bash -lc '{{.NIX_LOAD}}; …'` so `~/.nix-profile/bin` is on PATH
without a shell restart. Pinning is an installable override, for example
`ACTIONLINT_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#actionlint`.

`NIX_INSTALLABLE` is passed only as a **task-local** var into
`nix:install:profile`. Consuming modules never declare it at top-level.

Modules with no work tasks of their own — jq is the only one — are not an
exception: they keep `default` and export `[install, version]` like everything
else, so `task jq:install` works the same way as `task actionlint:install`.

The nix module is unchanged: it still installs Nix itself (`nix:install`) and
exposes `install:profile` / `install:shell`.

**Out of scope:** JS local-devDep `install` / `upgrade` / `{TOOL}_VERSION`;
npm / yarn / pnpm project install (installing *project* dependencies, not the
CLI itself).

**Exceptions:** docker keeps `install` / `install:undo` / `upgrade` / `version`,
because its installer is Docker Desktop rather than a Nix profile add.

**Node.js stack:** `nodejs` installs Node.js via `nixpkgs#nodejs`. `npm` depends
on `nodejs:install`. `yarn` and `pnpm` install their CLIs from nix profile and
also depend on `nodejs:install` for the runtime.

**Task naming:** the per-module installer was originally the private `_ensure`
task; it is now the public `install` task described above. `pnpm` and `yarn`
already reserve `install` and `version` for project dependencies, so there the
Nix installer and its version report are named `install:tool` and
`version:tool`; folding them into the existing `install` would create a
dependency cycle through the `_pnpm:*` / `_yarn:*` shims.

**Enforcement:** the shared integration suite
(`internal/taskintegration`) fails any module that includes nix and owns a
`{TOOL}_NIX_INSTALLABLE` var but does not declare and export both tasks. The
check accepts the `:tool` variants, and skips modules that own no installable.

Native Windows: `nix:install:profile` already errors via `_windows:unsupported`.
Runtime Windows commands on CLI modules may remain; auto-install on native
Windows will fail. Use WSL2.

## Consequences

- One installer path for CLI tools; module Taskfiles shrink to the work the
  tool does (`ci`, `fmt`, …) plus a nix include, an owned installable, and the
  two-task installer surface.
- Auto-install remains, and `task {tool}:install` is the supported way to
  install without doing any work. The installable is named once per module, in
  `install`, so a pin override reaches every task that needs the tool.
- The surface is uniform and machine-checked: every Nix-backed module answers
  `task {tool}:install` and `task {tool}:version`, and the integration suite
  fails a module that drifts from that.
- `{TOOL}_VERSION` is gone on converted modules. Pinning moves to
  `{TOOL}_NIX_INSTALLABLE`, which satisfies ADR 0002.
- JS and project-manager modules keep their existing install semantics.
- docker stays on its current installer so we do not pretend a Nix profile add
  replaces Docker Desktop.
- Native Windows users of converted modules must run under WSL2 for
  auto-install.
