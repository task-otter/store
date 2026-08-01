# Module dependency tree

This document reflects the module dependencies declared in [`.deps.yml`](.deps.yml).

**106 modules** total.

## Standalone

Modules with no `includes:` dependencies.

- [`bash-exec`](taskfiles/bash-exec/README.md)
- [`bencher`](taskfiles/bencher/README.md)
- [`bun`](taskfiles/bun/README.md)
- [`docker`](taskfiles/docker/README.md)
- [`fnm`](taskfiles/fnm/README.md)
- [`internal/skipfiles`](taskfiles/internal/skipfiles/Taskfile.yml)
- [`jq`](taskfiles/jq/README.md)
- [`nvm`](taskfiles/nvm/README.md)
- [`uv`](taskfiles/uv/README.md)

## Forward tree

### Node.js stacks

### Depth 0

- `bun`
- `fnm`
- `nvm`

### Depth 1

- `corepack/fnm` → `fnm`
- `corepack/nvm` → `nvm`

### Depth 2

- `npm/fnm` → `corepack/fnm`, `fnm`
- `npm/nvm` → `corepack/nvm`, `nvm`
- `pnpm/fnm` → `corepack/fnm`, `fnm`
- `pnpm/nvm` → `corepack/nvm`, `nvm`
- `yarn/fnm` → `corepack/fnm`, `fnm`
- `yarn/nvm` → `corepack/nvm`, `nvm`

**`bun`**

```
bun
```

**`corepack/fnm`**

```
corepack/fnm
    └── fnm
```

**`corepack/nvm`**

```
corepack/nvm
    └── nvm
```

**`fnm`**

```
fnm
```

**`npm/fnm`**

```
npm/fnm
    ├── corepack/fnm
    │   └── fnm
    └── fnm
```

**`npm/nvm`**

```
npm/nvm
    ├── corepack/nvm
    │   └── nvm
    └── nvm
```

**`nvm`**

```
nvm
```

**`pnpm/fnm`**

```
pnpm/fnm
    ├── corepack/fnm
    │   └── fnm
    └── fnm
```

**`pnpm/nvm`**

```
pnpm/nvm
    ├── corepack/nvm
    │   └── nvm
    └── nvm
```

**`yarn/fnm`**

```
yarn/fnm
    ├── corepack/fnm
    │   └── fnm
    └── fnm
```

**`yarn/nvm`**

```
yarn/nvm
    ├── corepack/nvm
    │   └── nvm
    └── nvm
```

### Other chains

### Depth 1

- `actionlint` → `internal/skipfiles`
- `ansible` → `internal/skipfiles`, `uv`
- `biome/bun` → `bun`
- `bruno/bun` → `bun`
- `buf` → `internal/skipfiles`
- `cargo` → `internal/skipfiles`
- `depcheck/bun` → `bun`
- `djlint` → `uv`
- `eslint/bun` → `bun`
- `gh` → `jq`
- `go` → `internal/skipfiles`
- `hadolint` → `internal/skipfiles`
- `jsonlint` → `internal/skipfiles`, `uv`
- `knip/bun` → `bun`
- `prettier/bun` → `bun`
- `python` → `uv`
- `rumdl` → `uv`
- `shellcheck` → `internal/skipfiles`
- `sqlfluff` → `internal/skipfiles`, `uv`
- `stylelint/bun` → `bun`
- `typescript/bun` → `bun`
- `vault` → `jq`
- `yamllint` → `internal/skipfiles`, `uv`
- `zizmor` → `internal/skipfiles`

### Depth 2

- `adrs` → `cargo`
- `dotenv-linter` → `cargo`, `internal/skipfiles`
- `git` → `gh`
- `golangci-lint` → `go`, `internal/skipfiles`
- `govulncheck` → `go`, `internal/skipfiles`
- `proto` → `go`
- `protolint` → `go`, `internal/skipfiles`
- `shfmt` → `go`, `internal/skipfiles`

### Depth 3

