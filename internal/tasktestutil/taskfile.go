// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package tasktestutil

import (
	yaml "go.yaml.in/yaml/v3"
)

type (
	// TaskNode wraps a YAML node with its task name for error messages.
	TaskNode = struct {
		Node *yaml.Node
		Name string
	}

	// LoadedTaskfile holds the parsed content of a Taskfile.
	LoadedTaskfile = struct {
		Root  TaskNode
		Tasks map[string]TaskNode
		Path  string
	}
)
