package stats

import (
	"sort"
)

// This file is NOT part of the upstream demo-anticheat project. It exposes the
// detector's published metrics in a presentation-neutral shape so the
// match report can consume them without duplicating scoring logic.
// Upstream: https://github.com/TimAnthonyAlexander/demo-anticheat
// Commit:   635e8c7a152f3026446463cbe8f0ba2e170ec818

// ReportChannel is one cheat-detection channel with its published values.
type ReportChannel struct {
	ID          string
	Label       string
	Score       float64
	Confidence  float64
	Zone        string
	Raw         float64
	Unit        string
	Samples     int
	Weight      float64
	HasData     bool
	Description string
}

// ReportGrade is a per-category or overall grade.
type ReportGrade struct {
	Category string
	Label    string
	Grade    string
}

// ReportAdjustment is one boost/discount/override applied to the verdict.
type ReportAdjustment struct {
	Label  string
	Detail string
}

// ReportMoment is one timestamped review cue recorded by a collector.
// Values mirror the channel that produced the signal.
type ReportMoment struct {
	Kind      string
	Label     string
	Rank      int
	ValueText string
	Tick      int
	Value     float64
	Unit      string
	Opponent  string
	Weapon    string
	Detail    string
}

// PlayerReport is the full detector verdict for one player.
type PlayerReport struct {
	SteamID64     uint64
	Name          string
	Likelihood    float64
	Flagged       bool
	FlagThreshold float64
	OverallGrade  string
	Grades        []ReportGrade
	Narrative     string
	Channels      []ReportChannel
	Adjustments   []ReportAdjustment
	Moments       []ReportMoment
}

// BuildPlayerReports converts the detector's DemoStats into per-player reports
// keyed by SteamID64. The special SteamID 0 placeholder (demo-wide metrics) is
// skipped.
func BuildPlayerReports(ds *DemoStats) map[uint64]*PlayerReport {
	out := map[uint64]*PlayerReport{}
	if ds == nil {
		return out
	}
	for sid, ps := range ds.Players {
		if sid == 0 || ps == nil {
			continue
		}
		rep := &PlayerReport{
			SteamID64:     sid,
			Name:          ps.Player.Name,
			FlagThreshold: flagThreshold,
		}
		if m, ok := ps.GetMetric(antiCheatCategory, Key("cheat_likelihood")); ok {
			rep.Likelihood = m.FloatValue
		}
		if m, ok := ps.GetMetric(antiCheatCategory, Key("cheater")); ok {
			rep.Flagged = m.StringValue == "Yes"
		}
		if m, ok := ps.GetMetric(Category("rating"), Key("overall")); ok {
			rep.OverallGrade = m.StringValue
		}
		rep.Grades = reportGrades(ps)
		rep.Narrative = buildCheatscoreNarrative(ps)
		rep.Channels = reportChannels(ps)
		rep.Adjustments = reportAdjustments(ps)
		rep.Moments = reportMoments(ps)
		out[sid] = rep
	}
	return out
}

func reportMoments(ps *PlayerStats) []ReportMoment {
	out := make([]ReportMoment, 0, len(ps.Moments))
	for _, m := range ps.Moments {
		out = append(out, ReportMoment{
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
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Tick != out[j].Tick {
			return out[i].Tick < out[j].Tick
		}
		return out[i].Kind < out[j].Kind
	})
	return out
}

func reportGrades(ps *PlayerStats) []ReportGrade {
	specs := []struct {
		category Category
		label    string
	}{
		{Category("kills"), "Combat"},
		{Category("reaction"), "Reaction"},
		{Category("recoil"), "Recoil"},
		{Category("grenade"), "Grenades"},
	}
	out := make([]ReportGrade, 0, len(specs))
	for _, spec := range specs {
		if m, ok := ps.GetMetric(spec.category, Key("grade")); ok && m.StringValue != "" {
			out = append(out, ReportGrade{Category: string(spec.category), Label: spec.label, Grade: m.StringValue})
		}
	}
	return out
}

func reportChannels(ps *PlayerStats) []ReportChannel {
	out := make([]ReportChannel, 0, len(detections))
	for _, det := range detections {
		ch := ReportChannel{
			ID:          det.ID,
			Label:       det.Label,
			Unit:        det.Unit,
			Description: det.Description,
			Weight:      det.Weight,
		}
		if det.Raw != nil {
			ch.Raw, ch.Samples, _ = det.Raw(ps)
		}
		if m, ok := ps.GetMetric(antiCheatCategory, Key(det.ID+"_score")); ok {
			ch.Score = m.FloatValue
			ch.HasData = true
		}
		if m, ok := ps.GetMetric(antiCheatCategory, Key(det.ID+"_confidence")); ok {
			ch.Confidence = m.FloatValue
		}
		if m, ok := ps.GetMetric(antiCheatCategory, Key(det.ID+"_zone")); ok {
			ch.Zone = m.StringValue
		}
		if !ch.HasData {
			ch.Score = 0
			ch.Confidence = 0
			ch.Zone = "no_data"
		}
		out = append(out, ch)
	}
	return out
}

func reportAdjustments(ps *PlayerStats) []ReportAdjustment {
	out := make([]ReportAdjustment, 0, len(modifiers))
	for _, m := range modifiers {
		if m.Adjustments != nil {
			out = append(out, m.Adjustments(ps)...)
		}
	}
	return out
}
