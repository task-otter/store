// REPLACE_ME 2026
// SPDX-License-Identifier: Apache-2.0

package taskfiles_test

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/task-otter/store/internal/tasktest"
)

const (
	biomeLintSkipVar     = "BIOME_LINT_SKIP_PATTERN"
	biomeFmtSkipVar      = "BIOME_FMT_SKIP_PATTERN"
	depcheckLintSkipVar  = "DEPCHECK_LINT_SKIP_PATTERN"
	eslintLintSkipVar    = "ESLINT_LINT_SKIP_PATTERN"
	htmlhintLintSkipVar  = "HTMLHINT_LINT_SKIP_PATTERN"
	knipLintSkipVar      = "KNIP_LINT_SKIP_PATTERN"
	prettierFmtSkipVar   = "PRETTIER_FMT_SKIP_PATTERN"
	spectralLintSkipVar  = "SPECTRAL_LINT_SKIP_PATTERN"
	sqlfluffModule       = "sqlfluff"
	stylelintLintSkipVar = "STYLELINT_LINT_SKIP_PATTERN"
	taskfileFlag         = "--taskfile"
	skipTaskfileYML      = "Taskfile.yml"
)

type skipPatternModule struct {
	name string
	vars []string
}

type sharedSkipFileMatcherCase struct {
	name     string
	pattern  string
	paths    []string
	retained []string
}

func skipPatternModules() []skipPatternModule {
	return append(skipPatternFlatModules(), skipPatternVariantModules()...)
}

func skipPatternFlatModules() []skipPatternModule {
	return []skipPatternModule{
		{name: "actionlint", vars: []string{"ACTIONLINT_LINT_SKIP_PATTERN"}},
		{name: "ansible", vars: []string{"ANSIBLE_LINT_SKIP_PATTERN"}},
		{name: "buf", vars: []string{"BUF_LINT_SKIP_PATTERN", "BUF_FMT_SKIP_PATTERN"}},
		{name: "cargo", vars: []string{"CARGO_LINT_SKIP_PATTERN", "CARGO_FMT_SKIP_PATTERN"}},
		{name: "djlint", vars: []string{"DJLINT_LINT_SKIP_PATTERN", "DJLINT_FMT_SKIP_PATTERN"}},
		{name: "dotenv-linter", vars: []string{"DOTENV_LINTER_LINT_SKIP_PATTERN"}},
		{name: "go", vars: []string{"GO_LINT_SKIP_PATTERN", "GO_FMT_SKIP_PATTERN"}},
		{name: "hadolint", vars: []string{"HADOLINT_LINT_SKIP_PATTERN"}},
		{name: "jsonlint", vars: []string{"JSONLINT_LINT_SKIP_PATTERN"}},
		{name: "protolint", vars: []string{"PROTOLINT_LINT_SKIP_PATTERN"}},
		{name: "rumdl", vars: []string{"RUMDL_LINT_SKIP_PATTERN", "RUMDL_FMT_SKIP_PATTERN"}},
		{name: "shellcheck", vars: []string{"SHELLCHECK_LINT_SKIP_PATTERN"}},
		{name: "shfmt", vars: []string{"SHFMT_FMT_SKIP_PATTERN"}},
		{name: sqlfluffModule, vars: []string{"SQLFLUFF_LINT_SKIP_PATTERN"}},
		{name: "staticcheck", vars: []string{"STATICCHECK_LINT_SKIP_PATTERN"}},
		{name: "yamllint", vars: []string{"YAMLLINT_LINT_SKIP_PATTERN"}},
		{name: "zizmor", vars: []string{"ZIZMOR_LINT_SKIP_PATTERN"}},
	}
}

