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
task nix:install:profile NIX_INSTALLABLE=nixpkgs#python3Packages.demjson3
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

## Variables

| Variable | Default | Description |
|---|---|---|
| `JSONLINT_NIX_INSTALLABLE` | `nixpkgs#python3Packages.demjson3` | Flake installable passed to `nix:install:profile` |
| `JSONLINT_TARGETS` | `.` | File or directory to validate; directories are scanned recursively for `*.json` |
| `JSONLINT_EXTRA_ARGS` | `""` | Extra flags forwarded to jsonlint |

Pin a revision by overriding the installable, for example
`JSONLINT_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#python3Packages.demjson3`.

## Notes

- The PyPI package named `jsonlint` is an unrelated validation library that
  ships no command-line tool; this Taskfile installs `demjson3`, which
  provides the actual `jsonlint` CLI.
- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.
