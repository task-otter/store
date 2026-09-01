# Yarn Taskfile Public Tasks

## What is this Taskfile?

This Taskfile wraps common Yarn operations — installing dependencies, running scripts, auditing, and cleaning — behind consistent, cross-platform task commands. Node.js is provided via the [`nodejs`](../nodejs/) module (Nix profile); Yarn is installed from nixpkgs through this module's `install:tool` task.

## Usage

```yaml
includes:
  nodejs:
    taskfile: taskfiles/nodejs/Taskfile.yml
  yarn:
    taskfile: taskfiles/yarn/Taskfile.yml
```

```bash
task yarn:install
task yarn:install:clean
task yarn:run SCRIPT=build
```

Override the Node.js version by setting `NODEJS_NIX_INSTALLABLE` on the nodejs module before dependents run `nodejs:install`.

## Public Tasks

| Task              | Variables                                  | Description                                                               |
| ----------------- | ------------------------------------------ | ------------------------------------------------------------------------- |
| `add`             | Required `PACKAGES`; optional `EXTRA_ARGS` | Add packages as devDependencies with `yarn add -D`.                       |
| `version`         | —                                          | Show the active Node.js and Yarn versions.                                |
| `install`         | —                                          | Run `yarn install` to install all dependencies from `package.json`.     |
| `install:undo`    | —                                          | Explain how to remove the Yarn Nix profile installable.                   |
| `upgrade`         | —                                          | Upgrade Yarn via the configured Nix installable.                          |
| `install:clean`   | —                                          | Run `yarn install --immutable` and require an unchanged lockfile.         |
| `ci:fix`          | —                                          | Run `yarn run format`.                                                    |
| `clean`           | —                                          | Remove `node_modules`.                                                    |
| `clean:all`       | —                                          | Remove `node_modules` and `yarn.lock`.                                    |
| `dev`             | —                                          | Run `yarn run dev`.                                                       |
| `build`           | —                                          | Run `yarn run build`.                                                     |
| `test`            | —                                          | Run `yarn test`.                                                          |
| `lint`            | —                                          | Run `yarn run lint`.                                                      |
| `typecheck`       | —                                          | Run `yarn run typecheck`.                                                 |
| `run`             | Required `SCRIPT`; optional CLI args       | Run a `package.json` script.                                              |
| `exec`            | Required `BINARY`                          | Execute a local binary via `yarn exec`.                                   |
| `remove`          | Required `PACKAGES`                        | Uninstall packages.                                                       |
| `update`          | —                                          | Update packages within declared ranges.                                   |
| `audit`           | —                                          | Run `yarn npm audit` (strict).                                            |
| `audit:report`    | —                                          | Run audit reporting without failing.                                      |
| `audit:json`      | —                                          | Output audit results as JSON.                                             |
| `cache:clean`     | —                                          | Clear the Yarn cache.                                                     |
| `install:tool`    | —                                          | Install the Yarn binary via the Nix profile.                              |
| `version:tool`    | —                                          | Show the version of the Yarn binary itself.                               |

## Runtime

Project commands depend on `nodejs:install` and run `yarn` with `NIX_LOAD` so the Nix profile tools are on PATH. Must be run from a directory containing `package.json`.

Native Windows auto-install requires WSL2 (same as other Nix profile modules).
