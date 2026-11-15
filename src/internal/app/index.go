package app

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Kandru/cs2-demo-analyzer/internal/demofile"
	"github.com/Kandru/cs2-demo-analyzer/internal/report/html"
	"github.com/Kandru/cs2-demo-analyzer/internal/store"
	"github.com/Kandru/cs2-demo-analyzer/internal/units"
	"github.com/Kandru/cs2-demo-analyzer/internal/version"
)

const reportOptionDemoDownload = "demo-download"

// RefreshIndexes rewrites index.html for cached reports whose output directory
// needs one. It runs on watch startup before new demos are processed.
func (p *Processor) RefreshIndexes() {
	dirs := map[string]struct{}{}
	for _, e := range p.store.Entries() {
		if e.HTMLPath == "" || !p.indexEnabled(e.HTMLPath, e.Path) {
			continue
		}
		dirs[filepath.Dir(e.HTMLPath)] = struct{}{}
	}
	ordered := make([]string, 0, len(dirs))
	for dir := range dirs {
		ordered = append(ordered, dir)
	}
	sort.Strings(ordered)
	for _, dir := range ordered {
		if err := p.writeIndex(dir); err != nil {
			p.log.Warn("could not write index", "dir", dir, "error", err)
		}
	}
}

// reportOptions identifies rendered report features outside the data schema.
// The watch cache uses it to regenerate reports when flags change.
func (p *Processor) reportOptions(htmlPath, demoPath string) string {
	if p.demoLinkEnabled(htmlPath, demoPath) {
		return reportOptionDemoDownload
	}
	return ""
}

// indexEnabled reports whether a directory should receive an index page. It is
// on explicitly via --index or when reports are written next to their demos.
func (p *Processor) indexEnabled(htmlPath, demoPath string) bool {
	return p.cfg.Index || sameDir(htmlPath, demoPath)
}

// demoLinkEnabled reports whether a report may link its source demo. Linking is
// always safe in the same directory; --demo-download opts in across directories.
func (p *Processor) demoLinkEnabled(htmlPath, demoPath string) bool {
	return p.cfg.DemoLinks || sameDir(htmlPath, demoPath)
}

// demoLink builds the download link for a demo relative to a report.
func (p *Processor) demoLink(htmlPath, demoPath string) *html.DemoLink {
	if demoPath == "" || !p.demoLinkEnabled(htmlPath, demoPath) {
		return nil
	}
	rel, err := filepath.Rel(filepath.Dir(htmlPath), demoPath)
	if err != nil || rel == "" {
		return nil
	}
	info, err := os.Stat(demoPath)
	if err != nil {
		return nil
	}
	return &html.DemoLink{
		Href: filepath.ToSlash(rel),
		Name: filepath.Base(demoPath),
		Size: units.Bytes(info.Size()),
	}
}

func (p *Processor) writeIndex(dir string) error {
	p.indexMu.Lock()
	defer p.indexMu.Unlock()

	entries, err := p.indexEntries(dir)
	if err != nil {
		return err
	}
	view := &html.IndexView{
		Title:       "Match reports",
		Version:     version.Short(),
		GeneratedAt: time.Now().Format("2006-01-02 15:04 MST"),
		Entries:     entries,
	}
	var buf bytes.Buffer
	if err := html.RenderIndex(&buf, view); err != nil {
		return err
	}
	return store.WriteAtomic(filepath.Join(dir, html.IndexFileName), func(w io.Writer) error {
		_, err := w.Write(buf.Bytes())
		return err
	}, 0o644)
}

func (p *Processor) indexEntries(dir string) ([]html.IndexEntry, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	entries := make([]html.IndexEntry, 0, len(files))
	for _, f := range files {
		name := f.Name()
		if f.IsDir() || strings.EqualFold(name, html.IndexFileName) || !strings.EqualFold(filepath.Ext(name), ".html") {
			continue
		}
		info, err := f.Info()
		if err != nil {
			continue
		}
		full := filepath.Join(dir, name)
		demoPath, ok := p.store.DemoForReport(full)
		if !ok {
			demoPath = demofile.FindSibling(full)
		}
		entries = append(entries, html.IndexEntry{
			Href:     name,
			Name:     strings.TrimSuffix(name, filepath.Ext(name)),
			Modified: info.ModTime().Format("2006-01-02 15:04"),
			Demo:     p.demoLink(full, demoPath),
		})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Modified != entries[j].Modified {
			return entries[i].Modified > entries[j].Modified
		}
		return entries[i].Href < entries[j].Href
	})
	return entries, nil
}

func sameDir(a, b string) bool {
	dirA, err := filepath.Abs(filepath.Dir(a))
	if err != nil {
		return false
	}
	dirB, err := filepath.Abs(filepath.Dir(b))
	if err != nil {
		return false
	}
	return dirA == dirB
}
