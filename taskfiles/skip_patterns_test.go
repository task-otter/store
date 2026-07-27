package taskfiles_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mostafakhairy0305-dot/TaskOtter/internal/tasktest"
)

type skipPatternModule struct {
	name string
	vars []string
}

var skipPatternModules = []skipPatternModule{
	{name: "actionlint", vars: []string{"ACTIONLINT_LINT_SKIP_PATTERN"}},
	{name: "ansible", vars: []string{"ANSIBLE_LINT_SKIP_PATTERN"}},
	{name: "biome/bun", vars: []string{"BIOME_LINT_SKIP_PATTERN", "BIOME_FMT_SKIP_PATTERN"}},
	{name: "biome/node/fnm/npm", vars: []string{"BIOME_LINT_SKIP_PATTERN", "BIOME_FMT_SKIP_PATTERN"}},
	{name: "biome/node/nvm/npm", vars: []string{"BIOME_LINT_SKIP_PATTERN", "BIOME_FMT_SKIP_PATTERN"}},
	{name: "biome/node/fnm/pnpm", vars: []string{"BIOME_LINT_SKIP_PATTERN", "BIOME_FMT_SKIP_PATTERN"}},
	{name: "biome/node/nvm/pnpm", vars: []string{"BIOME_LINT_SKIP_PATTERN", "BIOME_FMT_SKIP_PATTERN"}},
	{name: "biome/node/fnm/yarn", vars: []string{"BIOME_LINT_SKIP_PATTERN", "BIOME_FMT_SKIP_PATTERN"}},
	{name: "biome/node/nvm/yarn", vars: []string{"BIOME_LINT_SKIP_PATTERN", "BIOME_FMT_SKIP_PATTERN"}},
	{name: "buf", vars: []string{"BUF_LINT_SKIP_PATTERN", "BUF_FMT_SKIP_PATTERN"}},
	{name: "cargo", vars: []string{"CARGO_LINT_SKIP_PATTERN", "CARGO_FMT_SKIP_PATTERN"}},
	{name: "depcheck/bun", vars: []string{"DEPCHECK_LINT_SKIP_PATTERN"}},
	{name: "depcheck/node/fnm/npm", vars: []string{"DEPCHECK_LINT_SKIP_PATTERN"}},
	{name: "depcheck/node/nvm/npm", vars: []string{"DEPCHECK_LINT_SKIP_PATTERN"}},
	{name: "depcheck/node/fnm/pnpm", vars: []string{"DEPCHECK_LINT_SKIP_PATTERN"}},
	{name: "depcheck/node/nvm/pnpm", vars: []string{"DEPCHECK_LINT_SKIP_PATTERN"}},
	{name: "depcheck/node/fnm/yarn", vars: []string{"DEPCHECK_LINT_SKIP_PATTERN"}},
	{name: "depcheck/node/nvm/yarn", vars: []string{"DEPCHECK_LINT_SKIP_PATTERN"}},
	{name: "djlint", vars: []string{"DJLINT_LINT_SKIP_PATTERN", "DJLINT_FMT_SKIP_PATTERN"}},
	{name: "dotenv-linter", vars: []string{"DOTENV_LINTER_LINT_SKIP_PATTERN"}},
	{name: "eslint/bun", vars: []string{"ESLINT_LINT_SKIP_PATTERN"}},
	{name: "eslint/node/fnm/npm", vars: []string{"ESLINT_LINT_SKIP_PATTERN"}},
	{name: "eslint/node/nvm/npm", vars: []string{"ESLINT_LINT_SKIP_PATTERN"}},
	{name: "eslint/node/fnm/pnpm", vars: []string{"ESLINT_LINT_SKIP_PATTERN"}},
	{name: "eslint/node/nvm/pnpm", vars: []string{"ESLINT_LINT_SKIP_PATTERN"}},
	{name: "eslint/node/fnm/yarn", vars: []string{"ESLINT_LINT_SKIP_PATTERN"}},
	{name: "eslint/node/nvm/yarn", vars: []string{"ESLINT_LINT_SKIP_PATTERN"}},
	{name: "go", vars: []string{"GO_LINT_SKIP_PATTERN", "GO_FMT_SKIP_PATTERN"}},
	{name: "hadolint", vars: []string{"HADOLINT_LINT_SKIP_PATTERN"}},
	{name: "htmlhint/node/fnm/npm", vars: []string{"HTMLHINT_LINT_SKIP_PATTERN"}},
	{name: "htmlhint/node/nvm/npm", vars: []string{"HTMLHINT_LINT_SKIP_PATTERN"}},
	{name: "htmlhint/node/fnm/pnpm", vars: []string{"HTMLHINT_LINT_SKIP_PATTERN"}},
	{name: "htmlhint/node/nvm/pnpm", vars: []string{"HTMLHINT_LINT_SKIP_PATTERN"}},
	{name: "jsonlint", vars: []string{"JSONLINT_LINT_SKIP_PATTERN"}},
	{name: "knip/bun", vars: []string{"KNIP_LINT_SKIP_PATTERN"}},
	{name: "knip/node/fnm/npm", vars: []string{"KNIP_LINT_SKIP_PATTERN"}},
	{name: "knip/node/nvm/npm", vars: []string{"KNIP_LINT_SKIP_PATTERN"}},
	{name: "knip/node/fnm/pnpm", vars: []string{"KNIP_LINT_SKIP_PATTERN"}},
	{name: "knip/node/nvm/pnpm", vars: []string{"KNIP_LINT_SKIP_PATTERN"}},
	{name: "knip/node/fnm/yarn", vars: []string{"KNIP_LINT_SKIP_PATTERN"}},
	{name: "knip/node/nvm/yarn", vars: []string{"KNIP_LINT_SKIP_PATTERN"}},
	{name: "prettier/bun", vars: []string{"PRETTIER_FMT_SKIP_PATTERN"}},
	{name: "prettier/node/fnm/npm", vars: []string{"PRETTIER_FMT_SKIP_PATTERN"}},
	{name: "prettier/node/nvm/npm", vars: []string{"PRETTIER_FMT_SKIP_PATTERN"}},
	{name: "prettier/node/fnm/pnpm", vars: []string{"PRETTIER_FMT_SKIP_PATTERN"}},
	{name: "prettier/node/nvm/pnpm", vars: []string{"PRETTIER_FMT_SKIP_PATTERN"}},
	{name: "prettier/node/fnm/yarn", vars: []string{"PRETTIER_FMT_SKIP_PATTERN"}},
	{name: "prettier/node/nvm/yarn", vars: []string{"PRETTIER_FMT_SKIP_PATTERN"}},
	{name: "protolint", vars: []string{"PROTOLINT_LINT_SKIP_PATTERN"}},
	{name: "rumdl", vars: []string{"RUMDL_LINT_SKIP_PATTERN", "RUMDL_FMT_SKIP_PATTERN"}},
	{name: "shellcheck", vars: []string{"SHELLCHECK_LINT_SKIP_PATTERN"}},
	{name: "shfmt", vars: []string{"SHFMT_FMT_SKIP_PATTERN"}},
	{name: "spectral/node/fnm/npm", vars: []string{"SPECTRAL_LINT_SKIP_PATTERN"}},
	{name: "spectral/node/nvm/npm", vars: []string{"SPECTRAL_LINT_SKIP_PATTERN"}},
	{name: "spectral/node/fnm/pnpm", vars: []string{"SPECTRAL_LINT_SKIP_PATTERN"}},
	{name: "spectral/node/nvm/pnpm", vars: []string{"SPECTRAL_LINT_SKIP_PATTERN"}},
	{name: "sqlfluff", vars: []string{"SQLFLUFF_LINT_SKIP_PATTERN"}},
	{name: "staticcheck", vars: []string{"STATICCHECK_LINT_SKIP_PATTERN"}},
	{name: "stylelint/bun", vars: []string{"STYLELINT_LINT_SKIP_PATTERN"}},
	{name: "stylelint/node/fnm/npm", vars: []string{"STYLELINT_LINT_SKIP_PATTERN"}},
	{name: "stylelint/node/nvm/npm", vars: []string{"STYLELINT_LINT_SKIP_PATTERN"}},
	{name: "stylelint/node/fnm/pnpm", vars: []string{"STYLELINT_LINT_SKIP_PATTERN"}},
	{name: "stylelint/node/nvm/pnpm", vars: []string{"STYLELINT_LINT_SKIP_PATTERN"}},
	{name: "stylelint/node/fnm/yarn", vars: []string{"STYLELINT_LINT_SKIP_PATTERN"}},
	{name: "stylelint/node/nvm/yarn", vars: []string{"STYLELINT_LINT_SKIP_PATTERN"}},
	{name: "yamllint", vars: []string{"YAMLLINT_LINT_SKIP_PATTERN"}},
	{name: "zizmor", vars: []string{"ZIZMOR_LINT_SKIP_PATTERN"}},
}

