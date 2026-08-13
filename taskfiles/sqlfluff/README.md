# sqlfluff Taskfile

## What is this Taskfile?

A cross-platform Taskfile for installing sqlfluff, managing upgrades, linting
and auto-fixing SQL files, and generating a project configuration.

sqlfluff is installed via [uv](../uv/) into an isolated tool environment so it
never conflicts with project dependencies.

## Usage

### Standalone

```sh
task -t taskfiles/sqlfluff/Taskfile.yml install
task -t taskfiles/sqlfluff/Taskfile.yml config:init
task -t taskfiles/sqlfluff/Taskfile.yml ci
```

### Included

```yaml
includes:
  sqlfluff: ./taskfiles/sqlfluff/Taskfile.yml
```

Then run:

```sh
task sqlfluff:install
task sqlfluff:ci
task sqlfluff:lint:fix DIALECT_OVERRIDE=postgres
```

## Public Tasks

| Task           | Description                                     | Key variables                                |
| -------------- | ----------------------------------------------- | -------------------------------------------- |
| `install`      | Install sqlfluff on the current OS if missing   | none                                         |
| `install:undo` | Remove sqlfluff from the current OS             | none                                         |
| `upgrade`      | Upgrade sqlfluff to the latest release          | none                                         |
| `version`      | Show the installed sqlfluff version             | none                                         |
| `ci`         | Lint SQL files with sqlfluff                    | `TARGETS_OVERRIDE`, `CONFIG_OVERRIDE`, `DIALECT_OVERRIDE`, `EXTRA_ARGS_OVERRIDE` |
| `lint:fix`          | Auto-fix SQL lint violations                    | `TARGETS_OVERRIDE`, `CONFIG_OVERRIDE`, `DIALECT_OVERRIDE`, `EXTRA_ARGS_OVERRIDE` |
| `ci:fix` | Run `lint:fix` for CI fixing | — |
| `parse`        | Print the sqlfluff parse tree for SQL files     | `TARGETS_OVERRIDE`, `CONFIG_OVERRIDE`, `DIALECT_OVERRIDE`, `EXTRA_ARGS_OVERRIDE` |
| `config:init`  | Create a default `.sqlfluff` configuration file | none                                         |
| `config:skip`  | Write the skip-pattern config overlay (run automatically by `ci`, `lint:fix`, and `parse`) | `CONFIG_OVERRIDE` |

## Variables

| Variable              | Default   | Description                                                  |
| --------------------- | --------- | ------------------------------------------------------------ |
| `SQLFLUFF_VERSION`    | `4.2.2`   | Pinned sqlfluff version installed and enforced by `install`/`upgrade` |
| `TARGETS_OVERRIDE`    | _(empty)_ | Files or directories to lint/fix/parse (overrides task default `.`) |
| `CONFIG_OVERRIDE`     | _(empty)_ | Path to a sqlfluff config file passed via `--config`         |
| `DIALECT_OVERRIDE`    | _(empty)_ | SQL dialect passed via `--dialect` (e.g. `ansi`, `postgres`) |
| `EXTRA_ARGS_OVERRIDE` | _(empty)_ | Extra flags forwarded to sqlfluff                            |
| `SQLFLUFF_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint, fix, and parse tasks |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Notes

**`config:init`** writes a `.sqlfluff` file in the current directory and is
skipped if the file already exists. To regenerate, delete `.sqlfluff` first.

**`config:skip`** merges your config with `SQLFLUFF_LINT_SKIP_PATTERN` appended
to `[sqlfluff] ignore_paths` and writes `.taskotter-sqlfluff-skip.cfg`, which
`ci`, `lint:fix`, and `parse` then pass via `--config` so the rest of your settings
stay active. The overlay is rewritten on every run and is not deleted
afterwards, so **add `.taskotter-sqlfluff-skip.cfg` to your `.gitignore`**.
Running `config:skip` with no skip pattern set deletes it.

**Dialect:** sqlfluff requires a dialect to lint most SQL. Either set `DIALECT_OVERRIDE`
on the CLI or declare it in `.sqlfluff` under `[sqlfluff] dialect = <name>`.