func skipPatternVariantModules() []skipPatternModule {
	return []skipPatternModule{
		{name: "biome/bun", vars: []string{biomeLintSkipVar, biomeFmtSkipVar}},
		{name: "biome/node/fnm/npm", vars: []string{biomeLintSkipVar, biomeFmtSkipVar}},
		{name: "biome/node/nvm/npm", vars: []string{biomeLintSkipVar, biomeFmtSkipVar}},
		{name: "biome/node/fnm/pnpm", vars: []string{biomeLintSkipVar, biomeFmtSkipVar}},
		{name: "biome/node/nvm/pnpm", vars: []string{biomeLintSkipVar, biomeFmtSkipVar}},
		{name: "biome/node/fnm/yarn", vars: []string{biomeLintSkipVar, biomeFmtSkipVar}},
		{name: "biome/node/nvm/yarn", vars: []string{biomeLintSkipVar, biomeFmtSkipVar}},
		{name: "depcheck/bun", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/fnm/npm", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/nvm/npm", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/fnm/pnpm", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/nvm/pnpm", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/fnm/yarn", vars: []string{depcheckLintSkipVar}},
		{name: "depcheck/node/nvm/yarn", vars: []string{depcheckLintSkipVar}},
		{name: "eslint/bun", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/fnm/npm", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/nvm/npm", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/fnm/pnpm", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/nvm/pnpm", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/fnm/yarn", vars: []string{eslintLintSkipVar}},
		{name: "eslint/node/nvm/yarn", vars: []string{eslintLintSkipVar}},
		{name: "htmlhint/node/fnm/npm", vars: []string{htmlhintLintSkipVar}},
		{name: "htmlhint/node/nvm/npm", vars: []string{htmlhintLintSkipVar}},
		{name: "htmlhint/node/fnm/pnpm", vars: []string{htmlhintLintSkipVar}},
		{name: "htmlhint/node/nvm/pnpm", vars: []string{htmlhintLintSkipVar}},
		{name: "knip/bun", vars: []string{knipLintSkipVar}},
		{name: "knip/node/fnm/npm", vars: []string{knipLintSkipVar}},
		{name: "knip/node/nvm/npm", vars: []string{knipLintSkipVar}},
		{name: "knip/node/fnm/pnpm", vars: []string{knipLintSkipVar}},
		{name: "knip/node/nvm/pnpm", vars: []string{knipLintSkipVar}},
		{name: "knip/node/fnm/yarn", vars: []string{knipLintSkipVar}},
		{name: "knip/node/nvm/yarn", vars: []string{knipLintSkipVar}},
		{name: "prettier/bun", vars: []string{prettierFmtSkipVar}},
		{name: "prettier/node/fnm/npm", vars: []string{prettierFmtSkipVar}},
		{name: "prettier/node/nvm/npm", vars: []string{prettierFmtSkipVar}},
		{name: "prettier/node/fnm/pnpm", vars: []string{prettierFmtSkipVar}},
		{name: "prettier/node/nvm/pnpm", vars: []string{prettierFmtSkipVar}},
		{name: "prettier/node/fnm/yarn", vars: []string{prettierFmtSkipVar}},
		{name: "prettier/node/nvm/yarn", vars: []string{prettierFmtSkipVar}},
		{name: "spectral/node/fnm/npm", vars: []string{spectralLintSkipVar}},
		{name: "spectral/node/nvm/npm", vars: []string{spectralLintSkipVar}},
		{name: "spectral/node/fnm/pnpm", vars: []string{spectralLintSkipVar}},
		{name: "spectral/node/nvm/pnpm", vars: []string{spectralLintSkipVar}},
		{name: "stylelint/bun", vars: []string{stylelintLintSkipVar}},
		{name: "stylelint/node/fnm/npm", vars: []string{stylelintLintSkipVar}},
		{name: "stylelint/node/nvm/npm", vars: []string{stylelintLintSkipVar}},
		{name: "stylelint/node/fnm/pnpm", vars: []string{stylelintLintSkipVar}},
		{name: "stylelint/node/nvm/pnpm", vars: []string{stylelintLintSkipVar}},
		{name: "stylelint/node/fnm/yarn", vars: []string{stylelintLintSkipVar}},
		{name: "stylelint/node/nvm/yarn", vars: []string{stylelintLintSkipVar}},
	}
}

// Modules that still delegate to the shared skipfiles helper. Biome and Knip
// dropped out when their config overlays moved into their own config:skip
// tasks; SQLFluff and Go stay because they still use filter / go-packages.
func sharedSkipfilesConsumers() []string {
	return []string{
		"actionlint", "ansible", "buf", "cargo", "dotenv-linter", "go", "hadolint",
		"jsonlint", "protolint", "shellcheck", "shfmt", sqlfluffModule, "yamllint", "zizmor",
	}
}

// Modules that own their config overlay through a local config:skip task.
func configSkipModules() []string {
	return []string{
		"biome/bun", "biome/node/fnm/npm", "biome/node/nvm/npm",
		"biome/node/fnm/pnpm", "biome/node/nvm/pnpm", "biome/node/fnm/yarn", "biome/node/nvm/yarn",
		"go", "knip/bun",
		"knip/node/fnm/npm", "knip/node/nvm/npm", "knip/node/fnm/pnpm", "knip/node/nvm/pnpm",
		"knip/node/fnm/yarn", "knip/node/nvm/yarn", sqlfluffModule,
	}
}

func TestSkipPatternContract(t *testing.T) {
	t.Parallel()

	if len(skipPatternModules()) != 67 {
		t.Fatalf("skip-pattern module count = %d, want 67", len(skipPatternModules()))
	}

	root := tasktest.RepoRoot(t)

	for _, module := range skipPatternModules() {
		t.Run(module.name, func(t *testing.T) {
			t.Parallel()

			taskfile := tasktest.LoadTaskfile(t, module.name)
			taskfileContent := readFile(
				t,
				filepath.Join(root, "taskfiles", module.name, skipTaskfileYML),
			)
			// Tool families keep a single README at the tool root; flat modules
			// keep their own. Resolve the documentation module accordingly.
			docModule := module.name

			if index := strings.IndexByte(docModule, '/'); index >= 0 {
				docModule = docModule[:index]
			}

			readmeContent := readFile(t, filepath.Join(root, "taskfiles", docModule, "README.md"))

			for _, variable := range module.vars {
				value, exists := taskfile.Vars[variable]

				if !exists {
					t.Errorf("%s is not defined", variable)
				} else if value != "" {
					t.Errorf("%s default = %#v, want empty", variable, value)
				}

				if strings.Count(taskfileContent, variable) < 2 {
					t.Errorf("%s is declared but not used by a task", variable)
				}

				if !strings.Contains(readmeContent, "`"+variable+"`") {
					t.Errorf("README does not document %s", variable)
				}
			}
		})
	}
}

func TestSkipPatternVariantParity(t *testing.T) {
	t.Parallel()

	families := map[string][]string{
		"biome":     {biomeLintSkipVar, biomeFmtSkipVar},
		"depcheck":  {depcheckLintSkipVar},
		"eslint":    {eslintLintSkipVar},
		"htmlhint":  {htmlhintLintSkipVar},
		"knip":      {knipLintSkipVar},
		"prettier":  {prettierFmtSkipVar},
		"spectral":  {spectralLintSkipVar},
		"stylelint": {stylelintLintSkipVar},
	}

	root := tasktest.RepoRoot(t)

	for family, variables := range families {
		t.Run(family, func(t *testing.T) {
			t.Parallel()

			paths := variantLeaves(t, root, family)

			if len(paths) < 2 {
				t.Fatalf("found %d variants, want at least 2", len(paths))
			}

			for _, variable := range variables {
				want := strings.Count(readFile(t, paths[0]), variable)

				for _, path := range paths[1:] {
					if got := strings.Count(readFile(t, path), variable); got != want {
						t.Errorf(
							"%s uses %s %d times, want %d",
							filepath.Base(filepath.Dir(path)),
							variable,
							got,
							want,
						)
					}
				}
			}
		})
	}
}

func TestSharedSkipFileMatcher(t *testing.T) {
	t.Parallel()

	root := tasktest.RepoRoot(t)
	filter := filepath.Join(root, "taskfiles", "internal", "skipfiles", skipTaskfileYML)

	for _, test := range sharedSkipFileMatcherCases() {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertSharedSkipFileMatcher(t, root, filter, test)
		})
	}
}

