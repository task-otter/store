# Ansible Taskfile Public Tasks

## What is this Taskfile?

A Taskfile for installing Ansible and ansible-lint, linting Ansible YAML files,
running playbooks, testing connectivity, managing Ansible Galaxy dependencies,
and encrypting/decrypting secrets with Ansible Vault.

Ansible and ansible-lint are installed via uv as isolated tools.

> **Note:** Ansible does not support Windows as a control node. All tasks are
> macOS and Linux only.

## Usage

### Standalone

```sh
task -t taskfiles/ansible/Taskfile.yml install
task -t taskfiles/ansible/Taskfile.yml ci ANSIBLE_PLAYBOOK=site.yml
task -t taskfiles/ansible/Taskfile.yml run ANSIBLE_PLAYBOOK=site.yml ANSIBLE_INVENTORY=hosts
```

### Included

```yaml
includes:
  ansible: ./taskfiles/ansible/Taskfile.yml
```

Then run:

```sh
task ansible:install
task ansible:ci ANSIBLE_PLAYBOOK=site.yml
task ansible:run ANSIBLE_PLAYBOOK=site.yml ANSIBLE_INVENTORY=hosts
```

## Public Tasks

| Task             | Description                                            | Key variables                         |
| ---------------- | ------------------------------------------------------ | ------------------------------------- |
| `install`        | Install Ansible and ansible-lint via uv                | `ANSIBLE_VERSION`, `ANSIBLE_LINT_VERSION` |
| `install:undo`   | Remove Ansible and ansible-lint                        | none                                  |
| `upgrade`        | Upgrade Ansible and ansible-lint to the latest release | none                                  |
| `version`        | Show Ansible and ansible-lint versions                 | none                                  |
| `lint:fix`         | Auto-fix Ansible YAML files with ansible-lint --fix     | `ANSIBLE_TARGETS`, `ANSIBLE_EXTRA_ARGS`               |
| `ci`             | Run ansible-lint then `syntax:check`                    | `ANSIBLE_TARGETS`, `ANSIBLE_PLAYBOOK`, `ANSIBLE_INVENTORY` |
| `ci:fix` | Run `lint:fix` for CI fixing | — |
| `syntax:check`   | Check playbook syntax without executing                | `ANSIBLE_PLAYBOOK`, `ANSIBLE_INVENTORY`               |
| `run`            | Run an Ansible playbook                                | `ANSIBLE_PLAYBOOK`, `ANSIBLE_INVENTORY`, `ANSIBLE_EXTRA_ARGS` |
| `ping`           | Test connectivity to inventory hosts                   | `ANSIBLE_INVENTORY`, `ANSIBLE_PATTERN`, `ANSIBLE_EXTRA_ARGS`  |
| `list:hosts`     | List hosts matching PATTERN from INVENTORY             | `ANSIBLE_INVENTORY`, `ANSIBLE_PATTERN`, `ANSIBLE_EXTRA_ARGS`  |
| `galaxy:install` | Install roles and collections from a requirements file | `ANSIBLE_REQUIREMENTS`, `ANSIBLE_EXTRA_ARGS`          |
| `vault:encrypt`  | Encrypt a file with Ansible Vault                      | `ANSIBLE_FILE`, `ANSIBLE_EXTRA_ARGS`                  |
| `vault:decrypt`  | Decrypt a file with Ansible Vault                      | `ANSIBLE_FILE`, `ANSIBLE_EXTRA_ARGS`                  |

## Variables

| Variable       | Default                                | Description                                                      |
| -------------- | -------------------------------------- | ---------------------------------------------------------------- |
| `ANSIBLE_PLAYBOOK`     | _(empty)_                              | Playbook path; required by `run` and `syntax:check`              |
| `ANSIBLE_INVENTORY`    | _(empty)_                              | Inventory file or directory; required by `ping` and `list:hosts` |
| `ANSIBLE_PATTERN`      | `all`                                  | Host pattern for `ping` and `list:hosts`                         |
| `ANSIBLE_TARGETS`      | `.`                                    | Files or directories to lint with `ci` / `lint:fix`            |
| `ANSIBLE_FILE`         | _(empty)_                              | File path; required by `vault:encrypt` and `vault:decrypt`       |
| `ANSIBLE_REQUIREMENTS` | `requirements.yml`                     | Requirements file for `galaxy:install`                           |
| `ANSIBLE_EXTRA_ARGS`   | _(empty)_                              | Extra flags forwarded to the underlying Ansible command          |
| `ANSIBLE_VERSION` | _(empty)_                           | Pin a specific ansible release for `install`/`upgrade`; empty installs latest |
| `ANSIBLE_LINT_VERSION` | _(empty)_                      | Pin a specific ansible-lint release for `install`/`upgrade`; empty installs latest |
| `UV_LOAD`      | `export PATH="$HOME/.local/bin:$PATH"` | Shell snippet that ensures uv-managed binaries are in PATH       |
| `ANSIBLE_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint, fix, and syntax-check tasks |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Notes

**`ci`** runs ansible-lint (best practices and YAML checks) then
`syntax:check`. Configure linting rules with an `.ansible-lint` file in your
project root.

**`vault:decrypt`** prompts for confirmation before decrypting to prevent
accidental plaintext exposure. Both vault tasks prompt interactively for the
vault password.

**`galaxy:install`** installs roles under `~/.ansible/roles` and collections
under `~/.ansible/collections` by default. Override with `ANSIBLE_EXTRA_ARGS` or a
`ansible.cfg` in your project.
