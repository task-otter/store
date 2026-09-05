# Module dependency tree

This document reflects the module dependencies declared in [`.deps.yml`](.deps.yml).

**94 modules** total.

## Standalone

Modules with no `includes:` dependencies.

- [`docker`](taskfiles/docker/README.md)
- [`nix`](taskfiles/nix/README.md)

## Forward tree

### Node.js stacks

### Depth 1

- `bun` → `nix`
- `nodejs` → `nix`

### Depth 2

- `biome/bun` → `bun`
- `depcheck/bun` → `bun`
- `eslint/bun` → `bun`
- `htmlhint/bun` → `bun`
- `knip/bun` → `bun`
- `npm` → `nix`, `nodejs`
- `pnpm` → `nix`, `nodejs`
- `prettier/bun` → `bun`
- `spectral/bun` → `bun`
- `stylelint/bun` → `bun`
- `typescript/bun` → `bun`
- `yarn` → `nix`, `nodejs`

### Depth 3

- `biome/node/npm` → `npm`
- `biome/node/pnpm` → `pnpm`
- `biome/node/yarn` → `yarn`
- `depcheck/node/npm` → `npm`
- `depcheck/node/pnpm` → `pnpm`
- `depcheck/node/yarn` → `yarn`
- `eslint/node/npm` → `npm`
- `eslint/node/pnpm` → `pnpm`
- `eslint/node/yarn` → `yarn`
- `htmlhint/node/npm` → `npm`
- `htmlhint/node/pnpm` → `pnpm`
- `htmlhint/node/yarn` → `yarn`
- `knip/node/npm` → `npm`
- `knip/node/pnpm` → `pnpm`
- `knip/node/yarn` → `yarn`
- `prettier/node/npm` → `npm`
- `prettier/node/pnpm` → `pnpm`
- `prettier/node/yarn` → `yarn`
- `spectral/node/npm` → `npm`
- `spectral/node/pnpm` → `pnpm`
- `spectral/node/yarn` → `yarn`
- `stylelint/node/npm` → `npm`
- `stylelint/node/pnpm` → `pnpm`
- `stylelint/node/yarn` → `yarn`
- `typescript/node/npm` → `npm`
- `typescript/node/pnpm` → `pnpm`
- `typescript/node/yarn` → `yarn`

### Depth 4

- `biome/node` → `biome/node/npm`, `biome/node/pnpm`, `biome/node/yarn`
- `depcheck/node` → `depcheck/node/npm`, `depcheck/node/pnpm`, `depcheck/node/yarn`
- `eslint/node` → `eslint/node/npm`, `eslint/node/pnpm`, `eslint/node/yarn`
- `htmlhint/node` → `htmlhint/node/npm`, `htmlhint/node/pnpm`, `htmlhint/node/yarn`
- `knip/node` → `knip/node/npm`, `knip/node/pnpm`, `knip/node/yarn`
- `prettier/node` → `prettier/node/npm`, `prettier/node/pnpm`, `prettier/node/yarn`
- `spectral/node` → `spectral/node/npm`, `spectral/node/pnpm`, `spectral/node/yarn`
- `stylelint/node` → `stylelint/node/npm`, `stylelint/node/pnpm`, `stylelint/node/yarn`
- `typescript/node` → `typescript/node/npm`, `typescript/node/pnpm`, `typescript/node/yarn`

### Depth 5

- `biome` → `biome/bun`, `biome/node`
- `depcheck` → `depcheck/bun`, `depcheck/node`
- `eslint` → `eslint/bun`, `eslint/node`
- `htmlhint` → `htmlhint/bun`, `htmlhint/node`
- `knip` → `knip/bun`, `knip/node`
- `prettier` → `prettier/bun`, `prettier/node`
- `spectral` → `spectral/bun`, `spectral/node`
- `stylelint` → `stylelint/bun`, `stylelint/node`
- `typescript` → `typescript/bun`, `typescript/node`

### Other chains

### Depth 0

- `docker` → *(none)*
- `nix` → *(none)*

### Depth 1

- `actionlint` → `nix`
- `adrs` → `nix`
- `ansible` → `nix`
- `ansible-lint` → `nix`
- `bencher` → `nix`
- `bruno-cli` → `nix`
- `bruno-gui` → `nix`
- `buf` → `nix`
- `cargo` → `nix`
- `djlint` → `nix`
- `dotenv-linter` → `nix`
- `go` → `nix`
- `golangci-lint` → `nix`
- `govulncheck` → `nix`
- `hadolint` → `nix`
- `jq` → `nix`
- `jsonlint` → `nix`
- `proto` → `nix`
- `protolint` → `nix`
- `pulumi` → `nix`
- `python` → `nix`
- `rumdl` → `nix`
- `shellcheck` → `nix`
- `shfmt` → `nix`
- `sqlfluff` → `nix`
- `uv` → `nix`
- `yamlfix` → `nix`
- `yamllint` → `nix`
- `zizmor` → `nix`

### Depth 2

- `gh` → `jq`, `nix`
- `go-junit-report` → `go`, `nix`
- `vault` → `jq`, `nix`

### Depth 3

- `git` → `gh`, `nix`

## Reverse tree

Who depends on each module:

