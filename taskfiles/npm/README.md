# NPM Taskfile Public Tasks

## What is this Taskfile?

This Taskfile wraps common `npm` operations — installing dependencies, running scripts, auditing, and cleaning — behind consistent, cross-platform task commands. It delegates Node.js installation to a **node version manager**, keeping each concern in its own taskfile.

## Variants

npm ships as two variants, one per node version manager. Both expose exactly the
same tasks; pick the one matching your stack.

| Variant | Node version manager | Taskfile                        |
| ------- | -------------------- | ------------------------------- |
| `fnm`   | [fnm](../fnm/)       | `taskfiles/npm/fnm/Taskfile.yml` |
| `nvm`   | [nvm](../nvm/)       | `taskfiles/npm/nvm/Taskfile.yml` |

Include the family and address a variant through its namespace:

```yaml
includes:
  npm: taskfiles/npm/Taskfile.yml
```

```bash
task npm:fnm:install
task npm:nvm:install
```

Or include a single variant directly to keep the task names short:

```yaml
# fnm stack
includes:
  fnm:
    taskfile: taskfiles/fnm/Taskfile.yml
  corepack:
    taskfile: taskfiles/corepack/fnm/Taskfile.yml
  npm:
    taskfile: taskfiles/npm/fnm/Taskfile.yml
```

```yaml
# nvm stack
includes:
  nvm:
    taskfile: taskfiles/nvm/Taskfile.yml
  corepack:
    taskfile: taskfiles/corepack/nvm/Taskfile.yml
  npm:
    taskfile: taskfiles/npm/nvm/Taskfile.yml
```

The examples below use the single-variant form (`task npm:<task>`). With the
family include, prefix them with the variant: `task npm:fnm:<task>`.

---

## Installing Node.js

Before using npm tasks on a fresh machine, install Node.js via your node version manager:

```bash
task npm:node:setup
task npm:node:setup NODE_VERSION=20
task npm:node:setup NODE_VERSION=22.0.0
```

This delegates to `task fnm:node:install` or `task nvm:node:install` depending on the variant.

---

## Public Tasks

| Task              | Variables                                                  | Description                                                               |
| ----------------- | ---------------------------------------------------------- | ------------------------------------------------------------------------- |
| `add`             | Required `PACKAGES`; optional `EXTRA_ARGS`                 | Add packages as devDependencies with `npm install -D`.                    |
| `node:setup`      | Optional `NODE_VERSION`                    | Install Node.js and npm via fnm or nvm.             |
| `version`         | Optional `NODE_VERSION`                    | Show the active Node.js and npm versions.                                 |
| `install`         | Optional `NODE_VERSION`                    | Run `npm install` to install all dependencies from `package.json`.        |
| `install:undo`    | —                                                           | Explain how to remove npm, which ships bundled with Node.js.              |
| `upgrade`         | —                                                           | Upgrade npm with `npm install -g npm@latest`.                             |
| `remove`          | Required `PACKAGES`; optional `EXTRA_ARGS`                 | Remove local packages with `npm uninstall`.                               |
| `ci`              | Optional `NODE_VERSION`                    | Run `npm ci` for a clean, reproducible install from `package-lock.json`.  |
| `run`             | Required `SCRIPT` | Run a `package.json` script by name. Example: `SCRIPT=dev`.               |
| `dev`             | Optional `NODE_VERSION`                    | Run `npm run dev`.                                                        |
| `exec`            | Required `BINARY`; optional `ARGS`, `EXTRA_ARGS`           | Execute a local project binary via `npm exec --`.                         |
| `test`            | Optional `NODE_VERSION`                    | Run `npm test`.                                                           |
| `build`           | Optional `NODE_VERSION`                    | Run `npm run build`.                                                      |
| `lint`            | Optional `NODE_VERSION`                    | Run `npm run lint`.                                                       |
| `format`          | Optional `NODE_VERSION`                    | Run `npm run format`.                                                     |
| `typecheck`       | Optional `NODE_VERSION`                    | Run `npm run typecheck`.                                                  |
| `manager:setup`   | Optional `NODE_VERSION`                    | Install Corepack when needed and enable its shims.                        |
| `manager:pin`     | Required `PACKAGE_MANAGER_VERSION`                         | Pin npm in `package.json` with Corepack.                                  |
| `outdated`        | Optional `NODE_VERSION`                    | List newer package versions without failing the task.                     |
| `outdated:strict` | Optional `NODE_VERSION`                    | List newer package versions and propagate `npm outdated` failures for CI. |
| `update`          | Optional `NODE_VERSION`                    | Update all packages within declared version ranges.                       |
| `audit`           | Optional `NODE_VERSION`                    | Audit vulnerabilities and propagate `npm audit` failures for CI.          |
| `audit:report`    | Optional `NODE_VERSION`                    | Report vulnerabilities without failing the task.                          |
| `audit:fix`       | Optional `NODE_VERSION`                    | Auto-fix vulnerabilities where a non-breaking fix exists.                 |
| `audit:json`      | Optional `NODE_VERSION`                    | Emit `npm audit --json` output for tooling.                               |
| `doctor`          | Optional `NODE_VERSION`                    | Check npm environment health with `npm doctor`.                           |
| `cache:clean`     | Optional `NODE_VERSION`                    | Clear the npm cache with `npm cache clean --force`.                       |
| `clean`           | —                                                          | Remove `node_modules`.                                                    |
| `clean:all`       | —                                                          | Remove `node_modules` and `package-lock.json`.                            |

---

## Dependency workflow

```bash
# Install or restore dependencies
task npm:install

# Clean install (CI-style, from lock file)
task npm:ci

# Remove and restore from scratch
task npm:clean --yes
task npm:install

# Nuclear clean
task npm:clean:all --yes
task npm:install
```

## Running scripts

```bash
task npm:run SCRIPT=dev          # npm run dev
task npm:run SCRIPT=start        # npm run start
task npm:dev                     # npm run dev
task npm:test                    # npm test
task npm:build                   # npm run build
task npm:lint                    # npm run lint
task npm:format                  # npm run format
task npm:typecheck               # npm run typecheck
task npm:manager:setup           # ensure Corepack is available
task npm:manager:pin PACKAGE_MANAGER_VERSION=latest
```

## Keeping dependencies healthy

```bash
task npm:outdated                # see what has newer versions
task npm:outdated:strict         # fail when npm outdated exits non-zero
task npm:update                  # update within declared ranges
task npm:audit                   # fail when npm audit finds vulnerabilities
task npm:audit:report            # report vulnerabilities without failing
task npm:audit:json              # emit npm audit output as JSON
task npm:audit:fix               # auto-fix where possible
task npm:doctor                  # inspect npm environment health
task npm:cache:clean             # clear npm cache data
```

---

## Security notes

All npm project commands run through `corepack npm` inside `bash -c` on Unix or PowerShell on Windows with the selected Node.js runtime loaded first. The runners set `COREPACK_ENABLE_AUTO_PIN=1`, so Corepack adds a missing `packageManager` field the first time the manager is used in a project. `manager:pin` remains available when you want to pick the package manager version explicitly. No credentials, tokens, or registry configuration are set by this Taskfile; those are managed by your npm config (`~/.npmrc`) and the project's `.npmrc` as usual.
