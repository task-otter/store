# ansible-lint Taskfile Public Tasks

## What is this Taskfile?

A Taskfile for linting Ansible YAML files with
[ansible-lint](https://github.com/ansible/ansible-lint). The `ci` and `ci:fix`
tasks auto-install ansible-lint via `nix:install:profile`.

> **Note:** Ansible does not support Windows as a control node. All tasks are
> macOS and Linux only.

## Usage

### Standalone

```sh
task -t taskfiles/ansible-lint/Taskfile.yml ci
task -t taskfiles/ansible-lint/Taskfile.yml ci:fix ANSIBLE_LINT_TARGETS=roles/
```

Install only:

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#ansible-lint
```

### Included

```yaml
includes:
  ansible-lint: ./taskfiles/ansible-lint/Taskfile.yml
```

Then run:

```sh
task ansible-lint:ci
task ansible-lint:ci:fix ANSIBLE_LINT_TARGETS=playbooks/
```

## Public Tasks

| Task     | Description                                              |
| -------- | -------------------------------------------------------- |
| `ci`     | Lint Ansible YAML files with ansible-lint                |
| `ci:fix` | Auto-fix Ansible YAML files with ansible-lint `--fix` |

## Variables

| Variable                         | Default                | Description |
| -------------------------------- | ---------------------- | ----------- |
| `ANSIBLE_LINT_NIX_INSTALLABLE`   | `nixpkgs#ansible-lint` | Flake installable passed to `nix:install:profile` |
| `ANSIBLE_LINT_CONFIG`            | `ansible/ansible.cfg`  | Value exported as `ANSIBLE_CONFIG` for ansible-lint |
| `ANSIBLE_LINT_TARGETS`           | `.`                    | Files or directories to lint with `ci` / `ci:fix` |
| `ANSIBLE_LINT_EXTRA_ARGS`        | _(empty)_              | Extra flags forwarded to ansible-lint |
| `ANSIBLE_LINT_LINT_SKIP_PATTERN` | _(empty)_              | Path glob passed to ansible-lint `--exclude` |

Pin a revision by overriding the installable, for example
`ANSIBLE_LINT_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#ansible-lint`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.
- Configure linting rules with an `.ansible-lint` file in your project root.
- Pair with the [`ansible`](../ansible/README.md) module when you also need playbook syntax checks or execution.
