# Yarn Taskfile Public Tasks

This Taskfile wraps common modern Yarn project work while reusing the same Node.js loading model as the npm Taskfile. Yarn commands run through
Corepack so the project's `packageManager` field controls the Yarn version.

## Variants

Yarn ships as two variants, one per node version manager. Both expose exactly the
same tasks; pick the one matching your stack.

| Variant | Node version manager | Taskfile                          |
| ------- | -------------------- | --------------------------------- |
| `fnm`   | [fnm](../fnm/)       | `taskfiles/yarn/fnm/Taskfile.yml` |
| `nvm`   | [nvm](../nvm/)       | `taskfiles/yarn/nvm/Taskfile.yml` |

## Setup

Include the family and address a variant through its namespace:

```yaml
includes:
  yarn: taskfiles/yarn/Taskfile.yml
```

```bash
task yarn:fnm:install
task yarn:nvm:install
```

Or include a single variant directly to keep the task names short (`task yarn:install`).
Swap `fnm` for `nvm` throughout to use the nvm stack:

```yaml
includes:
  fnm:
    taskfile: taskfiles/fnm/Taskfile.yml
  corepack:
    taskfile: taskfiles/corepack/fnm/Taskfile.yml
  yarn:
    taskfile: taskfiles/yarn/fnm/Taskfile.yml
```

Install Node.js when needed:

```bash
task yarn:node:setup
task yarn:manager:setup
task yarn:node:setup NODE_VERSION=22
```

## Public Tasks

| Task            | Variables                                                  | Description                                                 |
| --------------- | ---------------------------------------------------------- | ----------------------------------------------------------- |
| `add`           | Required `PACKAGES`; optional `EXTRA_ARGS`                 | Add packages as devDependencies with `yarn add -D`.         |
| `node:setup`    | Optional `NODE_VERSION`                    | Install Node.js via fnm or nvm.                      |
| `manager:setup` | Optional `NODE_VERSION`                    | Install Corepack when needed and enable its shims.          |
| `manager:pin`   | Required `PACKAGE_MANAGER_VERSION`                         | Pin Yarn in `package.json` with Corepack.                   |
| `version`       | Optional `NODE_VERSION`                    | Show active Node.js and Yarn versions.                      |
| `install`       | Optional `NODE_VERSION`                    | Run `yarn install`.                                         |
| `install:undo`  | —                                                           | Disable the Yarn Corepack shim.                              |
| `upgrade`       | —                                                           | Upgrade Yarn via `corepack prepare yarn@latest --activate`. |
| `remove`        | Required `PACKAGES`; optional `EXTRA_ARGS`                 | Remove globally-installed packages with `yarn global remove`. |
| `ci`            | Optional `NODE_VERSION`                    | Run `yarn install --immutable` with a required `yarn.lock`. |
| `run`           | Required `SCRIPT` | Run a script via `yarn run`.                                |
| `dev`           | Optional `NODE_VERSION`                    | Run `yarn run dev`.                                         |
| `exec`          | Required `BINARY`; optional `ARGS`, `EXTRA_ARGS`           | Execute a local project binary via `yarn exec`.             |
| `test`          | Optional `NODE_VERSION`                    | Run `yarn test`.                                            |
| `build`         | Optional `NODE_VERSION`                    | Run `yarn run build`.                                       |
| `lint`          | Optional `NODE_VERSION`                    | Run `yarn run lint`.                                        |
| `format`        | Optional `NODE_VERSION`                    | Run `yarn run format`.                                      |
| `typecheck`     | Optional `NODE_VERSION`                    | Run `yarn run typecheck`.                                   |
| `update`        | Optional `NODE_VERSION`                    | Run `yarn up '*'` for modern Yarn dependency updates.       |
| `audit`         | Optional `NODE_VERSION`                    | Run strict `yarn npm audit`.                                |
| `audit:report`  | Optional `NODE_VERSION`                    | Report audit findings without failing.                      |
| `audit:json`    | Optional `NODE_VERSION`                    | Emit `yarn npm audit --json`.                               |
| `cache:clean`   | Optional `NODE_VERSION`                    | Run `yarn cache clean`.                                     |
| `clean`         | -                                                          | Remove `node_modules` when present.                         |
| `clean:all`     | -                                                          | Remove `node_modules` and `yarn.lock`.                      |

## Examples

```bash
task yarn:install
task yarn:manager:pin PACKAGE_MANAGER_VERSION=stable
task yarn:ci
task yarn:run SCRIPT=test -- --watch
task yarn:audit:report
task yarn:clean --yes
```

Yarn commands run through `corepack yarn`, so Corepack resolves the version
declared by the project's `packageManager` field. The runners set
`COREPACK_ENABLE_AUTO_PIN=1`, so Corepack writes a missing `packageManager`
field when Yarn is first used in a project. The CI task uses modern Yarn's
immutable install mode. Projects pinned to Yarn Classic can use
`task yarn:run SCRIPT=install -- --frozen-lockfile` or adjust the Taskfile for
their legacy install policy.
