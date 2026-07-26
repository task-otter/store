# ESLint

## What is this Taskfile?

This Taskfile wraps ESLint for JavaScript and TypeScript projects. It installs
ESLint as a local dev dependency, runs cached checks by default, supports strict
CI mode.

This variant uses the `npm-fnm` stack (`npm-fnm`) package manager.


## Variants

Every runtime + package-manager combination ships as its own leaf Taskfile under
`taskfiles/eslint/`. They all expose the identical public interface documented
below; only the underlying runtime and package manager differ. Include the tool
family once in your root Taskfile:

```yaml
includes:
  eslint: taskfiles/eslint/Taskfile.yml
```

Then run the leaf that matches your project through its namespace (replace
`{TASK}` with any public task):

```bash
task eslint:bun:{TASK}                 # Bun runtime + Bun as package manager
task eslint:node:fnm:npm:{TASK}        # Node via fnm, npm as package manager
task eslint:node:nvm:pnpm:{TASK}       # Node via nvm, pnpm as package manager
```

Available leaves: `bun`, `node/{fnm,nvm}/{npm,pnpm,yarn}`.

## Public Tasks

| Task          | Variables                                                             | Description                                                                              |
| ------------- | --------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `install`     | Optional `VERSION`, `EXTRA_ARGS`, `CLI_ARGS`                    | Install `eslint` as a local dev dependency. Pass `VERSION=x.y.z` to pin a release. |
| `install:undo`| Optional `EXTRA_ARGS`                                           | Remove the locally installed `eslint` devDependency.                                     |
| `upgrade`     | Optional `EXTRA_ARGS`                                           | Reinstall `eslint` at the latest version.                                                |
| `init`        | Optional `EXTRA_ARGS`, `CLI_ARGS`                               | Alias for `config:init`.                                                                 |
| `config:init` | Optional `EXTRA_ARGS`, `CLI_ARGS`                               | Run the ESLint configuration wizard. Skipped if a recognized config file already exists. |
| `lint`        | Optional `TARGETS`, `CONFIG`, `CACHE`, `EXTRA_ARGS`, `CLI_ARGS` | Lint targets with cache enabled by default.                                              |
| `lint:fix`    | Optional `TARGETS`, `CONFIG`, `CACHE`, `EXTRA_ARGS`, `CLI_ARGS` | Run ESLint with `--fix`.                                                                 |
| `ci`          | Optional `TARGETS`, `CONFIG`, `CACHE`, `EXTRA_ARGS`, `CLI_ARGS` | Run ESLint with `--max-warnings=0`.                                                      |
| `cache:clean` | —                                                                     | Remove `.cache/eslint`.                                                                  |
| `version`     | — | Show the resolved ESLint version.                                                        |
| `help`        | Optional `EXTRA_ARGS`, `CLI_ARGS`                               | Show ESLint CLI help.                                                                    |

## Variables

`TARGETS` defaults to `src/**/*.{js,jsx,ts,tsx}`. `CONFIG` adds
`--config <path>`. `CACHE` defaults to `true`; set `CACHE=false` to omit cache
flags. `EXTRA_ARGS` and arguments after `--` are appended to the command.

- `ESLINT_LINT_SKIP_PATTERN` (default empty): forward-slash path glob for files skipped by lint checks and fixes.

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Examples

```bash
task eslint:node:fnm:npm:install
task eslint:node:fnm:npm:install VERSION=8.57.0
task eslint:node:fnm:npm:lint
task eslint:node:fnm:npm:lint:fix TARGETS="src test"
```
