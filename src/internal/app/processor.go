package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Kandru/cs2-demo-analyzer/internal/analysis"
	"github.com/Kandru/cs2-demo-analyzer/internal/anticheat"
	"github.com/Kandru/cs2-demo-analyzer/internal/config"
	"github.com/Kandru/cs2-demo-analyzer/internal/model"
	"github.com/Kandru/cs2-demo-analyzer/internal/parser"
	"github.com/Kandru/cs2-demo-analyzer/internal/report"
	"github.com/Kandru/cs2-demo-analyzer/internal/report/html"
	"github.com/Kandru/cs2-demo-analyzer/internal/store"
	"github.com/Kandru/cs2-demo-analyzer/internal/version"
)

// outputPlan is the reserved output layout for one demo.
type outputPlan struct {
	HTMLPath string
	JSONPath string
}

// Processor runs the full analysis pipeline for one demo.
type Processor struct {
	cfg     config.Config
	store   *store.Store
	log     *slog.Logger
	indexMu sync.Mutex

	planMu sync.Mutex
	plans  map[string]outputPlan

	// deferIndex keeps per-demo Process calls from writing index.html; analyze
	// refreshes every directory once after all workers finish.
	deferIndex bool
}

// SetDeferredIndex controls whether Process refreshes the directory index.
// Analyze defers index writes to one pass at the end; watch refreshes after
// each successful analysis.
func (p *Processor) SetDeferredIndex(deferIndex bool) {
	p.deferIndex = deferIndex
}

// NewProcessor creates a processor writing through st.
func NewProcessor(cfg config.Config, st *store.Store, log *slog.Logger) *Processor {
	return &Processor{cfg: cfg, store: st, log: log, plans: map[string]outputPlan{}}
}

// Plan reserves report output paths for a demo. Analyze calls it for every
// discovered demo before workers start so shared-output basename collisions
// are resolved deterministically, including for demos that are then skipped.
func (p *Processor) Plan(demoPath string) {
	p.plan(demoPath)
}

func (p *Processor) plan(demoPath string) outputPlan {
	p.planMu.Lock()
	defer p.planMu.Unlock()
	if plan, ok := p.plans[demoPath]; ok {
		return plan
	}
	htmlPath := p.store.PlanOutput(demoPath, p.cfg.Output, ".html")
	plan := outputPlan{
		HTMLPath: htmlPath,
		JSONPath: strings.TrimSuffix(htmlPath, ".html") + ".json",
	}
	p.plans[demoPath] = plan
	return plan
}

// Skip reports whether path was already analyzed successfully with matching
// artifacts, or is a cached failure. Artifacts carry their own version header,
// so a missing manifest does not force a re-parse.
func (p *Processor) Skip(path string, size int64, mod time.Time, force bool) bool {
	if force {
		return false
	}
	if p.store.CachedFailure(path, size, mod) {
		return true
	}
	if version.CacheKey() == "" {
		return false
	}
	plan := p.plan(path)
	want := p.headerFor(path, size, mod, plan)
	if !report.Fresh(plan.HTMLPath, want) {
		return false
	}
	return !p.cfg.JSON || report.Fresh(plan.JSONPath, want)
}

// Save flushes the manifest.
func (p *Processor) Save() error { return p.store.Save() }

