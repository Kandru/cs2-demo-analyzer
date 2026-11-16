package report

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Kandru/cs2-demo-analyzer/internal/model"
)

func testHeader() Header {
	return Header{
		Format:          HeaderFormat,
		ToolVersion:     "26.11.1",
		SchemaVersion:   model.SchemaVersion,
		Options:         Options{Cheat: true, Aim: true, DemoDownload: true}.String(),
		DemoKey:         DemoKey("/demos/match.dem"),
		DemoName:        "match.dem",
		DemoSize:        1234,
		DemoModUnixNano: 987654321,
		GeneratedAt:     time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC),
		Map:             "de_mirage",
		ServerName:      "server",
		Score:           [2]int{13, 7},
		HasScore:        true,
		PlayedAt:        time.Date(2026, 11, 15, 20, 0, 0, 0, time.UTC),
	}
}

func testDocument() *Document {
	m := &model.Match{
		SchemaVersion: model.SchemaVersion,
		Map:           "de_mirage",
		Players:       []*model.Player{{ID: "1", Name: "kalle"}},
	}
	return New(m, testHeader())
}

func TestWriteJSONHeaderFirstAndRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, testDocument()); err != nil {
		t.Fatal(err)
	}
	section := buf.String()
	if !strings.HasPrefix(strings.TrimSpace(section), `{`) {
		t.Fatalf("not a JSON object: %q", section[:20])
	}
	key := "\"header\""
	if i, j := strings.Index(section, key), strings.Index(section, "\"schemaVersion\""); i < 0 || (j >= 0 && i > j) {
		t.Fatalf("header is not the first key:\n%s", section[:200])
	}

	h, err := ReadHeaderFrom(strings.NewReader(section))
	if err != nil {
		t.Fatal(err)
	}
	if !h.Matches(testHeader()) {
		t.Fatalf("header mismatch: %+v", h)
	}

	doc, err := Decode(strings.NewReader(section))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Match == nil || doc.Map != "de_mirage" || len(doc.Players) != 1 {
		t.Fatalf("round-trip lost match data: %+v", doc.Match)
	}
	if !doc.Header.Matches(testHeader()) {
		t.Fatalf("round-trip lost header data: %+v", doc.Header)
	}
}

func TestReadHeaderLegacyJSONIsStale(t *testing.T) {
	m := &model.Match{SchemaVersion: model.SchemaVersion, Map: "de_mirage"}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ReadHeaderFrom(bytes.NewReader(raw)); err == nil {
		t.Fatal("legacy raw match JSON should not expose a header")
	}
}

func TestReadHTMLHeader(t *testing.T) {
	comment, err := HeaderComment(testHeader())
	if err != nil {
		t.Fatal(err)
	}
	page := "<!doctype html>\n" + comment + "\n<html></html>"
	h, err := ReadHeaderFrom(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	if !h.Matches(testHeader()) {
		t.Fatalf("header mismatch: %+v", h)
	}
	if _, err := ReadHeaderFrom(strings.NewReader("<!doctype html><html></html>")); err == nil {
		t.Fatal("report without metadata must be stale")
	}
}

func TestHeaderCommentEscapesUntrustedDemoName(t *testing.T) {
	h := testHeader()
	h.DemoName = `--><script>alert(1)</script>`
	comment, err := HeaderComment(h)
	if err != nil {
		t.Fatal(err)
	}
	body := strings.TrimSuffix(strings.TrimPrefix(comment, HeaderPrefix), HeaderSuffix)
	if strings.Contains(body, "-->") || strings.Contains(body, "<script>") {
		t.Fatalf("unsafe comment payload: %s", comment)
	}
	if strings.Contains(comment, "\n") {
		t.Fatal("comment must stay on one line")
	}
}

func TestMatchesIgnoresPresentationFields(t *testing.T) {
	want := testHeader()
	got := want
	got.GeneratedAt = want.GeneratedAt.Add(time.Hour)
	got.Map = "de_dust2"
	got.ServerName = "other"
	got.Score = [2]int{0, 0}
	got.PlayedAt = time.Time{}
	if !got.Matches(want) {
		t.Fatal("presentation fields must not affect freshness")
	}
	for name, mutate := range map[string]func(*Header){
		"format":       func(h *Header) { h.Format = HeaderFormat + 1 },
		"tool":         func(h *Header) { h.ToolVersion = "older" },
		"schema":       func(h *Header) { h.SchemaVersion++ },
		"options":      func(h *Header) { h.Options = "cheat=false" },
		"demo key":     func(h *Header) { h.DemoKey = "other" },
		"demo size":    func(h *Header) { h.DemoSize++ },
		"demo modtime": func(h *Header) { h.DemoModUnixNano++ },
	} {
		changed := want
		mutate(&changed)
		if changed.Matches(want) {
			t.Errorf("%s change must invalidate", name)
		}
	}
}

func TestFresh(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "match.html")
	comment, err := HeaderComment(testHeader())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("<!doctype html>\n"+comment+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !Fresh(path, testHeader()) {
		t.Fatal("expected fresh report")
	}
	want := testHeader()
	want.ToolVersion = "26.12.1"
	if Fresh(path, want) {
		t.Fatal("version change must invalidate")
	}
	if Fresh(filepath.Join(dir, "missing.html"), testHeader()) {
		t.Fatal("missing report must not be fresh")
	}
}

func TestDemoKeyIsStableAndDistinct(t *testing.T) {
	a := DemoKey("/demos/a.dem")
	if a != DemoKey("/demos/a.dem") {
		t.Fatal("demo key must be stable")
	}
	if a == DemoKey("/demos/b.dem") {
		t.Fatal("different demos must have different keys")
	}
}

func TestOptionsString(t *testing.T) {
	got := Options{Cheat: true, Aim: false, DemoDownload: true}.String()
	want := "cheat=true;aim=false;demo-download=true"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
