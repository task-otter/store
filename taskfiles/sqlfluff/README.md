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
| `config:skip`  | Write the skip-pattern config overlay (run automatically by `ci`, `ci:fix`, and `parse`) |

## Variables

| Variable              | Default   | Description                                                  |
| --------------------- | --------- | ------------------------------------------------------------ |
| `SQLFLUFF_NIX_INSTALLABLE` | `nixpkgs#sqlfluff` | Flake installable passed to `nix:install:profile` |
| `SQLFLUFF_INTERNAL_SKIP_CONFIG` | `.taskotter-sqlfluff-skip.cfg` | Overlay config written by `config:skip` |
| `SQLFLUFF_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint, fix, and parse tasks |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

Pin a revision by overriding the installable, for example
`SQLFLUFF_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#sqlfluff`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.

**`config:init`** writes a `.sqlfluff` file in the current directory and is
skipped if the file already exists. To regenerate, delete `.sqlfluff` first.

**`config:skip`** merges your config with `SQLFLUFF_LINT_SKIP_PATTERN` appended
to `[sqlfluff] ignore_paths` and writes `.taskotter-sqlfluff-skip.cfg`, which
`ci`, `ci:fix`, and `parse` then pass via `--config` so the rest of your settings
stay active. The overlay is rewritten on every run and is not deleted
afterwards, so **add `.taskotter-sqlfluff-skip.cfg` to your `.gitignore`**.
Running `config:skip` with no skip pattern set deletes it.

**Dialect:** sqlfluff requires a dialect to lint most SQL. Either set `DIALECT_OVERRIDE`
on the CLI or declare it in `.sqlfluff` under `[sqlfluff] dialect = <name>`.
