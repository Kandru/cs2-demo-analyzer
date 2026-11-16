package store

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOutputPathDefaultsNextToDemo(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	demo := filepath.Join(dir, "match.dem")
	got := s.OutputPath(demo, "", ".html")
	want := filepath.Join(dir, "match.html")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	got = s.OutputPath(filepath.Join(dir, "match.dem.gz"), "", ".json")
	if got != filepath.Join(dir, "match.json") {
		t.Fatalf("compressed demo: %q", got)
	}
}

func TestOutputPathCollision(t *testing.T) {
	out := t.TempDir()
	s, _ := Open(out)
	a := filepath.Join(out, "a", "match.dem")
	b := filepath.Join(out, "b", "match.dem")
	pa := s.OutputPath(a, out, ".html")
	s.MarkOK(a, 1, time.Unix(1, 0), "", "", pa, "", "v", ReportMeta{})
	pb := s.OutputPath(b, out, ".html")
	if pa == pb {
		t.Fatalf("collision not disambiguated: %q", pa)
	}
	if !strings.Contains(pb, "match-") {
		t.Fatalf("unexpected collision name: %q", pb)
	}
}

func TestSkip(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir)
	demo := filepath.Join(dir, "m.dem")
	if err := os.WriteFile(demo, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	html := filepath.Join(dir, "m.html")
	if err := os.WriteFile(html, []byte("report"), 0o644); err != nil {
		t.Fatal(err)
	}
	mod := time.Unix(100, 0)
	s.MarkOK(demo, 1, mod, "h", "", html, "", "1.0", ReportMeta{})
	if !s.Skip(demo, 1, mod, "1.0", "", false) {
		t.Fatal("expected skip for unchanged successful demo")
	}
	if s.Skip(demo, 2, mod, "1.0", "", false) {
		t.Fatal("size change must invalidate")
	}
	if s.Skip(demo, 1, mod, "1.0", "", true) {
		t.Fatal("force must bypass skip")
	}
	if s.Skip(demo, 1, mod, "2.0", "", false) {
		t.Fatal("tool version change must invalidate")
	}
	if s.Skip(demo, 1, mod, "1.0", "demo-download", false) {
		t.Fatal("report option change must invalidate")
	}
}

func TestSkipCachesFailures(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir)
	mod := time.Unix(100, 0)
	s.MarkError(filepath.Join(dir, "bad.dem"), 10, mod, os.ErrInvalid, "1.0")
	if !s.Skip(filepath.Join(dir, "bad.dem"), 10, mod, "1.0", "", false) {
		t.Fatal("unchanged failed demos should not be retried on every scan")
	}
	if s.Skip(filepath.Join(dir, "bad.dem"), 11, mod, "1.0", "", false) {
		t.Fatal("changed failed demos should be retried")
	}
}

func TestReportMetaPersists(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	demo := filepath.Join(dir, "m.dem")
	report := filepath.Join(dir, "m.html")
	played := time.Date(2026, 11, 16, 12, 11, 0, 0, time.UTC)
	s.MarkOK(demo, 1, played, "h", "", report, "", "1.0", ReportMeta{
		Map: "de_mirage", ServerName: "Test Server", Score: [2]int{13, 7}, HasScore: true, PlayedAt: played,
	})
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	entry, ok := reopened.Entry(demo)
	if !ok || entry.Meta == nil {
		t.Fatalf("metadata missing after reopen: %+v", entry)
	}
	if entry.Meta.Map != "de_mirage" || entry.Meta.ServerName != "Test Server" {
		t.Fatalf("metadata = %+v", entry.Meta)
	}
	if entry.Meta.Score != [2]int{13, 7} || !entry.Meta.HasScore {
		t.Fatalf("score = %+v", entry.Meta)
	}
	if !entry.Meta.PlayedAt.Equal(played) {
		t.Fatalf("played = %v, want %v", entry.Meta.PlayedAt, played)
	}
}

func TestWriteAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "out.html")
	err := WriteAtomic(path, func(w io.Writer) error {
		_, err := w.Write([]byte("hello"))
		return err
	}, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "hello" {
		t.Fatalf("read: %q %v", data, err)
	}
}
