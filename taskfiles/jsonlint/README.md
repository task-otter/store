# jsonlint Taskfile

## What is this Taskfile?

This Taskfile wraps the `jsonlint` command-line JSON validator with automation
tasks for installing the tool and validating JSON files on macOS, Linux, and
Windows. The CLI is provided by the [demjson3](https://pypi.org/project/demjson3/)
package and is installed through [uv](https://docs.astral.sh/uv/) in an
isolated environment so it never conflicts with project dependencies.

## Usage

### Standalone

```bash
task --taskfile taskfiles/jsonlint/Taskfile.yml lint JSONLINT_TARGETS=config.json
```

### Included

```yaml
includes:
  jsonlint:
    taskfile: taskfiles/jsonlint/Taskfile.yml
```

```bash
task jsonlint:lint JSONLINT_TARGETS=config.json
task jsonlint:lint JSONLINT_TARGETS=data/   # validates every *.json under data/
task jsonlint:install JSONLINT_VERSION=3.0.6
```

## Public Tasks

| Task | Description | Key variables |
|---|---|---|
| `install` | Install the jsonlint CLI on the current operating system | `JSONLINT_VERSION` |
| `install:undo` | Remove the jsonlint CLI (alias: `uninstall`) | |
| `upgrade` | Upgrade the jsonlint CLI to the latest release | `JSONLINT_VERSION` |
| `lint` | Validate JSON files with jsonlint | `JSONLINT_TARGETS`, `JSONLINT_EXTRA_ARGS` |
| `version` | Show the installed jsonlint version | |

## Variables

| Variable | Default | Description |
|---|---|---|
| `JSONLINT_VERSION` | `""` (latest) | Pin the demjson3 release that provides the jsonlint CLI |
| `JSONLINT_TARGETS` | `.` | File or directory to validate; directories are scanned recursively for `*.json` |
| `JSONLINT_EXTRA_ARGS` | `""` | Extra flags forwarded to jsonlint |
| `UV_LOAD` | `export PATH="$HOME/.local/bin:$PATH"` | Shell snippet that puts uv-managed tools on PATH (unix) |
| `JSONLINT_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Notes

- The PyPI package named `jsonlint` is an unrelated validation library that
  ships no command-line tool; this Taskfile installs `demjson3`, which
  provides the actual `jsonlint` CLI.
- Auto-install: every run task depends on `install`, so the tool is installed
  on first use. Installs are idempotent and version-aware — changing
  `JSONLINT_VERSION` triggers a reinstall.
- Windows tasks invoke the uv-managed shims directly; macOS and Linux tasks
  load `~/.local/bin` onto PATH first.
