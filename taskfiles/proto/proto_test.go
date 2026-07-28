package proto_test

import (
	"slices"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

var publicTasks = []string{
	"gen",
	"install",
	"install:undo",
	"upgrade",
	"ungen",
	"version",
}

var publicVars = []string{
	"GO_CMD",
	"GLOBAL_GO_BIN",
	"PROTO_PATH",
	"PROTO_PATTERN",
	"PROTOC_GEN_GO_GRPC_VERSION",
	"PROTOC_GEN_GO_VERSION",
	"PROTOC_VERSION",
}

func TestTaskfileModuleContract(t *testing.T) {
	t.Parallel()

	tasktest.AssertModule(t, "proto", publicTasks, publicVars)
}

func TestPluginWorkflowsInstallGoFirst(t *testing.T) {
	t.Parallel()

	tf := tasktest.LoadTaskfile(t, "proto")

	for _, taskName := range []string{"install", "upgrade"} {
		deps, ok := tf.Tasks[taskName].Deps.([]any)
		if !ok {
			t.Fatalf("%s deps have type %T, want []any", taskName, tf.Tasks[taskName].Deps)
		}

		if !slices.Contains(deps, any("go:install")) {
			t.Errorf("%s must depend on go:install; deps: %v", taskName, deps)
		}
	}
}