func sharedSkipFileMatcherCases() []sharedSkipFileMatcherCase {
	return []sharedSkipFileMatcherCase{
		{
			name:     "single star stays in segment",
			pattern:  "*.go",
			paths:    []string{"main.go", "cmd/main.go"},
			retained: []string{"cmd/main.go"},
		},
		{
			name:     "double star crosses directories",
			pattern:  "**/generated/**",
			paths:    []string{"generated/a.go", "src/generated/a.go", "src/main.go"},
			retained: []string{"src/main.go"},
		},
		{
			name:     "question mark and spaces",
			pattern:  "src/?ock/*.go",
			paths:    []string{"src/mock/a.go", "src/lock/file with space.go", "src/tools/a.go"},
			retained: []string{"src/tools/a.go"},
		},
		{
			name:     "windows separators normalize",
			pattern:  "**/generated/**",
			paths:    []string{`src\generated\a.go`, `src\main.go`},
			retained: []string{`src\main.go`},
		},
	}
}

func assertSharedSkipFileMatcher(
	t *testing.T,
	root string,
	filter string,
	test sharedSkipFileMatcherCase,
) {

	t.Helper()

	separator := "\x00"

	if runtime.GOOS == "windows" {
		separator = "\n"
	}

	input := []byte(strings.Join(test.paths, separator) + separator)
	commandContext := exec.CommandContext
	command := commandContext(
		t.Context(),
		"task", "--silent", "--taskfile", filter, "filter",
		"SKIPFILES_PATTERN="+test.pattern,
	)

	command.Dir = root
	command.Stdin = bytes.NewReader(input)

	output, err := command.Output()
	if err != nil {
		t.Fatalf("run filter: %v", err)
	}

	outputText := strings.ReplaceAll(string(output), "\r\n", "\n")

	actual := strings.Split(strings.TrimSuffix(outputText, separator), separator)

	if len(output) == 0 {
		actual = nil
	}

	if strings.Join(actual, "\x00") != strings.Join(test.retained, "\x00") {
		t.Fatalf("retained paths = %q, want %q", actual, test.retained)
	}
}

// runConfigSkip runs a module's config:skip task inside project, which is the
// USER_WORKING_DIR the task writes its overlay relative to.
func runConfigSkip(t *testing.T, project, module string, vars ...string) {
	t.Helper()

	root := tasktest.RepoRoot(t)
	arguments := append([]string{
		"--silent", taskfileFlag,
		filepath.Join(root, "taskfiles", module, skipTaskfileYML),
		"config:skip",
	}, vars...)
	runCommand(t, project, "task", arguments...)
}

// writeFixture creates a file inside a fresh project directory.
func writeFixture(t *testing.T, project, name, content string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(project, name), []byte(content), 0o600)
	if err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// assertOverlayContains reads an overlay written into project and checks tokens.
func assertOverlayContains(t *testing.T, project, name string, tokens ...string) {
	t.Helper()

	content := readFile(t, filepath.Join(project, name))

	for _, token := range tokens {
		if !strings.Contains(content, token) {
			t.Fatalf("%s does not contain %q:\n%s", name, token, content)
		}
	}
}

