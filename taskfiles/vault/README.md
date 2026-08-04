# Vault Taskfile Public Tasks

## What is this Taskfile?

A cross-platform Taskfile for installing the HashiCorp Vault CLI and running
common Vault operator workflows such as status checks, initialization, unseal,
login, Raft peer inspection, snapshots, and restores.

macOS uses Homebrew with `hashicorp/tap`. Linux uses the official HashiCorp APT
or DNF repositories. Windows uses winget to install `Hashicorp.Vault`.

## Usage

### Standalone

```sh
task -t taskfiles/vault/Taskfile.yml install
task -t taskfiles/vault/Taskfile.yml status
task -t taskfiles/vault/Taskfile.yml health
```

### Included

```yaml
includes:
  vault: ./taskfiles/vault/Taskfile.yml
```

Then run:

```sh
task vault:install
task vault:status
task vault:health VAULT_ADDR=http://127.0.0.1:8200
task vault:snapshot VAULT_FILE=backup.snap
```

## Public Tasks

| Task           | Description                                  | Key variables                    |
| -------------- | -------------------------------------------- | -------------------------------- |
| `install`      | Install the Vault CLI on the current OS      | `VERSION`                        |
| `install:undo` | Remove the Vault CLI from the current OS     | none                             |
| `upgrade`      | Upgrade the Vault CLI                        | none                             |
| `version`      | Show the installed Vault CLI version         | none                             |
| `verify`       | Verify CLI installation and server status    | `VAULT_ADDR`                     |
| `status`       | Show Vault seal and HA status                | `VAULT_ADDR`                     |
| `health`       | Query the Vault HTTP health endpoint as JSON | `VAULT_ADDR`                     |
| `init`         | Initialize Vault and save unseal keys        | `VAULT_KEYS_FILE`, `VAULT_SHARES`, `VAULT_THRESHOLD` |
| `unseal`       | Unseal Vault using saved keys                | `VAULT_KEYS_FILE`, `VAULT_THRESHOLD`         |
| `seal`         | Seal the active Vault server                 | `VAULT_ADDR`                     |
| `login`            | Log in using the saved root token            | `VAULT_KEYS_FILE`                      |
| `login:root-token` | Log in using a token directly                | `VAULT_ROOT_TOKEN`                     |
| `login:approle`    | Log in using the AppRole auth method         | `VAULT_ROLE_ID`, `VAULT_SECRET_ID`, `VAULT_APPROLE_MOUNT` |
| `root-token`       | Print the saved root token                   | `VAULT_KEYS_FILE`                      |
| `token:issue:approle` | Exchange AppRole credentials for a token (printed to stdout) | `VAULT_ROLE_ID`, `VAULT_SECRET_ID`, `VAULT_APPROLE_MOUNT` |
| `token:revoke-self`   | Revoke the current Vault token              | `VAULT_TOKEN` (env)              |
| `kv:get`              | Read a KV v2 secret and print JSON to stdout | `KV_MOUNT`, `SECRET_PATH`, `SECRET_VERSION` |
| `peers`        | List Vault Raft cluster peers                | `VAULT_ADDR`                     |
| `snapshot`     | Save a Vault Raft snapshot                   | `FILE`, `VAULT_SNAPSHOT_FILE`; root: `VAULT_FILE` |
| `restore`      | Restore a Vault Raft snapshot                | `FILE`, `VAULT_SNAPSHOT_FILE`; root: `VAULT_FILE` |

## Variables

| Variable        | Default                 | Description                                      |
| --------------- | ----------------------- | ------------------------------------------------ |
| `VAULT_ADDR`    | `http://127.0.0.1:8200` | Vault server address used by CLI and HTTP tasks  |
| `VAULT_KEYS_FILE`     | `.vault-init-keys.json` | File used for init output, unseal keys, and token |
| `VAULT_SHARES`        | `5`                     | Number of unseal key shares for `init`           |
| `VAULT_THRESHOLD`     | `3`                     | Unseal key threshold for `init` and `unseal`     |
| `VAULT_SNAPSHOT_FILE` | `vault-snapshot.snap`   | Default Raft snapshot path                       |
| `FILE`          | _(empty)_               | Snapshot path override for `snapshot`/`restore` |
| `EXTRA_ARGS`    | _(empty)_               | Reserved for root include compatibility          |
| `VAULT_ROOT_TOKEN`    | _(empty)_               | Token for `login:root-token`                     |
| `VAULT_ROLE_ID`       | _(empty)_               | AppRole role_id for `login:approle` and `token:issue:approle` |
| `VAULT_SECRET_ID`     | _(empty)_               | AppRole secret_id for `login:approle` and `token:issue:approle` |
| `VAULT_APPROLE_MOUNT` | `approle`               | AppRole mount path for `login:approle` and `token:issue:approle` |
| `KV_MOUNT`      | _(empty)_               | KV v2 engine mount path for `kv:get`             |
| `SECRET_PATH`   | _(empty)_               | Secret path within the KV mount for `kv:get`     |
| `SECRET_VERSION`| _(empty)_               | Optional KV version to pin for `kv:get`          |
| `VERSION`       | _(empty)_               | Pin a specific Vault CLI release for `install`; empty installs latest. Exact availability depends on the platform's package manager/repository. |

## Notes

`init` writes the generated unseal keys and root token to `VAULT_KEYS_FILE` with mode
`600` under `umask 077` and does not echo the JSON payload to stdout. It refuses
to overwrite an existing `VAULT_KEYS_FILE`; move or remove the file before initializing
again. The default `.vault-init-keys.json` and `vault-snapshot.snap` files are
ignored by the repo.

`restore` is destructive and requires confirmation before it runs. It validates
the snapshot file before installing or invoking the Vault CLI.

`token:issue:approle` exchanges AppRole credentials for a client token and prints
it to stdout — unlike `login:approle`, the token is not stored in the token helper.
Pipe or capture the output for use by the caller.

`token:revoke-self` revokes the token in `VAULT_TOKEN`. The variable must be set
in the caller's environment before running this task.

`kv:get` requires both `VAULT_TOKEN` and `VAULT_ADDR` to be set in the caller's
environment. `KV_MOUNT` and `SECRET_PATH` must be provided as task variables.
Pass `SECRET_VERSION=<n>` to pin to a specific KV version.

When using this repository's root Taskfile include, pass `VAULT_FILE=path`
instead of `FILE=path` for `vault:snapshot` and `vault:restore`. The standalone
Vault Taskfile continues to use `FILE=path`.
