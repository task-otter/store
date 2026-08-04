# bash-exec

A [TaskOtter](https://github.com/task-otter/store) module for executing and syntax-checking Bash scripts already available on the host.

## What is this Taskfile?

This module is a small Bash runner rather than an installer. It executes a script or a supplied command string with the system `bash` binary, and provides syntax validation without execution.

## Usage

### Standalone

```sh
task -t taskfiles/bash-exec/Taskfile.yml run BASH_EXEC_SCRIPT=scripts/build.sh BASH_EXEC_ARGS="--release"
task -t taskfiles/bash-exec/Taskfile.yml check BASH_EXEC_SCRIPT=scripts/build.sh
task -t taskfiles/bash-exec/Taskfile.yml exec BASH_EXEC_COMMAND='printf "hello\\n"'
```

### Included in your Taskfile

```yaml
includes:
  bash-exec:
    taskfile: taskfiles/bash-exec/Taskfile.yml
    vars:
      BASH_EXEC_SCRIPT_OVERRIDE: "{{.BASH_EXEC_SCRIPT}}"
      BASH_EXEC_COMMAND_OVERRIDE: "{{.BASH_EXEC_COMMAND}}"
      BASH_EXEC_ARGS_OVERRIDE: "{{.BASH_EXEC_ARGS}}"
```

Then run:

```sh
task bash-exec:run BASH_EXEC_SCRIPT=scripts/build.sh
task bash-exec:check BASH_EXEC_SCRIPT=scripts/build.sh
```

## Public Tasks

| Task | Description |
|---|---|
| `run` | Run a Bash script (`BASH_EXEC_SCRIPT=path`) |
| `check` | Check a Bash script for syntax errors (`BASH_EXEC_SCRIPT=path`) |
| `exec` | Execute a Bash command string (`BASH_EXEC_COMMAND=...`) |
| `version` | Show the installed Bash version |

## Variables

| Variable | Default | Description |
|---|---|---|
| `BASH_EXEC_SCRIPT` | _(empty)_ | Bash script path; required by `run` and `check` |
| `BASH_EXEC_ARGS` | _(empty)_ | Positional arguments passed to the script by `run` |
| `BASH_EXEC_FLAGS` | _(empty)_ | Options passed to Bash, for example `-euo pipefail` |
| `BASH_EXEC_COMMAND` | _(empty)_ | Command string executed by `exec` |

## Notes

- Bash must already be installed and available in `PATH`; this module never installs or removes it.
- `check` uses `bash -n`, which checks syntax but does not execute the script.
- `exec` intentionally accepts a command string. Treat its value as trusted input, just as you would a command entered directly in a shell.
