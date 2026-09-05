# Node.js Taskfile

## What is this module?

Installs Node.js via Nix on Unix (`nixpkgs#nodejs` by default) and WinGet on Windows (`OpenJS.NodeJS`).
Package managers (`npm`, `yarn`, `pnpm`) and JS tool Taskfiles depend on
`nodejs:install` before running Node-backed commands.

## Usage

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#nodejs
```

Or include this module and depend on `nodejs:install`.

## Public Tasks

| Task | Description |
|---|---|
| `install` | Install Node.js via Nix (Unix) or WinGet (Windows) |
| `version` | Show the active Node.js version |

Dependents auto-install Node.js via `nodejs:install`.

## Variables

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `NODEJS_NIX_INSTALLABLE` | `nixpkgs#nodejs` | Flake installable for `nix:install:profile` |
| `NODEJS_WINGET_INSTALLABLE` | `OpenJS.NodeJS` | WinGet package ID for `winget:install:package` |

## Notes

- Install uses Nix on Linux and macOS (`NODEJS_NIX_INSTALLABLE`) and WinGet on Windows (`NODEJS_WINGET_INSTALLABLE`, default `OpenJS.NodeJS`).