- `biome/node/fnm/npm` → `npm/fnm`
- `biome/node/fnm/pnpm` → `pnpm/fnm`
- `biome/node/fnm/yarn` → `yarn/fnm`
- `biome/node/nvm/npm` → `npm/nvm`
- `biome/node/nvm/pnpm` → `pnpm/nvm`
- `biome/node/nvm/yarn` → `yarn/nvm`
- `bruno/node/fnm/npm` → `npm/fnm`
- `bruno/node/fnm/pnpm` → `pnpm/fnm`
- `bruno/node/fnm/yarn` → `yarn/fnm`
- `bruno/node/nvm/npm` → `npm/nvm`
- `bruno/node/nvm/pnpm` → `pnpm/nvm`
- `bruno/node/nvm/yarn` → `yarn/nvm`
- `depcheck/node/fnm/npm` → `npm/fnm`
- `depcheck/node/fnm/pnpm` → `pnpm/fnm`
- `depcheck/node/fnm/yarn` → `yarn/fnm`
- `depcheck/node/nvm/npm` → `npm/nvm`
- `depcheck/node/nvm/pnpm` → `pnpm/nvm`
- `depcheck/node/nvm/yarn` → `yarn/nvm`
- `eslint/node/fnm/npm` → `npm/fnm`
- `eslint/node/fnm/pnpm` → `pnpm/fnm`
- `eslint/node/fnm/yarn` → `yarn/fnm`
- `eslint/node/nvm/npm` → `npm/nvm`
- `eslint/node/nvm/pnpm` → `pnpm/nvm`
- `eslint/node/nvm/yarn` → `yarn/nvm`
- `htmlhint/node/fnm/npm` → `npm/fnm`
- `htmlhint/node/fnm/pnpm` → `pnpm/fnm`
- `htmlhint/node/nvm/npm` → `npm/nvm`
- `htmlhint/node/nvm/pnpm` → `pnpm/nvm`
- `knip/node/fnm/npm` → `npm/fnm`
- `knip/node/fnm/pnpm` → `pnpm/fnm`
- `knip/node/fnm/yarn` → `yarn/fnm`
- `knip/node/nvm/npm` → `npm/nvm`
- `knip/node/nvm/pnpm` → `pnpm/nvm`
- `knip/node/nvm/yarn` → `yarn/nvm`
- `prettier/node/fnm/npm` → `npm/fnm`
- `prettier/node/fnm/pnpm` → `pnpm/fnm`
- `prettier/node/fnm/yarn` → `yarn/fnm`
- `prettier/node/nvm/npm` → `npm/nvm`
- `prettier/node/nvm/pnpm` → `pnpm/nvm`
- `prettier/node/nvm/yarn` → `yarn/nvm`
- `spectral/node/fnm/npm` → `npm/fnm`
- `spectral/node/fnm/pnpm` → `pnpm/fnm`
- `spectral/node/nvm/npm` → `npm/nvm`
- `spectral/node/nvm/pnpm` → `pnpm/nvm`
- `stylelint/node/fnm/npm` → `npm/fnm`
- `stylelint/node/fnm/pnpm` → `pnpm/fnm`
- `stylelint/node/fnm/yarn` → `yarn/fnm`
- `stylelint/node/nvm/npm` → `npm/nvm`
- `stylelint/node/nvm/pnpm` → `pnpm/nvm`
- `stylelint/node/nvm/yarn` → `yarn/nvm`
- `typescript/node/fnm/npm` → `npm/fnm`
- `typescript/node/fnm/pnpm` → `pnpm/fnm`
- `typescript/node/fnm/yarn` → `yarn/fnm`
- `typescript/node/nvm/npm` → `npm/nvm`
- `typescript/node/nvm/pnpm` → `pnpm/nvm`
- `typescript/node/nvm/yarn` → `yarn/nvm`

**`actionlint`**

```
actionlint
    └── internal/skipfiles
```

**`adrs`**

```
adrs
    └── cargo
        └── internal/skipfiles
```

**`ansible`**

```
ansible
    ├── internal/skipfiles
    └── uv
```

**`biome/bun`**

```
biome/bun
    └── bun
```

**`biome/node/fnm/npm`**

