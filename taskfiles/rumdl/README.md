# rumdl Taskfile

## What is this Taskfile?

This Taskfile wraps [rumdl](https://github.com/rvben/rumdl), a fast Markdown
linter and formatter written in Rust, with automation tasks for installing the
tool and linting, fixing, and formatting Markdown files on macOS, Linux, and
Windows. rumdl is installed through [uv](https://docs.astral.sh/uv/) in an
isolated environment so it never conflicts with project dependencies.

## Usage

### Standalone

```bash
task --taskfile taskfiles/rumdl/Taskfile.yml ci RUMDL_TARGETS=docs
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
task rumdl:install RUMDL_VERSION=0.0.145
```

## Public Tasks

| Task | Description | Key variables |
|---|---|---|
| `install` | Install rumdl on the current operating system | `RUMDL_VERSION` |
| `install:undo` | Remove rumdl (alias: `uninstall`) | |
| `upgrade` | Upgrade rumdl to the latest release | `RUMDL_VERSION` |
| `ci` | Lint Markdown files with rumdl check | `RUMDL_TARGETS`, `RUMDL_EXTRA_ARGS` |
| `lint:fix` | Apply automatic fixes with rumdl check --fix | `RUMDL_TARGETS`, `RUMDL_EXTRA_ARGS` |
| `ci:fix` | Run `fmt` then `lint:fix` for CI | — |
| `fmt` | Format Markdown files with rumdl fmt | `RUMDL_TARGETS`, `RUMDL_EXTRA_ARGS` |
| `version` | Show the installed rumdl version | |

## Variables

| Variable | Default | Description |
|---|---|---|
| `RUMDL_VERSION` | `""` (latest) | Pin a specific rumdl release |
| `RUMDL_TARGETS` | `.` | File or directory rumdl operates on |
| `RUMDL_EXTRA_ARGS` | `""` | Extra flags forwarded to rumdl |
| `UV_LOAD` | `export PATH="$HOME/.local/bin:$PATH"` | Shell snippet that puts uv-managed tools on PATH (unix) |
| `RUMDL_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint checks and fixes |
| `RUMDL_FMT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by formatting checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Notes

- `lint:fix` (rumdl check --fix) exits non-zero when unfixable violations remain,
  which suits pre-commit hooks and CI. `fmt` (rumdl fmt) uses formatter-style
  exit codes and exits zero after formatting, which suits editor integration.
- Auto-install: every run task depends on `install`, so the tool is installed
  on first use. Installs are idempotent and version-aware — changing
  `RUMDL_VERSION` triggers a reinstall.
- Windows tasks invoke the uv-managed shims directly; macOS and Linux tasks
  load `~/.local/bin` onto PATH first.