// Modules that still delegate to the shared skipfiles helper. Biome and Knip
// dropped out when their config overlays moved into their own config:skip
// tasks; SQLFluff and Go stay because they still use filter / go-packages.
var sharedSkipfilesConsumers = []string{
	"actionlint", "ansible", "buf", "cargo", "dotenv-linter", "go", "hadolint",
	"jsonlint", "protolint", "shellcheck", "shfmt", "sqlfluff", "yamllint", "zizmor",
}

// Modules that own their config overlay through a local config:skip task.
var configSkipModules = []string{
	"biome/bun", "biome/node/fnm/npm", "biome/node/nvm/npm",
	"biome/node/fnm/pnpm", "biome/node/nvm/pnpm", "biome/node/fnm/yarn", "biome/node/nvm/yarn",
	"go", "knip/bun",
	"knip/node/fnm/npm", "knip/node/nvm/npm", "knip/node/fnm/pnpm", "knip/node/nvm/pnpm",
	"knip/node/fnm/yarn", "knip/node/nvm/yarn", "sqlfluff",
}

func TestSkipPatternContract(t *testing.T) {
	if len(skipPatternModules) != 67 {
		t.Fatalf("skip-pattern module count = %d, want 67", len(skipPatternModules))
	}

	root := tasktest.RepoRoot(t)
	for _, module := range skipPatternModules {
		t.Run(module.name, func(t *testing.T) {
			taskfile := tasktest.LoadTaskfile(t, module.name)
			taskfileContent := readFile(t, filepath.Join(root, "taskfiles", module.name, "Taskfile.yml"))
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
	families := map[string][]string{
		"biome":     {"BIOME_LINT_SKIP_PATTERN", "BIOME_FMT_SKIP_PATTERN"},
		"depcheck":  {"DEPCHECK_LINT_SKIP_PATTERN"},
		"eslint":    {"ESLINT_LINT_SKIP_PATTERN"},
		"htmlhint":  {"HTMLHINT_LINT_SKIP_PATTERN"},
		"knip":      {"KNIP_LINT_SKIP_PATTERN"},
		"prettier":  {"PRETTIER_FMT_SKIP_PATTERN"},
		"spectral":  {"SPECTRAL_LINT_SKIP_PATTERN"},
		"stylelint": {"STYLELINT_LINT_SKIP_PATTERN"},
	}
	root := tasktest.RepoRoot(t)
	for family, variables := range families {
		t.Run(family, func(t *testing.T) {
			paths := variantLeaves(t, root, family)
			if len(paths) < 2 {
				t.Fatalf("found %d variants, want at least 2", len(paths))
			}
			for _, variable := range variables {
				want := strings.Count(readFile(t, paths[0]), variable)
				for _, path := range paths[1:] {
					if got := strings.Count(readFile(t, path), variable); got != want {
						t.Errorf("%s uses %s %d times, want %d", filepath.Base(filepath.Dir(path)), variable, got, want)
					}
				}
			}
		})
	}
}

func TestSkipPatternRepresentativeDryRuns(t *testing.T) {
	const pattern = "**/generated/**"
	tests := []struct {
		module   string
		args     []string
		expected []string
	}{
		{module: "eslint/bun", args: []string{"lint", "ESLINT_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"--ignore-pattern", pattern}},
		{module: "prettier/bun", args: []string{"fmt:check", "PRETTIER_FMT_SKIP_PATTERN=" + pattern}, expected: []string{"!" + pattern}},
		{module: "biome/bun", args: []string{"ci", "BIOME_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"config:skip"}},
		{module: "knip/bun", args: []string{"lint", "KNIP_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"config:skip"}},
		{module: "actionlint", args: []string{"lint", "ACTIONLINT_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"internal/skipfiles/Taskfile.yml", pattern}},
		{module: "ansible", args: []string{"syntax:check", "PLAYBOOK_OVERRIDE=site.yml", "ANSIBLE_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"internal/skipfiles/Taskfile.yml", pattern}},
		{module: "buf", args: []string{"breaking", "BUF_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"internal/skipfiles/Taskfile.yml", pattern}},
		{module: "dotenv-linter", args: []string{"diff", "DOTENV_LINTER_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"internal/skipfiles/Taskfile.yml", pattern}},
		{module: "shellcheck", args: []string{"lint", "SHELLCHECK_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"internal/skipfiles/Taskfile.yml", pattern}},
		{module: "shfmt", args: []string{"fmt:check", "SHFMT_FMT_SKIP_PATTERN=" + pattern}, expected: []string{"internal/skipfiles/Taskfile.yml", pattern}},
		{module: "yamllint", args: []string{"ci", "YAMLLINT_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"internal/skipfiles/Taskfile.yml", pattern}},
		{module: "jsonlint", args: []string{"lint", "JSONLINT_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"internal/skipfiles/Taskfile.yml", pattern}},
		{module: "protolint", args: []string{"lint", "PROTOLINT_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"internal/skipfiles/Taskfile.yml", pattern}},
		{module: "sqlfluff", args: []string{"lint", "SQLFLUFF_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"config:skip"}},
		{module: "sqlfluff", args: []string{"parse", "SQLFLUFF_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"config:skip"}},
		{module: "cargo", args: []string{"lint", "CARGO_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"internal/skipfiles/Taskfile.yml", pattern}},
		{module: "staticcheck", args: []string{"lint", "STATICCHECK_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"internal/skipfiles/Taskfile.yml", pattern}},
		{module: "go", args: []string{"govulncheck:lint", "GO_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"internal/skipfiles/Taskfile.yml", pattern}},
		{module: "go", args: []string{"gosec:lint", "GO_LINT_SKIP_PATTERN=" + pattern}, expected: []string{"internal/skipfiles/Taskfile.yml", pattern}},
	}

	for _, test := range tests {
		t.Run(test.module, func(t *testing.T) {
			tasktest.AssertDryRunContains(t, test.module, test.args, test.expected...)
		})
	}
}

func TestSharedSkipFileMatcher(t *testing.T) {
	root := tasktest.RepoRoot(t)
	filter := filepath.Join(root, "taskfiles", "internal", "skipfiles", "Taskfile.yml")
	tests := []struct {
		name     string
		pattern  string
		paths    []string
		retained []string
	}{
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			separator := "\x00"
			if runtime.GOOS == "windows" {
				separator = "\n"
			}
			input := []byte(strings.Join(test.paths, separator) + separator)
			command := exec.Command(
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
		})
	}
}

// runConfigSkip runs a module's config:skip task inside project, which is the
// USER_WORKING_DIR the task writes its overlay relative to.
func runConfigSkip(t *testing.T, project, module string, vars ...string) {
	t.Helper()
	root := tasktest.RepoRoot(t)
	arguments := append([]string{
		"--silent", "--taskfile",
		filepath.Join(root, "taskfiles", module, "Taskfile.yml"),
		"config:skip",
	}, vars...)
	runCommand(t, project, "task", arguments...)
}

// writeFixture creates a file inside a fresh project directory.
func writeFixture(t *testing.T, project, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(project, name), []byte(content), 0o644); err != nil {
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
	const overlay = ".taskotter-biome-bun-skip.json"

	t.Run("both patterns", func(t *testing.T) {
		project := t.TempDir()
		runConfigSkip(t, project, "biome/bun",
			"BIOME_LINT_SKIP_PATTERN=**/generated/**",
			"BIOME_FMT_SKIP_PATTERN=**/vendor/**")
		assertOverlayContains(t, project, overlay, `"!**/generated/**"`, `"!**/vendor/**"`)
	})

	t.Run("lint scope excludes fmt pattern", func(t *testing.T) {
		project := t.TempDir()
		runConfigSkip(t, project, "biome/bun", "SKIP_SCOPE=lint",
			"BIOME_LINT_SKIP_PATTERN=**/generated/**",
			"BIOME_FMT_SKIP_PATTERN=**/vendor/**")
		assertOverlayContains(t, project, overlay, `"!**/generated/**"`)
		if content := readFile(t, filepath.Join(project, overlay)); strings.Contains(content, "vendor") {
			t.Fatalf("lint-scoped overlay leaked the fmt pattern:\n%s", content)
		}
	})

	t.Run("extends a discovered config", func(t *testing.T) {
		project := t.TempDir()
		writeFixture(t, project, "biome.json", `{"linter":{"enabled":true}}`)
		runConfigSkip(t, project, "biome/bun", "BIOME_LINT_SKIP_PATTERN=**/generated/**")
		assertOverlayContains(t, project, overlay, `"extends":["biome.json"]`)
	})

	t.Run("no pattern removes a stale overlay", func(t *testing.T) {
		project := t.TempDir()
		writeFixture(t, project, overlay, "stale\n")
		runConfigSkip(t, project, "biome/bun")
		if _, err := os.Stat(filepath.Join(project, overlay)); !os.IsNotExist(err) {
			t.Fatal("empty skip pattern did not remove the stale overlay")
		}
	})
}

// jsRuntimeVar pins the Knip overlay generator to a JS runtime that exists on
// this machine. The bun module defaults to bun, which CI does not install.
func jsRuntimeVar(t *testing.T) string {
	t.Helper()
	for _, runtime := range []string{"bun", "node"} {
		if _, err := exec.LookPath(runtime); err == nil {
			return "KNIP_INTERNAL_JS_RUNTIME=" + runtime
		}
	}
	t.Skip("neither bun nor node is installed")
	return ""
}

func TestKnipConfigSkipTask(t *testing.T) {
	const overlay = ".taskotter-knip-bun-skip.json"

	t.Run("no project config", func(t *testing.T) {
		project := t.TempDir()
		runConfigSkip(t, project, "knip/bun", "KNIP_LINT_SKIP_PATTERN=**/generated/**", jsRuntimeVar(t))
		assertOverlayContains(t, project, overlay, "**/generated/**")
	})

	t.Run("merges jsonc", func(t *testing.T) {
		project := t.TempDir()
		writeFixture(t, project, "knip.jsonc",
			"{\n  // keep this entry\n  \"entry\": [\"src/index.ts\"],\n}\n")
		runConfigSkip(t, project, "knip/bun", "KNIP_LINT_SKIP_PATTERN=**/generated/**", jsRuntimeVar(t))
		assertOverlayContains(t, project, overlay, "src/index.ts", "**/generated/**")
	})

	t.Run("merges the package.json knip section", func(t *testing.T) {
		project := t.TempDir()
		writeFixture(t, project, "package.json",
			`{"name":"fixture","knip":{"ignore":["existing/**"]}}`)
		runConfigSkip(t, project, "knip/bun", "KNIP_LINT_SKIP_PATTERN=**/generated/**", jsRuntimeVar(t))
		assertOverlayContains(t, project, overlay, "existing/**", "**/generated/**")
	})

	t.Run("rejects a dynamic JS config", func(t *testing.T) {
		project := t.TempDir()
		writeFixture(t, project, "knip.config.js", "export default {};\n")
		command := exec.Command("task", "--silent", "--taskfile",
			filepath.Join(tasktest.RepoRoot(t), "taskfiles", "knip", "bun", "Taskfile.yml"),
			"config:skip", "KNIP_LINT_SKIP_PATTERN=**/generated/**")
		command.Dir = project
		output, err := command.CombinedOutput()
		if err == nil || !strings.Contains(string(output), "dynamic JS/TS Knip config") {
			t.Fatalf("dynamic Knip config was not rejected clearly: err=%v\n%s", err, output)
		}
	})

	t.Run("no pattern removes a stale overlay", func(t *testing.T) {
		project := t.TempDir()
		writeFixture(t, project, overlay, "stale\n")
		runConfigSkip(t, project, "knip/bun")
		if _, err := os.Stat(filepath.Join(project, overlay)); !os.IsNotExist(err) {
			t.Fatal("empty skip pattern did not remove the stale overlay")
		}
	})
}

func TestSQLFluffConfigSkipTask(t *testing.T) {
	const overlay = ".taskotter-sqlfluff-skip.cfg"

	t.Run("merges ignore_paths and normalizes separators", func(t *testing.T) {
		project := t.TempDir()
		writeFixture(t, project, "source.cfg",
			"[sqlfluff]\ndialect = postgres\nignore_paths = build/**\n")
		runConfigSkip(t, project, "sqlfluff",
			`SQLFLUFF_LINT_SKIP_PATTERN=**\generated\**`, "CONFIG_OVERRIDE=source.cfg")
		assertOverlayContains(t, project, overlay,
			"dialect = postgres", "ignore_paths = build/**,**/generated/**")
	})

	t.Run("no project config", func(t *testing.T) {
		project := t.TempDir()
		runConfigSkip(t, project, "sqlfluff", "SQLFLUFF_LINT_SKIP_PATTERN=**/generated/**")
		assertOverlayContains(t, project, overlay, "[sqlfluff]", "ignore_paths = **/generated/**")
	})

	t.Run("no pattern removes a stale overlay", func(t *testing.T) {
		project := t.TempDir()
		writeFixture(t, project, overlay, "stale\n")
		runConfigSkip(t, project, "sqlfluff")
		if _, err := os.Stat(filepath.Join(project, overlay)); !os.IsNotExist(err) {
			t.Fatal("empty skip pattern did not remove the stale overlay")
		}
	})
}

func TestGoConfigSkipTask(t *testing.T) {
	const overlay = ".golangci-taskotter-skip.yml"

	t.Run("translates the glob into an exclusion regex", func(t *testing.T) {
		project := t.TempDir()
		runConfigSkip(t, project, "go", "GO_LINT_SKIP_PATTERN=**/generated/**")
		assertOverlayContains(t, project, overlay,
			"linters:", "exclusions:", "paths:", `^(?:.*/)?generated/.*$`)
	})

	t.Run("rewrites rather than accumulating overlays", func(t *testing.T) {
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
		project := t.TempDir()
		writeFixture(t, project, overlay, "stale\n")
		runConfigSkip(t, project, "go")
		if _, err := os.Stat(filepath.Join(project, overlay)); !os.IsNotExist(err) {
			t.Fatal("empty skip pattern did not remove the stale overlay")
		}
	})
}

func TestSharedSkipfilesTaskfileContract(t *testing.T) {
	root := tasktest.RepoRoot(t)
	helperDirectory := filepath.Join(root, "taskfiles", "internal", "skipfiles")
	entries, err := os.ReadDir(helperDirectory)
	if err != nil {
		t.Fatalf("read shared skipfiles directory: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "Taskfile.yml" {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Fatalf("shared skipfiles directory contains %v, want only Taskfile.yml", names)
	}

	// The helper is down to filter and go-packages; overlay generation now lives
	// in each tool's own config:skip task.
	for _, removed := range []string{"prepare-overlay:", "cleanup:"} {
		if strings.Contains(readFile(t, filepath.Join(helperDirectory, "Taskfile.yml")), removed) {
			t.Errorf("shared skipfiles Taskfile still defines %s", removed)
		}
	}

	if len(sharedSkipfilesConsumers) != 14 {
		t.Fatalf("shared skipfiles consumer count = %d, want 14", len(sharedSkipfilesConsumers))
	}
	for _, module := range sharedSkipfilesConsumers {
		content := readFile(t, filepath.Join(root, "taskfiles", module, "Taskfile.yml"))
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

	if len(configSkipModules) != 16 {
		t.Fatalf("config:skip module count = %d, want 16", len(configSkipModules))
	}
	for _, module := range configSkipModules {
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
	if testing.Short() {
		t.Skip("skipping task integration test in short mode")
	}
	root := tasktest.RepoRoot(t)
	project := t.TempDir()
	binDirectory := filepath.Join(project, "bin")
	workflowDirectory := filepath.Join(project, ".github", "workflows")
	generatedDirectory := filepath.Join(workflowDirectory, "generated")
	cliWorkflowDirectory := filepath.Join(project, "custom workflows")
	for _, directory := range []string{binDirectory, generatedDirectory, cliWorkflowDirectory} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatalf("create directory: %v", err)
		}
	}
	goodPath := filepath.Join(workflowDirectory, "good workflow.yml")
	skippedPath := filepath.Join(generatedDirectory, "bad.yml")
	cliGoodPath := filepath.Join(cliWorkflowDirectory, "cli good.yml")
	for _, path := range []string{goodPath, skippedPath, cliGoodPath} {
		if err := os.WriteFile(path, []byte("name: test\n"), 0o644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
	}
	logPath := filepath.Join(project, "actionlint.args")
	stub := `#!/usr/bin/env bash
if [[ "${1-}" == "--version" ]]; then
  echo "1.7.12"
  exit 0
fi
printf '%s\n' "$@" >"$TASKOTTER_ACTIONLINT_LOG"
`
	if err := os.WriteFile(filepath.Join(binDirectory, "actionlint"), []byte(stub), 0o755); err != nil {
		t.Fatalf("write actionlint stub: %v", err)
	}

	command := exec.Command("task", "--taskfile", filepath.Join(root, "taskfiles", "actionlint", "Taskfile.yml"),
		"--yes", "lint", "ACTIONLINT_LINT_SKIP_PATTERN=**/generated/**")
	command.Dir = project
	command.Env = append(os.Environ(),
		"PATH="+binDirectory+":"+os.Getenv("PATH"),
		"TASKOTTER_ACTIONLINT_LOG="+logPath,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run actionlint task: %v\n%s", err, output)
	}
	argumentBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read actionlint log: %v\ntask output:\n%s", err, output)
	}
	arguments := string(argumentBytes)
	if !strings.Contains(arguments, "good workflow.yml") {
		t.Fatalf("actionlint arguments do not contain retained file:\n%s", arguments)
	}
	if strings.Contains(arguments, "generated/bad.yml") {
		t.Fatalf("actionlint arguments contain skipped file:\n%s", arguments)
	}

	if err := os.Remove(logPath); err != nil {
		t.Fatalf("remove first actionlint log: %v", err)
	}
	cliTargets := exec.Command("task", "--taskfile", filepath.Join(root, "taskfiles", "actionlint", "Taskfile.yml"),
		"--yes", "lint", "ACTIONLINT_LINT_SKIP_PATTERN=**/generated/**", "--",
		filepath.ToSlash(cliGoodPath), filepath.ToSlash(skippedPath), "-oneline")
	cliTargets.Dir = project
	cliTargets.Env = command.Env
	if output, err := cliTargets.CombinedOutput(); err != nil {
		t.Fatalf("run actionlint task with CLI targets: %v\n%s", err, output)
	}
	argumentBytes, err = os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read actionlint CLI-target log: %v", err)
	}
	arguments = string(argumentBytes)
	if !strings.Contains(arguments, "cli good.yml") || !strings.Contains(arguments, "-oneline") {
		t.Fatalf("actionlint CLI targets or options were not retained:\n%s", arguments)
	}
	if strings.Contains(arguments, "generated/bad.yml") {
		t.Fatalf("actionlint CLI target bypassed skip filtering:\n%s", arguments)
	}
	if err := os.Remove(logPath); err != nil {
		t.Fatalf("remove CLI-target actionlint log: %v", err)
	}

	allSkipped := exec.Command("task", "--taskfile", filepath.Join(root, "taskfiles", "actionlint", "Taskfile.yml"),
		"--yes", "lint", "ACTIONLINT_LINT_SKIP_PATTERN=**")
	allSkipped.Dir = project
	allSkipped.Env = command.Env
	if output, err := allSkipped.CombinedOutput(); err != nil {
		t.Fatalf("run all-skipped actionlint task: %v\n%s", err, output)
	}
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Fatalf("actionlint ran even though every workflow was skipped")
	}
}

func TestCargoSkipPatternExcludesWorkspacePackages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping task integration test in short mode")
	}
	root := tasktest.RepoRoot(t)
	project := t.TempDir()
	binDirectory := filepath.Join(project, "bin")
	goodSourceDirectory := filepath.Join(project, "good", "src")
	generatedDirectory := filepath.Join(project, "generated package")
	generatedSourceDirectory := filepath.Join(generatedDirectory, "src")
	for _, directory := range []string{binDirectory, goodSourceDirectory, generatedSourceDirectory} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatalf("create Cargo fixture directory: %v", err)
		}
	}
	files := map[string]string{
		filepath.Join(project, "Cargo.toml"):              "[workspace]\nmembers = [\"good\", \"generated package\"]\n",
		filepath.Join(project, "good", "Cargo.toml"):      "[package]\nname = \"good_package\"\nversion = \"0.1.0\"\n",
		filepath.Join(generatedDirectory, "Cargo.toml"):   "[package]\nname = \"generated_package\"\nversion = \"0.1.0\"\n",
		filepath.Join(goodSourceDirectory, "lib.rs"):      "pub fn good() {}\n",
		filepath.Join(generatedSourceDirectory, "lib.rs"): "pub fn generated() {}\n",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write Cargo fixture: %v", err)
		}
	}
	logPath := filepath.Join(project, "cargo.args")
	stub := `#!/usr/bin/env bash
printf '%s\n' "$@" >"$TASKOTTER_CARGO_LOG"
`
	if err := os.WriteFile(filepath.Join(binDirectory, "cargo"), []byte(stub), 0o755); err != nil {
		t.Fatalf("write Cargo stub: %v", err)
	}

	command := exec.Command("task", "--taskfile", filepath.Join(root, "taskfiles", "cargo", "Taskfile.yml"),
		"--yes", "lint", "CARGO_LINT_SKIP_PATTERN=**/generated package/**")
	command.Dir = project
	command.Env = append(os.Environ(),
		"PATH="+binDirectory+":"+os.Getenv("PATH"),
		"TASKOTTER_CARGO_LOG="+logPath,
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("run Cargo task: %v\n%s", err, output)
	}
	arguments := readFile(t, logPath)
	if !strings.Contains(arguments, "clippy") || !strings.Contains(arguments, "good_package") {
		t.Fatalf("Cargo did not lint retained package:\n%s", arguments)
	}
	if strings.Contains(arguments, "generated_package") {
		t.Fatalf("Cargo lint included skipped package:\n%s", arguments)
	}

	if err := os.Remove(logPath); err != nil {
		t.Fatalf("remove Cargo log: %v", err)
	}
	allSkipped := exec.Command("task", "--taskfile", filepath.Join(root, "taskfiles", "cargo", "Taskfile.yml"),
		"--yes", "lint", "CARGO_LINT_SKIP_PATTERN=**/*.rs")
	allSkipped.Dir = project
	allSkipped.Env = command.Env
	if output, err := allSkipped.CombinedOutput(); err != nil {
		t.Fatalf("run all-skipped Cargo task: %v\n%s", err, output)
	}
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Fatalf("Cargo ran even though every workspace package was skipped")
	}
}

func TestGoAnalysisSkipPatternExcludesPackages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping task integration test in short mode")
	}
	root := tasktest.RepoRoot(t)
	project := t.TempDir()
	binDirectory := filepath.Join(project, "bin")
	goodDirectory := filepath.Join(project, "good")
	generatedDirectory := filepath.Join(project, "generated")
	for _, directory := range []string{binDirectory, goodDirectory, generatedDirectory} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatalf("create Go analysis fixture directory: %v", err)
		}
	}
	files := map[string]string{
		filepath.Join(project, "go.mod"):            "module example.com/skipfixture\n\ngo 1.25\n",
		filepath.Join(goodDirectory, "good.go"):     "package good\n",
		filepath.Join(generatedDirectory, "bad.go"): "package generated\n",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write Go analysis fixture: %v", err)
		}
	}
	logPath := filepath.Join(project, "govulncheck.args")
	stub := `#!/usr/bin/env bash
printf '%s\n' "$@" >"$TASKOTTER_GO_ANALYSIS_LOG"
`
	if err := os.WriteFile(filepath.Join(binDirectory, "govulncheck"), []byte(stub), 0o755); err != nil {
		t.Fatalf("write govulncheck stub: %v", err)
	}

	command := exec.Command("task", "--taskfile", filepath.Join(root, "taskfiles", "go", "Taskfile.yml"),
		"--yes", "govulncheck:lint", "GO_LINT_SKIP_PATTERN=generated/**")
	command.Dir = project
	command.Env = append(os.Environ(),
		"GOBIN="+binDirectory,
		"TASKOTTER_GO_ANALYSIS_LOG="+logPath,
		"GOCACHE=/private/tmp/taskotter-gocache",
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("run govulncheck task: %v\n%s", err, output)
	}
	arguments := readFile(t, logPath)
	if !strings.Contains(arguments, "example.com/skipfixture/good") {
		t.Fatalf("govulncheck did not receive retained package:\n%s", arguments)
	}
	if strings.Contains(arguments, "example.com/skipfixture/generated") {
		t.Fatalf("govulncheck received skipped package:\n%s", arguments)
	}

	if err := os.Remove(logPath); err != nil {
		t.Fatalf("remove govulncheck log: %v", err)
	}
	allSkipped := exec.Command("task", "--taskfile", filepath.Join(root, "taskfiles", "go", "Taskfile.yml"),
		"--yes", "govulncheck:lint", "GO_LINT_SKIP_PATTERN=**/*.go")
	allSkipped.Dir = project
	allSkipped.Env = command.Env
	if output, err := allSkipped.CombinedOutput(); err != nil {
		t.Fatalf("run all-skipped govulncheck task: %v\n%s", err, output)
	}
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Fatalf("govulncheck ran even though every Go package was skipped")
	}
}

// variantLeaves returns every concrete leaf Taskfile of a nested tool family
// (the bun leaf plus each node/<vm>/<pm> leaf), excluding aggregators.
func variantLeaves(t *testing.T, root, family string) []string {
	t.Helper()
	var paths []string
	for _, pattern := range []string{
		filepath.Join(root, "taskfiles", family, "bun", "Taskfile.yml"),
		filepath.Join(root, "taskfiles", family, "node", "*", "*", "Taskfile.yml"),
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
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func runCommand(t *testing.T, directory string, name string, arguments ...string) {
	t.Helper()
	command := exec.Command(name, arguments...)
	command.Dir = directory
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("run %s: %v\n%s", name, err, output)
	}
}
