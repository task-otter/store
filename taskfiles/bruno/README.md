# Bruno

## What is this Taskfile?

This Taskfile wraps the [Bruno](https://www.usebruno.com/) CLI (`@usebruno/cli`)
for running API collections from the command line. It installs Bruno as a local
dev dependency.

## Variants

Every runtime + package-manager combination ships as its own leaf Taskfile under
`taskfiles/bruno/`. They all expose the identical public interface documented
below; only the underlying runtime and package manager differ. Include the tool
family once in your root Taskfile:

```yaml
includes:
  bruno: taskfiles/bruno/Taskfile.yml
```

Then run the leaf that matches your project through its namespace (replace
`{TASK}` with any public task):

```bash
task bruno:bun:{TASK}                 # Bun runtime + Bun as package manager
task bruno:node:fnm:npm:{TASK}        # Node via fnm, npm as package manager
task bruno:node:nvm:pnpm:{TASK}       # Node via nvm, pnpm as package manager
```

Available leaves: `bun`, `node/{fnm,nvm}/{npm,pnpm,yarn}`.

## Public Tasks

| Task      | Variables                                                    | Description                                              |
| --------- | ------------------------------------------------------------ | -------------------------------------------------------- |
| `install` | Optional `BRUNO_VERSION`, `BRUNO_EXTRA_ARGS`, `CLI_ARGS`           | Install `@usebruno/cli` as a local dev dependency. Pass `BRUNO_VERSION=x.y.z` to pin a release. |
| `install:undo` | Optional `BRUNO_EXTRA_ARGS`                             | Remove the locally installed `@usebruno/cli` devDependency. |
| `upgrade` | Optional `BRUNO_EXTRA_ARGS`                                  | Reinstall `@usebruno/cli` at the latest version.          |
| `run`     | Optional `BRUNO_COLLECTION`, `BRUNO_ENV`, `BRUNO_EXTRA_ARGS`, `CLI_ARGS` | Run all requests in the Bruno collection.                |
| `ci`      | Optional `BRUNO_COLLECTION`, `BRUNO_ENV`, `BRUNO_EXTRA_ARGS`, `CLI_ARGS` | Run collection and stop on the first failure (`--bail`). |
| `version` | — | Show the locally resolved `bru` version.                 |
| `help`    | Optional `BRUNO_EXTRA_ARGS`, `CLI_ARGS`                      | Show Bruno CLI help.                                     |

## Variables

`BRUNO_COLLECTION` is the path to the Bruno collection directory. Defaults to `.`
(the current directory). `BRUNO_ENV` activates a named Bruno environment via
`--env <name>`. `BRUNO_EXTRA_ARGS` and arguments after `--` are appended to the
command.

## Examples

```bash
task bruno:node:fnm:npm:install
task bruno:node:fnm:npm:install BRUNO_VERSION=1.34.1
task bruno:node:fnm:npm:run
task bruno:node:fnm:npm:run BRUNO_COLLECTION=./api BRUNO_ENV=staging
```
