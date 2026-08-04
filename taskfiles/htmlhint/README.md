# HTMLHint

## What is this Taskfile?

This Taskfile wraps [HTMLHint](https://htmlhint.com/), a static analysis tool
for HTML, with automation tasks for installing the tool and linting HTML files.
HTMLHint is managed as a local devDependency. All tasks work on macOS, Linux,
and Windows through the underlying package-manager module.

## Variants

Every package-manager + Node-manager combination ships as its own leaf Taskfile
under `taskfiles/htmlhint/`. They all expose the identical public interface
documented below; only the underlying package manager and Node provisioning
differ. Include the tool family once in your root Taskfile:

```yaml
includes:
  htmlhint: taskfiles/htmlhint/Taskfile.yml
```

Then run the leaf that matches your project through its namespace (replace
`{TASK}` with any public task):

```bash
task htmlhint:node:fnm:npm:{TASK}        # Node via fnm, npm as package manager
task htmlhint:node:nvm:pnpm:{TASK}       # Node via nvm, pnpm as package manager
```

Available leaves: `node/{fnm,nvm}/{npm,pnpm}`. Run the tasks from the Node.js
project root (where `package.json` lives).

## Public Tasks

| Task | Description | Key variables |
|---|---|---|
| `install` | Install HTMLHint as a local devDependency | `HTMLHINT_VERSION` |
| `install:undo` | Remove the HTMLHint devDependency | |
| `upgrade` | Upgrade HTMLHint to the latest release | |
| `lint` | Lint HTML files with HTMLHint | `HTMLHINT_TARGETS`, `HTMLHINT_CONFIG`, `HTMLHINT_EXTRA_ARGS` |
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
| `HTMLHINT_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Notes

- Requires a package-manager stack: run `task htmlhint:node:fnm:npm:install` and
  it auto-installs HTMLHint on first use; on a fresh machine provision Node.js
  first (e.g. `task fnm:node:install`).
- The install `status:` guard keeps repeat runs idempotent — changing `HTMLHINT_VERSION`
  triggers a reinstall.
- HTMLHint is lint-only; it has no autofix mode.
