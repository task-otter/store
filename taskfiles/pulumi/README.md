# Pulumi Taskfile

## What is this Taskfile?

A cross-platform Taskfile for installing Pulumi and running the common project
lifecycle commands — login, project scaffolding, and stack deploys. macOS
installs via Homebrew (`pulumi/tap/pulumi`), Linux uses the official install
script at `https://get.pulumi.com` (drops the binary into `~/.pulumi/bin`),
and Windows uses Chocolatey (`choco install pulumi`).

## Usage

### Standalone

```sh
task -t taskfiles/pulumi/Taskfile.yml install
task -t taskfiles/pulumi/Taskfile.yml version
task -t taskfiles/pulumi/Taskfile.yml up PULUMI_STACK=dev
```

### Included

```yaml
includes:
  pulumi: ./taskfiles/pulumi/Taskfile.yml
```

Then run:

```sh
task pulumi:install
task pulumi:login PULUMI_LOGIN_URL=s3://my-pulumi-state
task pulumi:new PULUMI_TEMPLATE=aws-typescript
task pulumi:up PULUMI_STACK=dev PULUMI_EXTRA_ARGS=--yes
```

## Auto-install behaviour

Every task that requires Pulumi automatically installs it first if it is not
already present — you do not need to run `task install` manually before using
`version`, `upgrade`, `login`, `new`, or `up`.

Installs are **idempotent**: the internal install task has a `status` check
that exits early when Pulumi is already available on `PATH` (or, on Linux,
already at `~/.pulumi/bin/pulumi` for the requested `PULUMI_VERSION`).

## Public Tasks

| Task           | Variables                                              | Description                                                                                     |
| `install`      | Optional `PULUMI_VERSION` (Linux only)                 | Install Pulumi for the current OS. Homebrew on macOS, install script on Linux, Chocolatey on Windows. |
| `install:undo` | —                                                      | Remove Pulumi for the current OS.                                                              |
| `upgrade`      | —                                                      | Upgrade Pulumi to the latest release. Auto-installs Pulumi if missing.                          |
| `version`      | —                                                      | Show the installed Pulumi version via `pulumi version`. Auto-installs Pulumi if missing.        |
| `login`        | Optional `PULUMI_LOGIN_URL`, `PULUMI_EXTRA_ARGS`       | Run `pulumi login`. Empty `PULUMI_LOGIN_URL` uses the default Pulumi Cloud backend.             |
| `new`          | Required `PULUMI_TEMPLATE`; optional `PULUMI_EXTRA_ARGS` | Scaffold a new Pulumi project from a named template (for example `aws-typescript`).           |
| `up`           | Optional `PULUMI_STACK`, `PULUMI_EXTRA_ARGS`           | Preview and deploy the current Pulumi stack in the caller's working directory.                  |

## Variables

| Variable                 | Default                              | Description                                                                    |
| `PULUMI_VERSION`         | _(empty)_                            | Pin a specific Pulumi release for the Linux install script; empty installs latest. |
| `PULUMI_INSTALL_URL`     | `https://get.pulumi.com`             | Linux install script URL.                                                      |
| `PULUMI_INSTALL_PS1_URL` | `https://get.pulumi.com/install.ps1` | Reference URL for the official PowerShell installer; Windows uses Chocolatey by default. |
| `PULUMI_LOAD`            | `export PATH="$HOME/.pulumi/bin:$PATH"` | Shell snippet that adds `~/.pulumi/bin` to `PATH` on Unix.                     |
| `PULUMI_LOGIN_URL`       | _(empty)_                            | Backend URL passed to `pulumi login`; empty uses the default Pulumi Cloud backend. |
| `PULUMI_TEMPLATE`        | _(empty)_                            | Template name; required by `new`.                                              |
| `PULUMI_STACK`           | _(empty)_                            | Stack name; optional for `up`. When set, passed as `--stack <name>`.           |
| `PULUMI_ARGS`            | _(empty)_                            | Positional arguments forwarded to underlying Pulumi commands.                  |
| `PULUMI_EXTRA_ARGS`      | _(empty)_                            | Extra flags forwarded to the underlying `pulumi` command (for example `--yes`). |

## Notes

**After installation** on Linux, the Pulumi binary is placed at
`~/.pulumi/bin/pulumi`. Restart your shell for it to be available on `PATH`,
or let tasks locate it automatically via `PULUMI_LOAD`.

**`install:undo`** on macOS uses `brew uninstall pulumi` when Pulumi was
installed via Homebrew, otherwise falls back to removing `~/.pulumi`
directly. Linux always removes `~/.pulumi` and reports any shell profile
files that still reference `~/.pulumi/bin`. Windows uses
`choco uninstall pulumi -y`.

## Security Notes

**Linux install script**: The `install` and `upgrade` tasks download the
official Pulumi install script from `PULUMI_INSTALL_URL` to a temporary file
using `curl -fsSL`, then execute it with `sh`. The temporary file is removed
after execution via a shell `trap`. This relies on HTTPS transport for
integrity — no additional checksum verification is performed. Review the
script at `PULUMI_INSTALL_URL` before running in security-sensitive
environments.
