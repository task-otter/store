# rumdl Taskfile

## What Is This Taskfile?

This Taskfile wraps [rumdl](https://github.com/rvben/rumdl), a fast Markdown
linter and formatter written in Rust, with automation tasks for linting, fixing,
and formatting Markdown files. Run tasks auto-install rumdl via
`nix:install:profile`.

## Usage

### Standalone

```bash
task --taskfile taskfiles/rumdl/Taskfile.yml ci RUMDL_TARGETS=docs
```

Install only, without linting:

```sh
task --taskfile taskfiles/rumdl/Taskfile.yml install
```

### Included

```yaml
includes:
  rumdl:
    taskfile: taskfiles/rumdl/Taskfile.yml
```

```bash
task rumdl:ci RUMDL_TARGETS=docs
task rumdl:ci:fix RUMDL_TARGETS=README.md
```

## Public Tasks

| Task      | Description                                             | Key variables                                |
| --- | --- | --- |
| --------- | ------------------------------------------------------- | -------------------------------------------- |
| `ci`      | Lint Markdown files with rumdl check                    | `RUMDL_TARGETS`, `RUMDL_EXTRA_ARGS`          |
| `ci:fix`  | Format Markdown files with rumdl fmt                    | `RUMDL_TARGETS`, `RUMDL_EXTRA_ARGS`          |
| `install` | Install rumdl via Nix (Unix) or WinGet (Windows) | `RUMDL_NIX_INSTALLABLE`, `RUMDL_WINGET_INSTALLABLE` |
| `version` | Show the active rumdl version                           | —                                            |

## Variables

| Variable                | Default         | Description                                       |
| --- | --- | --- |
| ----------------------- | --------------- | ------------------------------------------------- |
| `RUMDL_NIX_INSTALLABLE` | `nixpkgs#rumdl` | Flake installable passed to `nix:install:profile` |
| `RUMDL_WINGET_INSTALLABLE` | `rvben.rumdl` | Package ID for Windows `winget:install:package` |
| `RUMDL_TARGETS`         | `.`             | File or directory rumdl operates on               |
| `RUMDL_EXTRA_ARGS`      | `""`            | Extra flags forwarded to rumdl                    |

Pin a revision by overriding the installable, for example
`RUMDL_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#rumdl`.

## Notes

* Unix install uses Nix (`RUMDL_NIX_INSTALLABLE`). Windows installs via
  `winget:install:package` (`RUMDL_WINGET_INSTALLABLE`).
* `ci:fix` (rumdl fmt) uses formatter-style exit codes and exits zero after
  successful formatting even if unfixable violations remain. Run `ci` to
  check for remaining violations.
* Auto-install: every run task depends on the module's `install` task, so the
  tool is installed on first use.
