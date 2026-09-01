# dotenv-linter Taskfile

## What is this Taskfile?

This Taskfile wraps [dotenv-linter](https://dotenv-linter.github.io/), a
lightning-fast linter for .env files written in Rust. Remaining tasks
auto-install dotenv-linter via `nix:install:profile`.

## Usage

### Standalone

```bash
task --taskfile taskfiles/dotenv-linter/Taskfile.yml ci DOTENV_LINTER_TARGETS=.env.example
```

Install only:

```sh
task --taskfile taskfiles/dotenv-linter/Taskfile.yml install
```

### Included

```yaml
includes:
  dotenv-linter:
    taskfile: taskfiles/dotenv-linter/Taskfile.yml
```

```bash
task dotenv-linter:ci DOTENV_LINTER_TARGETS=.env.example
task dotenv-linter:ci:fix DOTENV_LINTER_TARGETS=.env
```

## Public Tasks

| Task | Description |
|---|---|
| `ci` | Lint dotenv files with dotenv-linter check |
| `ci:fix` | Apply automatic fixes with dotenv-linter fix |
| `diff` | Compare .env files to ensure matching key sets |
| `install` | Install dotenv-linter via the Nix profile |
| `version` | Show the active dotenv-linter version |

## Variables

| Variable | Default | Description |
|---|---|---|
| `DOTENV_LINTER_NIX_INSTALLABLE` | `nixpkgs#dotenv-linter` | Flake installable passed to `nix:install:profile` |
| `DOTENV_LINTER_TARGETS` | `.env` | File or directory dotenv-linter operates on |
| `DOTENV_LINTER_EXTRA_ARGS` | `""` | Extra flags forwarded to dotenv-linter (e.g. `--recursive`, `--skip`) |

Pin a revision by overriding the installable, for example
`DOTENV_LINTER_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#dotenv-linter`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.
- `ci:fix` writes changes in place; dotenv-linter creates a backup of each
  changed file.
- The tasks target the dotenv-linter 4.x CLI, which uses subcommands
  (`check`, `fix`, `diff`).
