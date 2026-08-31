# bencher

A [TaskOtter](https://github.com/task-otter/store) module for the [Bencher CLI](https://bencher.dev/docs/how-to/install-cli/), which uploads and tracks benchmark results.

## What is this Taskfile?

This module runs the Bencher CLI. `run` and `exec` auto-install the CLI via `nix:install:profile`.

## Usage

### Standalone

```sh
task -t taskfiles/bencher/Taskfile.yml run -- --project my-project "make benchmarks"
task -t taskfiles/bencher/Taskfile.yml exec -- mock
```

Install only:

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#bencher
```

### Included in your Taskfile

```yaml
includes:
  bencher:
    taskfile: taskfiles/bencher/Taskfile.yml
```

Then run:

```sh
task bencher:run -- --project my-project "make benchmarks"
task bencher:exec -- mock
```

## Public Tasks

| Task | Description |
|---|---|
| `run` | Execute a benchmark command and track its results with `bencher run` |
| `exec` | Run any Bencher CLI subcommand (e.g. `mock`, `project`, `report`) |

## Variables

| Variable | Default | Description |
|---|---|---|
| `BENCHER_NIX_INSTALLABLE` | `nixpkgs#bencher` | Flake installable passed to `nix:install:profile` |
| `BENCHER_EXTRA_ARGS` | `""` | Arguments and flags appended to `run` and `exec` invocations |

Pin a revision by overriding the installable, for example
`BENCHER_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#bencher`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.
- The `run` and `exec` tasks auto-install the Bencher CLI if it is not already present in `PATH`.
