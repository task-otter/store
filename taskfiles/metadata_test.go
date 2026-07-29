package taskfiles_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
	"gopkg.in/yaml.v3"
)

const (
	constMetadataTestTaskfileYml = "Taskfile.yml"
)

const metadataSchema = "taskotter.dev/taskfile-metadata/v1"

func toolFamilies() map[string]bool {
	return map[string]bool{
		"eslint": true, "prettier": true, "biome": true, "bruno": true,
		"depcheck": true, "knip": true, "stylelint": true, "typescript": true,
		"htmlhint": true, "spectral": true,
		// Package managers are families too: taskfiles/<pm>/{fnm,nvm}.
		"npm": true, "pnpm": true, "yarn": true, "corepack": true,
	}
}

type moduleMetadata struct {
	Schema        string   `yaml:"schema"`
	Module        string   `yaml:"module"`
	Taskfile      string   `yaml:"taskfile"`
	ExportedTasks []string `yaml:"-"`
	Variants      []string `yaml:"variants"`
}

type discoveredMetadataModules struct {
	flatModules []string
	toolLeaves  map[string][]string
}

// exportedTasks returns the sorted public task names for a module (its path
// relative to taskfiles/, e.g. "jq" or "prettier/node/fnm/npm").
func exportedTasks(t *testing.T, module string) []string {
	t.Helper()
	taskfile := tasktest.LoadTaskfile(t, module)

	tasks := make([]string, 0, len(taskfile.Tasks))
	for name, task := range taskfile.Tasks {
		if name == "default" || strings.HasPrefix(name, "_") || task.Internal {
			continue
		}

		tasks = append(tasks, name)
	}

	slices.Sort(tasks)

	return tasks
}

// TestModuleMetadataListsEveryExportedTask verifies each module's metadata.yml
// is in sync with its Taskfile. Flat modules are checked directly; each JS-tool
// family carries one metadata.yml at its root whose exported_tasks must match
// every one of its variant leaves (the shared-interface guarantee) and whose
// variants list must match the leaves actually present on disk.
func TestModuleMetadataListsEveryExportedTask(t *testing.T) {
	t.Parallel()

	root := tasktest.RepoRoot(t)
	taskfilesDir := filepath.Join(root, "taskfiles")
	discovered := discoverMetadataModules(t, taskfilesDir)

	for _, module := range discovered.flatModules {
		t.Run(module, func(t *testing.T) {
			t.Parallel()

			assertMetadata(t, taskfilesDir, module, module, exportedTasks(t, module), nil)
		})
	}

	for _, tool := range sortedKeys(discovered.toolLeaves) {
		leaves := discovered.toolLeaves[tool]
		slices.Sort(leaves)
		t.Run(tool, func(t *testing.T) {
			t.Parallel()

			assertToolMetadata(t, taskfilesDir, tool, leaves)
		})
	}
}

func discoverMetadataModules(t *testing.T, taskfilesDir string) discoveredMetadataModules {
	t.Helper()

	discovered := discoveredMetadataModules{flatModules: nil, toolLeaves: map[string][]string{}}

	err := filepath.WalkDir(taskfilesDir, func(path string, dirEntry fs.DirEntry, err error) error {
		return discoverMetadataModule(t, taskfilesDir, path, dirEntry, err, &discovered)
	})
	if err != nil {
		t.Fatalf("walk taskfiles: %v", err)
	}

	if len(discovered.flatModules) == 0 || len(discovered.toolLeaves) == 0 {
		t.Fatal("no modules discovered")
	}

	return discovered
}

func discoverMetadataModule(
	t *testing.T,
	taskfilesDir string,
	path string,
	dirEntry fs.DirEntry,
	walkErr error,
	discovered *discoveredMetadataModules,
) error {
	t.Helper()

	if walkErr != nil {
		return walkErr
	}

	if dirEntry.IsDir() || dirEntry.Name() != constMetadataTestTaskfileYml {
		return nil
	}

	rel, err := filepath.Rel(taskfilesDir, filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("metadata module path relative to taskfiles: %w", err)
	}

	module := filepath.ToSlash(rel)
	segments := strings.Split(module, "/")

	if toolFamilies()[segments[0]] {
		addToolLeaf(t, discovered.toolLeaves, segments[0], module)

		return nil
	}

	if len(segments) == 1 {
		discovered.flatModules = append(discovered.flatModules, module)
	}

	return nil
}

func addToolLeaf(t *testing.T, toolLeaves map[string][]string, tool, module string) {
	t.Helper()

	if tool == module || len(exportedTasks(t, module)) == 0 {
		return
	}

	toolLeaves[tool] = append(toolLeaves[tool], module)
}

func assertToolMetadata(t *testing.T, taskfilesDir, tool string, leaves []string) {
	t.Helper()

	base := exportedTasks(t, leaves[0])
	variants := make([]string, 0, len(leaves))

	for _, leaf := range leaves {
		got := exportedTasks(t, leaf)
		if !slices.Equal(got, base) {
			t.Fatalf(
				"exported-task interface drift within %q:\n  %s: %v\n  %s: %v",
				tool, leaves[0], base, leaf, got,
			)
		}

		variants = append(variants, strings.TrimPrefix(leaf, tool+"/"))
	}

	assertMetadata(t, taskfilesDir, tool, tool, base, variants)
}

