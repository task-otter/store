# TaskOtter

[![CI](https://github.com/task-otter/store/actions/workflows/main.yml/badge.svg)](https://github.com/task-otter/store/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/task-otter/store/graph/badge.svg)](https://codecov.io/gh/task-otter/store)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Ftask-otter%2Fstore.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Ftask-otter%2Fstore?ref=badge_shield)

Reusable, tested [Taskfile](https://taskfile.dev) modules for installing and
running development tools. Include the modules you need in your project's
Taskfile and run their tasks.

## Quick Start

Install [Task](https://taskfile.dev), then add this repository to your project:

```sh
git submodule add https://github.com/task-otter/store.git taskotter
```

Create or update your project's `Taskfile.yml`:

```yaml
version: '3'

includes:
  go: ./taskotter/taskfiles/go/Taskfile.yml
```

List the available tasks and verify your Go installation:

```sh
task --list
task go:verify
```

The verification task installs Go if it is missing. Keep the `taskfiles/` tree
intact: modules include shared dependencies using relative paths.

To try a module directly from a clone of this repository:

```sh
git clone https://github.com/task-otter/store.git
cd store
task -t taskfiles/go/Taskfile.yml verify
```

## Tool Installation

Modules use Nix on Linux and macOS, and WinGet on native Windows where supported.
The [Nix module](taskfiles/nix/README.md) installs Nix on first use.
Check each module's README for platform support, tasks, and variables;
[Ansible](taskfiles/ansible/README.md) and
[ansible-lint](taskfiles/ansible-lint/README.md) support Linux and macOS only.

Most tool tasks install their CLI automatically when needed. You can also install
a tool or check its version explicitly:

```sh
task go:install
task go:version
```

Override a module's variables to choose a package. For example, replace `<rev>`
with a nixpkgs revision to select a specific Go package:

```sh
task go:install 'GO_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#go'
```

Installation differs for a few module types:

* JavaScript lint and format tools install local development dependencies.
* [npm](taskfiles/npm/README.md), [pnpm](taskfiles/pnpm/README.md), and
  [Yarn](taskfiles/yarn/README.md) use `install` for project dependencies and
  `install:tool` for their CLI.
* [Docker](taskfiles/docker/README.md) manages its own installation.

## Modules

Each link below lists the module's public tasks and configuration.

| Category | Modules |
| --- | --- |
| Node runtimes | [`nodejs`](taskfiles/nodejs/README.md), [`bun`](taskfiles/bun/README.md) |
| Package managers | [`npm`](taskfiles/npm/README.md), [`pnpm`](taskfiles/pnpm/README.md), [`yarn`](taskfiles/yarn/README.md) |
| JS lint / format / check | [`biome`](taskfiles/biome/README.md), [`depcheck`](taskfiles/depcheck/README.md), [`eslint`](taskfiles/eslint/README.md), [`htmlhint`](taskfiles/htmlhint/README.md), [`knip`](taskfiles/knip/README.md), [`prettier`](taskfiles/prettier/README.md), [`spectral`](taskfiles/spectral/README.md), [`typescript`](taskfiles/typescript/README.md) |
| Languages & runtimes | [`go`](taskfiles/go/README.md), [`go-junit-report`](taskfiles/go-junit-report/README.md), [`golangci-lint`](taskfiles/golangci-lint/README.md), [`python`](taskfiles/python/README.md), [`uv`](taskfiles/uv/README.md), [`cargo`](taskfiles/cargo/README.md), [`proto`](taskfiles/proto/README.md), [`nix`](taskfiles/nix/README.md), [`winget`](taskfiles/winget/README.md) |
| CI & infra | [`actionlint`](taskfiles/actionlint/README.md), [`adrs`](taskfiles/adrs/README.md), [`ansible`](taskfiles/ansible/README.md), [`ansible-lint`](taskfiles/ansible-lint/README.md), [`bruno-cli`](taskfiles/bruno-cli/README.md), [`buf`](taskfiles/buf/README.md), [`dasel`](taskfiles/dasel/README.md), [`docker`](taskfiles/docker/README.md), [`gh`](taskfiles/gh/README.md), [`git`](taskfiles/git/README.md), [`hadolint`](taskfiles/hadolint/README.md), [`protolint`](taskfiles/protolint/README.md), [`rumdl`](taskfiles/rumdl/README.md), [`sqlfluff`](taskfiles/sqlfluff/README.md), [`trivy`](taskfiles/trivy/README.md), [`vault`](taskfiles/vault/README.md), [`yamlfix`](taskfiles/yamlfix/README.md), [`yamllint`](taskfiles/yamllint/README.md), [`zizmor`](taskfiles/zizmor/README.md) |
| Desktop | [`bruno-gui`](taskfiles/bruno-gui/README.md) |

### JavaScript Tools

Include a tool family, then choose the runtime and package manager in the task
name. For example, add ESLint alongside Go:

```yaml
version: '3'

includes:
  go: ./taskotter/taskfiles/go/Taskfile.yml
  eslint: ./taskotter/taskfiles/eslint/Taskfile.yml
```

Run the variant that matches your project:

```sh
task eslint:bun:ci
task eslint:node:npm:ci
task eslint:node:pnpm:ci
task eslint:node:yarn:ci
```

Modules compose through Taskfile `includes:`. See the
[dependency graph](deps-tree.md) for their shared dependencies.

## Contributing

Use the Go version declared in [go.mod](go.mod) and Task 3.54.0, matching CI:

```sh
go install github.com/go-task/task/v3/cmd/task@v3.54.0
task --version
go test ./...
```

Ensure `GOBIN` (or `$(go env GOPATH)/bin` when unset) comes first on `PATH` so the
installed Task version is used.

When changing a module:

1. Update its `Taskfile.yml`, `metadata.yml`, and README public-task table together.
2. Add or update its contract and integration tests, then run `go test ./...`.
3. Update [`.deps.yml`](.deps.yml) and [deps-tree.md](deps-tree.md) if dependencies change.

Preview smoke checks before running them:

```sh
go run ./cmd/tasksmoke --list
go run ./cmd/tasksmoke --module go
```

Smoke checks execute public tasks and may install tools. They use an isolated
home and copied work directory, capture output, and limit each task to three
minutes.

See the architecture decisions for
[variable naming](doc/adr/0002-prefix-top-level-taskfile-vars-with-the-module-name.md),
[integration tests](doc/adr/0003-run-every-taskfile-folder-through-the-task-cli-in-tests.md),
and [tool installation](doc/adr/0004-install-cli-tools-via-nix-profile.md).
Dependency updates are configured in [Dependabot](.github/dependabot.yml).

## License

[MIT](LICENSE)