- `actionlint` — *(none)*
- `adrs` — *(none)*
- `ansible` — *(none)*
- `ansible-lint` — *(none)*
- `bencher` — *(none)*
- `biome` — *(none)*
- `biome/bun` ← `biome`
- `biome/node` ← `biome`
- `biome/node/npm` ← `biome/node`
- `biome/node/pnpm` ← `biome/node`
- `biome/node/yarn` ← `biome/node`
- `bruno-cli` — *(none)*
- `bruno-gui` — *(none)*
- `buf` — *(none)*
- `bun` ← `biome/bun`, `depcheck/bun`, `eslint/bun`, `htmlhint/bun`, `knip/bun`, `prettier/bun`, `spectral/bun`, `stylelint/bun`, `typescript/bun`
- `cargo` — *(none)*
- `depcheck` — *(none)*
- `depcheck/bun` ← `depcheck`
- `depcheck/node` ← `depcheck`
- `depcheck/node/npm` ← `depcheck/node`
- `depcheck/node/pnpm` ← `depcheck/node`
- `depcheck/node/yarn` ← `depcheck/node`
- `djlint` — *(none)*
- `docker` — *(none)*
- `dotenv-linter` — *(none)*
- `eslint` — *(none)*
- `eslint/bun` ← `eslint`
- `eslint/node` ← `eslint`
- `eslint/node/npm` ← `eslint/node`
- `eslint/node/pnpm` ← `eslint/node`
- `eslint/node/yarn` ← `eslint/node`
- `gh` ← `git`
- `git` — *(none)*
- `go` ← `go-junit-report`
- `go-junit-report` — *(none)*
- `golangci-lint` — *(none)*
- `govulncheck` — *(none)*
- `hadolint` — *(none)*
- `htmlhint` — *(none)*
- `htmlhint/bun` ← `htmlhint`
- `htmlhint/node` ← `htmlhint`
- `htmlhint/node/npm` ← `htmlhint/node`
- `htmlhint/node/pnpm` ← `htmlhint/node`
- `htmlhint/node/yarn` ← `htmlhint/node`
- `jq` ← `gh`, `vault`
- `jsonlint` — *(none)*
- `knip` — *(none)*
- `knip/bun` ← `knip`
- `knip/node` ← `knip`
- `knip/node/npm` ← `knip/node`
- `knip/node/pnpm` ← `knip/node`
- `knip/node/yarn` ← `knip/node`
- `nix` ← `actionlint`, `adrs`, `ansible`, `ansible-lint`, `bencher`, `bruno-cli`, `bruno-gui`, `buf`, `bun`, `cargo`, `djlint`, `dotenv-linter`, `gh`, `git`, `go`, `go-junit-report`, `golangci-lint`, `govulncheck`, `hadolint`, `jq`, `jsonlint`, `nodejs`, `npm`, `pnpm`, `proto`, `protolint`, `pulumi`, `python`, `rumdl`, `shellcheck`, `shfmt`, `sqlfluff`, `uv`, `vault`, `yamlfix`, `yamllint`, `yarn`, `zizmor`
- `nodejs` ← `npm`, `pnpm`, `yarn`
- `npm` ← `biome/node/npm`, `depcheck/node/npm`, `eslint/node/npm`, `htmlhint/node/npm`, `knip/node/npm`, `prettier/node/npm`, `spectral/node/npm`, `stylelint/node/npm`, `typescript/node/npm`
- `pnpm` ← `biome/node/pnpm`, `depcheck/node/pnpm`, `eslint/node/pnpm`, `htmlhint/node/pnpm`, `knip/node/pnpm`, `prettier/node/pnpm`, `spectral/node/pnpm`, `stylelint/node/pnpm`, `typescript/node/pnpm`
- `prettier` — *(none)*
- `prettier/bun` ← `prettier`
- `prettier/node` ← `prettier`
- `prettier/node/npm` ← `prettier/node`
- `prettier/node/pnpm` ← `prettier/node`
- `prettier/node/yarn` ← `prettier/node`
- `proto` — *(none)*
- `protolint` — *(none)*
- `pulumi` — *(none)*
- `python` — *(none)*
- `rumdl` — *(none)*
- `shellcheck` — *(none)*
- `shfmt` — *(none)*
- `spectral` — *(none)*
- `spectral/bun` ← `spectral`
- `spectral/node` ← `spectral`
- `spectral/node/npm` ← `spectral/node`
- `spectral/node/pnpm` ← `spectral/node`
- `spectral/node/yarn` ← `spectral/node`
- `sqlfluff` — *(none)*
- `stylelint` — *(none)*
- `stylelint/bun` ← `stylelint`
- `stylelint/node` ← `stylelint`
- `stylelint/node/npm` ← `stylelint/node`
- `stylelint/node/pnpm` ← `stylelint/node`
- `stylelint/node/yarn` ← `stylelint/node`
- `typescript` — *(none)*
- `typescript/bun` ← `typescript`
- `typescript/node` ← `typescript`
- `typescript/node/npm` ← `typescript/node`
- `typescript/node/pnpm` ← `typescript/node`
- `typescript/node/yarn` ← `typescript/node`
- `uv` — *(none)*
- `vault` — *(none)*
- `yamlfix` — *(none)*
- `yamllint` — *(none)*
- `yarn` ← `biome/node/yarn`, `depcheck/node/yarn`, `eslint/node/yarn`, `htmlhint/node/yarn`, `knip/node/yarn`, `prettier/node/yarn`, `spectral/node/yarn`, `stylelint/node/yarn`, `typescript/node/yarn`
- `zizmor` — *(none)*
