package stats

import (
	"sort"
)

// This file is NOT part of the upstream demo-anticheat project. It exposes the
// detector's published metrics in a presentation-neutral shape so the
// battlelog report can consume them without duplicating scoring logic.
// Upstream: https://github.com/TimAnthonyAlexander/demo-anticheat
// Commit:   635e8c7a152f3026446463cbe8f0ba2e170ec818

// BattlelogChannel is one cheat-detection channel with its published values.
type BattlelogChannel struct {
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

// BattlelogGrade is a per-category or overall grade.
type BattlelogGrade struct {
	Category string
	Label    string
	Grade    string
}

// BattlelogAdjustment is one boost/discount/override applied to the verdict.
type BattlelogAdjustment struct {
	Label  string
	Detail string
}

// BattlelogMoment is one timestamped review cue recorded by a collector.
// Values mirror the channel that produced the signal.
type BattlelogMoment struct {
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

// BattlelogPlayerReport is the full detector verdict for one player.
type BattlelogPlayerReport struct {
	SteamID64     uint64
	Name          string
	Likelihood    float64
	Flagged       bool
	FlagThreshold float64
	OverallGrade  string
	Grades        []BattlelogGrade
	Narrative     string
	Channels      []BattlelogChannel
	Adjustments   []BattlelogAdjustment
	Moments       []BattlelogMoment
}

// BattlelogReports converts the detector's DemoStats into per-player reports
// keyed by SteamID64. The special SteamID 0 placeholder (demo-wide metrics) is
// skipped.
func BattlelogReports(ds *DemoStats) map[uint64]*BattlelogPlayerReport {
	out := map[uint64]*BattlelogPlayerReport{}
	if ds == nil {
		return out
	}
	for sid, ps := range ds.Players {
		if sid == 0 || ps == nil {
			continue
		}
		rep := &BattlelogPlayerReport{
			SteamID64:     sid,
			Name:          ps.Player.Name,
			FlagThreshold: cheatscoreFlagThreshold,
		}
		if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key("cheat_likelihood")); ok {
			rep.Likelihood = m.FloatValue
		}
		if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key("cheater")); ok {
			rep.Flagged = m.StringValue == "Yes"
		}
		if m, ok := ps.GetMetric(Category("rating"), Key("overall")); ok {
			rep.OverallGrade = m.StringValue
		}
		rep.Grades = battlelogGrades(ps)
		rep.Narrative = buildCheatscoreNarrative(ps)
		rep.Channels = battlelogChannels(ps)
		rep.Adjustments = battlelogAdjustments(ps)
		rep.Moments = battlelogMoments(ps)
		out[sid] = rep
	}
	return out
}

func battlelogMoments(ps *PlayerStats) []BattlelogMoment {
	out := make([]BattlelogMoment, 0, len(ps.Moments))
	for _, m := range ps.Moments {
		out = append(out, BattlelogMoment{
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

func battlelogGrades(ps *PlayerStats) []BattlelogGrade {
	specs := []struct {
		category Category
		label    string
	}{
		{Category("kills"), "Combat"},
		{Category("reaction"), "Reaction"},
		{Category("recoil"), "Recoil"},
		{Category("grenade"), "Grenades"},
	}
	out := make([]BattlelogGrade, 0, len(specs))
	for _, spec := range specs {
		if m, ok := ps.GetMetric(spec.category, Key("grade")); ok && m.StringValue != "" {
			out = append(out, BattlelogGrade{Category: string(spec.category), Label: spec.label, Grade: m.StringValue})
		}
	}
	return out
}

func battlelogChannels(ps *PlayerStats) []BattlelogChannel {
	out := make([]BattlelogChannel, 0, len(detections))
	for _, det := range detections {
		ch := BattlelogChannel{
			ID:          det.ID,
			Label:       det.Label,
			Unit:        det.Unit,
			Description: det.Description,
		}
		if det.Raw != nil {
			ch.Raw, ch.Samples, _ = det.Raw(ps)
		}
		if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key(det.ID+"_score")); ok {
			ch.Score = m.FloatValue
			ch.HasData = true
		}
		if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key(det.ID+"_confidence")); ok {
			ch.Confidence = m.FloatValue
		}
		if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key(det.ID+"_zone")); ok {
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

func battlelogAdjustments(ps *PlayerStats) []BattlelogAdjustment {
	out := make([]BattlelogAdjustment, 0, 6)
	for _, m := range modifiers {
		if m.Adjustments != nil {
			out = append(out, m.Adjustments(ps)...)
		}
	}
	return out
}
