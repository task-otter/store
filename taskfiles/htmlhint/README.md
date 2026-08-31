# HTMLHint

## What is this Taskfile?

This Taskfile wraps [HTMLHint](https://htmlhint.com/), a static analysis tool
for HTML, with automation tasks for installing the tool and linting HTML files.
HTMLHint is managed as a local devDependency. All tasks work on macOS, Linux,
and Windows through the underlying package-manager module.

## Variants

Every runtime + package-manager combination ships as its own leaf Taskfile under
`taskfiles/htmlhint/`. They all expose the identical public interface documented
below; only the underlying runtime and package manager differ. Include the tool
family once in your root Taskfile:

```yaml
includes:
  htmlhint: taskfiles/htmlhint/Taskfile.yml
```

Then run the leaf that matches your project through its namespace (replace
`{TASK}` with any public task):

```bash
task htmlhint:bun:{TASK}             # Bun runtime + Bun as package manager
task htmlhint:node:npm:{TASK}        # Node via the nodejs module, npm as package manager
task htmlhint:node:pnpm:{TASK}       # Node via the nodejs module, pnpm as package manager
task htmlhint:node:yarn:{TASK}       # Node via the nodejs module, yarn as package manager
```

Available leaves: `bun`, `node/{npm,pnpm,yarn}`. Run the tasks from the Node.js
project root (where `package.json` lives).

## Public Tasks

| Task | Description | Key variables |
|---|---|---|
| `install` | Install HTMLHint as a local devDependency | `HTMLHINT_VERSION` |
| `install:undo` | Remove the HTMLHint devDependency | |
| `upgrade` | Upgrade HTMLHint to the latest release | |
| `ci` | Lint HTML files with HTMLHint | `HTMLHINT_TARGETS`, `HTMLHINT_CONFIG`, `HTMLHINT_EXTRA_ARGS` |
| `config:init` | Create a default .htmlhintrc configuration file | |
| `help` | Show the HTMLHint CLI help | |
| `version` | Show the locally resolved HTMLHint version | |

## Variables

| Variable | Default | Description |
|---|---|---|
| `HTMLHINT_VERSION` | `""` (package manager default) | Pin a specific htmlhint release |
| `HTMLHINT_TARGETS` | `**/*.html` | Glob of HTML files to lint |
| `HTMLHINT_CONFIG` | `""` | Path to a custom HTMLHint configuration file |
| `HTMLHINT_EXTRA_ARGS` | `""` | Extra flags forwarded to htmlhint |

## Notes

- Requires a package-manager stack: run `task htmlhint:node:npm:install` (or
  `htmlhint:bun:install` / `htmlhint:node:yarn:install`) and it auto-installs
  HTMLHint on first use. Node.js is provisioned through the `nodejs` module
  (`nodejs:_ensure`) on first run for node leaves; bun leaves use `bun:_ensure`.
- The install `status:` guard keeps repeat runs idempotent — changing `HTMLHINT_VERSION`
  triggers a reinstall.
- HTMLHint is lint-only; it has no autofix mode.
