# rumdl Taskfile

## What is this Taskfile?

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
task nix:install:profile NIX_INSTALLABLE=nixpkgs#rumdl
```

### Included

```yaml
includes:
  rumdl:
    taskfile: taskfiles/rumdl/Taskfile.yml
```

```bash
task rumdl:ci RUMDL_TARGETS=docs
task rumdl:lint:fix RUMDL_TARGETS=README.md
task rumdl:fmt
```

## Public Tasks

| Task | Description | Key variables |
|---|---|---|
| `ci` | Lint Markdown files with rumdl check | `RUMDL_TARGETS`, `RUMDL_EXTRA_ARGS` |
| `lint:fix` | Apply automatic fixes with rumdl check --fix | `RUMDL_TARGETS`, `RUMDL_EXTRA_ARGS` |
| `ci:fix` | Run `fmt` then `lint:fix` for CI | — |
| `fmt` | Format Markdown files with rumdl fmt | `RUMDL_TARGETS`, `RUMDL_EXTRA_ARGS` |

## Variables

| Variable | Default | Description |
|---|---|---|
| `RUMDL_NIX_INSTALLABLE` | `nixpkgs#rumdl` | Flake installable passed to `nix:install:profile` |
| `RUMDL_TARGETS` | `.` | File or directory rumdl operates on |
| `RUMDL_EXTRA_ARGS` | `""` | Extra flags forwarded to rumdl |

Pin a revision by overriding the installable, for example
`RUMDL_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#rumdl`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.
- `lint:fix` (rumdl check --fix) exits non-zero when unfixable violations remain,
  which suits pre-commit hooks and CI. `fmt` (rumdl fmt) uses formatter-style
  exit codes and exits zero after formatting, which suits editor integration.
- Auto-install: every run task depends on `nix:install:profile`, so the tool is
  installed on first use.
