# Node.js Taskfile

## What is this module?

Installs Node.js via the Nix user profile (`nixpkgs#nodejs` by default).
Package managers (`npm`, `yarn`, `pnpm`) and JS tool Taskfiles depend on
`nodejs:_ensure` before running Node-backed commands.

## Usage

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#nodejs
```

Or include this module and depend on `nodejs:_ensure`.

## Public Tasks

This module has no public tasks. Dependents auto-install via `nodejs:_ensure`.

## Variables

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `NODEJS_NIX_INSTALLABLE` | `nixpkgs#nodejs` | Flake installable for `nix:install:profile` |

## Notes

Pin Node.js by overriding `NODEJS_NIX_INSTALLABLE`. Native Windows auto-install
requires WSL2 (same as other Nix profile modules).
