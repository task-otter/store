package vault_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
	"gopkg.in/yaml.v3"
)

const (
	initTask              = "init"
	loginTask             = "login"
	loginTaskApprole      = "login:approle"
	loginTaskRootToken    = "login:root-token"
	restoreTask           = "restore"
	tokenIssueApproleTask = "token:issue:approle"
	tokenRevokeSelfTask   = "token:revoke-self"
	unsealTask            = "unseal"
	roleIDVar             = "ROLE_ID"
	secretIDVar           = "SECRET_ID"
	methodTokenFlag       = "-method=token"
	noPrintFlag           = "-no-print"
	vaultLoginPipe        = `| vault login`
)

func publicTasks() []string {
	return []string{
		"health",
		initTask,
		"install",
		"install:undo",
		"kv:get",
		loginTask,
		loginTaskApprole,
		loginTaskRootToken,
		"peers",
		restoreTask,
		"root-token",
		"seal",
		"snapshot",
		"status",
		tokenIssueApproleTask,
		tokenRevokeSelfTask,
		unsealTask,
		"upgrade",
		"verify",
		"version",
	}
}

func publicVars() []string {
	return []string{
		"APPROLE_MOUNT",
		"EXTRA_ARGS",
		"FILE",
		"KEYS_FILE",
		roleIDVar,
		"ROOT_TOKEN",
		secretIDVar,
		"SHARES",
		"SNAPSHOT_FILE",
		"THRESHOLD",
		"VAULT_ADDR",
		"VERSION",
	}
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "vault", publicTasks(), publicVars())
}

func TestInputValidatedTasksDoNotInstallBeforePreconditions(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "vault")

	for _, name := range []string{
		initTask,
		loginTask,
		loginTaskApprole,
		loginTaskRootToken,
		restoreTask,
		tokenIssueApproleTask,
		tokenRevokeSelfTask,
		unsealTask,
	} {
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

func TestVerifyDoesNotMaskStatusFailures(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "vault")
	cmds := taskFieldYAML(t, taskfile.Tasks["verify"].Cmds)

	if !strings.Contains(cmds, "vault status") {
		t.Fatalf("verify should run vault status\ncmds:\n%s", cmds)
	}

	if strings.Contains(cmds, "vault status || true") {
		t.Fatalf("verify should fail when vault status fails\ncmds:\n%s", cmds)
	}
}

func TestInitDoesNotOverwriteExistingKeysFile(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "vault")
	task := taskfile.Tasks[initTask]
	preconditions := taskFieldYAML(t, task.Preconditions)
	cmds := taskFieldYAML(t, task.Cmds)

	for _, token := range []string{"test ! -e", "KEYS_FILE already exists"} {
		if !strings.Contains(preconditions, token) {
			t.Fatalf(
				"init should refuse an existing KEYS_FILE with %q\npreconditions:\n%s",
				token,
				preconditions,
			)
		}
	}

	for _, token := range []string{`TMP="${KF}.tmp.$$"`, `-format=json > "$TMP"`, `mv "$TMP" "$KF"`} {
		if !strings.Contains(cmds, token) {
			t.Fatalf("init should stage init output safely with %q\ncmds:\n%s", token, cmds)
		}
	}

	if strings.Contains(cmds, `-format=json > "$KF"`) {
		t.Fatalf("init should not redirect operator init directly to KEYS_FILE\ncmds:\n%s", cmds)
	}
}

func TestLoginDoesNotPassRootTokenAsCommandArgument(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "vault")
	cmds := taskFieldYAML(t, taskfile.Tasks[loginTask].Cmds)

	for _, token := range []string{`jq -r '.root_token' "$KF"`, vaultLoginPipe, methodTokenFlag, noPrintFlag} {
		if !strings.Contains(cmds, token) {
			t.Fatalf(
				"login should pipe the root token to vault login stdin (missing %q)\ncmds:\n%s",
				token,
				cmds,
			)
		}
	}

	for _, token := range []string{`vault login "$(jq`, `vault login token=`, `vault login "$TOKEN"`} {
		if strings.Contains(cmds, token) {
			t.Fatalf(
				"login should not expose the root token as a command argument\ncmds:\n%s",
				cmds,
			)
		}
	}
}

func TestLoginRootTokenPipesTokenViaStdin(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "vault")
	cmds := taskFieldYAML(t, taskfile.Tasks[loginTaskRootToken].Cmds)

	for _, token := range []string{
		`printf '%s' "$VAULT_LOGIN_ROOT_TOKEN"`,
		vaultLoginPipe,
		methodTokenFlag,
		noPrintFlag,
	} {
		if !strings.Contains(cmds, token) {
			t.Fatalf(
				"login:root-token should pipe token to vault login stdin (missing %q)\ncmds:\n%s",
				token,
				cmds,
			)
		}
	}

	for _, token := range []string{`vault login "$TOKEN"`, `vault login token=`} {
		if strings.Contains(cmds, token) {
			t.Fatalf(
				"login:root-token should not expose token as command argument\ncmds:\n%s",
				cmds,
			)
		}
	}
}

