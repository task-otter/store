# Spectral

## What is this Taskfile?

This Taskfile wraps [Spectral](https://stoplight.io/open-source/spectral), a
JSON/YAML linter for OpenAPI, AsyncAPI, and Arazzo documents, with automation
tasks for installing the tool and linting API descriptions.
`@stoplight/spectral-cli` is managed as a local devDependency. All tasks work on
macOS, Linux, and Windows through the underlying package-manager module.

## Variants

Every runtime + package-manager combination ships as its own leaf Taskfile
under `taskfiles/spectral/`. They all expose the identical public interface
documented below; only the underlying runtime and package manager differ.
Include the tool family once in your root Taskfile:

```yaml
includes:
  spectral: taskfiles/spectral/Taskfile.yml
```

Then run the leaf that matches your project through its namespace (replace
`{TASK}` with any public task):

```bash
task spectral:bun:{TASK}             # Bun runtime + Bun as package manager
task spectral:node:npm:{TASK}        # Node via the nodejs module, npm as package manager
task spectral:node:pnpm:{TASK}       # Node via the nodejs module, pnpm as package manager
task spectral:node:yarn:{TASK}       # Node via the nodejs module, yarn as package manager
```

Available leaves: `bun`, `node/{npm,pnpm,yarn}`. Run the tasks from the
project root (where `package.json` lives).

## Public Tasks

| Task | Description | Key variables |
|---|---|---|
| `install` | Install Spectral as a local devDependency | `SPECTRAL_VERSION` |
| `install:undo` | Remove the Spectral devDependency | |
| `upgrade` | Upgrade Spectral to the latest release | |
| `ci` | Lint API documents with Spectral | `SPECTRAL_TARGETS`, `SPECTRAL_RULESET`, `SPECTRAL_EXTRA_ARGS` |
| `config:init` | Create a default .spectral.yaml ruleset | |
| `help` | Show the Spectral CLI help | |
| `version` | Show the locally resolved Spectral version | |

## Variables

| Variable | Default | Description |
|---|---|---|
| `SPECTRAL_VERSION` | `""` (package manager default) | Pin a specific @stoplight/spectral-cli release |
| `SPECTRAL_TARGETS` | `""` | API document(s) to lint, e.g. `openapi.yaml` |
| `SPECTRAL_RULESET` | `""` | Path to a Spectral ruleset file passed via `--ruleset` |
| `SPECTRAL_EXTRA_ARGS` | `""` | Extra flags forwarded to spectral |
| `SPECTRAL_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

Spectral skips matching files as top-level lint targets, but may still load them when another document references them through `$ref`.

## Notes

- Requires a package-manager stack: `ci` auto-installs Spectral on first use;
  on a fresh machine, run a leaf task such as `task spectral:node:npm:ci` to provision Node.js via `nodejs:_ensure`.
- `ci` needs `SPECTRAL_TARGETS` — Spectral prints its usage message when no document
  is given. Without `SPECTRAL_RULESET`, Spectral discovers `.spectral.yaml` in the
  project automatically; `config:init` scaffolds one extending `spectral:oas`.
- The install `status:` guard keeps repeat runs idempotent — changing `SPECTRAL_VERSION`
  triggers a reinstall.