func TestBiomeConfigSkipTask(t *testing.T) {
	t.Parallel()

	const overlay = ".taskotter-biome-bun-skip.json"

	t.Run("both patterns", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		runConfigSkip(t, project, "biome/bun",
			"BIOME_LINT_SKIP_PATTERN=**/generated/**",
			"BIOME_FMT_SKIP_PATTERN=**/vendor/**")
		assertOverlayContains(t, project, overlay, `"!**/generated/**"`, `"!**/vendor/**"`)
	})

	t.Run("lint scope excludes fmt pattern", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		runConfigSkip(t, project, "biome/bun", "SKIP_SCOPE=lint",
			"BIOME_LINT_SKIP_PATTERN=**/generated/**",
			"BIOME_FMT_SKIP_PATTERN=**/vendor/**")
		assertOverlayContains(t, project, overlay, `"!**/generated/**"`)

		if content := readFile(
			t,
			filepath.Join(project, overlay),
		); strings.Contains(
			content,
			"vendor",
		) {

			t.Fatalf("lint-scoped overlay leaked the fmt pattern:\n%s", content)
		}
	})

	t.Run("extends a discovered config", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		writeFixture(t, project, "biome.json", `{"linter":{"enabled":true}}`)
		runConfigSkip(t, project, "biome/bun", "BIOME_LINT_SKIP_PATTERN=**/generated/**")
		assertOverlayContains(t, project, overlay, `"extends":["biome.json"]`)
	})

	t.Run("no pattern removes a stale overlay", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		writeFixture(t, project, overlay, "stale\n")
		runConfigSkip(t, project, "biome/bun")

		_, err := os.Stat(filepath.Join(project, overlay))

		if !os.IsNotExist(err) {
			t.Fatal("empty skip pattern did not remove the stale overlay")
		}
	})
}

// jsRuntimeVar pins the Knip overlay generator to a JS runtime that exists on
// this machine. The bun module defaults to bun, which CI does not install.
func jsRuntimeVar(t *testing.T) string {
	t.Helper()

	for _, runtime := range []string{"bun", "node"} {
		_, err := exec.LookPath(runtime)
		if err == nil {
			return "KNIP_INTERNAL_JS_RUNTIME=" + runtime
		}
	}

	t.Skip("neither bun nor node is installed")

	return ""
}

func TestKnipConfigSkipTask(t *testing.T) {
	t.Parallel()

	const overlay = ".taskotter-knip-bun-skip.json"

	testKnipConfigSkipNoProjectConfig(t, overlay)
	testKnipConfigSkipMergesJSONC(t, overlay)
	testKnipConfigSkipMergesPackageJSON(t, overlay)
	testKnipConfigSkipRejectsDynamicJS(t)
	testKnipConfigSkipRemovesStaleOverlay(t, overlay)
}

func testKnipConfigSkipNoProjectConfig(t *testing.T, overlay string) {
	t.Helper()

	t.Run("no project config", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		runConfigSkip(
			t,
			project,
			"knip/bun",
			"KNIP_LINT_SKIP_PATTERN=**/generated/**",
			jsRuntimeVar(t),
		)
		assertOverlayContains(t, project, overlay, "**/generated/**")
	})
}

func testKnipConfigSkipMergesJSONC(t *testing.T, overlay string) {
	t.Helper()

	t.Run("merges jsonc", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		writeFixture(t, project, "knip.jsonc",
			"{\n  // keep this entry\n  \"entry\": [\"src/index.ts\"],\n}\n")
		runConfigSkip(
			t,
			project,
			"knip/bun",
			"KNIP_LINT_SKIP_PATTERN=**/generated/**",
			jsRuntimeVar(t),
		)
		assertOverlayContains(t, project, overlay, "src/index.ts", "**/generated/**")
	})
}

func testKnipConfigSkipMergesPackageJSON(t *testing.T, overlay string) {
	t.Helper()

	t.Run("merges the package.json knip section", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		writeFixture(t, project, "package.json",
			`{"name":"fixture","knip":{"ignore":["existing/**"]}}`)
		runConfigSkip(
			t,
			project,
			"knip/bun",
			"KNIP_LINT_SKIP_PATTERN=**/generated/**",
			jsRuntimeVar(t),
		)
		assertOverlayContains(t, project, overlay, "existing/**", "**/generated/**")
	})
}

func testKnipConfigSkipRejectsDynamicJS(t *testing.T) {
	t.Helper()

	t.Run("rejects a dynamic JS config", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		writeFixture(t, project, "knip.config.js", "export default {};\n")

		commandContext := exec.CommandContext
		command := commandContext(t.Context(), "task", "--silent", "--taskfile",
			filepath.Join(tasktest.RepoRoot(t), "taskfiles", "knip", "bun", skipTaskfileYML),
			"config:skip", "KNIP_LINT_SKIP_PATTERN=**/generated/**")

		command.Dir = project

		output, err := command.CombinedOutput()

		if err == nil || !strings.Contains(string(output), "dynamic JS/TS Knip config") {
			t.Fatalf("dynamic Knip config was not rejected clearly: err=%v\n%s", err, output)
		}
	})
}

func testKnipConfigSkipRemovesStaleOverlay(t *testing.T, overlay string) {
	t.Helper()

	t.Run("no pattern removes a stale overlay", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		writeFixture(t, project, overlay, "stale\n")
		runConfigSkip(t, project, "knip/bun")

		_, err := os.Stat(filepath.Join(project, overlay))

		if !os.IsNotExist(err) {
			t.Fatal("empty skip pattern did not remove the stale overlay")
		}
	})
}

