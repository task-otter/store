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
do not publish `install`, `install:undo`, `upgrade`, or `version`, and they
delete the `_install:*` / `_upgrade:*` platform internals.

Each module includes nix and owns a prefixed installable (not
`{TOOL}_VERSION`):

```yaml
includes:
  nix:
    taskfile: ../nix/Taskfile.yml

vars:
  ACTIONLINT_NIX_INSTALLABLE: nixpkgs#actionlint

tasks:
  ci:
    deps:
      - task: nix:install:profile
        vars:
          NIX_INSTALLABLE: "{{.ACTIONLINT_NIX_INSTALLABLE}}"
    cmds:
      - bash -lc '{{.NIX_LOAD}}; actionlint ...'
```

`NIX_LOAD` comes from the nix include. Unix commands that invoke the installed
CLI wrap with `bash -lc '{{.NIX_LOAD}}; …'` so `~/.nix-profile/bin` is on PATH
without a shell restart. Pinning is an installable override, for example
`ACTIONLINT_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#actionlint`.

`NIX_INSTALLABLE` is passed only as a **task-local** var into
`nix:install:profile`. Consuming modules never declare it at top-level.

Installer-only leftovers keep `default` and set `exported_tasks: []`. Callers
install with `task nix:install:profile NIX_INSTALLABLE=nixpkgs#jq` (root
include) or the same task under a module that includes nix.

The nix module is unchanged: it still installs Nix itself (`nix:install`) and
exposes `install:profile` / `install:shell`.

**Out of scope:** JS local-devDep `install` / `upgrade` / `{TOOL}_VERSION`;
npm / yarn / pnpm project install (installing *project* dependencies, not the
CLI itself).

**Exceptions:** docker keeps `install` / `install:undo` / `upgrade` / `version`.

**Node.js stack:** `nodejs` installs Node.js via `nixpkgs#nodejs`. `npm` depends
on `nodejs:_ensure`. `yarn` and `pnpm` install their CLIs from nix profile and
also depend on `nodejs:_ensure` for the runtime.

Native Windows: `nix:install:profile` already errors via `_windows:unsupported`.
Runtime Windows commands on CLI modules may remain; auto-install on native
Windows will fail. Use WSL2.

## Consequences

- One installer path for CLI tools; module Taskfiles shrink to the work the
  tool does (`ci`, `fmt`, …) plus a nix include and an owned installable.
- Auto-install remains, but it is a dependency, not a public task. Callers who
  used `task {tool}:install` switch to `task nix:install:profile
  NIX_INSTALLABLE=…` or rely on the task that already depended on install.
- `{TOOL}_VERSION` is gone on converted modules. Pinning moves to
  `{TOOL}_NIX_INSTALLABLE`, which satisfies ADR 0002.
- JS and project-manager modules keep their existing install semantics.
- docker stays on its current installer so we do not pretend a Nix profile add
  replaces Docker Desktop.
- Native Windows users of converted modules must run under WSL2 for
  auto-install.
