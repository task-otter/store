# protolint Taskfile

## What is this Taskfile?

This Taskfile wraps [protolint](https://github.com/yoheimuta/protolint), a
pluggable linter and fixer for Protocol Buffer files. The `ci` and `ci:fix`
tasks auto-install protolint via `nix:install:profile`.

## Usage

### Standalone

```bash
task --taskfile taskfiles/protolint/Taskfile.yml ci PROTOLINT_TARGETS=api
```

Install only, without linting:

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#protolint
```

### Included

```yaml
includes:
  protolint:
    taskfile: taskfiles/protolint/Taskfile.yml
```

```bash
task protolint:ci PROTOLINT_TARGETS=api
task protolint:ci:fix PROTOLINT_TARGETS=api
```

## Public Tasks

| Task | Description |
|---|---|
| `ci` | Lint protobuf files with protolint |
| `ci:fix` | Apply automatic fixes with protolint lint -fix |

## Variables

| Variable | Default | Description |
|---|---|---|
| `PROTOLINT_NIX_INSTALLABLE` | `nixpkgs#protolint` | Flake installable passed to `nix:install:profile` |
| `PROTOLINT_TARGETS` | `.` | File or directory protolint operates on |
| `PROTOLINT_EXTRA_ARGS` | `""` | Extra flags forwarded to protolint (e.g. `-config_path`, `-reporter json`) |

Pin a revision by overriding the installable, for example
`PROTOLINT_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#protolint`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.
