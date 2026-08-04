# Python Taskfile Public Tasks

## What is this Taskfile?

A cross-platform Taskfile for installing Python, managing upgrades, and running
common project operations such as creating virtual environments, installing
dependencies, and executing scripts.

Python is installed via `uv python install --default`, which downloads the
pinned PYTHON_PIN_VERSION and exposes it as `python3`/`python` on PATH through
uv's shim directory. The `install` task also ensures uv itself is installed
through the local uv Taskfile before Python installation starts.

## Usage

### Standalone

```sh
task -t taskfiles/python/Taskfile.yml install
task -t taskfiles/python/Taskfile.yml version
task -t taskfiles/python/Taskfile.yml venv
```

### Included

```yaml
includes:
  python: ./taskfiles/python/Taskfile.yml
```

Then run:

```sh
task python:install
task python:venv
task python:pip:install
```

## Public Tasks

| Task           | Description                                | Key variables                |
| -------------- | ------------------------------------------- | ---------------------------- |
| `install`      | Install Python via uv if missing            | `PYTHON_PIN_VERSION`         |
| `install:undo` | Remove the uv-managed Python                 | `PYTHON_PIN_VERSION`         |
| `upgrade`      | Upgrade Python to the latest release         | `PYTHON_PIN_VERSION`         |
| `version`      | Show the installed Python version            | none                          |
| `verify`       | Show Python and pip versions                 | none                          |
| `venv`         | Create a virtual environment                 | `PYTHON_VENV`                        |
| `pip:install`  | Install packages from a requirements file    | `PYTHON_REQUIREMENTS`, `PYTHON_EXTRA_ARGS` |
| `run`          | Run a Python script                          | `PYTHON_FILE`, `PYTHON_ARGS`, `PYTHON_EXTRA_ARGS`  |

## Variables

| Variable             | Default                                | Description                                                       |
| -------------------- | --------------------------------------- | ------------------------------------------------------------------ |
| `PYTHON_PIN_VERSION` | `3.13`                                  | Python version installed by `install`, `install:undo`, `upgrade` |
| `PYTHON_VENV`                | `.venv`                                | Virtual environment directory used by `venv`                     |
| `PYTHON_REQUIREMENTS`        | `requirements.txt`                     | Requirements file used by `pip:install`                          |
| `PYTHON_FILE`                | _(empty)_                              | Script path; required by `run`                                   |
| `PYTHON_ARGS`                | _(empty)_                              | Positional arguments forwarded to the script in `run`            |
| `PYTHON_EXTRA_ARGS`          | _(empty)_                              | Extra flags forwarded to `pip install` or the Python interpreter |
| `UV_LOAD`             | `export PATH="$HOME/.local/bin:$PATH"` | Shell snippet that ensures the uv-managed Python is in PATH      |

## Notes

**All platforms** get Python from the same source: uv downloads a prebuilt
interpreter for PYTHON_PIN_VERSION and shims `python3`/`python` (and
`pip3`/`pip`) into `~/.local/bin` on macOS/Linux, or the uv-managed bin
directory on Windows. This keeps the installed Python isolated from any
OS-provided Python.

**`upgrade`** reinstalls PYTHON_PIN_VERSION with `--reinstall` to fetch the
latest available patch release for that version line.

**`install:undo`** is supported on macOS, Linux, and Windows since it only
removes the uv-managed Python, leaving any OS-provided Python untouched.
