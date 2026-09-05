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
task -t taskfiles/bencher/Taskfile.yml install
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
| `install` | Install the Bencher CLI via Nix (Unix) or cargo install (Windows) |
| `version` | Show the active Bencher CLI version |

## Variables

| Variable | Default | Description |
|---|---|---|
| `BENCHER_NIX_INSTALLABLE` | `nixpkgs#bencher` | Flake installable passed to `nix:install:profile` |
| `BENCHER_CARGO_CRATE` | `bencher_cli` | Crate name for Windows `cargo:install:crate` |
| `BENCHER_CARGO_EXTRA_ARGS` | `--git https://github.com/bencherdev/bencher --branch main --locked --force` | Extra `cargo install` flags for Windows (official from-source install; no crates.io package) |
| `BENCHER_EXTRA_ARGS` | `""` | Arguments and flags appended to `run` and `exec` invocations |

Pin a revision by overriding the installable, for example
`BENCHER_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#bencher`.

## Notes

- Unix install uses Nix (`BENCHER_NIX_INSTALLABLE`). Windows installs via
  `cargo:install:crate` from the Bencher GitHub repo (`BENCHER_CARGO_CRATE` /
  `BENCHER_CARGO_EXTRA_ARGS`).

- The `run` and `exec` tasks auto-install the Bencher CLI if it is not already present in `PATH`.
