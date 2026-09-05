## Learned User Preferences

- Prefer fixing golangci findings by refactoring code; do not use `//nolint` and do not edit `.golangci.yml` to silence issues.
- Prefer Taskfile vars as overridable templates like `'{{.VAR | default "..."}}'` so included Taskfiles do not lock parent or CLI overrides.
- Prefer composing Taskfiles via includes/deps rather than duplicating nix installables or shared install logic across modules.

## Learned Workspace Facts

- Custom golangci distance/reusability scoring: use `type T = struct{...}` aliases for DTO/param bags (excluded from scoring); keep behavioral types as defined types.
- zizmor `self-repository` expects local GitHub Action refs as `$/.github/actions/...`, not `./.github/actions/...`.
- This repo is a Taskfile module store under `taskfiles/`; dependent modules follow a yarn→nodejs-style include of the shared language Taskfile (e.g. go-junit-report depends on go) instead of bundling that language’s nix package.
