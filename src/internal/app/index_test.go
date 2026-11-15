package app

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kandru/cs2-demo-analyzer/internal/config"
	"github.com/Kandru/cs2-demo-analyzer/internal/store"
)

func testProcessor() *Processor {
	return NewProcessor(config.Default(), store.NewMemory(), slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestDemoLink(t *testing.T) {
	dir := t.TempDir()
	p := testProcessor()
	demo := filepath.Join(dir, "match.dem")
	if err := os.WriteFile(demo, []byte("demo"), 0o644); err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(dir, "match.html")
	if link := p.demoLink(report, demo); link == nil || link.Href != "match.dem" {
		t.Fatalf("same-dir link = %+v", link)
	}

	out := filepath.Join(dir, "reports")
	if link := p.demoLink(filepath.Join(out, "match.html"), demo); link != nil {
		t.Fatal("cross-directory link should require --demo-download")
	}
	p.cfg.DemoLinks = true
	if link := p.demoLink(filepath.Join(out, "match.html"), demo); link == nil || link.Href != "../match.dem" {
		t.Fatalf("cross-dir link = %+v", link)
	}
}

func TestCollectIndexEntries(t *testing.T) {
	dir := t.TempDir()
	p := testProcessor()
	demo := filepath.Join(dir, "match.dem")
	report := filepath.Join(dir, "match.html")
	if err := os.WriteFile(demo, []byte("demo"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(report, []byte("report"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "other.html"), []byte("other"), 0o644); err != nil {
		t.Fatal(err)
	}
	p.store.ReserveOutput(report, demo)

	entries, err := p.indexEntries(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2: %+v", len(entries), entries)
	}
	byName := map[string]int{}
	for i, e := range entries {
		byName[e.Href] = i
	}
	i, ok := byName["match.html"]
	if !ok {
		t.Fatalf("match.html missing: %+v", entries)
	}
	e := entries[i]
	if e.Demo == nil || e.Demo.Href != "match.dem" || e.Demo.Name != "match.dem" || e.Demo.Size != "4 B" {
		t.Fatalf("demo link not resolved: %+v", e)
	}
	j, ok := byName["other.html"]
	if !ok {
		t.Fatalf("other.html missing: %+v", entries)
	}
	if entries[j].Demo != nil {
		t.Fatalf("unexpected demo link: %+v", entries[j])
	}
}

func TestWriteIndex(t *testing.T) {
	dir := t.TempDir()
	p := testProcessor()
	demo := filepath.Join(dir, "match.dem")
	report := filepath.Join(dir, "match.html")
	if err := os.WriteFile(demo, []byte("demo"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(report, []byte("report"), 0o644); err != nil {
		t.Fatal(err)
	}
	p.store.ReserveOutput(report, demo)
	if err := p.writeIndex(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	page := string(data)
	for _, want := range []string{"Match reports", "match.html", "Download demo", "match.dem"} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in index", want)
		}
	}
}