```
biome/node/fnm/npm
    └── npm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`biome/node/fnm/pnpm`**

```
biome/node/fnm/pnpm
    └── pnpm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`biome/node/fnm/yarn`**

```
biome/node/fnm/yarn
    └── yarn/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`biome/node/nvm/npm`**

```
biome/node/nvm/npm
    └── npm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`biome/node/nvm/pnpm`**

```
biome/node/nvm/pnpm
    └── pnpm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`biome/node/nvm/yarn`**

```
biome/node/nvm/yarn
    └── yarn/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`bruno/bun`**

```
bruno/bun
    └── bun
```

**`bruno/node/fnm/npm`**

```
bruno/node/fnm/npm
    └── npm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`bruno/node/fnm/pnpm`**

```
bruno/node/fnm/pnpm
    └── pnpm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`bruno/node/fnm/yarn`**

```
bruno/node/fnm/yarn
    └── yarn/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`bruno/node/nvm/npm`**

```
bruno/node/nvm/npm
    └── npm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`bruno/node/nvm/pnpm`**

```
bruno/node/nvm/pnpm
    └── pnpm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`bruno/node/nvm/yarn`**

```
bruno/node/nvm/yarn
    └── yarn/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`buf`**

```
buf
    └── internal/skipfiles
```

**`cargo`**

```
cargo
    └── internal/skipfiles
```

**`depcheck/bun`**

```
depcheck/bun
    └── bun
```

**`depcheck/node/fnm/npm`**

```
depcheck/node/fnm/npm
    └── npm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`depcheck/node/fnm/pnpm`**

```
depcheck/node/fnm/pnpm
    └── pnpm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`depcheck/node/fnm/yarn`**

```
depcheck/node/fnm/yarn
    └── yarn/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`depcheck/node/nvm/npm`**

```
depcheck/node/nvm/npm
    └── npm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`depcheck/node/nvm/pnpm`**

```
depcheck/node/nvm/pnpm
    └── pnpm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`depcheck/node/nvm/yarn`**

```
depcheck/node/nvm/yarn
    └── yarn/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`djlint`**

```
djlint
    └── uv
```

**`dotenv-linter`**

```
dotenv-linter
    ├── cargo
    │   └── internal/skipfiles
    └── internal/skipfiles
```

**`eslint/bun`**

```
eslint/bun
    └── bun
```

**`eslint/node/fnm/npm`**

```
eslint/node/fnm/npm
    └── npm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`eslint/node/fnm/pnpm`**

```
eslint/node/fnm/pnpm
    └── pnpm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`eslint/node/fnm/yarn`**

```
eslint/node/fnm/yarn
    └── yarn/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`eslint/node/nvm/npm`**

```
eslint/node/nvm/npm
    └── npm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`eslint/node/nvm/pnpm`**

```
eslint/node/nvm/pnpm
    └── pnpm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`eslint/node/nvm/yarn`**

```
eslint/node/nvm/yarn
    └── yarn/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`gh`**

```
gh
    └── jq
```

**`git`**

```
git
    └── gh
        └── jq
```

**`go`**

```
go
    └── internal/skipfiles
```

**`golangci-lint`**

```
golangci-lint
    ├── go
    │   └── internal/skipfiles
    └── internal/skipfiles
```

**`govulncheck`**

```
govulncheck
    ├── go
    │   └── internal/skipfiles
    └── internal/skipfiles
```

**`hadolint`**

```
hadolint
    └── internal/skipfiles
```

**`htmlhint/node/fnm/npm`**

```
htmlhint/node/fnm/npm
    └── npm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`htmlhint/node/fnm/pnpm`**

```
htmlhint/node/fnm/pnpm
    └── pnpm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`htmlhint/node/nvm/npm`**

```
htmlhint/node/nvm/npm
    └── npm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`htmlhint/node/nvm/pnpm`**

```
htmlhint/node/nvm/pnpm
    └── pnpm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`jsonlint`**

```
jsonlint
    ├── internal/skipfiles
    └── uv
```

**`knip/bun`**

```
knip/bun
    └── bun
```

**`knip/node/fnm/npm`**

