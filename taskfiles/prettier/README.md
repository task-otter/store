# Prettier

## What is this Taskfile?

This Taskfile wraps Prettier checks and writes for JavaScript/TypeScript
projects and workspaces. It uses the project's package manager for local binary
execution and package-manager detection.

## Variants

Every runtime + package-manager combination ships as its own leaf Taskfile under
`taskfiles/prettier/`. They all expose the identical public interface documented
below; only the underlying runtime and package manager differ. Include the tool
family once in your root Taskfile:

```yaml
includes:
  prettier: taskfiles/prettier/Taskfile.yml
```

Then run the leaf that matches your project through its namespace (replace
`{TASK}` with any public task):

```bash
task prettier:bun:{TASK}                 # Bun runtime + Bun as package manager
task prettier:node:fnm:npm:{TASK}        # Node via fnm, npm as package manager
task prettier:node:nvm:pnpm:{TASK}       # Node via nvm, pnpm as package manager
```

Available leaves: `bun`, `node/{fnm,nvm}/{npm,pnpm,yarn}`.

## Public Tasks

| Task          | Variables                                                                   | Description                                               |
| ------------- | --------------------------------------------------------------------------- | --------------------------------------------------------- |
| `install`     | Optional `PRETTIER_VERSION`, `PRETTIER_EXTRA_ARGS`, `CLI_ARGS`                          | Install `prettier` as a local dev dependency. Pass `PRETTIER_VERSION=x.y.z` to pin a release. |
| `install:undo`| Optional `PRETTIER_EXTRA_ARGS`                                                 | Remove the locally installed `prettier` devDependency.    |
| `upgrade`     | Optional `PRETTIER_EXTRA_ARGS`                                                 | Reinstall `prettier` at the latest version.                |
| `config:init` | Optional `PRETTIER_CONFIG`                                                           | Create a starter Prettier config when one does not exist. |
| `fmt:check`   | Optional `PRETTIER_TARGETS`, `PRETTIER_CONFIG`, `PRETTIER_IGNORE_PATH`, `PRETTIER_EXTRA_ARGS`, `CLI_ARGS` | Run `prettier --check`.                                   |
| `fmt`         | Optional `PRETTIER_TARGETS`, `PRETTIER_CONFIG`, `PRETTIER_IGNORE_PATH`, `PRETTIER_EXTRA_ARGS`, `CLI_ARGS` | Run `prettier --write`.                                   |
| `fix`         | Optional `PRETTIER_TARGETS`, `PRETTIER_CONFIG`, `PRETTIER_IGNORE_PATH`, `PRETTIER_EXTRA_ARGS`, `CLI_ARGS` | Alias for `fmt`.                                           |
| `ci`          | Optional `PRETTIER_TARGETS`, `PRETTIER_CONFIG`, `PRETTIER_IGNORE_PATH`, `PRETTIER_EXTRA_ARGS`, `CLI_ARGS` | Alias for `fmt:check`.                                     |
| `version`     | — | Show the resolved Prettier version.                       |
| `help`        | Optional `PRETTIER_EXTRA_ARGS`, `CLI_ARGS`                                     | Show Prettier CLI help.                                   |

## Variables

`PRETTIER_CONFIG` adds `--config <path>`, and `PRETTIER_IGNORE_PATH` defaults to
`.prettierignore`. The ignore path is only passed when the file exists.

`PRETTIER_EXTRA_ARGS` and arguments after `--` are appended to the command.

- `PRETTIER_FMT_SKIP_PATTERN` (default empty): forward-slash path glob for files skipped by formatting checks and fixes.

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Examples

```bash
task prettier:node:fnm:npm:install
task prettier:node:fnm:npm:install PRETTIER_VERSION=3.3.3
task prettier:node:fnm:npm:fmt:check
task prettier:node:fnm:npm:fmt PRETTIER_TARGETS="src/**/*.ts"
```
