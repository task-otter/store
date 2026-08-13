# Buf Taskfile

## What is this Taskfile?

A cross-platform Taskfile for linting, formatting, breaking-change detection,
and code generation from [Protocol Buffer](https://protobuf.dev/) definitions
using [Buf](https://buf.build/), the modern proto toolchain.

Buf is installed globally via a pinned GitHub release binary to `/usr/local/bin`
on macOS and Linux, and Scoop on Windows. The install task is skipped
automatically when Buf is already present in PATH at the pinned version.

## Usage

### Standalone

```sh
task -t taskfiles/buf/Taskfile.yml install
task -t taskfiles/buf/Taskfile.yml lint
task -t taskfiles/buf/Taskfile.yml fmt:check
task -t taskfiles/buf/Taskfile.yml version
```

Lint a specific proto directory:

```sh
task -t taskfiles/buf/Taskfile.yml lint BUF_INPUT=api/v1
```

Check for breaking changes against a branch:

```sh
task -t taskfiles/buf/Taskfile.yml breaking BUF_AGAINST=.git#branch=main
```

Pass buf flags directly after `--`:

```sh
task -t taskfiles/buf/Taskfile.yml lint -- api/v1 --error-format json
```

### Included

```yaml
includes:
  buf:
    taskfile: taskfiles/buf/Taskfile.yml
    vars:
      BUF_CONFIG_OVERRIDE: "{{.BUF_CONFIG}}"
      BUF_INPUT_OVERRIDE: "{{.BUF_INPUT}}"
      BUF_AGAINST_OVERRIDE: "{{.BUF_AGAINST}}"
      BUF_EXTRA_ARGS_OVERRIDE: "{{.BUF_EXTRA_ARGS}}"
```

Then run:

```sh
task buf:lint
task buf:fmt:check
task buf:breaking BUF_AGAINST=.git#branch=main
task buf:generate BUF_INPUT=api/v1
```

## Public Tasks

| Task            | Description                                              | Key variables                  |
| --------------- | -------------------------------------------------------- | ------------------------------ |
| `breaking`      | Check proto files for breaking changes against AGAINST   | `BUF_INPUT`, `BUF_AGAINST`, `BUF_EXTRA_ARGS` |
| `ci`            | Run `fmt:check` then `lint`                              | `BUF_INPUT`, `BUF_CONFIG`, `BUF_EXTRA_ARGS` |
| `ci:fix` | Run `fmt` for CI fixing | — |
| `fmt:check`     | Check proto file formatting with Buf                     | `BUF_INPUT`, `BUF_EXTRA_ARGS`          |
| `fmt`           | Format proto files in place with Buf                     | `BUF_INPUT`, `BUF_EXTRA_ARGS`          |
| `generate`      | Generate code from proto files with Buf                  | `BUF_INPUT`, `BUF_EXTRA_ARGS`          |
| `install`       | Install Buf on the current operating system              | none                           |
| `install:undo`  | Remove Buf from the current operating system             | none                           |
| `lint`          | Lint proto files with Buf                                | `BUF_INPUT`, `BUF_CONFIG`, `BUF_EXTRA_ARGS` |
| `upgrade`       | Upgrade Buf to the latest release                        | `BUF_VERSION` (Linux; brew on macOS) |
| `version`       | Show the installed Buf version                           | none                           |

## Variables

| Variable      | Default              | Description                                              |
| ------------- | -------------------- | -------------------------------------------------------- |
| `BUF_AGAINST`     | `.git#branch=main`   | Baseline for `breaking`: a git ref, Buf module, or path |
| `BUF_VERSION` | `1.47.2`             | Buf release to download on macOS and Linux              |
| `BUF_CONFIG`      | empty                | Path to a `buf.yaml` config file passed via `--config`  |
| `BUF_EXTRA_ARGS`  | empty                | Extra arguments appended when `CLI_ARGS` is not provided |
| `BUF_INPUT`       | `.`                  | Proto source directory or Buf module passed to buf       |
| `BUF_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint and breaking checks |
| `BUF_FMT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by formatting checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Notes

On macOS and Linux, Buf is installed from the pinned `BUF_VERSION` GitHub
release binary into `/usr/local/bin`. macOS supports `x86_64` and `arm64`;
Linux supports `x86_64` and `aarch64`. Other architectures require a manual
installation; see the [Buf installation docs](https://buf.build/docs/installation).
`BUF_VERSION` is ignored on Windows — Scoop controls the installed version.
macOS `upgrade` still uses Homebrew (`brew upgrade bufbuild/buf/buf`).

The `generate` task requires a `buf.gen.yaml` file in the working tree. See the
[buf generate docs](https://buf.build/docs/generate/tutorial) for configuration
details.
