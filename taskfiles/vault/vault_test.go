// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package vault_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
	yaml "go.yaml.in/yaml/v3"
)

type (
	cmdsAssertion struct {
		content string
		msgFmt  string
		tokens  []string
	}

	approleCredentialsCheck struct {
		msgPrefix   string
		cmds        string
		extraTokens []string
	}
)

const (
	initTask                  = "init"
	loginTask                 = "login"
	loginTaskApprole          = "login:approle"
	loginTaskRootAuth         = "login:root-token"
	restoreTask               = "restore"
	issueApproleTask          = "token:issue:approle"
	revokeSelfTask            = "token:revoke-self"
	unsealTask                = "unseal"
	roleEnvVarName            = "VAULT_ROLE_ID"
	approleKeyEnvVarName      = "VAULT_SECRET_ID"
	authMethodFlag            = "-method=token"
	noPrintFlag               = "-no-print"
	vaultLoginPipe            = `| vault login`
	healthTask                = "health"
	kvGetTask                 = "kv:get"
	vaultAddrEnvVarName       = "VAULT_ADDR"
	vaultModuleName           = "vault"
	vaultLoginValueArg        = `vault login token=`
	vaultLoginValueVar        = `vault login "$TOKEN"`
	printApproleKeyValueStdin = `printf '%s' "$VAULT_LOGIN_SECRET_ID"`
	vaultWritePipe            = `| vault write`
	fieldAuthFlag             = `-field=token`
	approleLoginPath          = `"auth/${VAULT_LOGIN_APPROLE_MOUNT}/login"`
	roleReferenceArg          = `"$VAULT_LOGIN_ROLE_ID"`
	approleKeyStdinArg        = `secret_id=-`
	vaultAuthEnvVarName       = "VAULT_TOKEN"
)

// TestTaskfileModuleContract validates the behavior covered by this test case.
func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(
		t,
		vaultModuleName,
		&tasktest.ModuleExpectations{Tasks: publicTasks(), Vars: publicVars()},
	)
}

// TestInputValidatedTasksDoNotInstallBeforePreconditions validates the behavior covered by this test case.
func TestInputValidatedTasksDoNotInstallBeforePreconditions(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, vaultModuleName)

	names := inputValidatedTaskNames()

	for i := range names {
		name := names[i]
		task := taskfile.Tasks[name]

		if task.Deps != nil {
			t.Fatalf(
				"%s should run install from cmds after local preconditions, got deps: %#v",
				name,
				task.Deps,
			)
		}
	}
}

// TestVerifyDoesNotMaskStatusFailures validates the behavior covered by this test case.
func TestVerifyDoesNotMaskStatusFailures(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, vaultModuleName)
	cmds := taskFieldYAML(t, taskfile.Tasks["verify"].Cmds)

	if !strings.Contains(cmds, "vault status") {
		t.Fatalf("verify should run vault status\ncmds:\n%s", cmds)
	}

	if strings.Contains(cmds, "vault status || true") {
		t.Fatalf("verify should fail when vault status fails\ncmds:\n%s", cmds)
	}
}

// TestInitDoesNotOverwriteExistingKeysFile validates the behavior covered by this test case.
func TestInitDoesNotOverwriteExistingKeysFile(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, vaultModuleName)
	task := taskfile.Tasks[initTask]
	preconditions := taskFieldYAML(t, task.Preconditions)
	cmds := taskFieldYAML(t, task.Cmds)

	assertInitPreconditionsRefuseExistingKeysFile(t, preconditions)
	assertInitStagesOutputSafely(t, cmds)
}

// TestLoginDoesNotPassRootTokenAsCommandArgument validates the behavior covered by this test case.
func TestLoginDoesNotPassRootTokenAsCommandArgument(t *testing.T) {
	t.Parallel()

	cmds := loginTaskCmds(t, loginTask)

	assertCmdsContainAll(t, &cmdsAssertion{
		content: cmds,
		tokens: []string{
			`jq -r '.root_token' "$KF"`,
			vaultLoginPipe,
			authMethodFlag,
			noPrintFlag,
		},
		msgFmt: "login should pipe the root token to vault login stdin (missing %q)\ncmds:\n%s",
	})

	assertCmdsContainNone(t, &cmdsAssertion{
		content: cmds,
		tokens:  []string{`vault login "$(jq`, vaultLoginValueArg, vaultLoginValueVar},
		msgFmt:  "login should not expose the root token as a command argument (found %q)\ncmds:\n%s",
	})
}

// TestLoginRootTokenPipesTokenViaStdin validates the behavior covered by this test case.
func TestLoginRootTokenPipesTokenViaStdin(t *testing.T) {
	t.Parallel()

	cmds := loginTaskCmds(t, loginTaskRootAuth)

	assertCmdsContainAll(t, &cmdsAssertion{
		content: cmds,
		tokens: []string{
			`printf '%s' "$VAULT_LOGIN_ROOT_TOKEN"`,
			vaultLoginPipe,
			authMethodFlag,
			noPrintFlag,
		},
		msgFmt: "login:root-token should pipe token to vault login stdin (missing %q)\ncmds:\n%s",
	})

	assertCmdsContainNone(t, &cmdsAssertion{
		content: cmds,
		tokens:  []string{vaultLoginValueVar, vaultLoginValueArg},
		msgFmt:  "login:root-token should not expose token as command argument (found %q)\ncmds:\n%s",
	})
}

