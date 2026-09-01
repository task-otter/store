# Buf Taskfile

## What is this Taskfile?

A cross-platform Taskfile for linting, formatting, breaking-change detection,
and code generation from [Protocol Buffer](https://protobuf.dev/) definitions
using [Buf](https://buf.build/), the modern proto toolchain.

Operational tasks auto-install Buf via `nix:install:profile`.

## Usage

### Standalone

```sh
task -t taskfiles/buf/Taskfile.yml lint
task -t taskfiles/buf/Taskfile.yml fmt:check
```

Install only, without linting:

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#buf
```

Lint a specific proto directory:

```sh
task -t taskfiles/buf/Taskfile.yml lint BUF_INPUT=api/v1
```

Check for breaking changes against a branch:

```sh
task -t taskfiles/buf/Taskfile.yml breaking BUF_AGAINST=.git#branch=main
```

Pass buf flags directly after `--`:

```sh
task -t taskfiles/buf/Taskfile.yml lint -- api/v1 --error-format json
```

### Included

```yaml
includes:
  buf:
    taskfile: taskfiles/buf/Taskfile.yml
```

Then run:

```sh
task buf:lint
task buf:fmt:check
task buf:breaking BUF_AGAINST=.git#branch=main
task buf:generate BUF_INPUT=api/v1
```

## Public Tasks

| Task            | Description                                              | Key variables                  |
| --------------- | -------------------------------------------------------- | ------------------------------ |
| `breaking`  | Check proto files for breaking changes against AGAINST | `BUF_INPUT`, `BUF_AGAINST`, `BUF_EXTRA_ARGS` |
| `ci`        | Run `fmt:check` then `lint`                            | `BUF_INPUT`, `BUF_CONFIG`, `BUF_EXTRA_ARGS` |
| `ci:fix`    | Format proto files in place with Buf                   | `BUF_INPUT`, `BUF_EXTRA_ARGS` |
| `fmt:check` | Check proto file formatting with Buf                   | `BUF_INPUT`, `BUF_EXTRA_ARGS` |
| `generate`  | Generate code from proto files with Buf                | `BUF_INPUT`, `BUF_EXTRA_ARGS` |
| `lint`      | Lint proto files with Buf                              | `BUF_INPUT`, `BUF_CONFIG`, `BUF_EXTRA_ARGS` |

## Variables

| Variable      | Default              | Description                                              |
| ------------- | -------------------- | -------------------------------------------------------- |
| `BUF_AGAINST`            | `.git#branch=main` | Baseline for `breaking`: a git ref, Buf module, or path |
| `BUF_NIX_INSTALLABLE`    | `nixpkgs#buf`      | Flake installable passed to `nix:install:profile`       |
| `BUF_CONFIG`             | empty              | Path to a `buf.yaml` config file passed via `--config`  |
| `BUF_EXTRA_ARGS`  | empty                | Extra arguments appended when `CLI_ARGS` is not provided |
| `BUF_INPUT`       | `.`                  | Proto source directory or Buf module passed to buf       |

Pin a revision by overriding the installable, for example
`BUF_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#buf`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.
- The `generate` task requires a `buf.gen.yaml` file in the working tree. See the
  [buf generate docs](https://buf.build/docs/generate/tutorial) for configuration
  details.
