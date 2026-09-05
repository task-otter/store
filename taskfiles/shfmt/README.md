# shfmt

A [TaskOtter](https://github.com/task-otter/store) module for [shfmt](https://github.com/mvdan/sh), the shell formatter supporting POSIX shell, Bash, Zsh, and mksh.

## What is this Taskfile?

This module formats shell scripts in place or reports formatting differences. `ci:fix` and `fmt:check` auto-install shfmt via `nix:install:profile`.

## Usage

### Standalone

```sh
task -t taskfiles/shfmt/Taskfile.yml ci:fix SHFMT_TARGETS=scripts
task -t taskfiles/shfmt/Taskfile.yml fmt:check SHFMT_TARGETS=scripts SHFMT_EXTRA_ARGS="-i 2 -ci"
```

Install only, without formatting:

```sh
task -t taskfiles/shfmt/Taskfile.yml install
```

### Included in your Taskfile

```yaml
includes:
  shfmt:
    taskfile: taskfiles/shfmt/Taskfile.yml
```

Then run:

```sh
task shfmt:ci:fix
task shfmt:fmt:check SHFMT_TARGETS=scripts
```

## Public Tasks

| Task | Description |
|---|---|
| `ci:fix` | Format shell scripts in place (`SHFMT_TARGETS=path`) |
| `fmt:check` | Check shell script formatting without modifying files (`SHFMT_TARGETS=path`) |
| `install` | Install shfmt via Nix (Unix) or WinGet (Windows) |
| `version` | Show the active shfmt version |

## Variables

| Variable | Default | Description |
|---|---|---|
| `SHFMT_NIX_INSTALLABLE` | `nixpkgs#shfmt` | Flake installable passed to `nix:install:profile` |
| `SHFMT_WINGET_INSTALLABLE` | `mvdan.shfmt` | WinGet package ID for `winget:install:package` |
| `SHFMT_TARGETS` | `.` | File or directory to format or check |
| `SHFMT_EXTRA_ARGS` | _(empty)_ | Extra shfmt flags, for example `-i 2`, `-ci`, `-sr`, or `-ln bash` |

Pin a revision by overriding the installable, for example
`SHFMT_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#shfmt`.

## Notes

- Install uses Nix on Linux and macOS (`SHFMT_NIX_INSTALLABLE`) and WinGet on Windows (`SHFMT_WINGET_INSTALLABLE`, default `mvdan.shfmt`).

- `ci:fix` uses `shfmt -w`; `fmt:check` uses `shfmt -d` and exits non-zero when formatting differs.
- `SHFMT_TARGETS` may be a single shell script or a directory. Pass dialect and style preferences through `SHFMT_EXTRA_ARGS`.
