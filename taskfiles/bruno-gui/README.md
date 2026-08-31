# bruno-gui

A [TaskOtter](https://github.com/task-otter/store) module for the [Bruno](https://www.usebruno.com/) desktop application.

## What is this Taskfile?

This module launches the Bruno desktop app and shows its built-in help. The `open` and `help` tasks auto-install `bruno` via `nix:install:profile` when it is not already on `PATH`.

## Usage

### Standalone

```sh
task -t taskfiles/bruno-gui/Taskfile.yml open
task -t taskfiles/bruno-gui/Taskfile.yml open BRUNO_GUI_COLLECTION=./api
task -t taskfiles/bruno-gui/Taskfile.yml help
```

Install only, without launching the app:

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#bruno
```

### Included in your Taskfile

```yaml
includes:
  bruno-gui:
    taskfile: taskfiles/bruno-gui/Taskfile.yml
    vars:
      BRUNO_GUI_COLLECTION_OVERRIDE: "{{.BRUNO_GUI_COLLECTION}}"
      BRUNO_GUI_EXTRA_ARGS_OVERRIDE: "{{.BRUNO_GUI_EXTRA_ARGS}}"
```

Then run:

```sh
task bruno-gui:open
task bruno-gui:open BRUNO_GUI_COLLECTION=./api
```

## Public Tasks

| Task | Description |
|---|---|
| `open` | Launch the Bruno desktop app (returns immediately on Unix) |
| `help` | Show Bruno desktop app help |

## Variables

| Variable | Default | Description |
|---|---|---|
| `BRUNO_GUI_NIX_INSTALLABLE` | `nixpkgs#bruno` | Flake installable passed to `nix:install:profile` |
| `BRUNO_GUI_COLLECTION` | `""` | Optional path to a Bruno collection directory |
| `BRUNO_GUI_EXTRA_ARGS` | `""` | Additional flags passed to `bruno` |

Pin a revision by overriding the installable, for example
`BRUNO_GUI_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#bruno`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported for auto-install; use WSL2 or ensure `bruno` is on `PATH`.
- On macOS and Linux, `open` launches Bruno in the background (`&`) so the task exits immediately.