func TestSQLFluffConfigSkipTask(t *testing.T) {
	t.Parallel()

	const overlay = ".taskotter-sqlfluff-skip.cfg"

	t.Run("merges ignore_paths and normalizes separators", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		writeFixture(t, project, "source.cfg",
			"[sqlfluff]\ndialect = postgres\nignore_paths = build/**\n")
		runConfigSkip(t, project, sqlfluffModule,
			`SQLFLUFF_LINT_SKIP_PATTERN=**\generated\**`, "CONFIG_OVERRIDE=source.cfg")
		assertOverlayContains(t, project, overlay,
			"dialect = postgres", "ignore_paths = build/**,**/generated/**")
	})

	t.Run("no project config", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		runConfigSkip(t, project, sqlfluffModule, "SQLFLUFF_LINT_SKIP_PATTERN=**/generated/**")
		assertOverlayContains(t, project, overlay, "[sqlfluff]", "ignore_paths = **/generated/**")
	})

	t.Run("no pattern removes a stale overlay", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		writeFixture(t, project, overlay, "stale\n")
		runConfigSkip(t, project, sqlfluffModule)

		_, err := os.Stat(filepath.Join(project, overlay))

		if !os.IsNotExist(err) {
			t.Fatal("empty skip pattern did not remove the stale overlay")
		}
	})
}

func TestGoConfigSkipTask(t *testing.T) {
	t.Parallel()

	const overlay = ".golangci-taskotter-skip.yml"

	t.Run("translates the glob into an exclusion regex", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		runConfigSkip(t, project, "go", "GO_LINT_SKIP_PATTERN=**/generated/**")
		assertOverlayContains(t, project, overlay,
			"linters:", "exclusions:", "paths:", `^(?:.*/)?generated/.*$`)
	})

	t.Run("rewrites rather than accumulating overlays", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		runConfigSkip(t, project, "go", "GO_LINT_SKIP_PATTERN=**/generated/**")
		runConfigSkip(t, project, "go", "GO_LINT_SKIP_PATTERN=**/mocks/*.go")

		content := readFile(t, filepath.Join(project, overlay))

		if strings.Contains(content, "generated") {
			t.Fatalf("second run did not replace the first pattern:\n%s", content)
		}

		if !strings.Contains(content, `^(?:.*/)?mocks/[^/]*\.go$`) {
			t.Fatalf("second run did not write the new pattern:\n%s", content)
		}

		entries, err := os.ReadDir(project)
		if err != nil {
			t.Fatalf("read project: %v", err)
		}

		if len(entries) != 1 {
			t.Fatalf("expected exactly one overlay file, found %d", len(entries))
		}
	})

	t.Run("no pattern removes a stale overlay", func(t *testing.T) {
		t.Parallel()

		project := t.TempDir()
		writeFixture(t, project, overlay, "stale\n")
		runConfigSkip(t, project, "go")

		_, err := os.Stat(filepath.Join(project, overlay))

		if !os.IsNotExist(err) {
			t.Fatal("empty skip pattern did not remove the stale overlay")
		}
	})
}

func TestSharedSkipfilesTaskfileContract(t *testing.T) {
	t.Parallel()

	root := tasktest.RepoRoot(t)
	helperDirectory := filepath.Join(root, "taskfiles", "internal", "skipfiles")
	assertSharedSkipfilesDirectory(t, helperDirectory)
	assertSharedSkipfilesConsumers(t, root)
	assertConfigSkipModules(t)
}

func assertSharedSkipfilesDirectory(t *testing.T, helperDirectory string) {
	t.Helper()

	entries, err := os.ReadDir(helperDirectory)
	if err != nil {
		t.Fatalf("read shared skipfiles directory: %v", err)
	}

	if len(entries) != 1 || entries[0].Name() != skipTaskfileYML {
		names := make([]string, 0, len(entries))

		for _, entry := range entries {
			names = append(names, entry.Name())
		}

		t.Fatalf("shared skipfiles directory contains %v, want only Taskfile.yml", names)
	}

	// The helper is down to filter and go-packages; overlay generation now lives
	// in each tool's own config:skip task.
	for _, removed := range []string{"prepare-overlay:", "cleanup:"} {
		if strings.Contains(readFile(t, filepath.Join(helperDirectory, skipTaskfileYML)), removed) {
			t.Errorf("shared skipfiles Taskfile still defines %s", removed)
		}
	}
}

func assertSharedSkipfilesConsumers(t *testing.T, root string) {
	t.Helper()

	if len(sharedSkipfilesConsumers()) != 14 {
		t.Fatalf("shared skipfiles consumer count = %d, want 14", len(sharedSkipfilesConsumers()))
	}

	for _, module := range sharedSkipfilesConsumers() {
		content := readFile(t, filepath.Join(root, "taskfiles", module, skipTaskfileYML))

		if !strings.Contains(content, "internal/skipfiles/Taskfile.yml") ||
			!strings.Contains(content, "internal: true") {

			t.Errorf("%s does not include the shared skipfiles Taskfile internally", module)
		}

		for _, removed := range []string{
			"filter.sh", "Filter-SkipFiles.ps1", "prepare-overlay.sh",
			"Prepare-Overlay.ps1", "knip-config.mjs",
		} {
			if strings.Contains(content, removed) {
				t.Errorf("%s still references removed helper %s", module, removed)
			}
		}
	}
}

