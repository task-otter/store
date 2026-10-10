# Dasel Taskfile Public Tasks

## What Is This Taskfile?

A Taskfile module for [Dasel](https://daseldocs.tomwright.me/), a command-line
tool for querying, modifying, and converting structured data. Installation
uses the shared Nix profile module on Linux and macOS. On Windows, it uses
the shared Go module to build and install Dasel; Go is installed via WinGet
when needed.

## Usage

### Standalone

```sh
task -t taskfiles/dasel/Taskfile.yml install
task -t taskfiles/dasel/Taskfile.yml version
```

### Included

```yaml
includes:
  dasel: ./taskfiles/dasel/Taskfile.yml
```

Then run:

```sh
task dasel:install
task dasel:version
```

Override `DASEL_NIX_INSTALLABLE` to pin a flake, for example
`github:NixOS/nixpkgs/<rev>#dasel`.

On Windows, override `DASEL_GO_PKG` to pin the Go package version (default:
`github.com/tomwright/dasel/v3/cmd/dasel@latest`).

## Public Tasks

| Task      | Description                                  |
| --------- | -------------------------------------------- |
| `install` | Install Dasel via Nix (Unix) or Go (Windows) |
| `version` | Show the active Dasel version                |

## Variables

| Variable                | Default                                          | Description                                 |
| ----------------------- | ------------------------------------------------ | ------------------------------------------- |
| `DASEL_NIX_INSTALLABLE` | `nixpkgs#dasel`                                  | Flake installable for `nix:install:profile` |
| `DASEL_GO_PKG`          | `github.com/tomwright/dasel/v3/cmd/dasel@latest` | Go package for `go:install:pkg` on Windows  |

## Notes

* These tasks support Linux, macOS, and native Windows. Windows uses
  PowerShell and reloads the Go binary directory into PATH after installation.
* `install` skips installation when Dasel is already on PATH; `version`
  depends on `install`.
