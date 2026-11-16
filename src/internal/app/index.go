package app

import (
	"bytes"
	"fmt"
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
		GeneratedAt: time.Now().Format("2006.01.02 15:04"),
		Groups:      groupIndexEntries(entries),
		Count:       len(entries),
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
		meta := store.ReportMeta{}
		if demoPath != "" {
			if cached, ok := p.store.Entry(demoPath); ok && cached.Meta != nil {
				meta = *cached.Meta
			}
		}
		playedAt := meta.PlayedAt
		if playedAt.IsZero() && demoPath != "" {
			if demoInfo, err := os.Stat(demoPath); err == nil {
				playedAt = demoInfo.ModTime()
			}
		}
		if playedAt.IsZero() {
			playedAt = info.ModTime()
		}
		displayName := html.MapLabel(meta.Map)
		if displayName == "" {
			displayName = strings.TrimSuffix(name, filepath.Ext(name))
		}
		score := ""
		if meta.HasScore {
			score = fmt.Sprintf("%d : %d", meta.Score[0], meta.Score[1])
		}
		entries = append(entries, html.IndexEntry{
			Href:       name,
			Name:       displayName,
			Score:      score,
			ServerName: strings.TrimSpace(meta.ServerName),
			Played:     playedAt.Format("2006-01-02 15:04"),
			PlayedAt:   playedAt,
			Demo:       p.demoLink(full, demoPath),
		})
	}
	sortIndexEntries(entries)
	return entries, nil
}

func sortIndexEntries(entries []html.IndexEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if !a.PlayedAt.Equal(b.PlayedAt) {
			return a.PlayedAt.After(b.PlayedAt)
		}
		if a.ServerName != b.ServerName {
			return a.ServerName < b.ServerName
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.Href < b.Href
	})
}

// groupIndexEntries groups sorted entries by server, preserving the newest-first
// order. Unknown servers are kept together at the end of the page.
func groupIndexEntries(entries []html.IndexEntry) []html.IndexServerGroup {
	groups := make([]html.IndexServerGroup, 0, 4)
	byName := map[string]int{}
	for _, entry := range entries {
		key := entry.ServerName
		if key == "" {
			key = "Unknown server"
		}
		if i, ok := byName[key]; ok {
			groups[i].Entries = append(groups[i].Entries, entry)
			continue
		}
		byName[key] = len(groups)
		groups = append(groups, html.IndexServerGroup{Name: key, Entries: []html.IndexEntry{entry}})
	}
	unknown := -1
	for i := range groups {
		if groups[i].Name == "Unknown server" {
			unknown = i
			break
		}
	}
	if unknown >= 0 && unknown != len(groups)-1 {
		g := groups[unknown]
		groups = append(groups[:unknown], groups[unknown+1:]...)
		groups = append(groups, g)
	}
	return groups
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