func assertConfigSkipModules(t *testing.T) {
	t.Helper()

	if len(configSkipModules()) != 16 {
		t.Fatalf("config:skip module count = %d, want 16", len(configSkipModules()))
	}

	for _, module := range configSkipModules() {
		taskfile := tasktest.LoadTaskfile(t, module)

		task, exists := taskfile.Tasks["config:skip"]

		if !exists {
			t.Errorf("%s does not define a config:skip task", module)

			continue
		}

		if task.Internal {
			t.Errorf("%s config:skip is internal, want a public task", module)
		}

		// run: once would let one caller's overlay be reused by the next call in
		// the same run, which passes a different scope or pattern.
		if task.Run == "once" {
			t.Errorf("%s config:skip must not use run: once", module)
		}
	}
}

func TestActionlintSkipPatternFiltersFiles(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping task integration test in short mode")
	}

	fixture := newActionlintSkipFixture(t)
	fixture.assertDefaultDiscovery(t)
	fixture.assertCliTargets(t)
	fixture.assertAllSkipped(t)
}

type actionlintSkipFixture struct {
	taskfilePath string
	project      string
	binDirectory string
	logPath      string
	skippedPath  string
	cliGoodPath  string
	env          []string
}

func newActionlintSkipFixture(t *testing.T) actionlintSkipFixture {
	t.Helper()

	root := tasktest.RepoRoot(t)
	project := t.TempDir()
	binDirectory := filepath.Join(project, "bin")
	workflowDirectory := filepath.Join(project, ".github", "workflows")
	generatedDirectory := filepath.Join(workflowDirectory, "generated")

	cliWorkflowDirectory := filepath.Join(project, "custom workflows")

	for _, directory := range []string{binDirectory, generatedDirectory, cliWorkflowDirectory} {
		mustMkdirAll(t, directory)
	}

	goodPath := filepath.Join(workflowDirectory, "good workflow.yml")
	skippedPath := filepath.Join(generatedDirectory, "bad.yml")

	cliGoodPath := filepath.Join(cliWorkflowDirectory, "cli good.yml")

	for _, path := range []string{goodPath, skippedPath, cliGoodPath} {
		mustWriteFile(t, path, "name: test\n", 0o644)
	}

	logPath := filepath.Join(project, "actionlint.args")

	stub := `#!/usr/bin/env bash
if [[ "${1-}" == "--version" ]]; then
  echo "1.7.12"
  exit 0
fi
printf '%s\n' "$@" >"$TASKOTTER_ACTIONLINT_LOG"
`

	mustWriteFile(t, filepath.Join(binDirectory, "actionlint"), stub, 0o500)

	return actionlintSkipFixture{
		taskfilePath: filepath.Join(root, "taskfiles", "actionlint", skipTaskfileYML),
		project:      project,
		binDirectory: binDirectory,
		logPath:      logPath,
		skippedPath:  skippedPath,
		cliGoodPath:  cliGoodPath,
		env: append(os.Environ(),
			"PATH="+binDirectory+":"+os.Getenv("PATH"),
			"TASKOTTER_ACTIONLINT_LOG="+logPath,
		),
	}
}

func (fixture actionlintSkipFixture) taskCommand(args ...string) *exec.Cmd {
	commandContext := exec.CommandContext
	command := commandContext(
		context.Background(),
		"task",
		append([]string{taskfileFlag, fixture.taskfilePath}, args...)...)

	command.Dir = fixture.project
	command.Env = fixture.env

	return command
}

func (fixture actionlintSkipFixture) assertDefaultDiscovery(t *testing.T) {
	t.Helper()

	output := fixture.runTask(t, "--yes", "lint", "ACTIONLINT_LINT_SKIP_PATTERN=**/generated/**")
	arguments := fixture.readLog(t, output)

	if !strings.Contains(arguments, "good workflow.yml") {
		t.Fatalf("actionlint arguments do not contain retained file:\n%s", arguments)
	}

	if strings.Contains(arguments, "generated/bad.yml") {
		t.Fatalf("actionlint arguments contain skipped file:\n%s", arguments)
	}

	fixture.removeLog(t, "first")
}

func (fixture actionlintSkipFixture) assertCliTargets(t *testing.T) {
	t.Helper()

	output := fixture.runTask(t,
		"--yes", "lint", "ACTIONLINT_LINT_SKIP_PATTERN=**/generated/**", "--",
		filepath.ToSlash(fixture.cliGoodPath), filepath.ToSlash(fixture.skippedPath), "-oneline",
	)

	arguments := fixture.readLog(t, output)

	if !strings.Contains(arguments, "cli good.yml") || !strings.Contains(arguments, "-oneline") {
		t.Fatalf("actionlint CLI targets or options were not retained:\n%s", arguments)
	}

	if strings.Contains(arguments, "generated/bad.yml") {
		t.Fatalf("actionlint CLI target bypassed skip filtering:\n%s", arguments)
	}

	fixture.removeLog(t, "CLI-target")
}

