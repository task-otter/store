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
task depcheck:node:fnm:npm:{TASK}        # Node via fnm, npm as package manager
task depcheck:node:nvm:pnpm:{TASK}       # Node via nvm, pnpm as package manager
```

Available leaves: `bun`, `node/{fnm,nvm}/{npm,pnpm,yarn}`.

## Public Tasks

| Task           | Variables                                                                             | Description                                   |
| -------------- | ------------------------------------------------------------------------------------- | --------------------------------------------- |
| `install`      | Optional `VERSION`, `EXTRA_ARGS`, `CLI_ARGS`                                    | Install `depcheck` as a local dev dependency. Pass `VERSION=x.y.z` to pin a release. |
| `install:undo` | Optional `EXTRA_ARGS`                                                           | Remove the locally installed `depcheck` devDependency. |
| `upgrade`      | Optional `EXTRA_ARGS`                                                           | Reinstall `depcheck` at the latest version.   |
| `lint`         | Optional `DEPCHECK_PROJECT_PATH`, `TARGETS`, `EXTRA_ARGS`, `CLI_ARGS`                    | Run Depcheck.                                 |
| `json`         | Optional `DEPCHECK_PROJECT_PATH`, `TARGETS`, `EXTRA_ARGS`, `CLI_ARGS`                    | Run Depcheck with `--json`.                   |
| `ignores`      | Optional `DEPCHECK_PROJECT_PATH`, `TARGETS`, `DEPCHECK_IGNORE_PACKAGES`, `EXTRA_ARGS`, `CLI_ARGS` | Run Depcheck with ignored packages.           |
| `skip-missing` | Optional `DEPCHECK_PROJECT_PATH`, `TARGETS`, `EXTRA_ARGS`, `CLI_ARGS`                    | Run Depcheck with `--skip-missing=true`.      |
| `ci`           | Optional `DEPCHECK_PROJECT_PATH`, `TARGETS`, `EXTRA_ARGS`, `CLI_ARGS`                    | Run Depcheck and fail on findings.            |
| `version`      | — | Show the resolved Depcheck version.           |
| `help`         | Optional `EXTRA_ARGS`, `CLI_ARGS`                                               | Show Depcheck CLI help.                       |

## Variables

`.` and can be overridden for monorepo packages. `TARGETS` is accepted as an
alias for the project path when used from aggregate tasks.

`DEPCHECK_IGNORE_PACKAGES` is a comma-separated list for the `ignores` task. `EXTRA_ARGS`
and arguments after `--` are appended to the command.

- `DEPCHECK_LINT_SKIP_PATTERN` (default empty): forward-slash path glob for files skipped by lint checks and fixes.

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Examples

```bash
task depcheck:node:fnm:npm:install
task depcheck:node:fnm:npm:install VERSION=1.4.7
task depcheck:node:fnm:npm:lint
task depcheck:node:fnm:npm:json
task depcheck:node:fnm:npm:lint DEPCHECK_PROJECT_PATH=packages/app
task depcheck:node:fnm:npm:lint -- --ignores="@types/*,eslint-*"
```
