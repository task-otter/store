// Command gen_taskfile_metadata writes the public task metadata for every
// TaskOtter module.
//
// Flat modules (taskfiles/<name>/Taskfile.yml) get one metadata.yml describing
// their exported tasks. The JS-tool families are nested
// (taskfiles/<tool>/{bun|node/<vm>/<pm>}/Taskfile.yml); every leaf of a tool
// exposes the same interface, so a single metadata.yml is written at the tool
// root listing the shared exported_tasks plus the concrete variants. Generation
// fails loudly if a tool's leaves ever disagree on their exported tasks.
package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

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

type taskfile struct {
	Tasks map[string]task `yaml:"tasks"`
}

type task struct {
	Internal bool `yaml:"internal"`
}

type metadata struct {
	Schema        string   `yaml:"schema"`
	Module        string   `yaml:"module"`
	Taskfile      string   `yaml:"taskfile"`
	ExportedTasks []string `yaml:"exported_tasks"`
	Variants      []string `yaml:"variants,omitempty"`
}

type leaf struct {
	rel   string
	tasks []string
}

func main() {
	var paths []string
	err := filepath.WalkDir("taskfiles", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "Taskfile.yml" {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		fatalf("walk taskfiles: %v", err)
	}
	if len(paths) == 0 {
		fatalf("no module Taskfiles found; run this command from the repository root")
	}

	tools := map[string][]leaf{}
	count := 0
	for _, path := range paths {
		tasks, err := exportedTasks(path)
		if err != nil {
			fatalf("read %s: %v", path, err)
		}
		rel, err := filepath.Rel("taskfiles", path)
		if err != nil {
			fatalf("rel %s: %v", path, err)
		}
		segs := strings.Split(filepath.ToSlash(filepath.Dir(rel)), "/")

		switch {
		case toolFamilies[segs[0]]:
			// Skip the tool-root and intermediate aggregators (they expose no
			// tasks of their own); collect the real leaves.
			if len(segs) == 1 || len(tasks) == 0 {
				continue
			}
			tools[segs[0]] = append(tools[segs[0]], leaf{rel: strings.Join(segs[1:], "/"), tasks: tasks})
		case len(segs) == 1:
			if err := writeMetadata(filepath.Dir(path), metadata{
				Schema:        metadataSchema,
				Module:        segs[0],
				Taskfile:      "Taskfile.yml",
				ExportedTasks: tasks,
			}); err != nil {
				fatalf("generate metadata for %s: %v", path, err)
			}
			count++
		default:
			// Nested non-tool helper (e.g. internal/skipfiles): no metadata,
			// matching historical behavior.
		}
	}

	for _, tool := range sortedKeys(tools) {
		leaves := tools[tool]
		slices.SortFunc(leaves, func(a, b leaf) int { return strings.Compare(a.rel, b.rel) })
		base := leaves[0].tasks
		variants := make([]string, 0, len(leaves))
		for _, l := range leaves {
			if !slices.Equal(l.tasks, base) {
				fatalf(
					"exported-task interface drift in tool %q:\n  %s: %v\n  %s: %v",
					tool, leaves[0].rel, base, l.rel, l.tasks,
				)
			}
			variants = append(variants, l.rel)
		}
		if err := writeMetadata(filepath.Join("taskfiles", tool), metadata{
			Schema:        metadataSchema,
			Module:        tool,
			Taskfile:      "Taskfile.yml",
			ExportedTasks: base,
			Variants:      variants,
		}); err != nil {
			fatalf("generate metadata for tool %s: %v", tool, err)
		}
		count++
	}

	fmt.Printf("generated metadata for %d modules\n", count)
}

func exportedTasks(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Taskfile: %w", err)
	}
	var source taskfile
	if err := yaml.Unmarshal(content, &source); err != nil {
		return nil, fmt.Errorf("parse Taskfile: %w", err)
	}
	tasks := make([]string, 0, len(source.Tasks))
	for name, t := range source.Tasks {
		if name == "default" || strings.HasPrefix(name, "_") || t.Internal {
			continue
		}
		tasks = append(tasks, name)
	}
	slices.Sort(tasks)
	return tasks, nil
}

func writeMetadata(dir string, document metadata) error {
	var output bytes.Buffer
	output.WriteString("---\n")
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(document); err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("close metadata encoder: %w", err)
	}

	metadataPath := filepath.Join(dir, "metadata.yml")
	if err := os.WriteFile(metadataPath, output.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", metadataPath, err)
	}
	return nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
