# NPM Taskfile Public Tasks

## What is this Taskfile?

This Taskfile wraps common `npm` operations — installing dependencies, running scripts, auditing, and cleaning — behind consistent, cross-platform task commands. Node.js is provided via the [`nodejs`](../nodejs/) module (Nix profile); npm ships bundled with that runtime.

## Usage

```yaml
includes:
  nodejs:
    taskfile: taskfiles/nodejs/Taskfile.yml
  npm:
    taskfile: taskfiles/npm/Taskfile.yml
```

```bash
task npm:install
task npm:install:clean
task npm:run SCRIPT=build
```

Override the Node.js version by setting `NODEJS_NIX_INSTALLABLE` on the nodejs module before dependents run `nodejs:install`.

## Public Tasks

| Task              | Variables                                  | Description                                                               |
| ----------------- | ------------------------------------------ | ------------------------------------------------------------------------- |
| `add`             | Required `PACKAGES`; optional `EXTRA_ARGS` | Add packages as devDependencies with `npm install -D`.                    |
| `version`         | —                                          | Show the active Node.js and npm versions.                                 |
| `install`         | —                                          | Run `npm install` to install all dependencies from `package.json`.        |
| `install:undo`    | —                                          | Explain how to remove npm (bundled with Node.js).                         |
| `upgrade`         | —                                          | Upgrade npm to the latest release globally.                               |
| `install:clean`   | —                                          | Run `npm ci` for a clean lockfile-driven install.                         |
| `ci:fix`          | —                                          | Run `npm run format`.                                                     |
| `clean`           | —                                          | Remove `node_modules`.                                                    |
| `clean:all`       | —                                          | Remove `node_modules` and `package-lock.json`.                            |
| `dev`             | —                                          | Run `npm run dev`.                                                        |
| `build`           | —                                          | Run `npm run build`.                                                      |
| `test`            | —                                          | Run `npm test`.                                                           |
| `lint`            | —                                          | Run `npm run lint`.                                                       |
| `typecheck`       | —                                          | Run `npm run typecheck`.                                                  |
| `run`             | Required `SCRIPT`; optional CLI args       | Run a `package.json` script.                                              |
| `exec`            | Required `BINARY`                          | Execute a local binary via `npm exec`.                                    |
| `remove`          | Required `PACKAGES`                        | Uninstall packages.                                                       |
| `outdated`        | —                                          | List outdated packages (non-strict).                                      |
| `outdated:strict` | —                                          | List outdated packages and fail if any exist.                             |
| `audit`           | —                                          | Run `npm audit` (strict).                                                 |
| `audit:report`    | —                                          | Run `npm audit` without failing.                                        |
| `audit:fix`       | —                                          | Run `npm audit fix`.                                                      |
| `audit:json`      | —                                          | Output audit results as JSON.                                             |
| `update`          | —                                          | Update packages within declared ranges.                                   |
| `doctor`          | —                                          | Run `npm doctor`.                                                         |
| `cache:clean`     | —                                          | Clear the npm cache.                                                      |

## Runtime

Project commands depend on `nodejs:install` and run `npm` with `NIX_LOAD` so the Nix profile Node.js is on PATH. Must be run from a directory containing `package.json`.

Native Windows auto-install requires WSL2 (same as other Nix profile modules).
