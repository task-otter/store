# Bun Taskfile Public Tasks

## What is Bun?

Bun is an all-in-one JavaScript runtime and toolkit — a single binary that replaces Node.js, npm, a bundler, and a test runner. It is written in Zig and designed to be significantly faster than Node.js for startup, module resolution, and package installation.

This module installs Bun via Nix on Unix and WinGet on Windows. Tool Taskfiles that need the Bun CLI should depend on `bun:install` and invoke `bun` directly (for example `bun add -d`, `bun remove`, `bun x`).

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
task bun:install
```

Override `BUN_NIX_INSTALLABLE` to pin a flake (for example
`github:NixOS/nixpkgs/<rev>#bun`) and pass that value as `NIX_INSTALLABLE`.

## Public Tasks

| Task | Description |
|---|---|
| `install` | Install Bun via Nix (Unix) or WinGet (Windows) |
| `version` | Show the active Bun version |

Dependents auto-install Bun via `bun:install`.

## Variables

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `BUN_NIX_INSTALLABLE` | `nixpkgs#bun` | Flake installable for `nix:install:profile` |
| `BUN_WINGET_INSTALLABLE` | `Oven-sh.Bun` | WinGet package ID for `winget:install:package` |

## Notes

- Install uses Nix on Linux and macOS (`BUN_NIX_INSTALLABLE`) and WinGet on Windows (`BUN_WINGET_INSTALLABLE`, default `Oven-sh.Bun`).

