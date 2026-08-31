# jq Taskfile Public Tasks

## What is this Taskfile?

A Taskfile module for [jq](https://jqlang.org), the lightweight command-line
JSON processor. This module does not ship an installer. Install jq through
the store's Nix profile task.

## Usage

### Standalone

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#jq
```

Or, targeting the nix Taskfile directly:

```sh
task -t taskfiles/nix/Taskfile.yml install:profile NIX_INSTALLABLE=nixpkgs#jq
```

### Included

```yaml
includes:
  jq: ./taskfiles/jq/Taskfile.yml
```

Then run:

```sh
task jq:nix:install:profile NIX_INSTALLABLE=nixpkgs#jq
```

Override `JQ_NIX_INSTALLABLE` to pin a flake (for example
`github:NixOS/nixpkgs/<rev>#jq`) and pass that value as `NIX_INSTALLABLE`.

## Public Tasks

This module has no public tasks. Install jq with `nix:install:profile`.

## Variables

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `JQ_NIX_INSTALLABLE` | `nixpkgs#jq` | Flake installable for `nix:install:profile` |

## Notes

`nix:install:profile` auto-installs Nix if it is missing and adds jq to the
user profile (`~/.nix-profile`). Native Windows is not supported; use WSL2.
