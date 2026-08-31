# git — Git Taskfile

## What is this Taskfile?

A production-ready, cross-platform Taskfile for everyday git operations and GitHub-integrated
workflows. It wraps the `git` CLI with consistent defaults and integrates with the
[GitHub CLI (`gh`)](https://cli.github.com) for authentication, pull requests, and releases.

Run `task auth:setup` once to configure git credential delegation to gh — after that,
`clone`, `push`, `pull`, `pr:create`, and `release:create` all authenticate transparently.

## Usage

### Standalone

```sh
task -t taskfiles/git/Taskfile.yml auth:setup
task -t taskfiles/git/Taskfile.yml clone GIT_OWNER=github GIT_REPO=cli
task -t taskfiles/git/Taskfile.yml commit GIT_COMMIT_MSG="feat: add login page"
task -t taskfiles/git/Taskfile.yml pr:create GIT_TITLE="feat: login page" GIT_BASE=main
```

Install git only:

```sh
task nix:install:profile NIX_INSTALLABLE=nixpkgs#git
```

### Included (recommended)

```yaml
# Taskfile.yml
includes:
  git: ./taskfiles/git/Taskfile.yml
```

Then run:

```sh
task git:auth:setup
task git:commit GIT_COMMIT_MSG="feat: add feature"
task git:pr:create GIT_TITLE="feat: add feature" GIT_BASE=main
```

## Public Tasks

| Task             | Description                                                     | Key variables                     |
| ---------------- | --------------------------------------------------------------- | --------------------------------- |
| `auth:setup`     | Configure git to use gh as credential helper                    | —                                 |
| `init`           | Initialize a new git repository                                 | `GIT_BRANCH`                          |
| `clone`          | Clone a GitHub repository using the GitHub CLI                  | `GIT_REPO`, `GIT_OWNER`, `GIT_CLONE_DIR`      |
| `status`         | Show the current working tree status                            | —                                 |
| `add`            | Stage files for the next commit                                 | `GIT_FILES`                           |
| `add:all`        | Stage all changes including untracked files                     | —                                 |
| `commit`         | Create a commit from staged changes                             | `GIT_COMMIT_MSG`                      |
| `commit:amend`   | Amend the most recent commit                                    | `GIT_COMMIT_MSG`                      |
| `push`           | Push commits to the remote repository                           | `GIT_REMOTE`, `GIT_BRANCH`                |
| `push:force`     | Force-push using --force-with-lease                             | `GIT_REMOTE`, `GIT_BRANCH`                |
| `pull`           | Pull changes from the remote repository                         | `GIT_REMOTE`, `GIT_BRANCH`                |
| `fetch`          | Fetch branches from the remote without merging                  | `GIT_REMOTE`                          |
| `sync`           | Sync the current branch with its GitHub upstream                | `GIT_BRANCH`                          |
| `diff`           | Show unstaged changes in the working tree                       | `GIT_FILES`                           |
| `diff:staged`    | Show changes staged for the next commit                         | —                                 |
| `log`            | Show the commit history with author and date                    | `GIT_EXTRA_ARGS`                      |
| `log:graph`      | Show commit history as an ASCII branch graph                    | `GIT_EXTRA_ARGS`                      |
| `branch:list`    | List all local and remote branches                              | —                                 |
| `branch:create`  | Create and switch to a new branch from current HEAD             | `GIT_BRANCH`                          |
| `branch:switch`  | Switch to an existing branch                                    | `GIT_BRANCH`                          |
| `branch:delete`  | Delete a local branch                                           | `GIT_BRANCH`                          |
| `branch:rename`  | Rename the current branch to a new name                         | `GIT_BRANCH`                          |
| `tag:list`       | List all tags sorted by version descending                      | —                                 |
| `tag:create`     | Create an annotated tag at HEAD                                 | `GIT_TAG`, `GIT_MESSAGE`                  |
| `tag:push`       | Push a tag or all tags to the remote                            | `GIT_TAG`, `GIT_REMOTE`                   |
| `tag:delete`     | Delete a tag locally and from the remote                        | `GIT_TAG`, `GIT_REMOTE`                   |
| `stash`          | Stash uncommitted changes in the working tree                   | `GIT_MESSAGE`                         |
| `stash:pop`      | Apply and remove the latest stash entry                         | `GIT_STASH_INDEX`                     |
| `stash:list`     | List all stash entries                                          | —                                 |
| `stash:drop`     | Discard a stash entry without applying it                       | `GIT_STASH_INDEX`                     |
| `reset:soft`     | Soft-reset HEAD to a commit, preserving staged changes          | `GIT_COMMIT`                          |
| `reset:hard`     | Hard-reset HEAD to a commit, discarding all local changes       | `GIT_COMMIT`                          |
| `clean`          | Remove untracked files and directories from the working tree    | —                                 |
| `config:user`    | Set the global git user name and email address                  | `GIT_NAME`, `GIT_EMAIL`                   |
| `config:list`    | List all git configuration values                               | —                                 |
| `remote:list`    | List all configured remotes and their URLs                      | —                                 |
| `remote:add`     | Add a new remote to the repository                              | `GIT_NAME`, `GIT_URL`                     |
| `remote:remove`  | Remove a configured remote from the repository                  | `GIT_NAME`                            |
| `remote:set-url` | Update the URL of a configured remote                           | `GIT_NAME`, `GIT_URL`                     |
| `pr:create`      | Push the current branch and open a pull request on GitHub       | `GIT_TITLE`, `GIT_BASE`, `GIT_BODY`, `GIT_REMOTE` |
| `pr:open`        | Open the current pull request in the browser via the GitHub CLI | —                                 |
| `release:create` | Create a git tag and a GitHub release via the GitHub CLI        | `GIT_TAG`, `GIT_TITLE`, `GIT_NOTES`, `GIT_REMOTE` |
| `help`           | Show the git built-in help and command list                     | —                                 |

## Variables

| Variable       | Default   | Description                                           |
| -------------- | --------- | ----------------------------------------------------- |
| `GIT_REMOTE`       | `origin`  | Remote name for push, pull, fetch, and tag operations |
| `GIT_BASE`         | `main`    | Base branch for pull requests                         |
| `GIT_MERGE_METHOD` | `merge`   | PR merge strategy: `merge`, `squash`, `rebase`        |
| `GIT_FILES`        | `.`       | Files or globs for `add` and `diff`                   |
| `GIT_STASH_INDEX`  | `0`       | Stash entry index for `stash:pop` and `stash:drop`    |
| `GIT_BRANCH`       | _(empty)_ | Branch name                                           |
| `GIT_CLONE_DIR`    | _(empty)_ | Destination directory for `clone`                     |
| `GIT_COMMIT`       | _(empty)_ | Commit ref for `reset:soft` and `reset:hard`          |
| `GIT_COMMIT_MSG`   | _(empty)_ | Commit message                                        |
| `GIT_EMAIL`        | _(empty)_ | Git user email for `config:user`                      |
| `GIT_NAME`         | _(empty)_ | Git user name or remote name                          |
| `GIT_NOTES`        | _(empty)_ | Release notes body                                    |
| `GIT_OWNER`        | _(empty)_ | GitHub user or organisation for `clone`               |
| `GIT_REPO`         | _(empty)_ | Repository name for `clone`                           |
| `GIT_TAG`          | _(empty)_ | Tag name                                              |
| `GIT_TITLE`        | _(empty)_ | PR or release title                                   |
| `GIT_BODY`         | _(empty)_ | PR description                                        |
| `GIT_URL`          | _(empty)_ | Remote URL for `remote:add` and `remote:set-url`      |
| `GIT_MESSAGE`      | _(empty)_ | Tag annotation or stash description                   |
| `GIT_EXTRA_ARGS`       | _(empty)_ | Extra arguments appended to the underlying command    |
| `GIT_NIX_INSTALLABLE`  | `nixpkgs#git` | Flake installable passed to `nix:install:profile` |

## Examples

```sh
# Set up gh as git credential helper (run once)
task -t taskfiles/git/Taskfile.yml auth:setup

# Clone a repository
task -t taskfiles/git/Taskfile.yml clone GIT_OWNER=github GIT_REPO=cli
task -t taskfiles/git/Taskfile.yml clone GIT_OWNER=myorg GIT_REPO=private-repo GIT_CLONE_DIR=~/src/project

# Stage and commit
task -t taskfiles/git/Taskfile.yml add GIT_FILES=src/
task -t taskfiles/git/Taskfile.yml commit GIT_COMMIT_MSG="feat: add login page"
task -t taskfiles/git/Taskfile.yml commit:amend GIT_COMMIT_MSG="feat: add login page with tests"

# Push and sync
task -t taskfiles/git/Taskfile.yml push
task -t taskfiles/git/Taskfile.yml push:force GIT_REMOTE=origin
task -t taskfiles/git/Taskfile.yml pull GIT_REMOTE=origin GIT_BRANCH=main
task -t taskfiles/git/Taskfile.yml sync GIT_BRANCH=main

# Branches
task -t taskfiles/git/Taskfile.yml branch:create GIT_BRANCH=feature/my-feature
task -t taskfiles/git/Taskfile.yml branch:switch GIT_BRANCH=main
task -t taskfiles/git/Taskfile.yml branch:delete GIT_BRANCH=feature/old-feature
task -t taskfiles/git/Taskfile.yml branch:rename GIT_BRANCH=new-name

# Pull requests (push + gh pr create in one step)
task -t taskfiles/git/Taskfile.yml pr:create GIT_TITLE="feat: login page" GIT_BASE=main
task -t taskfiles/git/Taskfile.yml pr:create GIT_TITLE="fix: auth bug" GIT_BASE=develop GIT_BODY="Closes #42"
task -t taskfiles/git/Taskfile.yml pr:open

# Tags and releases
task -t taskfiles/git/Taskfile.yml tag:create GIT_TAG=v1.0.0 GIT_MESSAGE="Release v1.0.0"
task -t taskfiles/git/Taskfile.yml tag:push GIT_TAG=v1.0.0
task -t taskfiles/git/Taskfile.yml release:create GIT_TAG=v1.0.0 GIT_TITLE="v1.0.0" GIT_NOTES="Initial release"
task -t taskfiles/git/Taskfile.yml tag:delete GIT_TAG=v0.1.0

# Stash
task -t taskfiles/git/Taskfile.yml stash GIT_MESSAGE="WIP: refactor auth"
task -t taskfiles/git/Taskfile.yml stash:pop
task -t taskfiles/git/Taskfile.yml stash:list

# Reset and clean
task -t taskfiles/git/Taskfile.yml reset:soft GIT_COMMIT=HEAD~1
task -t taskfiles/git/Taskfile.yml reset:hard GIT_COMMIT=HEAD~1
task -t taskfiles/git/Taskfile.yml clean

# Config
task -t taskfiles/git/Taskfile.yml config:user GIT_NAME="Ada Lovelace" GIT_EMAIL=ada@example.com
task -t taskfiles/git/Taskfile.yml config:list

# Remotes
task -t taskfiles/git/Taskfile.yml remote:list
task -t taskfiles/git/Taskfile.yml remote:add GIT_NAME=upstream GIT_URL=https://github.com/org/repo.git
task -t taskfiles/git/Taskfile.yml remote:set-url GIT_NAME=origin GIT_URL=git@github.com:org/repo.git
```

Pin a revision by overriding the installable, for example
`GIT_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#git`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.
- Tasks that call `gh` auto-install the GitHub CLI via `gh:nix:install:profile`.