// TestLoginApproleRequiresBothCredentials validates the behavior covered by this test case.
func TestLoginApproleRequiresBothCredentials(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, vaultModuleName)
	task := taskfile.Tasks[loginTaskApprole]
	preconditions := taskFieldYAML(t, task.Preconditions)
	cmds := taskFieldYAML(t, task.Cmds)

	assertLoginApprolePreconditionsAndCmds(t, preconditions, cmds)
}

// TestStrictShellSetOnSensitiveTasks validates the behavior covered by this test case.
func TestStrictShellSetOnSensitiveTasks(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, vaultModuleName)

	names := strictShellTaskNames()

	for i := range names {
		assertTaskSetsStrictShellOptions(t, taskfile, names[i])
	}
}

// TestTokenIssueApRolePipesSecretViaStdinWithoutLogin validates the behavior covered by this test case.
func TestTokenIssueApRolePipesSecretViaStdinWithoutLogin(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, vaultModuleName)
	task := taskfile.Tasks[issueApproleTask]
	preconditions := taskFieldYAML(t, task.Preconditions)
	cmds := taskFieldYAML(t, task.Cmds)

	assertCmdsContainAll(t, &cmdsAssertion{
		content: preconditions,
		tokens:  []string{roleEnvVarName, approleKeyEnvVarName},
		msgFmt:  "token:issue:approle should require %s in preconditions\npreconditions:\n%s",
	})

	assertApproleCredentialsExchanged(
		t,
		&approleCredentialsCheck{msgPrefix: "token:issue:approle", cmds: cmds, extraTokens: nil},
	)

	assertTokenIssueApRoleSkipsVaultLogin(t, cmds)
}

// TestTokenRevokeSelfRequiresVaultToken validates the behavior covered by this test case.
func TestTokenRevokeSelfRequiresVaultToken(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, vaultModuleName)
	task := taskfile.Tasks[revokeSelfTask]
	preconditions := taskFieldYAML(t, task.Preconditions)
	cmds := taskFieldYAML(t, task.Cmds)

	if !strings.Contains(preconditions, vaultAuthEnvVarName) {
		t.Fatalf(
			"token:revoke-self should require VAULT_TOKEN in preconditions\npreconditions:\n%s",
			preconditions,
		)
	}

	if !strings.Contains(cmds, "-self") {
		t.Fatalf("token:revoke-self should pass -self to vault token revoke\ncmds:\n%s", cmds)
	}
}

// TestKvGetRequiresMountPathAndToken validates the behavior covered by this test case.
func TestKvGetRequiresMountPathAndToken(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, vaultModuleName)
	task := taskfile.Tasks[kvGetTask]
	preconditions := taskFieldYAML(t, task.Preconditions)
	cmds := taskFieldYAML(t, task.Cmds)

	assertCmdsContainAll(t, &cmdsAssertion{
		content: preconditions,
		tokens:  []string{"KV_GET_MOUNT", "KV_GET_PATH", vaultAuthEnvVarName, vaultAddrEnvVarName},
		msgFmt:  "kv:get should require %s in preconditions\npreconditions:\n%s",
	})

	assertKvGetCallsVaultWithJSON(t, cmds)
}

func publicTasksA() []string {
	return []string{
		healthTask,
		initTask,
		kvGetTask,
		loginTask,
		loginTaskApprole,
		loginTaskRootAuth,
		"peers",
		restoreTask,
	}
}

func publicTasksB() []string {
	return []string{
		"root-token",
		"seal",
		"snapshot",
		"status",
		issueApproleTask,
		revokeSelfTask,
		unsealTask,
		"verify",
	}
}

func publicTasks() []string {
	return append(publicTasksA(), publicTasksB()...)
}

func publicVars() []string {
	return []string{
		"VAULT_APPROLE_MOUNT",
		"VAULT_EXTRA_ARGS",
		"VAULT_FILE",
		"VAULT_KEYS_FILE",
		roleEnvVarName,
		"VAULT_ROOT_TOKEN",
		approleKeyEnvVarName,
		"VAULT_SHARES",
		"VAULT_SNAPSHOT_FILE",
		"VAULT_THRESHOLD",
		vaultAddrEnvVarName,
		"VAULT_NIX_INSTALLABLE",
	}
}

func inputValidatedTaskNames() []string {
	return []string{
		initTask,
		loginTask,
		loginTaskApprole,
		loginTaskRootAuth,
		restoreTask,
		issueApproleTask,
		revokeSelfTask,
		unsealTask,
	}
}

func assertInitPreconditionsRefuseExistingKeysFile(t *testing.T, preconditions string) {
	t.Helper()

	assertCmdsContainAll(t, &cmdsAssertion{
		content: preconditions,
		tokens:  []string{"test ! -e", "KEYS_FILE already exists"},
		msgFmt:  "init should refuse an existing KEYS_FILE with %q\npreconditions:\n%s",
	})
}

