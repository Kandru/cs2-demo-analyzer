package store

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPlanOutputDefaultsNextToDemo(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	demo := filepath.Join(dir, "match.dem")
	got := s.PlanOutput(demo, "", ".html", nil)
	want := filepath.Join(dir, "match.html")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if again := s.PlanOutput(demo, "", ".html", nil); again != want {
		t.Fatalf("replanning changed the path: %q", again)
	}
	other := s.PlanOutput(filepath.Join(dir, "other.dem"), "", ".json", nil)
	if other != filepath.Join(dir, "other.json") {
		t.Fatalf("json plan: %q", other)
	}
}

func TestPlanOutputNeverClaimsIndexName(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir)
	demo := filepath.Join(dir, "index.dem")
	got := s.PlanOutput(demo, "", ".html", nil)
	if strings.EqualFold(filepath.Base(got), "index.html") {
		t.Fatalf("report must not overwrite the index: %q", got)
	}
	if !strings.Contains(filepath.Base(got), "index-") {
		t.Fatalf("unexpected fallback name: %q", got)
	}
}

func TestPlanOutputCollision(t *testing.T) {
	out := t.TempDir()
	s, _ := Open(out)
	a := filepath.Join(out, "a", "match.dem")
	b := filepath.Join(out, "b", "match.dem")
	pa := s.PlanOutput(a, out, ".html", nil)
	s.MarkOK(a, 1, time.Unix(1, 0), "", "", pa, "", "v", ReportMeta{})
	pb := s.PlanOutput(b, out, ".html", nil)
	if pa == pb {
		t.Fatalf("collision not disambiguated: %q", pa)
	}
	if !strings.Contains(pb, "match-") {
		t.Fatalf("unexpected collision name: %q", pb)
	}
}

func TestPlanOutputOccupiedCallback(t *testing.T) {
	out := t.TempDir()
	s, _ := Open(out)
	a := filepath.Join(out, "a", "match.dem")
	b := filepath.Join(out, "b", "match.dem")
	pa := s.PlanOutput(a, out, ".html", nil)
	pb := s.PlanOutput(b, out, ".html", func(path string) bool { return path == pa })
	if pa == pb {
		t.Fatalf("occupied candidate was reused: %q", pa)
	}
	if !strings.Contains(pb, "match-") {
		t.Fatalf("unexpected collision name: %q", pb)
	}
}

func TestCachedFailure(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir)
	demo := filepath.Join(dir, "bad.dem")
	mod := time.Unix(100, 0)
	s.MarkError(demo, 10, mod, os.ErrInvalid, "1.0")
	if !s.CachedFailure(demo, 10, mod) {
		t.Fatal("unchanged failed demos should not be retried on every scan")
	}
	if s.CachedFailure(demo, 11, mod) {
		t.Fatal("changed failed demos should be retried")
	}
	if s.CachedFailure(filepath.Join(dir, "good.dem"), 10, mod) {
		t.Fatal("unknown demos are not cached failures")
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
	if !ok {
		t.Fatalf("missing entry after reopen: %+v", entry)
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
