# bruno-cli

A [TaskOtter](https://github.com/task-otter/store) module for the [Bruno](https://www.usebruno.com/) CLI (`bru`) — run API collections from the command line.

## What is this Taskfile?

This module runs Bruno API collections with `bru run` and `bru run --bail` for CI. The `run`, `ci`, and `help` tasks auto-install `bru` via `nix:install:profile` when it is not already on `PATH`.

## Usage

### Standalone

```sh
task -t taskfiles/bruno-cli/Taskfile.yml run
task -t taskfiles/bruno-cli/Taskfile.yml run BRUNO_CLI_COLLECTION=./api BRUNO_CLI_ENV=staging
task -t taskfiles/bruno-cli/Taskfile.yml ci
task -t taskfiles/bruno-cli/Taskfile.yml help
```

Install only, without running a collection:

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#bruno-cli
```

### Included in your Taskfile

```yaml
includes:
  bruno-cli:
    taskfile: taskfiles/bruno-cli/Taskfile.yml
```

Then run:

```sh
task bruno-cli:run
task bruno-cli:ci BRUNO_CLI_COLLECTION=./api
```

## Public Tasks

| Task | Description |
|---|---|
| `run` | Run all requests in a Bruno collection |
| `ci` | Run a collection and stop on the first failure (`--bail`) |
| `help` | Show Bruno CLI help |
| `install` | Install the Bruno CLI via the Nix profile |
| `version` | Show the active Bruno CLI version |

## Variables

| Variable | Default | Description |
|---|---|---|
| `BRUNO_CLI_NIX_INSTALLABLE` | `nixpkgs#bruno-cli` | Flake installable passed to `nix:install:profile` |
| `BRUNO_CLI_COLLECTION` | `"."` | Path to the Bruno collection directory |
| `BRUNO_CLI_ENV` | `""` | Named Bruno environment to activate via `--env` |
| `BRUNO_CLI_EXTRA_ARGS` | `""` | Additional flags passed to `bru` |

Pin a revision by overriding the installable, for example
`BRUNO_CLI_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#bruno-cli`.

## Breaking changes from `bruno:*`

| Before | After |
|---|---|
| `task bruno:bun:run` | `task bruno-cli:run` |
| `task bruno:node:npm:ci` | `task bruno-cli:ci` |
| `task bruno:bun:install` | `task nix:install:profile NIX_INSTALLABLE=nixpkgs#bruno-cli` |
| `BRUNO_COLLECTION` | `BRUNO_CLI_COLLECTION` |
| `BRUNO_VERSION` | Pin via `BRUNO_CLI_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#bruno-cli` |

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported for auto-install; use WSL2 or ensure `bru` is on `PATH`.
- Pass arguments after `--` to override collection, env, and extra flags directly.
