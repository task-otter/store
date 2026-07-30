// REPLACE_ME 2026
// SPDX-License-Identifier: Apache-2.0

package git_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

func publicTasks() []string {
	return []string{
		"add",
		"add:all",
		"auth:setup",
		"branch:create",
		"branch:delete",
		"branch:list",
		"branch:rename",
		"branch:switch",
		"clean",
		"clone",
		"commit",
		"commit:amend",
		"config:list",
		"config:user",
		"diff",
		"diff:staged",
		"fetch",
		"help",
		"init",
		"install",
		"install:undo",
		"log",
		"log:graph",
		"pr:create",
		"pr:open",
		"pull",
		"push",
		"push:force",
		"release:create",
		"remote:add",
		"remote:list",
		"remote:remove",
		"remote:set-url",
		"reset:hard",
		"reset:soft",
		"stash",
		"stash:drop",
		"stash:list",
		"stash:pop",
		"status",
		"sync",
		"tag:create",
		"tag:delete",
		"tag:list",
		"tag:push",
		"upgrade",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"BASE",
		"BODY",
		"BRANCH",
		"CLONE_DIR",
		"COMMIT",
		"COMMIT_MSG",
		"EMAIL",
		"EXTRA_ARGS",
		"FILES",
		"MERGE_METHOD",
		"MESSAGE",
		"NAME",
		"NOTES",
		"OWNER",
		"REMOTE",
		"REPO",
		"STASH_INDEX",
		"TAG",
		"TITLE",
		"URL",
		"VERSION",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "git", publicTasks(), publicVars())
}