// TestTaskCliLoadsEveryFamily forks the task CLI once per family to prove every
// shipped Taskfile still parses. Per-module contract tests only parse the YAML
// themselves; this is the single place that pays for the CLI round trip.
func TestTaskCliLoadsEveryFamily(t *testing.T) {
	t.Parallel()

	for _, module := range familyRepresentatives(t) {
		t.Run(module, func(t *testing.T) {
			t.Parallel()
			tasktest.AssertTaskCliCanLoad(t, module)
		})
	}
}

// familyRepresentatives returns one module per family: flat modules stand for
// themselves, and each tool family is represented by its first leaf, since all
// leaves are generated from the same template.
func familyRepresentatives(t *testing.T) []string {
	t.Helper()

	taskfilesDir := filepath.Join(tasktest.RepoRoot(t), "taskfiles")
	seen := map[string]string{}

	err := filepath.WalkDir(taskfilesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || d.Name() != constMetadataTestTaskfileYml {
			return nil
		}

		rel, err := filepath.Rel(taskfilesDir, filepath.Dir(path))
		if err != nil {
			return fmt.Errorf("family representative path relative to taskfiles: %w", err)
		}

		module := filepath.ToSlash(rel)

		family, _, _ := strings.Cut(module, "/")
		if existing, ok := seen[family]; !ok || module < existing {
			seen[family] = module
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk taskfiles: %v", err)
	}

	if len(seen) == 0 {
		t.Fatal("no modules discovered")
	}

	modules := make([]string, 0, len(seen))
	for _, module := range seen {
		modules = append(modules, module)
	}

	slices.Sort(modules)

	return modules
}

func assertMetadata(
	t *testing.T,
	taskfilesDir, dir, module string,
	expectedTasks, expectedVariants []string,
) {
	t.Helper()

	metadataPath := filepath.ToSlash(filepath.Join(dir, "metadata.yml"))

	content, err := fs.ReadFile(os.DirFS(taskfilesDir), metadataPath)
	if err != nil {
		t.Fatalf("read metadata.yml: %v; add or update it to match the module Taskfile", err)
	}

	if strings.Contains(string(content), "\r\n") {
		t.Fatal("metadata.yml must use LF line endings")
	}

	if strings.TrimRight(string(content), " \t\r\n") != strings.TrimRight(string(content), "\r\n") {
		t.Fatal("metadata.yml has trailing whitespace")
	}

	var metadata moduleMetadata

	err = yaml.Unmarshal(content, &metadata)
	if err != nil {
		t.Fatalf("parse metadata.yml: %v", err)
	}

	metadata.ExportedTasks = metadataStringSequence(t, content, "exported_tasks")

	assertMetadataHeader(t, metadata, module)
	assertMetadataTasks(t, metadata, expectedTasks)
	assertMetadataVariants(t, metadata, expectedVariants)
}

func assertMetadataHeader(t *testing.T, metadata moduleMetadata, module string) {
	t.Helper()

	if metadata.Schema != metadataSchema {
		t.Errorf("schema = %q, want %q", metadata.Schema, metadataSchema)
	}

	if metadata.Module != module {
		t.Errorf("module = %q, want %q", metadata.Module, module)
	}

	if metadata.Taskfile != constMetadataTestTaskfileYml {
		t.Errorf("taskfile = %q, want %q", metadata.Taskfile, constMetadataTestTaskfileYml)
	}
}

func assertMetadataTasks(t *testing.T, metadata moduleMetadata, expectedTasks []string) {
	t.Helper()

	if !slices.IsSorted(metadata.ExportedTasks) {
		t.Errorf("exported_tasks must be sorted: %v", metadata.ExportedTasks)
	}

	for i := 1; i < len(metadata.ExportedTasks); i++ {
		if metadata.ExportedTasks[i] == metadata.ExportedTasks[i-1] {
			t.Errorf("exported_tasks contains duplicate %q", metadata.ExportedTasks[i])
		}
	}

	if !slices.Equal(metadata.ExportedTasks, expectedTasks) {
		t.Fatalf(
			"exported task drift\nmetadata: %v\ntaskfile: %v\nupdate metadata.yml to match the Taskfile",
			metadata.ExportedTasks,
			expectedTasks,
		)
	}
}

func assertMetadataVariants(t *testing.T, metadata moduleMetadata, expectedVariants []string) {
	t.Helper()

	if !slices.Equal(metadata.Variants, expectedVariants) {
		t.Fatalf(
			"variant drift\nmetadata: %v\non disk:  %v\nupdate metadata.yml to match the variants on disk",
			metadata.Variants,
			expectedVariants,
		)
	}
}

func metadataStringSequence(t *testing.T, content []byte, key string) []string {
	t.Helper()

	var doc yaml.Node

	err := yaml.Unmarshal(content, &doc)
	if err != nil {
		t.Fatalf("parse metadata.yml: %v", err)
	}

	root := doc.Content[0]
	for i := 0; i < len(root.Content); i += 2 {
		if root.Content[i].Value == key {
			return scalarSequence(root.Content[i+1])
		}
	}

	return nil
}

func scalarSequence(node *yaml.Node) []string {
	if node == nil || node.Kind != yaml.SequenceNode {
		return nil
	}

	values := make([]string, 0, len(node.Content))
	for _, child := range node.Content {
		values = append(values, child.Value)
	}

	return values
}

func sortedKeys[V any](items map[string]V) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	return keys
}
