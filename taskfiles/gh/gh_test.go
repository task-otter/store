// Copyright 2026 task-otter
// SPDX-License-Identifier: Apache-2.0

package gh_test

import (
	"slices"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// expectedPublicTasks is the canonical list of public gh Taskfile tasks.
// It must stay in sync with the tasks: block in Taskfile.yml.
func expectedPublicTasks() []string {
	return slices.Concat(ghCoreTasks(), ghIssueAndPRTasks(), ghRepoAndWorkflowTasks())
}

func ghCoreTasks() []string {
	return []string{
		"alias:delete",
		"alias:list",
		"alias:set",
		"api:delete:danger",
		"api:get",
		"api:patch",
		"api:post",
		"auth:login",
		"auth:login:ssh",
		"auth:login:web",
		"auth:logout",
		"auth:refresh",
		"auth:setup-git",
		"auth:status",
		"browse",
		"config:get",
		"config:list",
		"config:set",
		"doctor",
		"extension:install",
		"extension:list",
		"extension:remove",
		"extension:upgrade",
		"gist:create",
		"gist:delete:danger",
		"gist:list",
		"gist:view",
		"help",
		"install",
		"install:undo",
	}
}

func ghIssueAndPRTasks() []string {
	return []string{
		"issue:assign",
		"issue:close",
		"issue:comment",
		"issue:create",
		"issue:label",
		"issue:list",
		"issue:reopen",
		"issue:view",
		"open",
		"pr:checkout",
		"pr:close",
		"pr:comment",
		"pr:create",
		"pr:diff",
		"pr:list",
		"pr:merge",
		"pr:ready",
		"pr:review",
		"pr:status",
		"pr:view",
		"project:create",
		"project:list",
		"project:view",
		"release:create",
		"release:delete:danger",
		"release:download",
		"release:download:all",
		"release:list",
		"release:upload",
		"release:view",
	}
}

func ghRepoAndWorkflowTasks() []string {
	return []string{
		"repo:archive",
		"repo:clone",
		"repo:create",
		"repo:delete:danger",
		"repo:fork",
		"repo:list",
		"repo:sync",
		"repo:view",
		"run:cancel",
		"run:list",
		"run:logs",
		"run:rerun",
		"run:view",
		"search:issues",
		"search:prs",
		"search:repos",
		"secret:delete:danger",
		"secret:list",
		"secret:set",
		"ssh-key:add",
		"ssh-key:delete:danger",
		"ssh-key:list",
		"upgrade",
		"variable:delete:danger",
		"variable:list",
		"variable:set",
		"verify",
		"version",
		"which",
		"workflow:list",
		"workflow:run",
		"workflow:view",
		"workflow:watch",
	}
}

// expectedVars is the list of required top-level vars with non-empty defaults.
func expectedVars() []string {
	return []string{
		"BASE",
		"CLONE_DIR",
		"DATA",
		"DOWNLOAD_DIR",
		"MERGE_METHOD",
		"PAT_TOKEN",
		"VERSION",
		"VISIBILITY",
	}
}

func TestModule(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "gh", expectedPublicTasks(), expectedVars())
}
