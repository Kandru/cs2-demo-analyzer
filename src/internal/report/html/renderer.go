package html

import (
	"embed"
	"fmt"
	"html/template"
	"io"

	"github.com/Kandru/cs2-demo-analyzer/internal/report"
)

//go:embed templates/*.gohtml
var templatesFS embed.FS

//go:embed assets/style.css
var styleCSS string

//go:embed assets/app.js
var appJS string

var reportTemplate = template.Must(
	template.New("report.gohtml").Funcs(FuncMap()).ParseFS(templatesFS, "templates/report.gohtml"),
)

// Options are render-only inputs that are not part of the match data.
type Options struct {
	// DemoDownload links the source demo when it is reachable.
	DemoDownload *DemoLink
	// IndexHref links back to the report index when one is written.
	IndexHref string
}

// Render writes the complete self-contained report to w. The metadata comment
// is written before the template so the artifact is self-describing.
func Render(w io.Writer, doc *report.Document, opts Options) error {
	if doc == nil || doc.Match == nil {
		return fmt.Errorf("report: nil document")
	}
	comment, err := report.HeaderComment(doc.Header)
	if err != nil {
		return err
	}
	v := newView(doc, opts)
	v.CSS = template.CSS(styleCSS)
	v.JS = template.JS(appJS)
	if _, err := io.WriteString(w, "<!doctype html>\n"+comment+"\n"); err != nil {
		return err
	}
	if err := reportTemplate.ExecuteTemplate(w, "report", v); err != nil {
		return fmt.Errorf("render report: %w", err)
	}
	return nil
}
