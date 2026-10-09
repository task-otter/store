# Trivy Taskfile Public Tasks

## What is this Taskfile?

A Taskfile for installing Trivy, showing its version and running strict CI scans.
Uses the shared Nix
installer on Linux and macOS and the shared WinGet installer on Windows.

## Usage

### Standalone

```sh
task -t taskfiles/trivy/Taskfile.yml install
task -t taskfiles/trivy/Taskfile.yml version
```

### Included

```yaml
includes:
  trivy: ./taskfiles/trivy/Taskfile.yml
```

```sh
task trivy:install
task trivy:version
```

## Public Tasks

| Task       | Description                                            |
| ---------- | ------------------------------------------------------ |
| `ci:image` | Scan a container image with strict checks              |
| `ci:fs`    | Scan a filesystem with strict checks                   |
| `install`  | Install Trivy if missing from PATH                     |
| `version`  | Show the active Trivy version; auto-install if missing |

The default task lists available tasks. Installation checks load the Nix profile
on Unix or refresh the Windows PATH and WinGet links before looking for Trivy.
Version commands use the same environment setup so a new installation is
available within the current Task process.

## Variables

| Variable                   | Default                             | Description                                                |
| -------------------------- | ----------------------------------- | ---------------------------------------------------------- |
| `TRIVY_IMAGE`              | _(empty; required for image scans)_ | Container image reference                                  |
| `TRIVY_FS_TARGET`          | `.`                                 | Filesystem path relative to the caller's working directory |
| `TRIVY_SCANNERS`           | `vuln,misconfig,secret,license`     | Scanners used by the default image and filesystem policies |
| `TRIVY_IMAGE_ARGS`         | Strict image flags below            | Replace image scan policy                                  |
| `TRIVY_FS_ARGS`            | Strict filesystem flags below       | Replace filesystem scan policy                             |
| `TRIVY_IMAGE_EXTRA_ARGS`   | _(empty)_                           | Append image flags; later flags can override policy        |
| `TRIVY_FS_EXTRA_ARGS`      | _(empty)_                           | Append filesystem flags; later flags can override policy   |
| `TRIVY_NIX_INSTALLABLE`    | `nixpkgs#trivy`                     | Nix flake installable                                      |
| `TRIVY_WINGET_INSTALLABLE` | `AquaSecurity.Trivy`                | WinGet package ID                                          |

All variables support parent Taskfile and CLI overrides. The shared WinGet
variables, including `WINGET_VERSION`, `WINGET_SCOPE`, and `WINGET_SOURCE`, are
also available. Installation preserves an existing Trivy binary on PATH;
package overrides apply when Trivy is missing.

Pin a Nix revision with
`TRIVY_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#trivy`, or pass
`WINGET_VERSION=<version>` to select a Windows release.

## CI scans

```sh
task trivy:ci:image TRIVY_IMAGE=registry.example.com/app:tag
task trivy:ci:fs
task trivy:ci:fs TRIVY_FS_TARGET="path with spaces" \
  TRIVY_FS_EXTRA_ARGS="--format json --output report.json"
```

The two scans have independent policy variables. Defaults enable `vuln`,
`misconfig`, `secret` and `license` scanners, every severity including UNKNOWN,
`--exit-code 1`, `--detection-priority comprehensive`, `--ignore-unfixed=false`,
`--license-full`, `--include-deprecated-checks` and `--no-progress`.
Comprehensive detection favors coverage and may produce more false positives.

Image scans additionally use `--image-config-scanners misconfig,secret`,
`--removed-pkgs` (supported for Alpine) and `--exit-on-eol 1`. Filesystem scans
add `--include-dev-deps` (supported for npm, yarn and Gradle).
Scan failures propagate to Task on all supported operating systems.

This repository's top-level `ci` task uses the module's strict defaults without
policy overrides. Any finding fails CI, including LOW license notices and
UNKNOWN licenses. Full license scanning and development dependency scanning
remain enabled.

Use `TRIVY_SCANNERS` to select scanners when using the default policy flags.
Repeating `--scanners` in extra arguments adds scanners rather than replacing
the default list. Custom `TRIVY_IMAGE_ARGS` or `TRIVY_FS_ARGS` must specify
their own scanner list.

Set variables on an include to customize the tasks for a project:

```yaml
includes:
  trivy:
    taskfile: ./taskfiles/trivy/Taskfile.yml
    vars:
      TRIVY_IMAGE: registry.example.com/app:tag
      TRIVY_FS_TARGET: ./src
      TRIVY_FS_EXTRA_ARGS: --format json --output trivy-fs.json
```

`TRIVY_IMAGE_ARGS` and `TRIVY_FS_ARGS` replace their respective default flag
sets. Extra arguments append to those flags and can change the policy, add a
config file, select custom checks, or control reporting. These argument strings
are trusted shell fragments: quote values containing spaces for your OS shell.
Targets are passed separately and support spaces. Scans run from the caller's
working directory and honor Trivy configuration, ignore files and suppressions;
the strict defaults do not bypass project exceptions or enable experimental
scanners. Use a current Trivy release that supports the documented flags.

Flag references: [image CLI](https://trivy.dev/docs/latest/references/configuration/cli/trivy_image/)
and [filesystem CLI](https://trivy.dev/docs/latest/references/configuration/cli/trivy_filesystem/).

Package references: [Trivy installation documentation](https://www.trivy.dev/docs/latest/getting-started/installation/)
and [Microsoft WinGet manifests](https://github.com/microsoft/winget-pkgs/tree/master/manifests/a/AquaSecurity/Trivy).
