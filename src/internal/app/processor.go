package app

import (
	"bytes"
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
	"github.com/Kandru/cs2-demo-analyzer/internal/report/html"
	"github.com/Kandru/cs2-demo-analyzer/internal/report/jsondata"
	"github.com/Kandru/cs2-demo-analyzer/internal/store"
	"github.com/Kandru/cs2-demo-analyzer/internal/version"
)

// Processor runs the full analysis pipeline for one demo.
type Processor struct {
	cfg     config.Config
	store   *store.Store
	log     *slog.Logger
	indexMu sync.Mutex
}

// NewProcessor creates a processor writing through st.
func NewProcessor(cfg config.Config, st *store.Store, log *slog.Logger) *Processor {
	return &Processor{cfg: cfg, store: st, log: log}
}

// Skip reports whether path was already analyzed successfully (or failed) with
// the same file signature.
func (p *Processor) Skip(path string, size int64, mod time.Time, force bool) bool {
	return p.store.Skip(path, size, mod, version.Short(), p.reportOptions(p.outputPath(path), path), force)
}

// outputPath returns the report path for a demo.
func (p *Processor) outputPath(demoPath string) string {
	return p.store.OutputPath(demoPath, p.cfg.Output, ".html")
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

	htmlPath := p.outputPath(path)
	p.store.ReserveOutput(htmlPath, path)

	view := html.NewView(m, version.Short())
	view.DemoDownload = p.demoLink(htmlPath, path)
	var buf bytes.Buffer
	if err := html.Render(&buf, view); err != nil {
		p.recordError(path, info, err)
		return err
	}

	if err := store.WriteAtomic(htmlPath, func(w io.Writer) error {
		_, err := w.Write(buf.Bytes())
		return err
	}, 0o644); err != nil {
		p.recordError(path, info, err)
		return fmt.Errorf("write report: %w", err)
	}

	jsonPath := ""
	if p.cfg.JSON {
		jsonPath = strings.TrimSuffix(htmlPath, ".html") + ".json"
		if err := store.WriteAtomic(jsonPath, func(w io.Writer) error {
			return jsondata.Write(w, m)
		}, 0o644); err != nil {
			p.recordError(path, info, err)
			return fmt.Errorf("write json: %w", err)
		}
	}

	p.store.MarkOK(path, info.Size(), info.ModTime(), hash, p.reportOptions(htmlPath, path), htmlPath, jsonPath, version.Short())
	if p.indexEnabled(htmlPath, path) {
		if err := p.writeIndex(filepath.Dir(htmlPath)); err != nil {
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
		"report", htmlPath,
		"duration", time.Since(start).Round(time.Millisecond),
	)
	if m.Partial {
		p.log.Warn("report is partial", "demo", filepath.Base(path), "error", m.ParseError)
	}
	return nil
}

func (p *Processor) recordError(path string, info os.FileInfo, err error) {
	if info != nil {
		p.store.MarkError(path, info.Size(), info.ModTime(), err, version.Short())
		if serr := p.store.Save(); serr != nil {
			p.log.Warn("could not save manifest", "error", serr)
		}
	}
	p.log.Error("analysis failed", "demo", path, "error", err)
}
