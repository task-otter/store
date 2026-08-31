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
task knip:node:npm:{TASK}        # Node via the nodejs module, npm as package manager
task knip:node:pnpm:{TASK}       # Node via the nodejs module, pnpm as package manager
```

Available leaves: `bun`, `node/{npm,pnpm,yarn}`.

## Public Tasks

| Task               | Variables                                         | Description                                               |
| ------------------ | ------------------------------------------------- | --------------------------------------------------------- |
| `install`          | Optional `KNIP_VERSION`, `KNIP_EXTRA_ARGS`, `CLI_ARGS` | Install `knip` as a local dev dependency. Pass `KNIP_VERSION=x.y.z` to pin a release. |
| `install:undo`     | Optional `KNIP_EXTRA_ARGS`                       | Remove the locally installed `knip` devDependency.         |
| `upgrade`          | Optional `KNIP_EXTRA_ARGS`                       | Reinstall `knip` at the latest version.                    |
| `config:init`      | Optional `KNIP_EXTRA_ARGS`, `CLI_ARGS`           | Initialize Knip configuration (writes starter `knip.json`). |
| `production`       | Optional `KNIP_CONFIG`, `KNIP_EXTRA_ARGS`, `CLI_ARGS` | Run Knip with `--production`.                             |
| `dependencies`     | Optional `KNIP_CONFIG`, `KNIP_EXTRA_ARGS`, `CLI_ARGS` | Report unused production dependencies.                    |
| `dev-dependencies` | Optional `KNIP_CONFIG`, `KNIP_EXTRA_ARGS`, `CLI_ARGS` | Report unused development dependencies.                   |
| `files`            | Optional `KNIP_CONFIG`, `KNIP_EXTRA_ARGS`, `CLI_ARGS` | Report unused files.                                      |
| `exports`          | Optional `KNIP_CONFIG`, `KNIP_EXTRA_ARGS`, `CLI_ARGS` | Report unused exports.                                    |
| `ci`               | Optional `KNIP_CONFIG`, `KNIP_EXTRA_ARGS`, `CLI_ARGS` | Run production checks for CI.                             |
| `ci:fix` | Optional `KNIP_CONFIG`, `KNIP_EXTRA_ARGS`, `CLI_ARGS` | Run `knip --fix` when supported by the installed version. |
| `version`          | — | Show the resolved Knip version.                           |
| `help`             | Optional `KNIP_EXTRA_ARGS`, `CLI_ARGS`           | Show Knip CLI help.                                       |

## Variables

`--config <path>`. `KNIP_EXTRA_ARGS` and arguments after `--` are appended to the
command.

Review Knip findings before deleting files or dependencies.

## Examples

```bash
task knip:node:npm:install
task knip:node:npm:install KNIP_VERSION=5.27.0
task knip:node:npm:ci
task knip:node:npm:production
task knip:node:npm:dependencies
task knip:node:npm:files
task knip:node:npm:exports
```
