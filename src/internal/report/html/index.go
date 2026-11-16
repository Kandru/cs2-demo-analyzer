package html

import (
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"
	"unicode"
)

// IndexFileName is the report index written into each output directory.
const IndexFileName = "index.html"

// RepoURL is the public project page linked from generated pages.
const RepoURL = "https://github.com/Kandru/cs2-demo-analyzer"

// IndexView is the template model for an output-directory index page.
type IndexView struct {
	Title       string
	Version     string
	GeneratedAt string
	RepoURL     string
	IndexFile   string
	CSS         template.CSS
	Groups      []IndexServerGroup
	Count       int
}

// IndexServerGroup lists the reports recorded for one server.
type IndexServerGroup struct {
	Name    string
	Entries []IndexEntry
}

// IndexEntry is one report listed on an index page.
type IndexEntry struct {
	Href       string
	Name       string
	Score      string
	ServerName string
	Played     string
	PlayedAt   time.Time
	Demo       *DemoLink
}

// MapLabel returns a human-friendly map name such as "Mirage" for "de_mirage".
// Unknown or empty names are returned unchanged.
func MapLabel(name string) string {
	name = strings.TrimSpace(name)
	lower := strings.ToLower(name)
	for _, prefix := range []string{"de_", "cs_", "gd_", "ar_"} {
		if strings.HasPrefix(lower, prefix) {
			name = name[len(prefix):]
			break
		}
	}
	words := strings.Fields(strings.ReplaceAll(name, "_", " "))
	for i, word := range words {
		runes := []rune(word)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

var indexTemplate = template.Must(
	template.New("index.gohtml").ParseFS(templatesFS, "templates/index.gohtml"),
)

// RenderIndex writes the self-contained report index to w.
func RenderIndex(w io.Writer, v *IndexView) error {
	if v == nil {
		return fmt.Errorf("index: nil view")
	}
	v.IndexFile = IndexFileName
	v.RepoURL = RepoURL
	v.CSS = template.CSS(styleCSS)
	if err := indexTemplate.ExecuteTemplate(w, "index", v); err != nil {
		return fmt.Errorf("render index: %w", err)
	}
	return nil
}
