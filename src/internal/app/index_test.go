package app

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Kandru/cs2-demo-analyzer/internal/config"
	"github.com/Kandru/cs2-demo-analyzer/internal/report/html"
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

func TestCollectIndexEntriesUsesMetadataAndSortsNewestFirst(t *testing.T) {
	dir := t.TempDir()
	p := testProcessor()

	addReport := func(name, demoName string, mod time.Time, meta *store.ReportMeta) {
		t.Helper()
		demo := filepath.Join(dir, demoName)
		report := filepath.Join(dir, name+".html")
		if err := os.WriteFile(demo, []byte("demo"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(report, []byte("report"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(report, mod, mod); err != nil {
			t.Fatal(err)
		}
		p.store.ReserveOutput(report, demo)
		if meta != nil {
			p.store.MarkOK(demo, 1, mod, "", "", report, "", "v", *meta)
		}
	}

	addReport("alpha", "alpha.dem", time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC), &store.ReportMeta{
		Map: "de_mirage", ServerName: "Alpha Server", Score: [2]int{13, 7}, HasScore: true,
		PlayedAt: time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC),
	})
	addReport("bravo", "bravo.dem", time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC), &store.ReportMeta{
		Map: "de_nuke", ServerName: "Bravo Server", Score: [2]int{10, 13}, HasScore: true,
		PlayedAt: time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC),
	})
	// Newest file mtime but no metadata: it still sorts by fallback time and is
	// grouped as unknown.
	addReport("unknown", "unknown.dem", time.Date(2026, 11, 17, 8, 0, 0, 0, time.UTC), nil)

	entries, err := p.indexEntries(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3: %+v", len(entries), entries)
	}
	if got := []string{entries[0].Href, entries[1].Href, entries[2].Href}; got[0] != "unknown.html" || got[1] != "bravo.html" || got[2] != "alpha.html" {
		t.Fatalf("unexpected order: %v", got)
	}
	bravo := entries[1]
	if bravo.Name != "Nuke" || bravo.Score != "10 : 13" || bravo.ServerName != "Bravo Server" {
		t.Fatalf("metadata not rendered into entry: %+v", bravo)
	}
	if bravo.Played != "2026-11-16 09:00" {
		t.Fatalf("played time = %q", bravo.Played)
	}
	if bravo.Demo == nil || bravo.Demo.Href != "bravo.dem" || bravo.Demo.Size != "4 B" {
		t.Fatalf("demo link not resolved: %+v", bravo)
	}
	if entries[0].Name != "unknown" || entries[0].ServerName != "" || entries[0].Score != "" {
		t.Fatalf("fallback entry should use the report basename: %+v", entries[0])
	}
}

func TestGroupIndexEntriesKeepsUnknownServerLast(t *testing.T) {
	entries := []html.IndexEntry{
		{Href: "a.html", Name: "A", ServerName: "", PlayedAt: time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)},
		{Href: "b.html", Name: "B", ServerName: "Bravo", PlayedAt: time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)},
		{Href: "c.html", Name: "C", ServerName: "Alpha", PlayedAt: time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)},
		{Href: "d.html", Name: "D", ServerName: "Bravo", PlayedAt: time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)},
	}
	groups := groupIndexEntries(entries)
	if len(groups) != 3 {
		t.Fatalf("got %d groups, want 3: %+v", len(groups), groups)
	}
	if groups[0].Name != "Bravo" || len(groups[0].Entries) != 2 {
		t.Fatalf("Bravo group = %+v", groups[0])
	}
	if groups[1].Name != "Alpha" {
		t.Fatalf("Alpha group = %+v", groups[1])
	}
	if groups[2].Name != "Unknown server" {
		t.Fatalf("unknown group should be last: %+v", groups[2])
	}
}

func TestWriteIndexSkipsEmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	p := testProcessor()
	if err := p.writeIndex(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, html.IndexFileName)); !os.IsNotExist(err) {
		t.Fatalf("empty directory should not get an index: %v", err)
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
	p.store.MarkOK(demo, 1, time.Now(), "", "", report, "", "v", store.ReportMeta{
		Map: "de_mirage", ServerName: "Test Server", Score: [2]int{13, 7}, HasScore: true,
		PlayedAt: time.Date(2026, 11, 16, 12, 11, 0, 0, time.UTC),
	})
	if err := p.writeIndex(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	page := string(data)
	for _, want := range []string{
		"Match reports", "match.html", "Download demo", "match.dem",
		"Mirage", "13 : 7", "Test Server", "2026-11-16 12:11",
		`href="https://github.com/Kandru/cs2-demo-analyzer"`, "generated ",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in index", want)
		}
	}
	for _, unwanted := range []string{"Self-contained index", "no external assets", "Generated 2026"} {
		if strings.Contains(page, unwanted) {
			t.Errorf("removed text %q still present", unwanted)
		}
	}
}
