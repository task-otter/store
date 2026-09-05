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
task --taskfile taskfiles/protolint/Taskfile.yml install
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
| `install` | Install protolint via Nix (Unix) or go install (Windows) |
| `version` | Show the active protolint version |

## Variables

| Variable | Default | Description |
|---|---|---|
| `PROTOLINT_NIX_INSTALLABLE` | `nixpkgs#protolint` | Flake installable passed to `nix:install:profile` |
| `PROTOLINT_GO_PKG` | `github.com/yoheimuta/protolint/cmd/protolint@latest` | Go module passed to `go:install:pkg` on Windows |
| `PROTOLINT_TARGETS` | `.` | File or directory protolint operates on |
| `PROTOLINT_EXTRA_ARGS` | `""` | Extra flags forwarded to protolint (e.g. `-config_path`, `-reporter json`) |

Pin a revision by overriding the installable, for example
`PROTOLINT_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#protolint`.

## Notes

- Unix install uses Nix (`PROTOLINT_NIX_INSTALLABLE`). On Windows, install uses `go:install:pkg` with `PROTOLINT_GO_PKG`.
- Go is provided by the included [`go`](../go/README.md) module. Operational tasks depend on `go:install`.