// Process parses one demo, runs anti-cheat, and writes the HTML/JSON reports.
func (p *Processor) Process(ctx context.Context, path string) error {
	start := time.Now()
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat demo: %w", err)
	}
	plan := p.plan(path)
	ctx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	hash, herr := parser.HashFile(path)
	if herr != nil {
		p.log.Warn("could not hash demo", "demo", path, "error", herr)
	}

	m, err := parser.Parse(ctx, path, parser.Options{Logger: p.log, SkipAim: !p.cfg.Aim})
	if err != nil {
		p.recordError(path, info, err)
		return err
	}
	if m == nil {
		err = fmt.Errorf("parser returned no match")
		p.recordError(path, info, err)
		return err
	}
	m.DemoHash = hash

	cheat := anticheat.Report{Players: map[uint64]*model.CheatReport{}}
	if p.cfg.Cheat {
		cheat = anticheat.Analyze(path, p.log)
	} else {
		cheat.Err = fmt.Errorf("anti-cheat analysis disabled with --cheat=false")
	}
	analysis.Apply(m, analysis.Options{ToolVersion: version.Short(), Cheat: cheat})
	analysis.Validate(m)

	header := p.headerFor(path, info.Size(), info.ModTime(), plan)
	fillMatchMeta(&header, m, info.ModTime())
	doc := report.New(m, header)

	htmlOpts := html.Options{DemoDownload: p.demoLink(plan.HTMLPath, path)}
	indexed := p.indexEnabled(plan.HTMLPath, path)
	if indexed {
		htmlOpts.IndexHref = html.IndexFileName
	}
	if err := store.WriteAtomic(plan.HTMLPath, func(w io.Writer) error {
		return html.Render(w, doc, htmlOpts)
	}, 0o644); err != nil {
		p.recordError(path, info, err)
		return fmt.Errorf("write report: %w", err)
	}

	jsonPath := ""
	if p.cfg.JSON {
		jsonPath = plan.JSONPath
		if err := store.WriteAtomic(jsonPath, func(w io.Writer) error {
			return report.WriteJSON(w, doc)
		}, 0o644); err != nil {
			p.recordError(path, info, err)
			return fmt.Errorf("write json: %w", err)
		}
	}

	p.store.MarkOK(path, info.Size(), info.ModTime(), hash, p.reportOptions(plan.HTMLPath, path).String(),
		plan.HTMLPath, jsonPath, reportVersion(), reportMeta(header))
	if indexed && !p.deferIndex {
		if err := p.writeIndex(filepath.Dir(plan.HTMLPath)); err != nil {
			p.log.Warn("could not write index", "demo", path, "error", err)
		}
	}
	if err := p.store.Save(); err != nil {
		p.log.Warn("could not save manifest", "error", err)
	}

	flags := 0
	if m.Summary != nil {
		flags = m.Summary.FlaggedPlayers
	}
	p.log.Info("analyzed demo",
		"demo", filepath.Base(path),
		"map", m.Map,
		"rounds", len(m.Rounds),
		"players", len(m.Players),
		"flagged", flags,
		"report", plan.HTMLPath,
		"duration", time.Since(start).Round(time.Millisecond),
	)
	if m.Partial {
		p.log.Warn("report is partial", "demo", filepath.Base(path), "error", m.ParseError)
	}
	return nil
}

// reportVersion is the version stamp written into artifacts. When no stable
// cache key exists (an unversioned dev build), the display version is written
// but Skip refuses to reuse anything.
func reportVersion() string {
	if key := version.CacheKey(); key != "" {
		return key
	}
	return version.Short()
}

func (p *Processor) headerFor(demoPath string, size int64, mod time.Time, plan outputPlan) report.Header {
	return report.Header{
		Format:          report.HeaderFormat,
		ToolVersion:     reportVersion(),
		SchemaVersion:   model.SchemaVersion,
		Options:         p.reportOptions(plan.HTMLPath, demoPath).String(),
		DemoSize:        size,
		DemoModUnixNano: mod.UnixNano(),
		GeneratedAt:     time.Now().UTC(),
	}
}

func fillMatchMeta(h *report.Header, m *model.Match, fallback time.Time) {
	h.Map = m.Map
	h.ServerName = m.ServerName
	h.PlayedAt = m.FileTime
	if h.PlayedAt.IsZero() {
		h.PlayedAt = fallback
	}
	if m.Summary != nil {
		h.Score = m.Summary.Score
		h.HasScore = true
	}
}

func reportMeta(h report.Header) store.ReportMeta {
	return store.ReportMeta{
		Map:        h.Map,
		ServerName: h.ServerName,
		Score:      h.Score,
		HasScore:   h.HasScore,
		PlayedAt:   h.PlayedAt,
	}
}

func (p *Processor) recordError(path string, info os.FileInfo, err error) {
	if info != nil {
		p.store.MarkError(path, info.Size(), info.ModTime(), err, reportVersion())
		if serr := p.store.Save(); serr != nil {
			p.log.Warn("could not save manifest", "error", serr)
		}
	}
	p.log.Error("analysis failed", "demo", path, "error", err)
}