```
knip/node/fnm/npm
    └── npm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`knip/node/fnm/pnpm`**

```
knip/node/fnm/pnpm
    └── pnpm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`knip/node/fnm/yarn`**

```
knip/node/fnm/yarn
    └── yarn/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`knip/node/nvm/npm`**

```
knip/node/nvm/npm
    └── npm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`knip/node/nvm/pnpm`**

```
knip/node/nvm/pnpm
    └── pnpm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`knip/node/nvm/yarn`**

```
knip/node/nvm/yarn
    └── yarn/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`prettier/bun`**

```
prettier/bun
    └── bun
```

**`prettier/node/fnm/npm`**

```
prettier/node/fnm/npm
    └── npm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`prettier/node/fnm/pnpm`**

```
prettier/node/fnm/pnpm
    └── pnpm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`prettier/node/fnm/yarn`**

```
prettier/node/fnm/yarn
    └── yarn/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`prettier/node/nvm/npm`**

```
prettier/node/nvm/npm
    └── npm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`prettier/node/nvm/pnpm`**

```
prettier/node/nvm/pnpm
    └── pnpm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`prettier/node/nvm/yarn`**

```
prettier/node/nvm/yarn
    └── yarn/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`proto`**

```
proto
    └── go
        └── internal/skipfiles
```

**`protolint`**

```
protolint
    ├── go
    │   └── internal/skipfiles
    └── internal/skipfiles
```

**`python`**

```
python
    └── uv
```

**`rumdl`**

```
rumdl
    └── uv
```

**`shellcheck`**

```
shellcheck
    └── internal/skipfiles
```

**`shfmt`**

```
shfmt
    ├── go
    │   └── internal/skipfiles
    └── internal/skipfiles
```

**`spectral/node/fnm/npm`**

```
spectral/node/fnm/npm
    └── npm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`spectral/node/fnm/pnpm`**

```
spectral/node/fnm/pnpm
    └── pnpm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`spectral/node/nvm/npm`**

```
spectral/node/nvm/npm
    └── npm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`spectral/node/nvm/pnpm`**

```
spectral/node/nvm/pnpm
    └── pnpm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`sqlfluff`**

```
sqlfluff
    ├── internal/skipfiles
    └── uv
```

**`stylelint/bun`**

```
stylelint/bun
    └── bun
```

**`stylelint/node/fnm/npm`**

```
stylelint/node/fnm/npm
    └── npm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`stylelint/node/fnm/pnpm`**

```
stylelint/node/fnm/pnpm
    └── pnpm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`stylelint/node/fnm/yarn`**

```
stylelint/node/fnm/yarn
    └── yarn/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`stylelint/node/nvm/npm`**

```
stylelint/node/nvm/npm
    └── npm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`stylelint/node/nvm/pnpm`**

```
stylelint/node/nvm/pnpm
    └── pnpm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`stylelint/node/nvm/yarn`**

```
stylelint/node/nvm/yarn
    └── yarn/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`typescript/bun`**

```
typescript/bun
    └── bun
```

**`typescript/node/fnm/npm`**

```
typescript/node/fnm/npm
    └── npm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`typescript/node/fnm/pnpm`**

```
typescript/node/fnm/pnpm
    └── pnpm/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`typescript/node/fnm/yarn`**

```
typescript/node/fnm/yarn
    └── yarn/fnm
        ├── corepack/fnm
        │   └── fnm
        └── fnm
```

**`typescript/node/nvm/npm`**

```
typescript/node/nvm/npm
    └── npm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`typescript/node/nvm/pnpm`**

```
typescript/node/nvm/pnpm
    └── pnpm/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`typescript/node/nvm/yarn`**

```
typescript/node/nvm/yarn
    └── yarn/nvm
        ├── corepack/nvm
        │   └── nvm
        └── nvm
```

**`vault`**

```
vault
    └── jq
```

**`yamllint`**

```
yamllint
    ├── internal/skipfiles
    └── uv
```

**`zizmor`**

```
zizmor
    └── internal/skipfiles
```

## Reverse tree

For each module, modules that depend on it (direct dependents only).

