// Package anticheat adapts the vendored demo-anticheat detector to the
// battlelog report model.
//
// The detector code lives in internal/anticheat/{analyzer,stats} and is a
// vendored copy of https://github.com/TimAnthonyAlexander/demo-anticheat at
// commit 635e8c7a152f3026446463cbe8f0ba2e170ec818. See docs/THIRD_PARTY.md.
package anticheat

import (
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Kandru/cs2-demo-analyzer/internal/anticheat/analyzer"
	"github.com/Kandru/cs2-demo-analyzer/internal/anticheat/stats"
	"github.com/Kandru/cs2-demo-analyzer/internal/model"
)

// Report is the result of one detector run.
type Report struct {
	Players  map[uint64]*model.CheatReport
	Duration time.Duration
	Err      error
}

// Analyze runs the vendored detector against one demo. A failure is returned
// in Report.Err and is never fatal for the battlelog report.
func Analyze(path string, logger *slog.Logger) (rep Report) {
	start := time.Now()
	rep.Players = map[uint64]*model.CheatReport{}
	defer func() {
		rep.Duration = time.Since(start)
		if r := recover(); r != nil {
			rep.Err = fmt.Errorf("anti-cheat detector panicked: %v", r)
			if logger != nil {
				logger.Warn("anti-cheat detector panicked", "demo", path, "panic", r)
			}
		}
	}()

	an := analyzer.NewAnalyzer(path)
	rd, closeFn, oerr := openDemo(path)
	if oerr != nil {
		rep.Err = fmt.Errorf("open demo for anti-cheat: %w", oerr)
		if logger != nil {
			logger.Warn("anti-cheat could not open demo", "demo", path, "error", oerr)
		}
		return rep
	}
	if closeFn != nil {
		defer closeFn()
	}
	demoStats, err := an.AnalyzeReader(rd)
	if err != nil {
		rep.Err = fmt.Errorf("anti-cheat analysis failed: %w", err)
		if logger != nil {
			logger.Warn("anti-cheat analysis failed", "demo", path, "error", err)
		}
		return rep
	}

	vendored := stats.BattlelogReports(demoStats)
	for sid, vr := range vendored {
		cr := &model.CheatReport{
			Available:     true,
			Likelihood:    vr.Likelihood,
			Flagged:       vr.Flagged,
			FlagThreshold: vr.FlagThreshold,
			OverallGrade:  vr.OverallGrade,
			Narrative:     vr.Narrative,
		}
		for _, g := range vr.Grades {
			cr.Grades = append(cr.Grades, model.Grade{Category: g.Category, Label: g.Label, Grade: g.Grade})
		}
		for _, c := range vr.Channels {
			cr.Channels = append(cr.Channels, model.CheatChannel{
				ID:          c.ID,
				Label:       c.Label,
				Score:       c.Score,
				Confidence:  c.Confidence,
				Zone:        c.Zone,
				Raw:         c.Raw,
				Unit:        c.Unit,
				Samples:     c.Samples,
				Weight:      c.Weight,
				HasData:     c.HasData,
				Description: c.Description,
			})
		}
		for _, a := range vr.Adjustments {
			cr.Adjustments = append(cr.Adjustments, model.Adjustment{Label: a.Label, Detail: a.Detail})
		}
		for _, m := range vr.Moments {
			cr.Moments = append(cr.Moments, model.CheatMoment{
				Kind:      m.Kind,
				Label:     m.Label,
				Rank:      m.Rank,
				ValueText: m.ValueText,
				Tick:      m.Tick,
				Value:     m.Value,
				Unit:      m.Unit,
				Opponent:  m.Opponent,
				Weapon:    m.Weapon,
				Detail:    m.Detail,
			})
		}
		rep.Players[sid] = cr
	}
	return rep
}

func openDemo(path string) (io.Reader, func() error, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	name := strings.ToLower(path)
	var magic [3]byte
	n, _ := io.ReadFull(f, magic[:])
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		f.Close()
		return nil, nil, err
	}
	if n >= 2 && magic[0] == 0x1f && magic[1] == 0x8b {
		name += ".gz"
	}
	if n >= 3 && magic[0] == 'B' && magic[1] == 'Z' && magic[2] == 'h' {
		name += ".bz2"
	}
	switch {
	case strings.HasSuffix(name, ".gz"):
		zr, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, nil, err
		}
		return zr, func() error { err := zr.Close(); _ = f.Close(); return err }, nil
	case strings.HasSuffix(name, ".bz2"):
		return bzip2.NewReader(f), f.Close, nil
	default:
		return f, f.Close, nil
	}
}
