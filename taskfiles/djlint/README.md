# djLint Taskfile

## What is this Taskfile?

This Taskfile wraps [djLint](https://www.djlint.com/), a linter and formatter
for HTML template languages (Django, Jinja, Nunjucks, Handlebars, and more).
Lint and format tasks auto-install djLint via `nix:install:profile`.

## Usage

### Standalone

```bash
task --taskfile taskfiles/djlint/Taskfile.yml lint DJLINT_TARGETS=templates
```

Install only:

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#djlint
```

### Included

```yaml
includes:
  djlint:
    taskfile: taskfiles/djlint/Taskfile.yml
```

```bash
task djlint:lint DJLINT_TARGETS=templates
task djlint:lint DJLINT_TARGETS=templates DJLINT_EXTRA_ARGS="--profile django"
task djlint:ci:fix DJLINT_TARGETS=templates
task djlint:fmt:check DJLINT_TARGETS=templates
```

## Public Tasks

| Task | Description |
|---|---|
| `lint` | Lint HTML templates with djlint --lint |
| `fmt:check` | Report formatting changes without modifying files (djlint --check) |
| `ci` | Run `fmt:check` then `lint` |
| `ci:fix` | Format HTML templates in place with djlint --reformat |

## Variables

| Variable | Default | Description |
|---|---|---|
| `DJLINT_NIX_INSTALLABLE` | `nixpkgs#djlint` | Flake installable passed to `nix:install:profile` |
| `DJLINT_TARGETS` | `.` | File or directory djLint operates on |
| `DJLINT_EXTRA_ARGS` | `""` | Extra flags forwarded to djLint (e.g. `--profile django`) |
| `DJLINT_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint checks and fixes |
| `DJLINT_FMT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by formatting checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

Pin a revision by overriding the installable, for example
`DJLINT_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#djlint`.

## Notes

- `lint` reports template lint rule violations (`--lint`); `fmt:check` is the
  dry-run counterpart of `ci:fix` and reports formatting differences (`--check`).
  They are distinct djLint modes.
- Pass `DJLINT_EXTRA_ARGS="--profile <name>"` to select the template dialect
  (django, jinja, nunjucks, handlebars, golang, angular).
- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.
