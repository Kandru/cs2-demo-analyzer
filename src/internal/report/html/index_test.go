package html

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderIndexListsReportsAndEscapesNames(t *testing.T) {
	v := &IndexView{
		Title:   "Match reports",
		Version: "test",
		Entries: []IndexEntry{
			{
				Href:     "match.html",
				Name:     `<b>de_mirage</b>`,
				Modified: "2026-11-15 12:00",
				Demo:     &DemoLink{Href: "match.dem", Name: `match"demo.dem`, Size: "12.0 MiB"},
			},
			{Href: "other.html", Name: "other", Modified: "2026-11-14 12:00"},
		},
	}
	var buf bytes.Buffer
	if err := RenderIndex(&buf, v); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	for _, want := range []string{"Match reports", "match.html", "other.html", "download=\"match&#34;demo.dem\"", "12.0 MiB", "Open report"} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in rendered index", want)
		}
	}
	if strings.Contains(page, "<b>de_mirage</b>") {
		t.Fatal("index name was not escaped")
	}
	if !strings.Contains(page, "&lt;b&gt;de_mirage&lt;/b&gt;") {
		t.Fatal("escaped index name missing")
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
