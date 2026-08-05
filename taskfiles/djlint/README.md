# djLint Taskfile

## What is this Taskfile?

This Taskfile wraps [djLint](https://www.djlint.com/), a linter and formatter
for HTML template languages (Django, Jinja, Nunjucks, Handlebars, and more),
with automation tasks for installing the tool and linting, formatting, and
format-checking templates on macOS, Linux, and Windows. djLint is installed
through [uv](https://docs.astral.sh/uv/) in an isolated environment so it
never conflicts with project dependencies.

## Usage

### Standalone

```bash
task --taskfile taskfiles/djlint/Taskfile.yml lint DJLINT_TARGETS=templates
```

### Included

```yaml
includes:
  djlint:
    taskfile: taskfiles/djlint/Taskfile.yml
```

```bash
task djlint:lint DJLINT_TARGETS=templates
task djlint:lint DJLINT_TARGETS=templates DJLINT_EXTRA_ARGS="--profile django"
task djlint:fmt DJLINT_TARGETS=templates
task djlint:fmt:check DJLINT_TARGETS=templates
task djlint:install DJLINT_VERSION=1.36.4
```

## Public Tasks

| Task | Description | Key variables |
|---|---|---|
| `install` | Install djLint on the current operating system | `DJLINT_VERSION` |
| `install:undo` | Remove djLint (alias: `uninstall`) | |
| `upgrade` | Upgrade djLint to the latest release | `DJLINT_VERSION` |
| `lint` | Lint HTML templates with djlint --lint | `DJLINT_TARGETS`, `DJLINT_EXTRA_ARGS` |
| `fmt` | Format HTML templates in place with djlint --reformat | `DJLINT_TARGETS`, `DJLINT_EXTRA_ARGS` |
| `fmt:check` | Report formatting changes without modifying files (djlint --check) | `DJLINT_TARGETS`, `DJLINT_EXTRA_ARGS` |
| `ci` | Run `fmt:check` then `lint` | `DJLINT_TARGETS`, `DJLINT_EXTRA_ARGS` |
| `ci:fix` | Run `fmt` for CI fixing | — |
| `version` | Show the installed djLint version | |

## Variables

| Variable | Default | Description |
|---|---|---|
| `DJLINT_VERSION` | `""` (latest) | Pin a specific djLint release |
| `DJLINT_TARGETS` | `.` | File or directory djLint operates on |
| `DJLINT_EXTRA_ARGS` | `""` | Extra flags forwarded to djLint (e.g. `--profile django`) |
| `UV_LOAD` | `export PATH="$HOME/.local/bin:$PATH"` | Shell snippet that puts uv-managed tools on PATH (unix) |
| `DJLINT_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint checks and fixes |
| `DJLINT_FMT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by formatting checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Notes

- `lint` reports template lint rule violations (`--lint`); `fmt:check` is the
  dry-run counterpart of `fmt` and reports formatting differences (`--check`).
  They are distinct djLint modes.
- Pass `DJLINT_EXTRA_ARGS="--profile <name>"` to select the template dialect
  (django, jinja, nunjucks, handlebars, golang, angular).
- Auto-install: every run task depends on `install`, so the tool is installed
  on first use. Installs are idempotent and version-aware — changing
  `DJLINT_VERSION` triggers a reinstall.
- Windows tasks invoke the uv-managed shims directly; macOS and Linux tasks
  load `~/.local/bin` onto PATH first.
