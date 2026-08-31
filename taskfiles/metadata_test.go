// Taskotter 2026.
// SPDX-License-Identifier: Apache-2.0.

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
	yaml "go.yaml.in/yaml/v3"
)

type (
	moduleMetadata struct {
		Schema        string   `yaml:"schema"`
		Module        string   `yaml:"module"`
		Taskfile      string   `yaml:"taskfile"`
		ExportedTasks []string `yaml:"-"`
		Variants      []string `yaml:"variants"`
	}

	discoveredMetadataModules struct {
		toolLeaves  map[string][]string
		flatModules []string
	}

	metadataDiscoverer struct {
		t            *testing.T
		discovered   *discoveredMetadataModules
		taskfilesDir string
	}

	toolLeaf struct {
		leaves map[string][]string
		tool   string
		module string
	}

	toolMetadataCheck struct {
		taskfilesDir string
		tool         string
		leaves       []string
	}

	familyRepScanner struct {
		seen         map[string]string
		taskfilesDir string
	}

	toolMetadataVariantCheck struct {
		check *toolMetadataCheck
		leaf  string
		base  []string
	}

	metadataCheck struct {
		taskfilesDir     string
		dir              string
		module           string
		expectedTasks    []string
		expectedVariants []string
	}
)

const (
	metadataSchema = "taskotter.dev/taskfile-metadata/v1"

	typescriptFamily = "typescript"
	npmFamily        = "npm"
	pnpmFamily       = "pnpm"
	yarnFamily       = "yarn"

	walkTaskfilesErrFormat = "walk taskfiles: %v"
	noModulesDiscoveredMsg = "no modules discovered"
	pathSeparator          = "/"
	parseMetadataErrFormat = "parse metadata.yml: %v"
)

// TestModuleMetadataListsEveryExportedTask verifies each module's metadata.yml
// is in sync with its Taskfile. Flat modules are checked directly; each JS-tool
// family carries one metadata.yml at its root whose exported_tasks must match
// every one of its variant leaves (the shared-interface guarantee) and whose
// variants list must match the leaves actually present on disk.
// TestModuleMetadataListsEveryExportedTask
func TestModuleMetadataListsEveryExportedTask(t *testing.T) {
	t.Parallel()

	root := tasktest.RepoRoot(t)
	taskfilesDir := filepath.Join(root, taskfilesDirName)
	discovered := discoverMetadataModules(t, taskfilesDir)

	assertFlatMetadataModules(t, taskfilesDir, discovered.flatModules)
	assertToolMetadataModules(t, taskfilesDir, discovered.toolLeaves)
}

// TestTaskCliLoadsEveryFamily forks the task CLI once per family to prove every
// shipped Taskfile still parses. Per-module contract tests only parse the YAML
// themselves; this is the single place that pays for the CLI round trip.
// TestTaskCliLoadsEveryFamily
func TestTaskCliLoadsEveryFamily(t *testing.T) {
	t.Parallel()

	for i := range familyRepresentatives(t) {
		module := familyRepresentatives(t)[i]
		t.Run(module, func(t *testing.T) {
			t.Parallel()
			tasktest.AssertTaskCliCanLoad(t, module)
		})
	}
}

func toolFamilies() map[string]bool {
	return map[string]bool{
		eslintFamily: true, prettierFamily: true, biomeFamily: true,
		depcheckFamily: true, knipFamily: true, stylelintFamily: true, typescriptFamily: true,
		htmlhintFamily: true, spectralFamily: true,
		// Package managers are flat nix-backed modules.
		npmFamily: true, pnpmFamily: true, yarnFamily: true,
	}
}

// exportedTasks returns the sorted public task names for a module (its path
// relative to taskfiles/, e.g. "jq" or "prettier/node/npm").
func exportedTasks(t *testing.T, module string) []string {
	t.Helper()

	taskfile := tasktest.LoadTaskfile(t, module)

	tasks := make([]string, constZero, len(taskfile.Tasks))

	for name := range taskfile.Tasks {
		task := taskfile.Tasks[name]

		if name == "default" || strings.HasPrefix(name, underscoreChar) || task.Internal {
			continue
		}

		tasks = append(tasks, name)
	}

	slices.Sort(tasks)

	return tasks
}

func assertFlatMetadataModules(t *testing.T, taskfilesDir string, modules []string) {
	t.Helper()

	for i := range modules {
		module := modules[i]
		t.Run(module, func(t *testing.T) {
			t.Parallel()

			assertMetadata(t, &metadataCheck{
				taskfilesDir:     taskfilesDir,
				dir:              module,
				module:           module,
				expectedTasks:    exportedTasks(t, module),
				expectedVariants: nil,
			})
		})
	}
}

