## Learned User Preferences

- Prefer fixing lint findings (golangci, yamllint, etc.) by refactoring code or wrapping YAML; do not use `//nolint`, yamllint ignores, or edit linter configs (e.g. `.golangci.yml`) to silence issues.
- Prefer Taskfile vars as overridable templates like `'{{.VAR | default "..."}}'` so included Taskfiles do not lock parent or CLI overrides.
- Prefer composing Taskfiles via includes/deps rather than duplicating nix installables or shared install logic across modules.
- Prefer Windows tool installs via a winget Taskfile module (reusable tasks like `install:package`) and `*_WINGET_INSTALLABLE` vars set to winget package IDs when a package exists.
- Prefer OS-split Taskfiles as `taskfiles/<module>/unix` (nix-only) and `taskfiles/<module>/windows` (winget-only), with cross-module includes between matching platforms (e.g. yarn/unix → nodejs/unix); remove or thin-parent the combined module Taskfile.

## Learned Workspace Facts

- Custom golangci distance/reusability scoring: use `type T = struct{...}` aliases for DTO/param bags (excluded from scoring); keep behavioral types as defined types.
- zizmor `self-repository` expects local GitHub Action refs as `$/.github/actions/...`, not `./.github/actions/...`.
- This repo is a Taskfile module store under `taskfiles/`; dependent modules follow a yarn→nodejs-style include of the shared language Taskfile (e.g. go-junit-report depends on go) instead of bundling that language’s nix package.
- Taskotter `collectModules` only catalogs directories under `taskfiles/` that contain a `Taskfile.yml`; nested `unix`/`windows` Taskfiles alone are not discovered without a parent Taskfile or catalog changes.
- `ansible` and `ansible-lint` are linux/darwin-only; the `nix` module has no native Windows install path (WSL2/docs only).
