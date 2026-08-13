# shfmt

A [TaskOtter](https://github.com/task-otter/store) module for [shfmt](https://github.com/mvdan/sh), the shell formatter supporting POSIX shell, Bash, Zsh, and mksh.

## What is this Taskfile?

This module installs `shfmt` from its official Go module into the global Go bin directory, then formats shell scripts in place or reports formatting differences.

## Usage

### Standalone

```sh
task -t taskfiles/shfmt/Taskfile.yml install
task -t taskfiles/shfmt/Taskfile.yml ci:fix SHFMT_TARGETS=scripts
task -t taskfiles/shfmt/Taskfile.yml fmt:check SHFMT_TARGETS=scripts SHFMT_EXTRA_ARGS="-i 2 -ci"
task -t taskfiles/shfmt/Taskfile.yml version
```

### Included in your Taskfile

```yaml
includes:
  shfmt:
    taskfile: taskfiles/shfmt/Taskfile.yml
    vars:
      SHFMT_TARGETS_OVERRIDE: "{{.SHFMT_TARGETS}}"
      SHFMT_EXTRA_ARGS_OVERRIDE: "{{.SHFMT_EXTRA_ARGS}}"
```

Then run:

```sh
task shfmt:ci:fix
task shfmt:fmt:check SHFMT_TARGETS=scripts
```

## Public Tasks

| Task | Description |
|---|---|
| `install` | Install shfmt into the global Go bin |
| `install:undo` | Remove shfmt from the global Go bin |
| `upgrade` | Upgrade shfmt to the requested version |
| `ci:fix` | Format shell scripts in place (`SHFMT_TARGETS=path`) |
| `fmt:check` | Check shell script formatting without modifying files (`SHFMT_TARGETS=path`) |
| `version` | Show the installed shfmt version |

## Variables

| Variable | Default | Description |
|---|---|---|
| `SHFMT_VERSION` | _(empty)_ | Exact module release to install; empty installs the latest stable v3 release |
| `SHFMT_TARGETS` | `.` | File or directory to format or check |
| `SHFMT_EXTRA_ARGS` | _(empty)_ | Extra shfmt flags, for example `-i 2`, `-ci`, `-sr`, or `-ln bash` |
| `GO_GLOBAL_BIN` | Go's `GOBIN` or `GOPATH/bin` | Directory where the shfmt binary is installed |
| `SHFMT_FMT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by formatting checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Notes

- `install` first ensures the Go toolchain is available through the local `go` Taskfile, then runs the [official installation command](https://github.com/mvdan/sh#shfmt).
- `ci:fix` uses `shfmt -w`; `fmt:check` uses `shfmt -d` and exits non-zero when formatting differs.
- `SHFMT_TARGETS` may be a single shell script or a directory. Pass dialect and style preferences through `SHFMT_EXTRA_ARGS`.
