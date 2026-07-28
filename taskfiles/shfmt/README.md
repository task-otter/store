# shfmt

A [TaskOtter](https://github.com/task-otter/store) module for [shfmt](https://github.com/mvdan/sh), the shell formatter supporting POSIX shell, Bash, Zsh, and mksh.

## What is this Taskfile?

This module installs `shfmt` from its official Go module into the global Go bin directory, then formats shell scripts in place or reports formatting differences.

## Usage

### Standalone

```sh
task -t taskfiles/shfmt/Taskfile.yml install
task -t taskfiles/shfmt/Taskfile.yml fmt TARGETS=scripts
task -t taskfiles/shfmt/Taskfile.yml fmt:check TARGETS=scripts EXTRA_ARGS="-i 2 -ci"
task -t taskfiles/shfmt/Taskfile.yml version
```

### Included in your Taskfile

```yaml
includes:
  shfmt:
    taskfile: taskfiles/shfmt/Taskfile.yml
    vars:
      TARGETS_OVERRIDE: "{{.TARGETS}}"
      EXTRA_ARGS_OVERRIDE: "{{.EXTRA_ARGS}}"
```

Then run:

```sh
task shfmt:fmt
task shfmt:fmt:check TARGETS=scripts
```

## Public Tasks

| Task | Description |
|---|---|
| `install` | Install shfmt into the global Go bin |
| `install:undo` | Remove shfmt from the global Go bin |
| `upgrade` | Upgrade shfmt to the requested version |
| `fmt` | Format shell scripts in place (`TARGETS=path`) |
| `fmt:check` | Check shell script formatting without modifying files (`TARGETS=path`) |
| `version` | Show the installed shfmt version |

## Variables

| Variable | Default | Description |
|---|---|---|
| `SHFMT_VERSION` | _(empty)_ | Exact module release to install; empty installs the latest stable v3 release |
| `TARGETS` | `.` | File or directory to format or check |
| `EXTRA_ARGS` | _(empty)_ | Extra shfmt flags, for example `-i 2`, `-ci`, `-sr`, or `-ln bash` |
| `GLOBAL_GO_BIN` | Go's `GOBIN` or `GOPATH/bin` | Directory where the shfmt binary is installed |
| `SHFMT_FMT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by formatting checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Notes

- `install` first ensures the Go toolchain is available through the local `go` Taskfile, then runs the [official installation command](https://github.com/mvdan/sh#shfmt).
- `fmt` uses `shfmt -w`; `fmt:check` uses `shfmt -d` and exits non-zero when formatting differs.
- `TARGETS` may be a single shell script or a directory. Pass dialect and style preferences through `EXTRA_ARGS`.
