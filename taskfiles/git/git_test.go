// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package git_test

import (
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

// TestTaskfileModuleContract
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		"git",
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

func publicTasks() []string {
	return append(publicTasksCore(), publicTasksReleaseAndSync()...)
}

func publicTasksCoreA() []string {
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
	}
}

func publicTasksCoreB() []string {
	return []string{
		"commit:amend",
		"config:list",
		"config:user",
		"diff",
		"diff:staged",
		"fetch",
		"help",
		"init",
		"log",
		"log:graph",
	}
}

func publicTasksCore() []string {
	return append(publicTasksCoreA(), publicTasksCoreB()...)
}

func publicTasksReleaseAndSyncA() []string {
	return []string{
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
	}
}

func publicTasksReleaseAndSyncB() []string {
	return []string{
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
	}
}

func publicTasksReleaseAndSync() []string {
	return append(publicTasksReleaseAndSyncA(), publicTasksReleaseAndSyncB()...)
}

func publicVars() []string {
	return append(publicVarsCore(), publicVarsExtra()...)
}

func publicVarsCore() []string {
	return []string{
		"GIT_BASE",
		"GIT_BODY",
		"GIT_BRANCH",
		"GIT_CLONE_DIR",
		"GIT_COMMIT",
		"GIT_COMMIT_MSG",
		"GIT_EMAIL",
		"GIT_EXTRA_ARGS",
		"GIT_FILES",
		"GIT_MERGE_METHOD",
	}
}

func publicVarsExtra() []string {
	return []string{
		"GIT_MESSAGE",
		"GIT_NAME",
		"GIT_NOTES",
		"GIT_OWNER",
		"GIT_REMOTE",
		"GIT_REPO",
		"GIT_STASH_INDEX",
		"GIT_TAG",
		"GIT_TITLE",
		"GIT_URL",
		"GIT_NIX_INSTALLABLE",
	}
}
