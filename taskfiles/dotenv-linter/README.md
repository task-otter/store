# dotenv-linter Taskfile

## What is this Taskfile?

This Taskfile wraps [dotenv-linter](https://dotenv-linter.github.io/), a
lightning-fast linter for .env files written in Rust, with automation tasks
for installing the tool and linting or fixing dotenv files on macOS, Linux,
and Windows. dotenv-linter is installed with `cargo install`, and the Rust
toolchain itself is bootstrapped through the cargo module when missing.

## Usage

### Standalone

```bash
task --taskfile taskfiles/dotenv-linter/Taskfile.yml ci DOTENV_LINTER_TARGETS=.env.example
```

### Included

```yaml
includes:
  dotenv-linter:
    taskfile: taskfiles/dotenv-linter/Taskfile.yml
```

```bash
task dotenv-linter:ci DOTENV_LINTER_TARGETS=.env.example
task dotenv-linter:ci:fix DOTENV_LINTER_TARGETS=.env
task dotenv-linter:install DOTENV_LINTER_VERSION=3.3.0
```

## Public Tasks

| Task | Description | Key variables |
|---|---|---|
| `install` | Install dotenv-linter on the current operating system | `DOTENV_LINTER_VERSION` |
| `install:undo` | Remove dotenv-linter (alias: `uninstall`) | |
| `upgrade` | Reinstall dotenv-linter at the requested version | `DOTENV_LINTER_VERSION` |
| `ci` | Lint dotenv files with dotenv-linter check | `DOTENV_LINTER_TARGETS`, `DOTENV_LINTER_EXTRA_ARGS` |
| `ci:fix` | Apply automatic fixes with dotenv-linter fix | `DOTENV_LINTER_TARGETS`, `DOTENV_LINTER_EXTRA_ARGS` |
| `diff` | Compare .env files to ensure matching key sets | `DOTENV_LINTER_TARGETS`, `DOTENV_LINTER_EXTRA_ARGS` |
| `version` | Show the installed dotenv-linter version | |

## Variables

| Variable | Default | Description |
|---|---|---|
| `DOTENV_LINTER_VERSION` | `""` (latest) | Pin a specific dotenv-linter release, e.g. `3.3.0` |
| `DOTENV_LINTER_TARGETS` | `.env` | File or directory dotenv-linter operates on |
| `DOTENV_LINTER_EXTRA_ARGS` | `""` | Extra flags forwarded to dotenv-linter (e.g. `--recursive`, `--skip`) |
| `CARGO_BIN_UNIX` | `$HOME/.cargo/bin` | Fallback cargo bin directory on macOS and Linux |
| `DOTENV_LINTER_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by check, fix, and diff tasks |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Notes

- Auto-install: every run task depends on `install`, and `install` bootstraps
  the Rust toolchain via the cargo module first, so `task dotenv-linter:ci`
  works on a fresh machine. Installs are idempotent and version-aware —
  changing `DOTENV_LINTER_VERSION` triggers a reinstall.
- Binaries are resolved from PATH first, falling back to `~/.cargo/bin`
  (`%USERPROFILE%\.cargo\bin` on Windows), so a fresh cargo install works
  without restarting the shell.
- `ci:fix` writes changes in place; dotenv-linter creates a backup of each
  changed file.
- The tasks target the dotenv-linter 4.x CLI, which uses subcommands
  (`check`, `fix`, `diff`). Pin a 4.x release with `DOTENV_LINTER_VERSION`
  if you need reproducible behavior.
