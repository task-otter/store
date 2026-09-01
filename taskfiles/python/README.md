# Python Taskfile Public Tasks

## What is this Taskfile?

A cross-platform Taskfile for creating virtual environments, installing
dependencies, running scripts, and verifying Python. Remaining tasks
auto-install Python via `nix:install:profile` (`nixpkgs#python3`).

## Usage

### Standalone

```sh
task -t taskfiles/python/Taskfile.yml venv
task -t taskfiles/python/Taskfile.yml pip:install
task -t taskfiles/python/Taskfile.yml verify
```

Install only:

```sh
task -t taskfiles/python/Taskfile.yml install
```

### Included

```yaml
includes:
  python: ./taskfiles/python/Taskfile.yml
```

Then run:

```sh
task python:venv
task python:pip:install
task python:run PYTHON_FILE=script.py
```

## Public Tasks

| Task           | Description                                |
| -------------- | ------------------------------------------- |
| `verify`       | Show Python and pip versions                 |
| `venv`         | Create a virtual environment                 |
| `pip:install`  | Install packages from a requirements file    |
| `run`          | Run a Python script                          |
| `install`      | Install Python via the Nix profile           |
| `version`      | Show the active Python version               |

## Variables

| Variable             | Default                                | Description                                                       |
| -------------------- | --------------------------------------- | ------------------------------------------------------------------ |
| `PYTHON_NIX_INSTALLABLE` | `nixpkgs#python3`                   | Flake installable passed to `nix:install:profile` |
| `PYTHON_VENV`                | `.venv`                                | Virtual environment directory used by `venv`                     |
| `PYTHON_REQUIREMENTS`        | `requirements.txt`                     | Requirements file used by `pip:install`                          |
| `PYTHON_FILE`                | _(empty)_                              | Script path; required by `run`                                   |
| `PYTHON_ARGS`                | _(empty)_                              | Positional arguments forwarded to the script in `run`            |
| `PYTHON_EXTRA_ARGS`          | _(empty)_                              | Extra flags forwarded to `pip install` or the Python interpreter |

Pin a Python version by overriding the installable, for example
`PYTHON_NIX_INSTALLABLE=nixpkgs#python313`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.
- `venv`, `pip:install`, `run`, and `verify` use `python3` from PATH (after `NIX_LOAD` on Unix). `pip:install` runs `python3 -m pip`.
