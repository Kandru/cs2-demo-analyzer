package aim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMetricDocsExist keeps metric specs and docs/aim pages in lockstep.
func TestMetricDocsExist(t *testing.T) {
	root := repoRoot(t)
	required := []string{
		"## Formula",
		"## Collection",
		"## Sample gate",
		"## Anchors and weight",
		"## Caveats",
		"## Files",
		"## Tuning checklist",
	}
	allowed := map[string]bool{"README.md": true, "_TEMPLATE.md": true}
	for _, spec := range metricSpecs {
		allowed[filepath.Base(spec.Doc)] = true
		body, err := os.ReadFile(filepath.Join(root, spec.Doc))
		if err != nil {
			t.Errorf("metric %q: read %s: %v", spec.ID, spec.Doc, err)
			continue
		}
		if !strings.Contains(string(body), "<!-- metric: "+spec.ID+" -->") {
			t.Errorf("metric %q doc is missing its marker", spec.ID)
		}
		for _, h := range required {
			if !strings.Contains(string(body), h) {
				t.Errorf("metric %q doc is missing heading %q", spec.ID, h)
			}
		}
	}

	entries, err := os.ReadDir(filepath.Join(root, "docs", "aim"))
	if err != nil {
		t.Fatalf("read docs/aim: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		if !allowed[entry.Name()] {
			t.Errorf("orphan aim doc: docs/aim/%s", entry.Name())
		}
	}
}

// repoRoot walks up from the test working directory until it finds docs/aim.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, "docs", "aim")); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repository root")
		}
		dir = parent
	}
}