func (fixture actionlintSkipFixture) assertAllSkipped(t *testing.T) {
	t.Helper()

	fixture.runTask(t, "--yes", "lint", "ACTIONLINT_LINT_SKIP_PATTERN=**")

	_, err := os.Stat(fixture.logPath)

	if !os.IsNotExist(err) {
		t.Fatalf("actionlint ran even though every workflow was skipped")
	}
}

func (fixture actionlintSkipFixture) runTask(t *testing.T, args ...string) []byte {
	t.Helper()

	output, err := fixture.taskCommand(args...).CombinedOutput()
	if err != nil {
		t.Fatalf("run actionlint task: %v\n%s", err, output)
	}

	return output
}

func (fixture actionlintSkipFixture) readLog(t *testing.T, output []byte) string {
	t.Helper()

	argumentBytes, err := os.ReadFile(fixture.logPath)
	if err != nil {
		t.Fatalf("read actionlint log: %v\ntask output:\n%s", err, output)
	}

	return string(argumentBytes)
}

func (fixture actionlintSkipFixture) removeLog(t *testing.T, name string) {
	t.Helper()

	err := os.Remove(fixture.logPath)
	if err != nil {
		t.Fatalf("remove %s actionlint log: %v", name, err)
	}
}

func TestCargoSkipPatternExcludesWorkspacePackages(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping task integration test in short mode")
	}

	fixture := newCargoSkipFixture(t)
	fixture.assertRetainedPackage(t)
	fixture.assertAllSkipped(t)
}

type cargoSkipFixture struct {
	taskfilePath string
	project      string
	logPath      string
	env          []string
}

func newCargoSkipFixture(t *testing.T) cargoSkipFixture {
	t.Helper()

	root := tasktest.RepoRoot(t)
	project := t.TempDir()
	binDirectory := filepath.Join(project, "bin")
	goodSourceDirectory := filepath.Join(project, "good", "src")
	generatedDirectory := filepath.Join(project, "generated package")

	generatedSourceDirectory := filepath.Join(generatedDirectory, "src")

	for _, directory := range []string{binDirectory, goodSourceDirectory, generatedSourceDirectory} {
		mustMkdirAll(t, directory)
	}

	files := map[string]string{
		filepath.Join(project, "Cargo.toml"):              "[workspace]\nmembers = [\"good\", \"generated package\"]\n",
		filepath.Join(project, "good", "Cargo.toml"):      "[package]\nname = \"good_package\"\nversion = \"0.1.0\"\n",
		filepath.Join(generatedDirectory, "Cargo.toml"):   "[package]\nname = \"generated_package\"\nversion = \"0.1.0\"\n",
		filepath.Join(goodSourceDirectory, "lib.rs"):      "pub fn good() {}\n",
		filepath.Join(generatedSourceDirectory, "lib.rs"): "pub fn generated() {}\n",
	}

	for path, content := range files {
		mustWriteFile(t, path, content, 0o644)
	}

	logPath := filepath.Join(project, "cargo.args")

	stub := `#!/usr/bin/env bash
printf '%s\n' "$@" >"$TASKOTTER_CARGO_LOG"
`

	mustWriteFile(t, filepath.Join(binDirectory, "cargo"), stub, 0o500)

	return cargoSkipFixture{
		taskfilePath: filepath.Join(root, "taskfiles", "cargo", skipTaskfileYML),
		project:      project,
		logPath:      logPath,
		env: append(os.Environ(),
			"PATH="+binDirectory+":"+os.Getenv("PATH"),
			"TASKOTTER_CARGO_LOG="+logPath,
		),
	}
}

func (fixture cargoSkipFixture) taskCommand(args ...string) *exec.Cmd {
	commandContext := exec.CommandContext
	command := commandContext(
		context.Background(),
		"task",
		append([]string{taskfileFlag, fixture.taskfilePath}, args...)...)

	command.Dir = fixture.project
	command.Env = fixture.env

	return command
}

func (fixture cargoSkipFixture) assertRetainedPackage(t *testing.T) {
	t.Helper()

	fixture.runTask(t, "--yes", "lint", "CARGO_LINT_SKIP_PATTERN=**/generated package/**")

	arguments := readFile(t, fixture.logPath)

	if !strings.Contains(arguments, "clippy") || !strings.Contains(arguments, "good_package") {
		t.Fatalf("Cargo did not lint retained package:\n%s", arguments)
	}

	if strings.Contains(arguments, "generated_package") {
		t.Fatalf("Cargo lint included skipped package:\n%s", arguments)
	}

	mustRemove(t, fixture.logPath)
}

func (fixture cargoSkipFixture) assertAllSkipped(t *testing.T) {
	t.Helper()

	fixture.runTask(t, "--yes", "lint", "CARGO_LINT_SKIP_PATTERN=**/*.rs")

	_, err := os.Stat(fixture.logPath)

	if !os.IsNotExist(err) {
		t.Fatalf("Cargo ran even though every workspace package was skipped")
	}
}