func assertToolMetadataModules(t *testing.T, taskfilesDir string, toolLeaves map[string][]string) {
	t.Helper()

	for i := range sortedKeys(toolLeaves) {
		tool := sortedKeys(toolLeaves)[i]
		leaves := toolLeaves[tool]
		slices.Sort(leaves)
		t.Run(tool, func(t *testing.T) {
			t.Parallel()

			assertToolMetadata(t, &toolMetadataCheck{
				taskfilesDir: taskfilesDir,
				tool:         tool,
				leaves:       leaves,
			})
		})
	}
}

func (discoverer *metadataDiscoverer) discover(path string, entry fs.DirEntry, err error) error {
	discoverer.t.Helper()

	if err != nil {
		return err
	}

	if entry.IsDir() || entry.Name() != skipTaskfileYML {
		return nil
	}

	moduleErr := discoverer.recordModuleFor(path)
	if moduleErr != nil {
		return fmt.Errorf("record metadata module: %w", moduleErr)
	}

	return nil
}

func (discoverer *metadataDiscoverer) moduleFor(path string) (string, error) {
	rel, err := filepath.Rel(discoverer.taskfilesDir, filepath.Dir(path))
	if err != nil {
		return "", fmt.Errorf("metadata module path relative to taskfiles: %w", err)
	}

	return filepath.ToSlash(rel), nil
}

func (discoverer *metadataDiscoverer) record(module string) {
	segments := strings.Split(module, pathSeparator)

	if toolFamilies()[segments[constZero]] {
		addToolLeaf(discoverer.t, &toolLeaf{
			leaves: discoverer.discovered.toolLeaves,
			tool:   segments[constZero],
			module: module,
		})

		return
	}

	if len(segments) == constOne {
		discoverer.discovered.flatModules = append(discoverer.discovered.flatModules, module)
	}
}

func (discoverer *metadataDiscoverer) recordModuleFor(path string) error {
	module, err := discoverer.moduleFor(path)
	if err != nil {
		return fmt.Errorf("discover metadata module: %w", err)
	}

	discoverer.record(module)

	return nil
}

func discoverMetadataModules(t *testing.T, taskfilesDir string) discoveredMetadataModules {
	t.Helper()

	discovered := discoveredMetadataModules{flatModules: nil, toolLeaves: map[string][]string{}}
	discoverer := metadataDiscoverer{t: t, taskfilesDir: taskfilesDir, discovered: &discovered}

	err := filepath.WalkDir(taskfilesDir, discoverer.discover)
	if err != nil {
		t.Fatalf(walkTaskfilesErrFormat, err)
	}

	if len(discovered.flatModules) == constZero || len(discovered.toolLeaves) == constZero {
		t.Fatal(noModulesDiscoveredMsg)
	}

	return discovered
}

func addToolLeaf(t *testing.T, leaf *toolLeaf) {
	t.Helper()

	if leaf.tool == leaf.module || len(exportedTasks(t, leaf.module)) == constZero {
		return
	}

	leaf.leaves[leaf.tool] = append(leaf.leaves[leaf.tool], leaf.module)
}

func assertToolMetadataVariantMatches(t *testing.T, variant *toolMetadataVariantCheck) {
	t.Helper()

	got := exportedTasks(t, variant.leaf)

	if !slices.Equal(got, variant.base) {
		t.Fatalf(
			"exported-task interface drift within %q:\n  %s: %v\n  %s: %v",
			variant.check.tool, variant.check.leaves[0], variant.base, variant.leaf, got,
		)
	}
}

func assertToolVariantsMatch(t *testing.T, check *toolMetadataCheck, base []string) []string {
	t.Helper()

	variants := make([]string, constZero, len(check.leaves))

	for i := range check.leaves {
		leaf := check.leaves[i]
		assertToolMetadataVariantMatches(
			t,
			&toolMetadataVariantCheck{check: check, base: base, leaf: leaf},
		)

		variants = append(variants, strings.TrimPrefix(leaf, check.tool+"/"))
	}

	return variants
}

func assertToolMetadata(t *testing.T, check *toolMetadataCheck) {
	t.Helper()

	base := exportedTasks(t, check.leaves[0])
	variants := assertToolVariantsMatch(t, check, base)

	assertMetadata(t, &metadataCheck{
		taskfilesDir:     check.taskfilesDir,
		dir:              check.tool,
		module:           check.tool,
		expectedTasks:    base,
		expectedVariants: variants,
	})
}

// familyRepresentatives returns one module per family: flat modules stand for
// themselves, and each tool family is represented by its first leaf, since all
// leaves are generated from the same template.
func familyRepresentatives(t *testing.T) []string {
	t.Helper()

	scanner := familyRepScanner{
		taskfilesDir: filepath.Join(tasktest.RepoRoot(t), taskfilesDirName),
		seen:         map[string]string{},
	}

	err := filepath.WalkDir(scanner.taskfilesDir, scanner.scan)
	if err != nil {
		t.Fatalf(walkTaskfilesErrFormat, err)
	}

	if len(scanner.seen) == constZero {
		t.Fatal(noModulesDiscoveredMsg)
	}

	return sortedFamilyModules(scanner.seen)
}

