# Ansible Taskfile Public Tasks

## What is this Taskfile?

A Taskfile for running playbooks, testing connectivity, managing Ansible Galaxy
dependencies, and encrypting/decrypting secrets with Ansible Vault. Tasks
auto-install Ansible via `nix:install:profile`.

For Ansible YAML linting, use the separate
[`ansible-lint`](../ansible-lint/README.md) module.

> **Note:** Ansible does not support Windows as a control node. All tasks are
> macOS and Linux only.

## Usage

### Standalone

```sh
task -t taskfiles/ansible/Taskfile.yml syntax:check ANSIBLE_PLAYBOOK=site.yml
task -t taskfiles/ansible/Taskfile.yml run ANSIBLE_PLAYBOOK=site.yml ANSIBLE_INVENTORY=hosts
```

Install only:

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#ansible
```

### Included

```yaml
includes:
  ansible: ./taskfiles/ansible/Taskfile.yml
```

Then run:

```sh
task ansible:syntax:check ANSIBLE_PLAYBOOK=site.yml
task ansible:run ANSIBLE_PLAYBOOK=site.yml ANSIBLE_INVENTORY=hosts
```

## Public Tasks

| Task             | Description                                            |
| ---------------- | ------------------------------------------------------ |
| `syntax:check`   | Check playbook syntax without executing                |
| `run`            | Run an Ansible playbook                                |
| `ping`           | Test connectivity to inventory hosts                   |
| `list:hosts`     | List hosts matching PATTERN from INVENTORY             |
| `galaxy:install` | Install roles and collections from a requirements file |
| `vault:encrypt`  | Encrypt a file with Ansible Vault                      |
| `vault:decrypt`  | Decrypt a file with Ansible Vault                      |

## Variables

| Variable       | Default | Description |
| -------------- | ------- | ----------- |
| `ANSIBLE_NIX_INSTALLABLE` | `nixpkgs#ansible` | Flake installable passed to `nix:install:profile` |
| `ANSIBLE_CONFIG` | `ansible/ansible.cfg` | Value exported as `ANSIBLE_CONFIG` for ansible commands |
| `ANSIBLE_PLAYBOOK`     | _(empty)_                              | Playbook path; required by `run` and `syntax:check`              |
| `ANSIBLE_INVENTORY`    | _(empty)_                              | Inventory file or directory; required by `ping` and `list:hosts` |
| `ANSIBLE_PATTERN`      | `all`                                  | Host pattern for `ping` and `list:hosts`                         |
| `ANSIBLE_FILE`         | _(empty)_                              | File path; required by `vault:encrypt` and `vault:decrypt`       |
| `ANSIBLE_REQUIREMENTS` | `requirements.yml`                     | Requirements file for `galaxy:install`                           |
| `ANSIBLE_EXTRA_ARGS`   | _(empty)_                              | Extra flags forwarded to the underlying Ansible command          |

Pin a revision by overriding the installable, for example
`ANSIBLE_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#ansible`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.

**`vault:decrypt`** prompts for confirmation before decrypting to prevent
accidental plaintext exposure. Both vault tasks prompt interactively for the
vault password.

**`galaxy:install`** installs roles under `~/.ansible/roles` and collections
under `~/.ansible/collections` by default. Override with `ANSIBLE_EXTRA_ARGS` or a
`ansible.cfg` in your project.

For linting, run `ansible-lint:ci` from the [`ansible-lint`](../ansible-lint/README.md)
module, optionally followed by `ansible:syntax:check`.
