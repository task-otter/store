package taskfiles_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
	"gopkg.in/yaml.v3"
)

const metadataSchema = "taskotter.dev/taskfile-metadata/v1"

var toolFamilies = map[string]bool{
	"eslint": true, "prettier": true, "biome": true, "bruno": true,
	"depcheck": true, "knip": true, "stylelint": true, "typescript": true,
	"htmlhint": true, "spectral": true,
	// Package managers are families too: taskfiles/<pm>/{fnm,nvm}.
	"npm": true, "pnpm": true, "yarn": true, "corepack": true,
}

type moduleMetadata struct {
	Schema        string   `yaml:"schema"`
	Module        string   `yaml:"module"`
	Taskfile      string   `yaml:"taskfile"`
	ExportedTasks []string `yaml:"exported_tasks"`
	Variants      []string `yaml:"variants"`
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

	var flatModules []string
	toolLeaves := map[string][]string{} // tool -> leaf module paths

	err := filepath.WalkDir(taskfilesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "Taskfile.yml" {
			return nil
		}
		rel, err := filepath.Rel(taskfilesDir, filepath.Dir(path))
		if err != nil {
			return err
		}
		module := filepath.ToSlash(rel)
		segs := strings.Split(module, "/")

		switch {
		case toolFamilies[segs[0]]:
			if len(segs) == 1 || len(exportedTasks(t, module)) == 0 {
				return nil // tool-root / intermediate aggregator
			}
			toolLeaves[segs[0]] = append(toolLeaves[segs[0]], module)
		case len(segs) == 1:
			flatModules = append(flatModules, module)
		default:
			// nested non-tool helper (e.g. internal/skipfiles): no metadata
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk taskfiles: %v", err)
	}
	if len(flatModules) == 0 || len(toolLeaves) == 0 {
		t.Fatal("no modules discovered")
	}

	for _, module := range flatModules {
		module := module
		t.Run(module, func(t *testing.T) {
			assertMetadata(t, taskfilesDir, module, module, exportedTasks(t, module), nil)
		})
	}

	for _, tool := range sortedKeys(toolLeaves) {
		leaves := toolLeaves[tool]
		slices.Sort(leaves)
		t.Run(tool, func(t *testing.T) {
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
		})
	}
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
		if d.IsDir() || d.Name() != "Taskfile.yml" {
			return nil
		}
		rel, err := filepath.Rel(taskfilesDir, filepath.Dir(path))
		if err != nil {
			return err
		}
		module := filepath.ToSlash(rel)
		family := strings.Split(module, "/")[0]
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

func assertMetadata(t *testing.T, taskfilesDir, dir, module string, expectedTasks, expectedVariants []string) {
	t.Helper()
	metadataPath := filepath.Join(taskfilesDir, dir, "metadata.yml")
	content, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("read metadata.yml: %v; regenerate with go run ./scripts/gen_taskfile_metadata.go", err)
	}
	if strings.Contains(string(content), "\r\n") {
		t.Fatal("metadata.yml must use LF line endings")
	}
	if strings.TrimRight(string(content), " \t\r\n") != strings.TrimRight(string(content), "\r\n") {
		t.Fatal("metadata.yml has trailing whitespace")
	}

	var metadata moduleMetadata
	if err := yaml.Unmarshal(content, &metadata); err != nil {
		t.Fatalf("parse metadata.yml: %v", err)
	}
	if metadata.Schema != metadataSchema {
		t.Errorf("schema = %q, want %q", metadata.Schema, metadataSchema)
	}
	if metadata.Module != module {
		t.Errorf("module = %q, want %q", metadata.Module, module)
	}
	if metadata.Taskfile != "Taskfile.yml" {
		t.Errorf("taskfile = %q, want %q", metadata.Taskfile, "Taskfile.yml")
	}
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
			"exported task drift\nmetadata: %v\ntaskfile: %v\nregenerate with go run ./scripts/gen_taskfile_metadata.go",
			metadata.ExportedTasks, expectedTasks,
		)
	}
	if !slices.Equal(metadata.Variants, expectedVariants) {
		t.Fatalf(
			"variant drift\nmetadata: %v\non disk:  %v\nregenerate with go run ./scripts/gen_taskfile_metadata.go",
			metadata.Variants, expectedVariants,
		)
	}
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
