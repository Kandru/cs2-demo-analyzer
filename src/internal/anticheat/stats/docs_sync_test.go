package stats

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDetectionDocsExist keeps code and documentation in lockstep: every
// registered detection/modifier must have a page with the expected marker and
// headings, and no orphan pages may exist.
func TestDetectionDocsExist(t *testing.T) {
	root := repoRoot(t)
	required := []string{
		"## Signal",
		"## Collection",
		"## Scoring",
		"## Evidence",
		"## Calibration",
		"## Interactions",
		"## Files",
		"## Tuning checklist",
	}
	for _, d := range detections {
		path := filepath.Join(root, d.Doc)
		body, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("detection %q: read %s: %v", d.ID, d.Doc, err)
			continue
		}
		assertDoc(t, d.ID, "detection", string(body), required)
	}

	modifierHeadings := []string{
		"## Signal",
		"## Apply",
		"## Published metrics",
		"## Report",
		"## Calibration",
		"## Files",
		"## Tuning checklist",
	}
	for _, m := range modifiers {
		path := filepath.Join(root, m.Doc)
		body, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("modifier %q: read %s: %v", m.ID, m.Doc, err)
			continue
		}
		assertDoc(t, m.ID, "modifier", string(body), modifierHeadings)
	}

	allowed := map[string]bool{
		"README.md":    true,
		"_TEMPLATE.md": true,
		"evidence.md":  true,
	}
	for _, d := range detections {
		allowed[filepath.Base(d.Doc)] = true
	}
	for _, m := range modifiers {
		allowed[filepath.Base(m.Doc)] = true
	}
	entries, err := os.ReadDir(filepath.Join(root, "docs", "detections"))
	if err != nil {
		t.Fatalf("read docs/detections: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		if !allowed[entry.Name()] {
			t.Errorf("orphan detection doc: docs/detections/%s", entry.Name())
		}
	}
}

func assertDoc(t *testing.T, id, kind, body string, headings []string) {
	t.Helper()
	wantMarker := "<!-- " + kind + ": " + id + " -->"
	if !strings.Contains(body, wantMarker) {
		t.Errorf("%s doc is missing marker %q", id, wantMarker)
	}
	for _, h := range headings {
		if !strings.Contains(body, h) {
			t.Errorf("%s doc is missing heading %q", id, h)
		}
	}
}

// repoRoot walks up from the test working directory until it finds the docs
// tree, so the test works from any package directory.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, "docs", "detections")); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repository root")
		}
		dir = parent
	}
}