func (fixture cargoSkipFixture) runTask(t *testing.T, args ...string) {
	t.Helper()

	output, err := fixture.taskCommand(args...).CombinedOutput()
	if err != nil {
		t.Fatalf("run Cargo task: %v\n%s", err, output)
	}
}

func TestGoAnalysisSkipPatternExcludesPackages(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping task integration test in short mode")
	}

	fixture := newGoAnalysisSkipFixture(t)
	fixture.assertRetainedPackage(t)
	fixture.assertAllSkipped(t)
}

type goAnalysisSkipFixture struct {
	taskfilePath string
	project      string
	logPath      string
	env          []string
}

func newGoAnalysisSkipFixture(t *testing.T) goAnalysisSkipFixture {
	t.Helper()

	root := tasktest.RepoRoot(t)
	project := t.TempDir()
	binDirectory := filepath.Join(project, "bin")
	goodDirectory := filepath.Join(project, "good")

	generatedDirectory := filepath.Join(project, "generated")

	for _, directory := range []string{binDirectory, goodDirectory, generatedDirectory} {
		mustMkdirAll(t, directory)
	}

	files := map[string]string{
		filepath.Join(project, "go.mod"):            "module example.com/skipfixture\n\ngo 1.25\n",
		filepath.Join(goodDirectory, "good.go"):     "package good\n",
		filepath.Join(generatedDirectory, "bad.go"): "package generated\n",
	}

	for path, content := range files {
		mustWriteFile(t, path, content, 0o644)
	}

	logPath := filepath.Join(project, "govulncheck.args")

	stub := `#!/usr/bin/env bash
printf '%s\n' "$@" >"$TASKOTTER_GO_ANALYSIS_LOG"
`

	mustWriteFile(t, filepath.Join(binDirectory, "govulncheck"), stub, 0o500)

	return goAnalysisSkipFixture{
		taskfilePath: filepath.Join(root, "taskfiles", "go", skipTaskfileYML),
		project:      project,
		logPath:      logPath,
		env: append(os.Environ(),
			"GOBIN="+binDirectory,
			"TASKOTTER_GO_ANALYSIS_LOG="+logPath,
			"GOCACHE=/private/tmp/taskotter-gocache",
		),
	}
}

func (fixture goAnalysisSkipFixture) taskCommand(args ...string) *exec.Cmd {
	commandContext := exec.CommandContext
	command := commandContext(
		context.Background(),
		"task",
		append([]string{taskfileFlag, fixture.taskfilePath}, args...)...)

	command.Dir = fixture.project
	command.Env = fixture.env

	return command
}

func (fixture goAnalysisSkipFixture) assertRetainedPackage(t *testing.T) {
	t.Helper()

	fixture.runTask(t, "--yes", "govulncheck:lint", "GO_LINT_SKIP_PATTERN=generated/**")

	arguments := readFile(t, fixture.logPath)

	if !strings.Contains(arguments, "example.com/skipfixture/good") {
		t.Fatalf("govulncheck did not receive retained package:\n%s", arguments)
	}

	if strings.Contains(arguments, "example.com/skipfixture/generated") {
		t.Fatalf("govulncheck received skipped package:\n%s", arguments)
	}

	mustRemove(t, fixture.logPath)
}

func (fixture goAnalysisSkipFixture) assertAllSkipped(t *testing.T) {
	t.Helper()

	fixture.runTask(t, "--yes", "govulncheck:lint", "GO_LINT_SKIP_PATTERN=**/*.go")

	_, err := os.Stat(fixture.logPath)

	if !os.IsNotExist(err) {
		t.Fatalf("govulncheck ran even though every Go package was skipped")
	}
}

func (fixture goAnalysisSkipFixture) runTask(t *testing.T, args ...string) {
	t.Helper()

	output, err := fixture.taskCommand(args...).CombinedOutput()
	if err != nil {
		t.Fatalf("run govulncheck task: %v\n%s", err, output)
	}
}

// variantLeaves returns every concrete leaf Taskfile of a nested tool family
// (the bun leaf plus each node/<vm>/<pm> leaf), excluding aggregators.
func variantLeaves(t *testing.T, root, family string) []string {
	t.Helper()

	var paths []string

	for _, pattern := range []string{
		filepath.Join(root, "taskfiles", family, "bun", skipTaskfileYML),
		filepath.Join(root, "taskfiles", family, "node", "*", "*", skipTaskfileYML),
	} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob variant leaves: %v", err)
		}

		paths = append(paths, matches...)
	}

	return paths
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return string(content)
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()

	err := os.MkdirAll(path, 0o700)
	if err != nil {
		t.Fatalf("create directory %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, content string, perm os.FileMode) {
	t.Helper()

	err := os.WriteFile(path, []byte(content), perm)
	if err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustRemove(t *testing.T, path string) {
	t.Helper()

	err := os.Remove(path)
	if err != nil {
		t.Fatalf("remove %s: %v", path, err)
	}
}

func runCommand(t *testing.T, directory string, name string, arguments ...string) {
	t.Helper()

	if name != "task" {
		t.Fatalf("unsupported command %q", name)
	}

	commandContext := exec.CommandContext
	command := commandContext(t.Context(), "task", arguments...)

	command.Dir = directory

	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run %s: %v\n%s", name, err, output)
	}
}
