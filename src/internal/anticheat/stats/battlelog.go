package stats

import (
	"fmt"
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
}

const battlelogFlagThreshold = cheatscoreFlagThreshold

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
			FlagThreshold: battlelogFlagThreshold,
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
		out[sid] = rep
	}
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
	n := func(cat Category, key Key) int {
		v, _ := psGetInt(ps, cat, key)
		return int(v)
	}
	f := func(cat Category, key Key) float64 {
		v, _ := psGetFloat(ps, cat, key)
		return v
	}

	type spec struct {
		id, label, unit, desc string
		raw                   float64
		samples               int
	}
	// Raw values are read from the same keys the channel evaluators use, so
	// the UI can show the human-readable metric behind every score.
	specs := []spec{
		{"hs", "Headshot rate", "%", "Share of kills that were headshots.", f(Category("kills"), Key("headshot_percentage")), n(Category("kills"), Key("total_kills"))},
		{"snap", "Snap velocity P95", "°/ms", "95th percentile view-angle snap speed.", f(Category("aiming"), Key("p95_snap_velocity")), n(Category("aiming"), Key("snap_count"))},
		{"reaction", "Damage speed (TTD)", "ms", "Median time from visibility to first damage.", f(Category("reaction"), Key("median_ttd")), n(Category("reaction"), Key("ttd_samples"))},
		{"ttd_sub100", "Sub-100ms damage", "%", "Share of engagements completed in under 100 ms.", f(Category("reaction"), Key("sub_100ms_ttd")), n(Category("reaction"), Key("ttd_samples"))},
		{"recoil", "Recoil deviation", "°", "Mean angular error against known spray patterns.", f(Category("recoil"), Key("mean_angular_error")), n(Category("recoil"), Key("total_counted_bullets"))},
		{"pre_fov", "Pre-FOV pre-aim", "°", "Median crosshair-to-enemy angle 200 ms before visibility.", f(Category("behavioral"), Key("pre_fov_aim_median_deg")), n(Category("behavioral"), Key("pre_fov_aim_samples"))},
		{"pre_fov_presence", "Pre-aim sample asymmetry", "samples", "Pre-aim samples relative to the rest of the lobby.", float64(n(Category("behavioral"), Key("pre_fov_aim_samples"))), n(Category("behavioral"), Key("pre_fov_aim_samples"))},
		{"attention", "Off-engagement attention", "°", "Median crosshair-to-nearest-enemy angle outside fights.", f(Category("behavioral"), Key("nearest_enemy_angle_median_deg")), n(Category("behavioral"), Key("nearest_enemy_angle_samples"))},
		{"back_killed", "Back-killed rate", "%", "Share of deaths where the player was looking away from the killer.", f(Category("behavioral"), Key("back_killed_pct")), n(Category("behavioral"), Key("back_killed_total_deaths"))},
		{"decoupling", "Fight/idle decoupling", "°", "Attention angle minus pre-FOV angle.", f(Category("behavioral"), Key("nearest_enemy_angle_median_deg")) - f(Category("behavioral"), Key("pre_fov_aim_median_deg")), n(Category("behavioral"), Key("pre_fov_aim_samples"))},
	}

	out := make([]BattlelogChannel, 0, len(specs))
	for _, s := range specs {
		ch := BattlelogChannel{
			ID:          s.id,
			Label:       s.label,
			Unit:        s.unit,
			Description: s.desc,
			Raw:         s.raw,
			Samples:     s.samples,
		}
		if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key(s.id+"_score")); ok {
			ch.Score = m.FloatValue
			ch.HasData = true
		} else if s.id == "hs" {
			if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key("hs_score")); ok {
				ch.Score = m.FloatValue
				ch.HasData = true
			}
		} else if s.id == "reaction" {
			if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key("reaction_score")); ok {
				ch.Score = m.FloatValue
				ch.HasData = true
			}
		} else if s.id == "recoil" {
			if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key("recoil_score")); ok {
				ch.Score = m.FloatValue
				ch.HasData = true
			}
		}
		if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key(s.id+"_confidence")); ok {
			ch.Confidence = m.FloatValue
		}
		if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key(s.id+"_zone")); ok {
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
	if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key("wingman_boost")); ok && m.StringValue == "Yes" {
		detail := ""
		if r, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key("wingman_kpr_boost_reason")); ok {
			detail = r.StringValue
		}
		out = append(out, BattlelogAdjustment{Label: "Wingman score boost", Detail: detail})
	}
	if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key("competitive_boost")); ok && m.StringValue == "Yes" {
		out = append(out, BattlelogAdjustment{Label: "Competitive score boost", Detail: ">39 kills in regulation time"})
	}
	if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key("position_discount")); ok && m.FloatValue > 0 {
		out = append(out, BattlelogAdjustment{Label: "Scoreboard-position discount", Detail: fmt.Sprintf("%.0f%% reduction", m.FloatValue)})
	}
	if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key("evidence_stacking_boost")); ok && m.StringValue != "" {
		out = append(out, BattlelogAdjustment{Label: "Evidence stacking boost", Detail: m.StringValue})
	}
	if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key("wallhack_co_occurrence_boost")); ok && m.StringValue == "Yes" {
		out = append(out, BattlelogAdjustment{Label: "Wallhack co-occurrence boost", Detail: "pre-FOV pre-aim and back-kill signal both elevated"})
	}
	if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, Key("ttd_sub100_high_floor")); ok && m.StringValue == "Yes" {
		out = append(out, BattlelogAdjustment{Label: "Sub-100ms floor", Detail: "verdict floored at 55%"})
	}
	for _, key := range []Key{"sniper_wallbang_override", "scout_precision_override"} {
		if m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, key); ok && m.StringValue == "Yes" {
			out = append(out, BattlelogAdjustment{Label: "Sniper anomaly override", Detail: string(key)})
		}
	}
	return out
}

// SortBattlelogReports orders reports by likelihood descending (stable by name).
func SortBattlelogReports(reports map[uint64]*BattlelogPlayerReport) []*BattlelogPlayerReport {
	out := make([]*BattlelogPlayerReport, 0, len(reports))
	for _, r := range reports {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Likelihood != out[j].Likelihood {
			return out[i].Likelihood > out[j].Likelihood
		}
		return out[i].Name < out[j].Name
	})
	return out
}
