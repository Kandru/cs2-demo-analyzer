package app

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kandru/cs2-demo-analyzer/internal/config"
	"github.com/Kandru/cs2-demo-analyzer/internal/model"
	"github.com/Kandru/cs2-demo-analyzer/internal/report"
	"github.com/Kandru/cs2-demo-analyzer/internal/report/html"
	"github.com/Kandru/cs2-demo-analyzer/internal/store"
	"github.com/Kandru/cs2-demo-analyzer/internal/version"
)

func testProcessorWithConfig(cfg config.Config) *Processor {
	return NewProcessor(cfg, store.NewMemory(), slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// withVersion pins the build version for the duration of a test.
func withVersion(t *testing.T, v, commit string) {
	t.Helper()
	oldV, oldC := version.Version, version.Commit
	version.Version, version.Commit = v, commit
	t.Cleanup(func() {
		version.Version, version.Commit = oldV, oldC
	})
}

func writeArtifacts(t *testing.T, p *Processor, demo string, info os.FileInfo) (outputPlan, report.Header) {
	t.Helper()
	plan := p.plan(demo)
	header := p.headerFor(demo, info.Size(), info.ModTime(), plan)
	doc := report.New(&model.Match{SchemaVersion: model.SchemaVersion, Map: "de_mirage"}, header)
	var page bytes.Buffer
	if err := html.Render(&page, doc, html.Options{}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan.HTMLPath, page.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	if p.cfg.JSON {
		var data bytes.Buffer
		if err := report.WriteJSON(&data, doc); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(plan.JSONPath, data.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return plan, header
}

func TestSkipUsesArtifactHeaders(t *testing.T) {
	withVersion(t, "26.11.1", "abc")
	dir := t.TempDir()
	demo := filepath.Join(dir, "match.dem")
	if err := os.WriteFile(demo, []byte("demo"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(demo)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Output = dir
	cfg.JSON = true
	p := testProcessorWithConfig(cfg)
	writeArtifacts(t, p, demo, info)

	if !p.Skip(demo, info.Size(), info.ModTime(), false) {
		t.Fatal("current HTML and JSON should be skipped")
	}
	if p.Skip(demo, info.Size()+1, info.ModTime(), false) {
		t.Fatal("changed size must invalidate")
	}
	if p.Skip(demo, info.Size(), info.ModTime(), true) {
		t.Fatal("--force must invalidate")
	}

	version.Version = "26.12.1"
	if p.Skip(demo, info.Size(), info.ModTime(), false) {
		t.Fatal("version change must invalidate")
	}
}

func TestSkipRequiresEnabledJSON(t *testing.T) {
	withVersion(t, "26.11.1", "abc")
	dir := t.TempDir()
	demo := filepath.Join(dir, "match.dem")
	if err := os.WriteFile(demo, []byte("demo"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(demo)
	cfg := config.Default()
	cfg.Output = dir
	cfg.JSON = true
	p := testProcessorWithConfig(cfg)
	plan, _ := writeArtifacts(t, p, demo, info)
	if err := os.Remove(plan.JSONPath); err != nil {
		t.Fatal(err)
	}
	if p.Skip(demo, info.Size(), info.ModTime(), false) {
		t.Fatal("missing requested JSON must trigger a rebuild")
	}

	// JSON is optional: without --json the HTML alone is sufficient.
	p.cfg.JSON = false
	if !p.Skip(demo, info.Size(), info.ModTime(), false) {
		t.Fatal("HTML-only report should be skipped when JSON is disabled")
	}
}

func TestSkipOptionsChangeInvalidates(t *testing.T) {
	withVersion(t, "26.11.1", "abc")
	dir := t.TempDir()
	demo := filepath.Join(dir, "match.dem")
	if err := os.WriteFile(demo, []byte("demo"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(demo)
	cfg := config.Default()
	cfg.Output = dir
	cfg.JSON = true
	p := testProcessorWithConfig(cfg)
	writeArtifacts(t, p, demo, info)
	p.cfg.Cheat = false
	if p.Skip(demo, info.Size(), info.ModTime(), false) {
		t.Fatal("cheat option change must invalidate")
	}
}

func TestSkipCachedFailure(t *testing.T) {
	withVersion(t, "dev", "")
	dir := t.TempDir()
	demo := filepath.Join(dir, "bad.dem")
	if err := os.WriteFile(demo, []byte("demo"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(demo)
	p := testProcessorWithConfig(config.Default())
	p.store.MarkError(demo, info.Size(), info.ModTime(), errors.New("test failure"), "dev")
	if !p.Skip(demo, info.Size(), info.ModTime(), false) {
		t.Fatal("cached failures are skipped even without a version stamp")
	}
	if p.Skip(demo, info.Size(), info.ModTime(), true) {
		t.Fatal("--force retries cached failures")
	}
}

func TestSkipDisabledWithoutStableVersion(t *testing.T) {
	withVersion(t, "dev", "")
	dir := t.TempDir()
	demo := filepath.Join(dir, "match.dem")
	if err := os.WriteFile(demo, []byte("demo"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(demo)
	p := testProcessorWithConfig(config.Default())
	writeArtifacts(t, p, demo, info)
	if p.Skip(demo, info.Size(), info.ModTime(), false) {
		t.Fatal("unversioned dev builds must always rebuild")
	}
}

func TestPlanOverwritesExistingReportOnRegeneration(t *testing.T) {
	out := t.TempDir()
	demo := filepath.Join(out, "match.dem")
	if err := os.WriteFile(demo, []byte("demo"), 0o644); err != nil {
		t.Fatal(err)
	}
	candidate := filepath.Join(out, "match.html")
	if err := os.WriteFile(candidate, []byte("old report"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Output = out
	p := testProcessorWithConfig(cfg)
	if got := p.plan(demo).HTMLPath; got != candidate {
		t.Fatalf("regeneration must overwrite %q, got %q", candidate, got)
	}
}

func TestIndexUsesHeaderMetadataWhenStoreIsEmpty(t *testing.T) {
	withVersion(t, "26.11.1", "abc")
	dir := t.TempDir()
	demo := filepath.Join(dir, "match.dem")
	if err := os.WriteFile(demo, []byte("demo"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(demo)
	cfg := config.Default()
	cfg.Output = dir
	p := testProcessorWithConfig(cfg)
	plan, header := writeArtifacts(t, p, demo, info)
	header.Map = "de_mirage"
	header.ServerName = "Test Server"
	header.Score = [2]int{13, 7}
	header.HasScore = true
	header.PlayedAt = time.Date(2026, 11, 15, 20, 0, 0, 0, time.UTC)
	doc := report.New(&model.Match{SchemaVersion: model.SchemaVersion, Map: "de_mirage", ServerName: "Test Server",
		Summary: &model.Summary{Score: [2]int{13, 7}}}, header)
	var page bytes.Buffer
	if err := html.Render(&page, doc, html.Options{}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan.HTMLPath, page.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, err := p.indexEntries(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %+v", entries)
	}
	if entries[0].Name != "Mirage" || entries[0].ServerName != "Test Server" || entries[0].Score != "13 : 7" {
		t.Fatalf("header metadata not used: %+v", entries[0])
	}
}
