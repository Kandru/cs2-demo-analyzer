package stats

// mod_sniper_overrides.go pins the score to 100 for high-confidence sniper
// anomalies:
//
//   - >10 sniper wallbang kills across the match
//   - >10 Scout kills with a ≥80% headshot rate
//
// These are rare, high-specificity patterns; the override is deliberately
// pinned rather than additive.
func applySniperOverrides(st *scoreState) Applied {
	var names []string
	if wb, ok := psGetInt(st.ps, sniperCategory, Key("sniper_wallbang_kills")); ok && wb > 10 {
		st.score = 100.0
		names = append(names, "sniper_wallbang_override")
	}
	if scoutKills, ok := psGetInt(st.ps, sniperCategory, Key("scout_kills")); ok && scoutKills > 10 {
		if scoutHS, ok := psGetFloat(st.ps, sniperCategory, Key("scout_hs_rate")); ok && scoutHS >= 80.0 {
			st.score = 100.0
			names = append(names, "scout_precision_override")
		}
	}
	return Applied{Fired: len(names) > 0, Names: names}
}

var modSniper = &Modifier{
	ID:    "sniper_overrides",
	Doc:   "docs/detections/sniper_overrides.md",
	Apply: applySniperOverrides,
	Publish: func(ps *PlayerStats, a Applied) {
		for _, name := range a.Names {
			ps.AddMetric(cheatscoreCategoryAntiCheat, Key(name), Metric{
				Type:        MetricString,
				StringValue: "Yes",
				Description: "Sniper-anomaly override — pinned to 100%",
			})
		}
	},
	Adjustments: func(ps *PlayerStats) []BattlelogAdjustment {
		var out []BattlelogAdjustment
		for _, key := range []Key{"sniper_wallbang_override", "scout_precision_override"} {
			if yesMetric(ps, key) {
				out = append(out, BattlelogAdjustment{Label: "Sniper anomaly override", Detail: string(key)})
			}
		}
		return out
	},
}
