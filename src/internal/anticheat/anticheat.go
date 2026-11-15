// Package anticheat adapts the vendored demo-anticheat detector to the
// battlelog report model.
//
// The detector code lives in internal/anticheat/{analyzer,stats} and is a
// vendored copy of https://github.com/TimAnthonyAlexander/demo-anticheat at
// commit 635e8c7a152f3026446463cbe8f0ba2e170ec818. See docs/THIRD_PARTY.md.
package anticheat

import (
	"fmt"
	"log/slog"
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

	results, err := analyzer.NewAnalyzer(path).Analyze()
	if err != nil {
		rep.Err = fmt.Errorf("anti-cheat analysis failed: %w", err)
		if logger != nil {
			logger.Warn("anti-cheat analysis failed", "demo", path, "error", err)
		}
		return rep
	}

	vendored := stats.BattlelogReports(results.DemoStats)
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
		rep.Players[sid] = cr
	}
	return rep
}
