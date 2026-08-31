# sqlfluff Taskfile

## What is this Taskfile?

A cross-platform Taskfile for linting and auto-fixing SQL files and generating
a project configuration. Remaining tasks auto-install sqlfluff via
`nix:install:profile`.

## Usage

### Standalone

```sh
task -t taskfiles/sqlfluff/Taskfile.yml config:init
task -t taskfiles/sqlfluff/Taskfile.yml ci
```

Install only:

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#sqlfluff
```

### Included

```yaml
includes:
  sqlfluff: ./taskfiles/sqlfluff/Taskfile.yml
```

Then run:

```sh
task sqlfluff:ci
task sqlfluff:ci:fix DIALECT_OVERRIDE=postgres
```

## Public Tasks

| Task           | Description                                     |
| -------------- | ----------------------------------------------- |
| `ci`           | Lint SQL files with sqlfluff                    |
| `ci:fix`       | Auto-fix SQL lint violations |
| `parse`        | Print the sqlfluff parse tree for SQL files     |
| `config:init`  | Create a default `.sqlfluff` configuration file |

## Variables

| Variable              | Default   | Description                                                  |
| --------------------- | --------- | ------------------------------------------------------------ |
| `SQLFLUFF_NIX_INSTALLABLE` | `nixpkgs#sqlfluff` | Flake installable passed to `nix:install:profile` |

Pin a revision by overriding the installable, for example
`SQLFLUFF_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#sqlfluff`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.

**`config:init`** writes a `.sqlfluff` file in the current directory and is
skipped if the file already exists. To regenerate, delete `.sqlfluff` first.

**Dialect:** sqlfluff requires a dialect to lint most SQL. Either set `DIALECT_OVERRIDE`
on the CLI or declare it in `.sqlfluff` under `[sqlfluff] dialect = <name>`.
