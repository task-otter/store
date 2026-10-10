# Vault Taskfile Public Tasks

## What Is This Taskfile?

A Taskfile for installing the HashiCorp Vault CLI, checking server health,
authenticating with tokens or AppRole, and reading KV secrets. Operational tasks
auto-install the Vault CLI via Nix on Unix or WinGet on Windows.

## Usage

### Standalone

```sh
task -t taskfiles/vault/Taskfile.yml healthy
```

Install only:

```sh
task -t taskfiles/vault/Taskfile.yml install
```

### Included

```yaml
includes:
  vault: ./taskfiles/vault/Taskfile.yml
```

Then run:

```sh
task vault:healthy VAULT_ADDR=http://127.0.0.1:8200
```

## Public Tasks

Generate an SSH key pair and sign it using an existing Vault SSH role:

```sh
task vault:ssh:keys VAULT_SSH_ROLE=my-role \
  VAULT_SSH_PRIVATE_KEY_PATH="$HOME/.ssh/vault_key" \
  VAULT_SSH_PUBLIC_KEY_PATH="$HOME/.ssh/vault_key.pub"
```

`ssh:keys` requires `ssh-keygen` on PATH and current Vault authentication
(`VAULT_TOKEN` or the token helper). Override `VAULT_SSH_MOUNT` (default `ssh`) for
another SSH secrets engine mount. It generates an unencrypted Ed25519 key,
reuses an existing pair, and derives a missing public key from the private key.
Each invocation signs the public key and writes `vault_key-cert.pub` beside it.
Failed signing preserves any existing certificate. Input preconditions run
before dependencies install Vault and prepare the key pair.

| Task               | Description                                              | Key variables                                                                                  |
| ------------------ | -------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `install`          | Install the Vault CLI via Nix (Unix) or WinGet (Windows) | `VAULT_NIX_INSTALLABLE`                                                                        |
| `version`          | Show the active Vault CLI version                        | —                                                                                              |
| `ssh:keys`         | Generate and sign SSH keys                               | `VAULT_SSH_ROLE`, `VAULT_SSH_PRIVATE_KEY_PATH`, `VAULT_SSH_PUBLIC_KEY_PATH`, `VAULT_SSH_MOUNT` |
| `healthy`          | Query the Vault HTTP health endpoint as JSON             | `VAULT_ADDR`                                                                                   |
| `login:root-token` | Log in using a token directly                            | `VAULT_ROOT_TOKEN`                                                                             |
| `login:approle`    | Log in using the AppRole auth method                     | `VAULT_ROLE_ID`, `VAULT_SECRET_ID`, `VAULT_APPROLE_MOUNT`                                      |
| `token:revoke`     | Revoke the current Vault token                           | `VAULT_TOKEN` (env)                                                                            |
| `kv:get`           | Read a KV v2 secret and print JSON to stdout             | `KV_MOUNT`, `SECRET_PATH`, `SECRET_VERSION`                                                    |

## Variables

| Variable                   | Default                 | Description                                       |
| -------------------------- | ----------------------- | ------------------------------------------------- |
| `VAULT_ADDR`               | `http://127.0.0.1:8200` | Vault server address used by CLI and HTTP tasks   |
| `VAULT_EXTRA_ARGS`         | _(empty)_               | Reserved for root include compatibility           |
| `VAULT_ROOT_TOKEN`         | _(empty)_               | Token for `login:root-token`                      |
| `VAULT_ROLE_ID`            | _(empty)_               | AppRole role_id for `login:approle`               |
| `VAULT_SECRET_ID`          | _(empty)_               | AppRole secret_id for `login:approle`             |
| `VAULT_APPROLE_MOUNT`      | `approle`               | AppRole mount path for `login:approle`            |
| `KV_MOUNT`                 | _(empty)_               | KV v2 engine mount path for `kv:get`              |
| `SECRET_PATH`              | _(empty)_               | Secret path within the KV mount for `kv:get`      |
| `SECRET_VERSION`           | _(empty)_               | Optional KV version to pin for `kv:get`           |
| `VAULT_NIX_INSTALLABLE`    | `nixpkgs#vault-bin`     | Flake installable passed to `nix:install:profile` |
| `VAULT_WINGET_INSTALLABLE` | `Hashicorp.Vault`       | WinGet package ID for `winget:install:package`    |

## Notes

* Install uses Nix on Linux and macOS (`VAULT_NIX_INSTALLABLE`, default `nixpkgs#vault-bin`) and WinGet on Windows (`VAULT_WINGET_INSTALLABLE`, default `Hashicorp.Vault`). Unix Nix
  install sets `NIXPKGS_ALLOW_UNFREE=1` and passes `--impure` to `nix:install:profile` because HashiCorp Vault is unfree in nixpkgs.

`token:revoke` revokes the token in `VAULT_TOKEN`. The variable must be set
in the caller's environment before running this task.

`kv:get` requires both `VAULT_TOKEN` and `VAULT_ADDR` to be set in the caller's
environment. `KV_MOUNT` and `SECRET_PATH` must be provided as task variables.
Pass `SECRET_VERSION=<n>` to pin to a specific KV version.

Pin a revision by overriding the installable, for example
`VAULT_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#vault-bin`.
