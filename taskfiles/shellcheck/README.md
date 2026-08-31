# ShellCheck

A [TaskOtter](https://github.com/task-otter/store) module for [ShellCheck](https://www.shellcheck.net) — a static analysis tool for shell scripts.

## What is this Taskfile?

This module lints Bash/sh scripts with [ShellCheck](https://www.shellcheck.net). ShellCheck finds syntax errors, quoting mistakes, deprecated constructs, and portability problems. The `ci` task auto-installs ShellCheck via `nix:install:profile`.

## Usage

### Standalone

```sh
task -t taskfiles/shellcheck/Taskfile.yml ci
task -t taskfiles/shellcheck/Taskfile.yml ci SHELLCHECK_TARGETS="scripts/*.sh"
task -t taskfiles/shellcheck/Taskfile.yml ci SHELLCHECK_EXTRA_ARGS="--shell=bash --severity=warning"
```

Install only, without linting:

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#shellcheck
```

### Included in your Taskfile

```yaml
includes:
  shellcheck:
    taskfile: taskfiles/shellcheck/Taskfile.yml
    vars:
      SHELLCHECK_TARGETS_OVERRIDE: "{{.SHELLCHECK_TARGETS}}"
      SHELLCHECK_EXTRA_ARGS_OVERRIDE: "{{.SHELLCHECK_EXTRA_ARGS}}"
```

Then run:

```sh
task shellcheck:ci
```

## Public Tasks

| Task | Description |
|---|---|
| `ci` | Lint shell scripts with ShellCheck (SHELLCHECK_TARGETS=glob) |

## Variables

| Variable | Default | Description |
|---|---|---|
| `SHELLCHECK_NIX_INSTALLABLE` | `nixpkgs#shellcheck` | Flake installable passed to `nix:install:profile` |
| `SHELLCHECK_EXTRA_ARGS` | `""` | Additional flags passed to `shellcheck` (e.g. `--shell`, `--severity`) |
| `SHELLCHECK_TARGETS` | `""` | Paths or globs of scripts to check; empty = discover all `*.sh` recursively |

Pin a revision by overriding the installable, for example
`SHELLCHECK_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#shellcheck`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.
- When `SHELLCHECK_TARGETS` is empty, all `*.sh` and `*.bash` files under the working tree are discovered recursively (excluding `.git`).
- Pass explicit paths or globs (e.g. `SHELLCHECK_TARGETS="scripts/*.sh"`) to limit the scope.
- The `ci` task auto-installs ShellCheck if it is not already present in `PATH`.