func assertInitStagesOutputSafely(t *testing.T, cmds string) {
	t.Helper()

	assertCmdsContainAll(t, &cmdsAssertion{
		content: cmds,
		tokens:  []string{`TMP="${KF}.tmp.$$"`, `-format=json > "$TMP"`, `mv "$TMP" "$KF"`},
		msgFmt:  "init should stage init output safely with %q\ncmds:\n%s",
	})

	if strings.Contains(cmds, `-format=json > "$KF"`) {
		t.Fatalf("init should not redirect operator init directly to KEYS_FILE\ncmds:\n%s", cmds)
	}
}

func loginTaskCmds(t *testing.T, taskName string) string {
	t.Helper()

	taskfile := tasktest.LoadTaskfile(t, vaultModuleName)

	return taskFieldYAML(t, taskfile.Tasks[taskName].Cmds)
}

func assertApproleCredentialsExchanged(t *testing.T, check *approleCredentialsCheck) {
	t.Helper()

	tokens := append([]string{
		printApproleKeyValueStdin,
		vaultWritePipe,
		fieldAuthFlag,
		approleLoginPath,
		roleReferenceArg,
		approleKeyStdinArg,
	}, check.extraTokens...)

	assertCmdsContainAll(t, &cmdsAssertion{
		content: check.cmds,
		tokens:  tokens,
		msgFmt:  check.msgPrefix + " should exchange credentials without exposing secret_id (missing %q)\ncmds:\n%s",
	})
}

func assertLoginApprolePreconditions(t *testing.T, preconditions string) {
	t.Helper()

	assertCmdsContainAll(t, &cmdsAssertion{
		content: preconditions,
		tokens:  []string{roleEnvVarName, approleKeyEnvVarName},
		msgFmt:  "login:approle should require %s in preconditions\npreconditions:\n%s",
	})
}

func assertLoginApprolePreconditionsAndCmds(t *testing.T, preconditions, cmds string) {
	t.Helper()

	assertLoginApprolePreconditions(t, preconditions)
	assertApproleCredentialsExchanged(t, &approleCredentialsCheck{
		msgPrefix:   "login:approle",
		cmds:        cmds,
		extraTokens: []string{vaultLoginPipe, authMethodFlag, noPrintFlag},
	})

	assertCmdsContainNone(t, &cmdsAssertion{
		content: cmds,
		tokens:  []string{`secret_id="{{`, `secret_id="$VAULT_LOGIN_SECRET_ID"`},
		msgFmt:  "login:approle should not expose secret_id as a command argument (found %q)\ncmds:\n%s",
	})
}

func strictShellTaskNames() []string {
	return []string{
		healthTask,
		initTask,
		kvGetTask,
		loginTask,
		loginTaskApprole,
		loginTaskRootAuth,
		restoreTask,
		issueApproleTask,
		revokeSelfTask,
		unsealTask,
	}
}

func assertTaskSetsStrictShellOptions(t *testing.T, taskfile *tasktest.Taskfile, name string) {
	t.Helper()

	task := taskfile.Tasks[name]

	options := []string{"errexit", "nounset", "pipefail"}

	for i := range options {
		option := options[i]

		if !slices.Contains(task.Set, option) {
			t.Fatalf("%s should set %s, got %#v", name, option, task.Set)
		}
	}
}

func assertTokenIssueApRoleSkipsVaultLogin(t *testing.T, cmds string) {
	t.Helper()

	assertCmdsContainNone(t, &cmdsAssertion{
		content: cmds,
		tokens:  []string{vaultLoginPipe, noPrintFlag},
		msgFmt:  "token:issue:approle should not call vault login (found %q, token must go to stdout)\ncmds:\n%s",
	})
}

func assertKvGetCallsVaultWithJSON(t *testing.T, cmds string) {
	t.Helper()

	assertCmdsContainAll(t, &cmdsAssertion{
		content: cmds,
		tokens:  []string{`vault kv get`, `-format=json`, `-mount=`},
		msgFmt:  "kv:get should call vault kv get with json format (missing %q)\ncmds:\n%s",
	})

	if !strings.Contains(cmds, "KV_GET_VERSION") {
		t.Fatalf("kv:get should handle optional SECRET_VERSION\ncmds:\n%s", cmds)
	}
}

func assertCmdsContainAll(t *testing.T, check *cmdsAssertion) {
	t.Helper()

	for i := range check.tokens {
		token := check.tokens[i]

		if !strings.Contains(check.content, token) {
			t.Fatalf(check.msgFmt, token, check.content)
		}
	}
}

func assertCmdsContainNone(t *testing.T, check *cmdsAssertion) {
	t.Helper()

	for i := range check.tokens {
		token := check.tokens[i]

		if strings.Contains(check.content, token) {
			t.Fatalf(check.msgFmt, token, check.content)
		}
	}
}

func taskFieldYAML(t *testing.T, value any) string {
	t.Helper()

	content, err := yaml.Marshal(value)
	if err != nil {
		t.Fatalf("marshal task field: %v", err)
	}

	return string(content)
}
