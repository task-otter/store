# Pulumi Taskfile

## What is this Taskfile?

A Taskfile for running the common Pulumi project lifecycle commands — login,
project scaffolding, and stack deploys. Operational tasks auto-install Pulumi
via `nix:install:profile`.

## Usage

### Standalone

```sh
task -t taskfiles/pulumi/Taskfile.yml up PULUMI_STACK=dev
```

Install only:

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#pulumi
```

### Included

```yaml
includes:
  pulumi: ./taskfiles/pulumi/Taskfile.yml
```

Then run:

```sh
task pulumi:login PULUMI_LOGIN_URL=s3://my-pulumi-state
task pulumi:new PULUMI_TEMPLATE=aws-typescript
task pulumi:up PULUMI_STACK=dev PULUMI_EXTRA_ARGS=--yes
```

## Public Tasks

| Task    | Variables                                                | Description                                                                           |
| ------- | -------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| `login` | Optional `PULUMI_LOGIN_URL`, `PULUMI_EXTRA_ARGS`         | Run `pulumi login`. Empty `PULUMI_LOGIN_URL` uses the default Pulumi Cloud backend.   |
| `new`   | Required `PULUMI_TEMPLATE`; optional `PULUMI_EXTRA_ARGS` | Scaffold a new Pulumi project from a named template (for example `aws-typescript`). |
| `up`    | Optional `PULUMI_STACK`, `PULUMI_EXTRA_ARGS`             | Preview and deploy the current Pulumi stack in the caller's working directory.        |

## Variables

| Variable                   | Default            | Description                                                                    |
| -------------------------- | ------------------ | ------------------------------------------------------------------------------ |
| `PULUMI_NIX_INSTALLABLE`   | `nixpkgs#pulumi`   | Flake installable passed to `nix:install:profile`                              |
| `PULUMI_LOGIN_URL`         | _(empty)_          | Backend URL passed to `pulumi login`; empty uses the default Pulumi Cloud backend. |
| `PULUMI_TEMPLATE`          | _(empty)_          | Template name; required by `new`.                                              |
| `PULUMI_STACK`             | _(empty)_          | Stack name; optional for `up`. When set, passed as `--stack <name>`.           |
| `PULUMI_ARGS`              | _(empty)_          | Positional arguments forwarded to underlying Pulumi commands.                  |
| `PULUMI_EXTRA_ARGS`        | _(empty)_          | Extra flags forwarded to the underlying `pulumi` command (for example `--yes`). |

Pin a revision by overriding the installable, for example
`PULUMI_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#pulumi`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.
- Every task that requires Pulumi automatically installs it first if it is not already present.
