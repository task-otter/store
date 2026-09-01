# Bun Taskfile Public Tasks

## What is Bun?

Bun is an all-in-one JavaScript runtime and toolkit — a single binary that replaces Node.js, npm, a bundler, and a test runner. It is written in Zig and designed to be significantly faster than Node.js for startup, module resolution, and package installation.

This module does not ship an installer. Install Bun through the store's Nix profile task. Tool Taskfiles that need the Bun CLI should depend on `bun:install` and invoke `bun` directly (for example `bun add -d`, `bun remove`, `bun x`).

## Usage

### Standalone

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#bun
```

Or, targeting the nix Taskfile directly:

```sh
task -t taskfiles/nix/Taskfile.yml install:profile NIX_INSTALLABLE=nixpkgs#bun
```

### Included

```yaml
includes:
  bun: ./taskfiles/bun/Taskfile.yml
```

Then run:

```sh
task bun:nix:install:profile NIX_INSTALLABLE=nixpkgs#bun
```

Override `BUN_NIX_INSTALLABLE` to pin a flake (for example
`github:NixOS/nixpkgs/<rev>#bun`) and pass that value as `NIX_INSTALLABLE`.

## Public Tasks

| Task | Description |
|---|---|
| `install` | Install Bun via the Nix profile |
| `version` | Show the active Bun version |

Dependents auto-install Bun via `bun:install`.

## Variables

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `BUN_NIX_INSTALLABLE` | `nixpkgs#bun` | Flake installable for `nix:install:profile` |

## Notes

`nix:install:profile` auto-installs Nix if it is missing and adds Bun to the
user profile (`~/.nix-profile`). Native Windows is not supported; use WSL2.
