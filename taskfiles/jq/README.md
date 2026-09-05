# jq Taskfile Public Tasks

## What is this Taskfile?

A Taskfile module for [jq](https://jqlang.org), the lightweight command-line
JSON processor. The `install` task adds jq to the user Nix profile; `version`
reports the resolved binary's version.

## Usage

### Standalone

```sh
task -t taskfiles/jq/Taskfile.yml install
task -t taskfiles/jq/Taskfile.yml version
```

### Included

```yaml
includes:
  jq: ./taskfiles/jq/Taskfile.yml
```

Then run:

```sh
task jq:install
task jq:version
```

Override `JQ_NIX_INSTALLABLE` to pin a flake, for example
`github:NixOS/nixpkgs/<rev>#jq`.

## Public Tasks

| Task | Description |
|---|---|
| `install` | Install jq via Nix (Unix) or WinGet (Windows) |
| `version` | Show the active jq version |

## Variables

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `JQ_NIX_INSTALLABLE` | `nixpkgs#jq` | Flake installable for `nix:install:profile` |
| `JQ_WINGET_INSTALLABLE` | `jqlang.jq` | WinGet package ID for `winget:install:package` |

## Notes

- Install uses Nix on Linux and macOS (`JQ_NIX_INSTALLABLE`) and WinGet on Windows (`JQ_WINGET_INSTALLABLE`, default `jqlang.jq`).

