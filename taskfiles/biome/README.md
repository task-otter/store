# Biome

## What is this Taskfile?

This Taskfile wraps Biome for formatting, linting, combined checks, and CI. It
installs `@biomejs/biome` locally and delegates package-manager behavior to the

## Variants

Every runtime + package-manager combination ships as its own leaf Taskfile under
`taskfiles/biome/`. They all expose the identical public interface documented
below; only the underlying runtime and package manager differ. Include the tool
family once in your root Taskfile:

```yaml
includes:
  biome: taskfiles/biome/Taskfile.yml
```

Then run the leaf that matches your project through its namespace (replace
`{TASK}` with any public task):

```bash
task biome:bun:{TASK}                 # Bun runtime + Bun as package manager
task biome:node:fnm:npm:{TASK}        # Node via fnm, npm as package manager
task biome:node:nvm:pnpm:{TASK}       # Node via nvm, pnpm as package manager
```

Available leaves: `bun`, `node/{fnm,nvm}/{npm,pnpm,yarn}`.

## Public Tasks

| Task           | Variables                                                    | Description                                                                |
| -------------- | ------------------------------------------------------------ | -------------------------------------------------------------------------- |
| `install`      | Optional `BIOME_VERSION`, `BIOME_EXTRA_ARGS`, `CLI_ARGS`           | Install `@biomejs/biome` as a local dev dependency. Pass `BIOME_VERSION=x.y.z` to pin a release. |
| `install:undo` | Optional `BIOME_EXTRA_ARGS`                                  | Remove the locally installed `@biomejs/biome` devDependency.               |
| `upgrade`      | Optional `BIOME_EXTRA_ARGS`                                  | Reinstall `@biomejs/biome` at the latest version.                          |
| `init`         | Optional `BIOME_EXTRA_ARGS`, `CLI_ARGS`                      | Alias for `config:init`.                                                   |
| `config:init`  | Optional `BIOME_EXTRA_ARGS`, `CLI_ARGS`                      | Run `biome init`. Skipped if `biome.json` or `biome.jsonc` already exists. |
| `config:skip`  | Optional `SKIP_SCOPE`                                  | Write the skip-pattern config overlay. Run automatically by the tasks below. |
| `check`        | Optional `BIOME_TARGETS`, `BIOME_CONFIG`, `BIOME_EXTRA_ARGS`, `CLI_ARGS` | Run `biome check`.                                                         |
| `check:write`  | Optional `BIOME_TARGETS`, `BIOME_CONFIG`, `BIOME_EXTRA_ARGS`, `CLI_ARGS` | Run `biome check --write`.                                                 |
| `fix`          | Optional `BIOME_TARGETS`, `BIOME_CONFIG`, `BIOME_EXTRA_ARGS`, `CLI_ARGS` | Alias for `check:write`.                                                   |
| `lint`         | Optional `BIOME_TARGETS`, `BIOME_CONFIG`, `BIOME_EXTRA_ARGS`, `CLI_ARGS` | Run `biome lint`.                                                          |
| `lint:fix`     | Optional `BIOME_TARGETS`, `BIOME_CONFIG`, `BIOME_EXTRA_ARGS`, `CLI_ARGS` | Run `biome lint --write`.                                                  |
| `fmt:check`    | Optional `BIOME_TARGETS`, `BIOME_CONFIG`, `BIOME_EXTRA_ARGS`, `CLI_ARGS` | Run `biome format`.                                                        |
| `fmt`          | Optional `BIOME_TARGETS`, `BIOME_CONFIG`, `BIOME_EXTRA_ARGS`, `CLI_ARGS` | Run `biome format --write`.                                                |
| `ci`           | Optional `BIOME_TARGETS`, `BIOME_CONFIG`, `BIOME_EXTRA_ARGS`, `CLI_ARGS` | Run `biome ci`.                                                            |
| `cache:clean`  | —                                                            | Remove common Biome cache directories.                                     |
| `version`      | — | Show the resolved Biome version.                                           |
| `help`         | Optional `BIOME_EXTRA_ARGS`, `CLI_ARGS`                      | Show Biome CLI help.                                                       |

## Variables

and `BIOME_CONFIG` adds `--config-path <path>`.

`BIOME_EXTRA_ARGS` and arguments after `--` are appended to the command.

- `BIOME_LINT_SKIP_PATTERN` (default empty): forward-slash path glob for files skipped by lint checks and fixes.
- `BIOME_FMT_SKIP_PATTERN` (default empty): forward-slash path glob for files skipped by formatting checks and fixes.

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

When a skip pattern is set, `config:skip` writes `.taskotter-biome-<pm>-skip.json`
next to your `biome.json` — a config that extends it and negates the pattern —
and the task runs Biome against that overlay so the rest of your configuration
stays active. Pass `SKIP_SCOPE=lint` or `SKIP_SCOPE=fmt` to negate only one of
the two patterns; the default negates both. The overlay is rewritten on every
run and is not deleted afterwards, so **add `.taskotter-biome-*-skip.json` to
your `.gitignore`**. Running `config:skip` with no skip pattern set deletes it.

## Examples

```bash
task biome:node:fnm:npm:install
task biome:node:fnm:npm:install BIOME_VERSION=1.9.4
task biome:node:fnm:npm:config:init
task biome:node:fnm:npm:check
task biome:node:fnm:npm:check:write
task biome:node:fnm:npm:lint
task biome:node:fnm:npm:fmt
```
