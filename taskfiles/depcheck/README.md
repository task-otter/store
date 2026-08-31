# Depcheck

## What is this Taskfile?

This Taskfile wraps Depcheck for unused and missing dependency reports. It runs
against the project root by default and uses the project's package manager for local
binary execution.

## Variants

Every runtime + package-manager combination ships as its own leaf Taskfile under
`taskfiles/depcheck/`. They all expose the identical public interface documented
below; only the underlying runtime and package manager differ. Include the tool
family once in your root Taskfile:

```yaml
includes:
  depcheck: taskfiles/depcheck/Taskfile.yml
```

Then run the leaf that matches your project through its namespace (replace
`{TASK}` with any public task):

```bash
task depcheck:bun:{TASK}                 # Bun runtime + Bun as package manager
task depcheck:node:npm:{TASK}        # Node via the nodejs module, npm as package manager
task depcheck:node:pnpm:{TASK}       # Node via the nodejs module, pnpm as package manager
```

Available leaves: `bun`, `node/{npm,pnpm,yarn}`.

## Public Tasks

| Task           | Variables                                                                             | Description                                   |
| -------------- | ------------------------------------------------------------------------------------- | --------------------------------------------- |
| `install`      | Optional `DEPCHECK_VERSION`, `DEPCHECK_EXTRA_ARGS`, `CLI_ARGS`                                    | Install `depcheck` as a local dev dependency. Pass `DEPCHECK_VERSION=x.y.z` to pin a release. |
| `install:undo` | Optional `DEPCHECK_EXTRA_ARGS`                                                           | Remove the locally installed `depcheck` devDependency. |
| `upgrade`      | Optional `DEPCHECK_EXTRA_ARGS`                                                           | Reinstall `depcheck` at the latest version.   |
| `json`         | Optional `DEPCHECK_PROJECT_PATH`, `DEPCHECK_TARGETS`, `DEPCHECK_EXTRA_ARGS`, `CLI_ARGS`                    | Run Depcheck with `--json`.                   |
| `ignores`      | Optional `DEPCHECK_PROJECT_PATH`, `DEPCHECK_TARGETS`, `DEPCHECK_IGNORE_PACKAGES`, `DEPCHECK_EXTRA_ARGS`, `CLI_ARGS` | Run Depcheck with ignored packages.           |
| `skip-missing` | Optional `DEPCHECK_PROJECT_PATH`, `DEPCHECK_TARGETS`, `DEPCHECK_EXTRA_ARGS`, `CLI_ARGS`                    | Run Depcheck with `--skip-missing=true`.      |
| `ci`           | Optional `DEPCHECK_PROJECT_PATH`, `DEPCHECK_TARGETS`, `DEPCHECK_EXTRA_ARGS`, `CLI_ARGS`                    | Run Depcheck and fail on findings.            |
| `version`      | — | Show the resolved Depcheck version.           |
| `help`         | Optional `DEPCHECK_EXTRA_ARGS`, `CLI_ARGS`                                               | Show Depcheck CLI help.                       |

## Variables

`.` and can be overridden for monorepo packages. `DEPCHECK_TARGETS` is accepted as an
alias for the project path when used from aggregate tasks.

`DEPCHECK_IGNORE_PACKAGES` is a comma-separated list for the `ignores` task. `DEPCHECK_EXTRA_ARGS`
and arguments after `--` are appended to the command.

## Examples

```bash
task depcheck:node:npm:install
task depcheck:node:npm:install DEPCHECK_VERSION=1.4.7
task depcheck:node:npm:ci
task depcheck:node:npm:json
task depcheck:node:npm:ci DEPCHECK_PROJECT_PATH=packages/app
task depcheck:node:npm:ci -- --ignores="@types/*,eslint-*"
```