- `actionlint` — *(none)*
- `adrs` — *(none)*
- `ansible` — *(none)*
- `bash-exec` — *(none)*
- `bencher` — *(none)*
- `biome/bun` — *(none)*
- `biome/node/fnm/npm` — *(none)*
- `biome/node/fnm/pnpm` — *(none)*
- `biome/node/fnm/yarn` — *(none)*
- `biome/node/nvm/npm` — *(none)*
- `biome/node/nvm/pnpm` — *(none)*
- `biome/node/nvm/yarn` — *(none)*
- `bruno/bun` — *(none)*
- `bruno/node/fnm/npm` — *(none)*
- `bruno/node/fnm/pnpm` — *(none)*
- `bruno/node/fnm/yarn` — *(none)*
- `bruno/node/nvm/npm` — *(none)*
- `bruno/node/nvm/pnpm` — *(none)*
- `bruno/node/nvm/yarn` — *(none)*
- `buf` — *(none)*
- `bun` ← `biome/bun`, `bruno/bun`, `depcheck/bun`, `eslint/bun`, `knip/bun`, `prettier/bun`, `stylelint/bun`, `typescript/bun`
- `cargo` ← `adrs`, `dotenv-linter`
- `corepack/fnm` ← `npm/fnm`, `pnpm/fnm`, `yarn/fnm`
- `corepack/nvm` ← `npm/nvm`, `pnpm/nvm`, `yarn/nvm`
- `depcheck/bun` — *(none)*
- `depcheck/node/fnm/npm` — *(none)*
- `depcheck/node/fnm/pnpm` — *(none)*
- `depcheck/node/fnm/yarn` — *(none)*
- `depcheck/node/nvm/npm` — *(none)*
- `depcheck/node/nvm/pnpm` — *(none)*
- `depcheck/node/nvm/yarn` — *(none)*
- `djlint` — *(none)*
- `docker` — *(none)*
- `dotenv-linter` — *(none)*
- `eslint/bun` — *(none)*
- `eslint/node/fnm/npm` — *(none)*
- `eslint/node/fnm/pnpm` — *(none)*
- `eslint/node/fnm/yarn` — *(none)*
- `eslint/node/nvm/npm` — *(none)*
- `eslint/node/nvm/pnpm` — *(none)*
- `eslint/node/nvm/yarn` — *(none)*
- `fnm` ← `corepack/fnm`, `npm/fnm`, `pnpm/fnm`, `yarn/fnm`
- `gh` ← `git`
- `git` — *(none)*
- `go` ← `golangci-lint`, `govulncheck`, `proto`, `protolint`, `shfmt`
- `golangci-lint` — *(none)*
- `govulncheck` — *(none)*
- `hadolint` — *(none)*
- `htmlhint/node/fnm/npm` — *(none)*
- `htmlhint/node/fnm/pnpm` — *(none)*
- `htmlhint/node/nvm/npm` — *(none)*
- `htmlhint/node/nvm/pnpm` — *(none)*
- `internal/skipfiles` ← `actionlint`, `ansible`, `buf`, `cargo`, `dotenv-linter`, `go`, `golangci-lint`, `govulncheck`, `hadolint`, `jsonlint`, `protolint`, `shellcheck`, `shfmt`, `sqlfluff`, `yamllint`, `zizmor`
- `jq` ← `gh`, `vault`
- `jsonlint` — *(none)*
- `knip/bun` — *(none)*
- `knip/node/fnm/npm` — *(none)*
- `knip/node/fnm/pnpm` — *(none)*
- `knip/node/fnm/yarn` — *(none)*
- `knip/node/nvm/npm` — *(none)*
- `knip/node/nvm/pnpm` — *(none)*
- `knip/node/nvm/yarn` — *(none)*
- `npm/fnm` ← `biome/node/fnm/npm`, `bruno/node/fnm/npm`, `depcheck/node/fnm/npm`, `eslint/node/fnm/npm`, `htmlhint/node/fnm/npm`, `knip/node/fnm/npm`, `prettier/node/fnm/npm`, `spectral/node/fnm/npm`, `stylelint/node/fnm/npm`, `typescript/node/fnm/npm`
- `npm/nvm` ← `biome/node/nvm/npm`, `bruno/node/nvm/npm`, `depcheck/node/nvm/npm`, `eslint/node/nvm/npm`, `htmlhint/node/nvm/npm`, `knip/node/nvm/npm`, `prettier/node/nvm/npm`, `spectral/node/nvm/npm`, `stylelint/node/nvm/npm`, `typescript/node/nvm/npm`
- `nvm` ← `corepack/nvm`, `npm/nvm`, `pnpm/nvm`, `yarn/nvm`
- `pnpm/fnm` ← `biome/node/fnm/pnpm`, `bruno/node/fnm/pnpm`, `depcheck/node/fnm/pnpm`, `eslint/node/fnm/pnpm`, `htmlhint/node/fnm/pnpm`, `knip/node/fnm/pnpm`, `prettier/node/fnm/pnpm`, `spectral/node/fnm/pnpm`, `stylelint/node/fnm/pnpm`, `typescript/node/fnm/pnpm`
- `pnpm/nvm` ← `biome/node/nvm/pnpm`, `bruno/node/nvm/pnpm`, `depcheck/node/nvm/pnpm`, `eslint/node/nvm/pnpm`, `htmlhint/node/nvm/pnpm`, `knip/node/nvm/pnpm`, `prettier/node/nvm/pnpm`, `spectral/node/nvm/pnpm`, `stylelint/node/nvm/pnpm`, `typescript/node/nvm/pnpm`
- `prettier/bun` — *(none)*
- `prettier/node/fnm/npm` — *(none)*
- `prettier/node/fnm/pnpm` — *(none)*
- `prettier/node/fnm/yarn` — *(none)*
- `prettier/node/nvm/npm` — *(none)*
- `prettier/node/nvm/pnpm` — *(none)*
- `prettier/node/nvm/yarn` — *(none)*
- `proto` — *(none)*
- `protolint` — *(none)*
- `python` — *(none)*
- `rumdl` — *(none)*
- `shellcheck` — *(none)*
- `shfmt` — *(none)*
- `spectral/node/fnm/npm` — *(none)*
- `spectral/node/fnm/pnpm` — *(none)*
- `spectral/node/nvm/npm` — *(none)*
- `spectral/node/nvm/pnpm` — *(none)*
- `sqlfluff` — *(none)*
- `stylelint/bun` — *(none)*
- `stylelint/node/fnm/npm` — *(none)*
- `stylelint/node/fnm/pnpm` — *(none)*
- `stylelint/node/fnm/yarn` — *(none)*
- `stylelint/node/nvm/npm` — *(none)*
- `stylelint/node/nvm/pnpm` — *(none)*
- `stylelint/node/nvm/yarn` — *(none)*
- `typescript/bun` — *(none)*
- `typescript/node/fnm/npm` — *(none)*
- `typescript/node/fnm/pnpm` — *(none)*
- `typescript/node/fnm/yarn` — *(none)*
- `typescript/node/nvm/npm` — *(none)*
- `typescript/node/nvm/pnpm` — *(none)*
- `typescript/node/nvm/yarn` — *(none)*
- `uv` ← `ansible`, `djlint`, `jsonlint`, `python`, `rumdl`, `sqlfluff`, `yamllint`
- `vault` — *(none)*
- `yamllint` — *(none)*
- `yarn/fnm` ← `biome/node/fnm/yarn`, `bruno/node/fnm/yarn`, `depcheck/node/fnm/yarn`, `eslint/node/fnm/yarn`, `knip/node/fnm/yarn`, `prettier/node/fnm/yarn`, `stylelint/node/fnm/yarn`, `typescript/node/fnm/yarn`
- `yarn/nvm` ← `biome/node/nvm/yarn`, `bruno/node/nvm/yarn`, `depcheck/node/nvm/yarn`, `eslint/node/nvm/yarn`, `knip/node/nvm/yarn`, `prettier/node/nvm/yarn`, `stylelint/node/nvm/yarn`, `typescript/node/nvm/yarn`
- `zizmor` — *(none)*
