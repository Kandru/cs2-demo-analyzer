package html

import (
	"embed"
	"fmt"
	"html/template"
	"io"
)

//go:embed templates/report.gohtml
var templatesFS embed.FS

//go:embed assets/style.css
var styleCSS string

//go:embed assets/app.js
var appJS string

var reportTemplate = template.Must(
	template.New("report.gohtml").Funcs(FuncMap()).ParseFS(templatesFS, "templates/report.gohtml"),
)

// Render writes the complete self-contained report to w.
func Render(w io.Writer, v *View) error {
	if v == nil || v.Match == nil {
		return fmt.Errorf("report: nil view")
	}
	v.CSS = template.CSS(styleCSS)
	v.JS = template.JS(appJS)
	if err := reportTemplate.ExecuteTemplate(w, "report", v); err != nil {
		return fmt.Errorf("render report: %w", err)
	}
	return nil
}
