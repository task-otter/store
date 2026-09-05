# jsonlint Taskfile

## What is this Taskfile?

This Taskfile wraps the `jsonlint` command-line JSON validator. The CLI is
provided by [demjson3](https://pypi.org/project/demjson3/) (`nixpkgs#python3Packages.demjson3`,
not the Node jsonlint package). The `ci` task auto-installs it via
`nix:install:profile`.

## Usage

### Standalone

```bash
task --taskfile taskfiles/jsonlint/Taskfile.yml ci JSONLINT_TARGETS=config.json
```

Install only:

```sh
task --taskfile taskfiles/jsonlint/Taskfile.yml install
```

### Included

```yaml
includes:
  jsonlint:
    taskfile: taskfiles/jsonlint/Taskfile.yml
```

```bash
task jsonlint:ci JSONLINT_TARGETS=config.json
task jsonlint:ci JSONLINT_TARGETS=data/   # validates every *.json under data/
```

## Public Tasks

| Task | Description |
|---|---|
| `ci` | Validate JSON files with jsonlint |
| `install` | Install jsonlint via Nix (Unix) or uv tool (Windows) |
| `version` | Show the active jsonlint version |

## Variables

| Variable | Default | Description |
|---|---|---|
| `JSONLINT_NIX_INSTALLABLE` | `nixpkgs#python3Packages.demjson3` | Flake installable passed to `nix:install:profile` |
| `JSONLINT_UV_TOOL` | `demjson3` | uv tool name passed to `uv:tool:install` on Windows (provides `jsonlint`) |
| `JSONLINT_TARGETS` | `.` | File or directory to validate; directories are scanned recursively for `*.json` |
| `JSONLINT_EXTRA_ARGS` | `""` | Extra flags forwarded to jsonlint |

Pin a revision by overriding the installable, for example
`JSONLINT_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#python3Packages.demjson3`.

## Notes

- Unix install uses Nix (`JSONLINT_NIX_INSTALLABLE`). Windows installs via `uv:tool:install` (`JSONLINT_UV_TOOL`, default `demjson3`).

- The PyPI package named `jsonlint` is an unrelated validation library that
  ships no command-line tool; this Taskfile installs `demjson3`, which
  provides the actual `jsonlint` CLI.
