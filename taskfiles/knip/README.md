# Knip

## What is this Taskfile?

This Taskfile wraps Knip for unused file, export, and dependency analysis. Knip
can report framework-specific false positives, so treat output as review input
instead of an instruction to delete files or packages automatically.

## Variants

Every runtime + package-manager combination ships as its own leaf Taskfile under
`taskfiles/knip/`. They all expose the identical public interface documented
below; only the underlying runtime and package manager differ. Include the tool
family once in your root Taskfile:

```yaml
includes:
  knip: taskfiles/knip/Taskfile.yml
```

Then run the leaf that matches your project through its namespace (replace
`{TASK}` with any public task):

```bash
task knip:bun:{TASK}                 # Bun runtime + Bun as package manager
task knip:node:fnm:npm:{TASK}        # Node via fnm, npm as package manager
task knip:node:nvm:pnpm:{TASK}       # Node via nvm, pnpm as package manager
```

Available leaves: `bun`, `node/{fnm,nvm}/{npm,pnpm,yarn}`.

## Public Tasks

| Task               | Variables                                         | Description                                               |
| ------------------ | ------------------------------------------------- | --------------------------------------------------------- |
| `install`          | Optional `VERSION`, `EXTRA_ARGS`, `CLI_ARGS` | Install `knip` as a local dev dependency. Pass `VERSION=x.y.z` to pin a release. |
| `install:undo`     | Optional `EXTRA_ARGS`                       | Remove the locally installed `knip` devDependency.         |
| `upgrade`          | Optional `EXTRA_ARGS`                       | Reinstall `knip` at the latest version.                    |
| `init`             | Optional `EXTRA_ARGS`, `CLI_ARGS`           | Initialize Knip configuration.                            |
| `config:init`      | Optional `EXTRA_ARGS`, `CLI_ARGS`           | Alias for `init`.                                         |
| `lint`             | Optional `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run the default Knip analysis.                            |
| `production`       | Optional `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run Knip with `--production`.                             |
| `dependencies`     | Optional `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Report unused production dependencies.                    |
| `dev-dependencies` | Optional `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Report unused development dependencies.                   |
| `files`            | Optional `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Report unused files.                                      |
| `exports`          | Optional `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Report unused exports.                                    |
| `lint:fix`         | Optional `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run `knip --fix` when supported by the installed version. |
| `ci`               | Optional `CONFIG`, `EXTRA_ARGS`, `CLI_ARGS` | Run production checks for CI.                             |
| `version`          | — | Show the resolved Knip version.                           |
| `help`             | Optional `EXTRA_ARGS`, `CLI_ARGS`           | Show Knip CLI help.                                       |

## Variables

`--config <path>`. `EXTRA_ARGS` and arguments after `--` are appended to the
command.

Review Knip findings before deleting files or dependencies.

- `KNIP_LINT_SKIP_PATTERN` (default empty): forward-slash path glob for files skipped by lint checks and fixes.

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

Knip still analyzes its project graph; the generated configuration suppresses findings for files matching `KNIP_LINT_SKIP_PATTERN`.

When a skip pattern is set, TaskOtter can merge JSON, JSONC, and
`package.json#knip` configurations. Dynamic JavaScript and TypeScript Knip
configurations must add the pattern to their own `ignore` array.

## Examples

```bash
task knip:node:fnm:npm:install
task knip:node:fnm:npm:install VERSION=5.27.0
task knip:node:fnm:npm:lint
task knip:node:fnm:npm:production
task knip:node:fnm:npm:dependencies
task knip:node:fnm:npm:files
task knip:node:fnm:npm:exports
```
