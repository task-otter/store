# protolint Taskfile

## What is this Taskfile?

This Taskfile wraps [protolint](https://github.com/yoheimuta/protolint), a
pluggable linter and fixer for Protocol Buffer files, with automation tasks
for installing the tool and linting or fixing .proto files on macOS, Linux,
and Windows. protolint is installed from its official Go module into the
global Go bin, so the Go toolchain is bootstrapped automatically when needed.

## Usage

### Standalone

```bash
task --taskfile taskfiles/protolint/Taskfile.yml ci PROTOLINT_TARGETS=api
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
task protolint:install PROTOLINT_VERSION=v0.55.6
```

## Public Tasks

| Task | Description | Key variables |
|---|---|---|
| `install` | Install protolint into the global Go bin | `PROTOLINT_VERSION` |
| `install:undo` | Remove protolint from the global Go bin (alias: `uninstall`) | |
| `upgrade` | Reinstall protolint at the requested version | `PROTOLINT_VERSION` |
| `ci` | Lint protobuf files with protolint | `PROTOLINT_TARGETS`, `PROTOLINT_EXTRA_ARGS` |
| `ci:fix` | Apply automatic fixes with protolint lint -fix | `PROTOLINT_TARGETS`, `PROTOLINT_EXTRA_ARGS` |
| `version` | Show the installed protolint version | |

## Variables

| Variable | Default | Description |
|---|---|---|
| `PROTOLINT_VERSION` | `""` (latest) | Pin an exact protolint module version, e.g. `v0.55.6` |
| `PROTOLINT_TARGETS` | `.` | File or directory protolint operates on |
| `PROTOLINT_EXTRA_ARGS` | `""` | Extra flags forwarded to protolint (e.g. `-config_path`, `-reporter json`) |
| `GO_GLOBAL_BIN` | GOBIN → GOPATH/bin → `$HOME/go/bin` | Resolved global Go bin directory |
| `PROTOLINT_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Notes

- Auto-install: every run task depends on `install`, so protolint (and the Go
  toolchain via the go module) is installed on first use. Installs are
  idempotent and version-aware — changing `PROTOLINT_VERSION` triggers a
  reinstall, verified with `go version -m`.
- Windows tasks invoke `protolint.exe` from the resolved Go bin; macOS and
  Linux invoke the binary directly, so no PATH changes are required.