func TestLoginApproleRequiresBothCredentials(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "vault")
	task := taskfile.Tasks[loginTaskApprole]
	preconditions := taskFieldYAML(t, task.Preconditions)
	cmds := taskFieldYAML(t, task.Cmds)

	for _, token := range []string{roleIDVar, secretIDVar} {
		if !strings.Contains(preconditions, token) {
			t.Fatalf(
				"login:approle should require %s in preconditions\npreconditions:\n%s",
				token,
				preconditions,
			)
		}
	}

	for _, token := range []string{
		`printf '%s' "$VAULT_LOGIN_SECRET_ID"`,
		`| vault write`,
		`-field=token`,
		`"auth/${VAULT_LOGIN_APPROLE_MOUNT}/login"`,
		`"$VAULT_LOGIN_ROLE_ID"`,
		`secret_id=-`,
		`| vault login`,
		methodTokenFlag,
		noPrintFlag,
	} {
		if !strings.Contains(cmds, token) {
			t.Fatalf(
				"login:approle should exchange AppRole credentials without exposing the secret_id (missing %q)\ncmds:\n%s",
				token,
				cmds,
			)
		}
	}

	for _, token := range []string{`secret_id="{{`, `secret_id="$VAULT_LOGIN_SECRET_ID"`} {
		if strings.Contains(cmds, token) {
			t.Fatalf(
				"login:approle should not expose secret_id as a command argument\ncmds:\n%s",
				cmds,
			)
		}
	}
}

func TestLinuxParentTasksGuardUnsupportedPackageManagers(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "vault")

	for _, name := range []string{"_install:linux", "_install:undo:linux", "_upgrade:linux"} {
		preconditions := taskFieldYAML(t, taskfile.Tasks[name].Preconditions)
		for _, token := range []string{"apt-get", "dnf"} {
			if !strings.Contains(preconditions, token) {
				t.Fatalf(
					"%s should guard unsupported Linux package managers with %s\npreconditions:\n%s",
					name,
					token,
					preconditions,
				)
			}
		}
	}
}

func TestStrictShellSetOnSensitiveTasks(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "vault")

	for _, name := range []string{
		"health",
		initTask,
		"kv:get",
		loginTask,
		loginTaskApprole,
		loginTaskRootToken,
		restoreTask,
		tokenIssueApproleTask,
		tokenRevokeSelfTask,
		unsealTask,
	} {
		task := taskfile.Tasks[name]
		for _, option := range []string{"errexit", "nounset", "pipefail"} {
			if !slices.Contains(task.Set, option) {
				t.Fatalf("%s should set %s, got %#v", name, option, task.Set)
			}
		}
	}
}

func TestTokenIssueApRolePipesSecretViaStdinWithoutLogin(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "vault")
	task := taskfile.Tasks[tokenIssueApproleTask]
	preconditions := taskFieldYAML(t, task.Preconditions)
	cmds := taskFieldYAML(t, task.Cmds)

	for _, token := range []string{roleIDVar, secretIDVar} {
		if !strings.Contains(preconditions, token) {
			t.Fatalf(
				"token:issue:approle should require %s in preconditions\npreconditions:\n%s",
				token,
				preconditions,
			)
		}
	}

	for _, token := range []string{
		`printf '%s' "$VAULT_LOGIN_SECRET_ID"`,
		`| vault write`,
		`-field=token`,
		`"auth/${VAULT_LOGIN_APPROLE_MOUNT}/login"`,
		`"$VAULT_LOGIN_ROLE_ID"`,
		`secret_id=-`,
	} {
		if !strings.Contains(cmds, token) {
			t.Fatalf(
				"token:issue:approle should exchange AppRole credentials without exposing secret_id (missing %q)\ncmds:\n%s",
				token,
				cmds,
			)
		}
	}

	for _, token := range []string{`| vault login`, noPrintFlag} {
		if strings.Contains(cmds, token) {
			t.Fatalf(
				"token:issue:approle should not call vault login (token must go to stdout)\ncmds:\n%s",
				cmds,
			)
		}
	}
}

func TestTokenRevokeSelfRequiresVaultToken(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "vault")
	task := taskfile.Tasks[tokenRevokeSelfTask]
	preconditions := taskFieldYAML(t, task.Preconditions)
	cmds := taskFieldYAML(t, task.Cmds)

	if !strings.Contains(preconditions, "VAULT_TOKEN") {
		t.Fatalf(
			"token:revoke-self should require VAULT_TOKEN in preconditions\npreconditions:\n%s",
			preconditions,
		)
	}

	if !strings.Contains(cmds, "-self") {
		t.Fatalf("token:revoke-self should pass -self to vault token revoke\ncmds:\n%s", cmds)
	}
}

func TestKvGetRequiresMountPathAndToken(t *testing.T) {
	t.Parallel()

	taskfile := tasktest.LoadTaskfile(t, "vault")
	task := taskfile.Tasks["kv:get"]
	preconditions := taskFieldYAML(t, task.Preconditions)
	cmds := taskFieldYAML(t, task.Cmds)

	for _, token := range []string{"KV_GET_MOUNT", "KV_GET_PATH", "VAULT_TOKEN", "VAULT_ADDR"} {
		if !strings.Contains(preconditions, token) {
			t.Fatalf(
				"kv:get should require %s in preconditions\npreconditions:\n%s",
				token,
				preconditions,
			)
		}
	}

	for _, token := range []string{`vault kv get`, `-format=json`, `-mount=`} {
		if !strings.Contains(cmds, token) {
			t.Fatalf(
				"kv:get should call vault kv get with json format (missing %q)\ncmds:\n%s",
				token,
				cmds,
			)
		}
	}

	if !strings.Contains(cmds, "KV_GET_VERSION") {
		t.Fatalf("kv:get should handle optional SECRET_VERSION\ncmds:\n%s", cmds)
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
