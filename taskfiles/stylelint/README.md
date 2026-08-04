# Stylelint

## What is this Taskfile?

This Taskfile wraps Stylelint for stylesheet checks and fixes. It installs
`stylelint` and `stylelint-config-standard` locally, enables cache by default,

## Variants

Every runtime + package-manager combination ships as its own leaf Taskfile under
`taskfiles/stylelint/`. They all expose the identical public interface documented
below; only the underlying runtime and package manager differ. Include the tool
family once in your root Taskfile:

```yaml
includes:
  stylelint: taskfiles/stylelint/Taskfile.yml
```

Then run the leaf that matches your project through its namespace (replace
`{TASK}` with any public task):

```bash
task stylelint:bun:{TASK}                 # Bun runtime + Bun as package manager
task stylelint:node:fnm:npm:{TASK}        # Node via fnm, npm as package manager
task stylelint:node:nvm:pnpm:{TASK}       # Node via nvm, pnpm as package manager
```

Available leaves: `bun`, `node/{fnm,nvm}/{npm,pnpm,yarn}`.

## Public Tasks

| Task          | Variables                                                                                  | Description                                                          |
| ------------- | ------------------------------------------------------------------------------------------ | -------------------------------------------------------------------- |
| `install`     | Optional `STYLELINT_VERSION`, `STYLELINT_EXTRA_ARGS`, `CLI_ARGS`                                         | Install Stylelint and the standard config as local dev dependencies. Pass `STYLELINT_VERSION=x.y.z` to pin a release. |
| `install:undo`| Optional `STYLELINT_EXTRA_ARGS`                                                                | Remove the locally installed Stylelint devDependencies.              |
| `upgrade`     | Optional `STYLELINT_EXTRA_ARGS`                                                                | Reinstall Stylelint and the standard config at their latest versions. |
| `config:init` | Optional `STYLELINT_CONFIG`                                                                          | Create a starter Stylelint config when one does not exist.           |
| `lint`        | Optional `STYLELINT_TARGETS`, `STYLELINT_CONFIG`, `STYLELINT_CACHE`, `STYLELINT_ALLOW_EMPTY_INPUT`, `STYLELINT_EXTRA_ARGS`, `CLI_ARGS` | Lint stylesheet targets.                                             |
| `lint:fix`    | Optional `STYLELINT_TARGETS`, `STYLELINT_CONFIG`, `STYLELINT_CACHE`, `STYLELINT_ALLOW_EMPTY_INPUT`, `STYLELINT_EXTRA_ARGS`, `CLI_ARGS` | Run Stylelint with `--fix`.                                          |
| `ci`          | Optional `STYLELINT_TARGETS`, `STYLELINT_CONFIG`, `STYLELINT_CACHE`, `STYLELINT_ALLOW_EMPTY_INPUT`, `STYLELINT_EXTRA_ARGS`, `CLI_ARGS` | Run Stylelint with `--max-warnings=0`.                               |
| `cache:clean` | —                                                                                          | Remove `.cache/stylelint`.                                           |
| `version`     | — | Show the resolved Stylelint version.                                 |
| `help`        | Optional `STYLELINT_EXTRA_ARGS`, `CLI_ARGS`                                                    | Show Stylelint CLI help.                                             |

## Variables

`**/*.{css,scss,sass,less,vue,svelte,astro}`. `STYLELINT_CONFIG` adds `--config <path>`.

`STYLELINT_CACHE` and `STYLELINT_ALLOW_EMPTY_INPUT` both default to `true`; set either to `false`
to omit that flag. `STYLELINT_EXTRA_ARGS` and arguments after `--` are appended to the
command.

- `STYLELINT_LINT_SKIP_PATTERN` (default empty): forward-slash path glob for files skipped by lint checks and fixes.

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Examples

```bash
task stylelint:node:fnm:npm:install
task stylelint:node:fnm:npm:install STYLELINT_VERSION=16.6.1
task stylelint:node:fnm:npm:lint
task stylelint:node:fnm:npm:lint:fix STYLELINT_TARGETS="src/**/*.scss"
```
