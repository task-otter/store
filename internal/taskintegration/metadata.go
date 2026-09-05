// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

package taskintegration

import (
	"path/filepath"
	"testing"

	"github.com/task-otter/store/internal/tasktestutil"
	yaml "go.yaml.in/yaml/v3"
)

type (
	// moduleMetadata is the subset of metadata.yml the integration suite reads.
	moduleMetadata = struct {
		ExportedTasks []string `yaml:"exported_tasks"`
		Variants      []string `yaml:"variants"`
	}
)

const (
	metadataName = "metadata.yml"
)

// loadMetadata reads the folder's metadata.yml. Nested family variants share the
// family root's metadata and carry none of their own, so a missing file yields
// empty expectations rather than a failure.
func loadMetadata(t *testing.T, dir string) *moduleMetadata {
	t.Helper()

	path := filepath.Join(dir, metadataName)

	if !tasktestutil.FileExists(path) {
		return &moduleMetadata{ExportedTasks: nil, Variants: nil}
	}

	return parseMetadata(t, path, tasktestutil.ReadFile(t, path))
}

func parseMetadata(t suiteT, path, content string) *moduleMetadata {
	t.Helper()

	metadata := new(moduleMetadata)

	err := yaml.Unmarshal([]byte(content), metadata)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	return metadata
}
