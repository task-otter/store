# Bruno

## What is this Taskfile?

This Taskfile wraps the [Bruno](https://www.usebruno.com/) CLI (`@usebruno/cli`)
for running API collections from the command line. It installs Bruno as a local
dev dependency This variant uses the `npm-fnm` stack (`npm-fnm`) package manager.


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
| `install` | Optional `VERSION`, `EXTRA_ARGS`, `CLI_ARGS`           | Install `@usebruno/cli` as a local dev dependency. Pass `VERSION=x.y.z` to pin a release. |
| `install:undo` | Optional `EXTRA_ARGS`                             | Remove the locally installed `@usebruno/cli` devDependency. |
| `upgrade` | Optional `EXTRA_ARGS`                                  | Reinstall `@usebruno/cli` at the latest version.          |
| `run`     | Optional `COLLECTION`, `ENV`, `EXTRA_ARGS`, `CLI_ARGS` | Run all requests in the Bruno collection.                |
| `ci`      | Optional `COLLECTION`, `ENV`, `EXTRA_ARGS`, `CLI_ARGS` | Run collection and stop on the first failure (`--bail`). |
| `version` | — | Show the locally resolved `bru` version.                 |
| `help`    | Optional `EXTRA_ARGS`, `CLI_ARGS`                      | Show Bruno CLI help.                                     |

## Variables

`COLLECTION` is the path to the Bruno collection directory. Defaults to `.`
(the current directory). `ENV` activates a named Bruno environment via
`--env <name>`. `EXTRA_ARGS` and arguments after `--` are appended to the
command.

## Examples

```bash
task bruno:node:fnm:npm:install
task bruno:node:fnm:npm:install VERSION=1.34.1
task bruno:node:fnm:npm:run
task bruno:node:fnm:npm:run COLLECTION=./api ENV=staging
```
