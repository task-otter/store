# bencher

A [TaskOtter](https://github.com/task-otter/store) module for the [Bencher CLI](https://bencher.dev/docs/how-to/install-cli/), which uploads and tracks benchmark results.

## What is this Taskfile?

This module installs and manages the Bencher CLI using Bencher's official installer. It supports Bencher Cloud's latest CLI by default and can pin a compatible release when using Bencher Self-Hosted.

## Usage

### Standalone

```sh
task -t taskfiles/bencher/Taskfile.yml install
task -t taskfiles/bencher/Taskfile.yml version
task -t taskfiles/bencher/Taskfile.yml install BENCHER_VERSION=0.6.8
```

### Included in your Taskfile

```yaml
includes:
  bencher:
    taskfile: taskfiles/bencher/Taskfile.yml
```

Then run:

```sh
task bencher:install
task bencher:version
task bencher:run -- --project my-project "make benchmarks"
task bencher:exec -- mock
```

## Public Tasks

| Task | Description |
|---|---|
| `install` | Install the Bencher CLI on the current operating system |
| `upgrade` | Upgrade the Bencher CLI with the official installer |
| `run` | Execute a benchmark command and track its results with `bencher run` |
| `exec` | Run any Bencher CLI subcommand (e.g. `mock`, `project`, `report`) |
| `version` | Show the installed Bencher CLI version, installing it if missing |

## Variables

| Variable | Default | Description |
|---|---|---|
| `BENCHER_INSTALL_URL` | `https://bencher.dev/download/install-cli.sh` | Unix installer URL; override with your self-hosted Bencher instance URL if needed |
| `BENCHER_INSTALL_URL_WINDOWS` | `https://bencher.dev/download/install-cli.ps1` | Windows installer URL; override with your self-hosted Bencher instance URL if needed |
| `BENCHER_VERSION` | `""` (latest) | Exact CLI release for Bencher Self-Hosted; leave empty for the latest Bencher Cloud CLI |
| `BENCHER_EXTRA_ARGS` | `""` | Arguments and flags appended to `run` and `exec` invocations |

## Notes

- macOS, Linux, and other Unix-like systems run Bencher's documented `curl --proto '=https' --tlsv1.2 -sSfL … | sh` installer.
- Windows runs Bencher's documented `irm … | iex` installer in PowerShell.
- Bencher recommends leaving `BENCHER_VERSION` empty for Bencher Cloud. Pin it only when using Bencher Self-Hosted.
- For Bencher Self-Hosted, set the installer URL to your instance's `/download/install-cli.sh` or `/download/install-cli.ps1` endpoint.
- The installer may update your shell profile. Restart the shell or terminal if `bencher` is not immediately available in `PATH`.
