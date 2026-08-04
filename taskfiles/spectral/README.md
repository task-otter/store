# Spectral

## What is this Taskfile?

This Taskfile wraps [Spectral](https://stoplight.io/open-source/spectral), a
JSON/YAML linter for OpenAPI, AsyncAPI, and Arazzo documents, with automation
tasks for installing the tool and linting API descriptions.
`@stoplight/spectral-cli` is managed as a local devDependency. All tasks work on
macOS, Linux, and Windows through the underlying package-manager module.

## Variants

Every package-manager + Node-manager combination ships as its own leaf Taskfile
under `taskfiles/spectral/`. They all expose the identical public interface
documented below; only the underlying package manager and Node provisioning
differ. Include the tool family once in your root Taskfile:

```yaml
includes:
  spectral: taskfiles/spectral/Taskfile.yml
```

Then run the leaf that matches your project through its namespace (replace
`{TASK}` with any public task):

```bash
task spectral:node:fnm:npm:{TASK}        # Node via fnm, npm as package manager
task spectral:node:nvm:pnpm:{TASK}       # Node via nvm, pnpm as package manager
```

Available leaves: `node/{fnm,nvm}/{npm,pnpm}`. Run the tasks from the Node.js
project root (where `package.json` lives).

## Public Tasks

| Task | Description | Key variables |
|---|---|---|
| `install` | Install Spectral as a local devDependency | `VERSION` |
| `install:undo` | Remove the Spectral devDependency | |
| `upgrade` | Upgrade Spectral to the latest release | |
| `lint` | Lint API documents with Spectral | `TARGETS`, `SPECTRAL_RULESET`, `EXTRA_ARGS` |
| `config:init` | Create a default .spectral.yaml ruleset | |
| `help` | Show the Spectral CLI help | |
| `version` | Show the locally resolved Spectral version | |

## Variables

| Variable | Default | Description |
|---|---|---|
| `VERSION` | `""` (package manager default) | Pin a specific @stoplight/spectral-cli release |
| `TARGETS` | `""` | API document(s) to lint, e.g. `openapi.yaml` |
| `SPECTRAL_RULESET` | `""` | Path to a Spectral ruleset file passed via `--ruleset` |
| `EXTRA_ARGS` | `""` | Extra flags forwarded to spectral |
| `SPECTRAL_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

Spectral skips matching files as top-level lint targets, but may still load them when another document references them through `$ref`.

## Notes

- Requires a package-manager stack: `lint` auto-installs Spectral on first use;
  on a fresh machine provision Node.js first (e.g. `task fnm:node:install`).
- `lint` needs `TARGETS` — Spectral prints its usage message when no document
  is given. Without `SPECTRAL_RULESET`, Spectral discovers `.spectral.yaml` in the
  project automatically; `config:init` scaffolds one extending `spectral:oas`.
- The install `status:` guard keeps repeat runs idempotent — changing `VERSION`
  triggers a reinstall.
