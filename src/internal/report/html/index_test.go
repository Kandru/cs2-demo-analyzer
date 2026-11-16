package html

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestMapLabel(t *testing.T) {
	cases := map[string]string{
		"de_mirage":    "Mirage",
		"de_dust2":     "Dust2",
		"cs_office":    "Office",
		"de_nuke":      "Nuke",
		"workshop_map": "Workshop Map",
		"":             "",
	}
	for in, want := range cases {
		if got := MapLabel(in); got != want {
			t.Errorf("MapLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRenderIndexListsReportsAndEscapesNames(t *testing.T) {
	v := &IndexView{
		Title:       "Match reports",
		Version:     "test",
		GeneratedAt: "2026.11.16 12:11",
		Count:       2,
		Groups: []IndexServerGroup{
			{
				Name: "Alpha Server",
				Entries: []IndexEntry{
					{
						Href:       "match.html",
						Name:       `<b>de_mirage</b>`,
						Score:      "13 : 7",
						ServerName: `<script>server</script>`,
						Played:     "2026-11-15 12:00",
						Demo:       &DemoLink{Href: "match.dem", Name: `match"demo.dem`, Size: "12.0 MiB"},
					},
				},
			},
			{
				Name: "Bravo Server",
				Entries: []IndexEntry{
					{Href: "other.html", Name: "other", Played: "2026-11-14 12:00"},
				},
			},
		},
	}
	var buf bytes.Buffer
	if err := RenderIndex(&buf, v); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	for _, want := range []string{
		"Match reports", "match.html", "other.html", `download="match&#34;demo.dem"`, "12.0 MiB",
		"13 : 7", "Alpha Server", "Bravo Server", "2026-11-15 12:00",
		`href="https://github.com/Kandru/cs2-demo-analyzer"`, "CS2 Demo Analyzer", "vtest",
		"generated 2026.11.16 12:11",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in rendered index", want)
		}
	}
	for _, unwanted := range []string{"Self-contained index", "no external assets", "Generated 2026"} {
		if strings.Contains(page, unwanted) {
			t.Errorf("removed text %q still present", unwanted)
		}
	}
	if strings.Contains(page, "<b>de_mirage</b>") || strings.Contains(page, "<script>server</script>") {
		t.Fatal("index content was not escaped")
	}
	if !strings.Contains(page, "&lt;b&gt;de_mirage&lt;/b&gt;") {
		t.Fatal("escaped index name missing")
	}
	if !strings.Contains(page, "&lt;script&gt;server&lt;/script&gt;") {
		t.Fatal("escaped server name missing")
	}
	if strings.Contains(page, "{{") {
		t.Fatal("unrendered template action")
	}
}

func TestRenderIndexEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderIndex(&buf, &IndexView{Title: "Match reports", Version: "test"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No match reports") {
		t.Fatal("empty index should explain that no reports exist")
	}
}

func TestIndexEntryPlayedAtNotRendered(t *testing.T) {
	v := &IndexView{
		Version: "test",
		Count:   1,
		Groups: []IndexServerGroup{{
			Name: "Server",
			Entries: []IndexEntry{{
				Href:     "match.html",
				Name:     "Mirage",
				Played:   "2026-11-15 12:00",
				PlayedAt: time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC),
			}},
		}},
	}
	var buf bytes.Buffer
	if err := RenderIndex(&buf, v); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "0001-01-01") {
		t.Fatal("sort-only PlayedAt leaked into the page")
	}
}
