package tasktest

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

var updateGolden = flag.Bool("update", false, "rewrite the dry-run goldens instead of comparing against them")

// TestGoldenDryRun pins the fully rendered command line of every public task
// across the base/override matrix.
//
// Regenerate with:
//
//	go test ./internal/tasktest -run TestGoldenDryRun -update
//
// Regenerating is only correct when a diff has been read and understood. A
// blanket -update to make the suite green defeats the entire purpose of the
// harness.
func TestGoldenDryRun(t *testing.T) {
	if _, err := os.Stat(filepath.Join(RepoRoot(t), "internal", "tasktest", "testdata", "golden", runtime.GOOS)); err != nil {
		if !*updateGolden {
			t.Skipf("no dry-run goldens recorded for GOOS=%s; capture them with -update on this platform", runtime.GOOS)
		}
	}

	for _, module := range GoldenModules(t) {
		for _, kase := range GoldenCases() {
			t.Run(module+"/"+string(kase), func(t *testing.T) {
				t.Parallel()

				got := GoldenRender(t, module, kase)
				path := GoldenPath(t, module, kase)

				if *updateGolden {
					if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
						t.Fatalf("create golden dir: %v", err)
					}
					if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
						t.Fatalf("write golden: %v", err)
					}
					return
				}

				want, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read golden %s: %v\n(regenerate with -update)", path, err)
				}

				if got != string(want) {
					t.Fatalf("dry-run output drift for %s/%s\n%s", module, kase, describeGoldenDiff(string(want), got))
				}
			})
		}
	}
}

// describeGoldenDiff reports the first differing line with surrounding context,
// rendering each side with visible line boundaries so that a lost or gained
// trailing space is legible in the failure output.
func describeGoldenDiff(want, got string) string {
	wantLines := splitLinesKeepEmpty(want)
	gotLines := splitLinesKeepEmpty(got)

	limit := min(len(wantLines), len(gotLines))
	for i := range limit {
		if wantLines[i] == gotLines[i] {
			continue
		}
		return "line " + itoa(i+1) + ":\n  want: |" + wantLines[i] + "|\n  got:  |" + gotLines[i] + "|"
	}

	if len(wantLines) != len(gotLines) {
		return "line count differs: want " + itoa(len(wantLines)) + ", got " + itoa(len(gotLines))
	}
	return "outputs differ but no line differs (trailing newline mismatch)"
}

func splitLinesKeepEmpty(text string) []string {
	var lines []string
	start := 0
	for i := range len(text) {
		if text[i] == '\n' {
			lines = append(lines, text[start:i])
			start = i + 1
		}
	}
	lines = append(lines, text[start:])
	return lines
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits []byte
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}
