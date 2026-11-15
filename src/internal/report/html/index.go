package html

import (
	"fmt"
	"html/template"
	"io"
)

// IndexFileName is the report index written into each output directory.
const IndexFileName = "index.html"

// IndexView is the template model for an output-directory index page.
type IndexView struct {
	Title       string
	Version     string
	GeneratedAt string
	IndexFile   string
	CSS         template.CSS
	Entries     []IndexEntry
}

// IndexEntry is one report listed on an index page.
type IndexEntry struct {
	Href     string
	Name     string
	Modified string
	Demo     *DemoLink
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
	v.CSS = template.CSS(styleCSS)
	if err := indexTemplate.ExecuteTemplate(w, "index", v); err != nil {
		return fmt.Errorf("render index: %w", err)
	}
	return nil
}