func (scanner familyRepScanner) moduleFor(path string) (string, error) {
	rel, err := filepath.Rel(scanner.taskfilesDir, filepath.Dir(path))
	if err != nil {
		return "", fmt.Errorf("family representative path relative to taskfiles: %w", err)
	}

	return filepath.ToSlash(rel), nil
}

func (scanner familyRepScanner) record(module string) {
	family := module

	if prefix, _, ok := strings.Cut(module, pathSeparator); ok {
		family = prefix
	}

	if existing, ok := scanner.seen[family]; !ok || module < existing {
		scanner.seen[family] = module
	}
}

func (scanner familyRepScanner) scan(path string, d fs.DirEntry, walkErr error) error {
	if walkErr != nil {
		return walkErr
	}

	if d.IsDir() || d.Name() != skipTaskfileYML {
		return nil
	}

	module, err := scanner.moduleFor(path)
	if err != nil {
		return fmt.Errorf("scan family representative module: %w", err)
	}

	scanner.record(module)

	return nil
}

func sortedFamilyModules(seen map[string]string) []string {
	modules := make([]string, constZero, len(seen))

	for i := range seen {
		module := seen[i]

		modules = append(modules, module)
	}

	slices.Sort(modules)

	return modules
}

func assertMetadata(t *testing.T, check *metadataCheck) {
	t.Helper()

	content := loadMetadataContent(t, check)
	assertMetadataContentFormat(t, content)

	metadata := parseMetadata(t, content)

	assertMetadataHeader(t, &metadata, check.module)
	assertMetadataTasks(t, &metadata, check.expectedTasks)
	assertMetadataVariants(t, &metadata, check.expectedVariants)
}

func parseMetadata(t *testing.T, content []byte) moduleMetadata {
	t.Helper()

	var metadata moduleMetadata

	err := yaml.Unmarshal(content, &metadata)
	if err != nil {
		t.Fatalf(parseMetadataErrFormat, err)
	}

	metadata.ExportedTasks = metadataStringSequence(t, content, "exported_tasks")

	return metadata
}

func loadMetadataContent(t *testing.T, check *metadataCheck) []byte {
	t.Helper()

	metadataPath := filepath.ToSlash(filepath.Join(check.dir, "metadata.yml"))

	content, err := fs.ReadFile(os.DirFS(check.taskfilesDir), metadataPath)
	if err != nil {
		t.Fatalf("read metadata.yml: %v; add or update it to match the module Taskfile", err)
	}

	return content
}

func assertMetadataContentFormat(t *testing.T, content []byte) {
	t.Helper()

	if strings.Contains(string(content), crlfSeparator) {
		t.Fatal("metadata.yml must use LF line endings")
	}

	trimmedAllWhitespace := strings.TrimRight(string(content), " \t\r\n")
	trimmedNewlinesOnly := strings.TrimRight(string(content), crlfSeparator)

	if trimmedAllWhitespace != trimmedNewlinesOnly {
		t.Fatal("metadata.yml has trailing whitespace")
	}
}

func assertMetadataHeader(t *testing.T, metadata *moduleMetadata, module string) {
	t.Helper()

	if metadata.Schema != metadataSchema {
		t.Errorf("schema = %q, want %q", metadata.Schema, metadataSchema)
	}

	if metadata.Module != module {
		t.Errorf("module = %q, want %q", metadata.Module, module)
	}

	if metadata.Taskfile != skipTaskfileYML {
		t.Errorf("taskfile = %q, want %q", metadata.Taskfile, skipTaskfileYML)
	}
}

func assertMetadataTasks(t *testing.T, metadata *moduleMetadata, expectedTasks []string) {
	t.Helper()

	if !slices.IsSorted(metadata.ExportedTasks) {
		t.Errorf("exported_tasks must be sorted: %v", metadata.ExportedTasks)
	}

	for i := constOne; i < len(metadata.ExportedTasks); i++ {
		if metadata.ExportedTasks[i] == metadata.ExportedTasks[i-constOne] {
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

func assertMetadataVariants(t *testing.T, metadata *moduleMetadata, expectedVariants []string) {
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
		t.Fatalf(parseMetadataErrFormat, err)
	}

	root := doc.Content[constZero]

	for i := constZero; i < len(root.Content); i += constTwo {
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

	values := make([]string, constZero, len(node.Content))

	for i := range node.Content {
		child := node.Content[i]

		values = append(values, child.Value)
	}

	return values
}

func sortedKeys[V any](items map[string]V) []string {
	keys := make([]string, constZero, len(items))

	for key := range items {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	return keys
}
