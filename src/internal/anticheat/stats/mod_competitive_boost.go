package stats

// mod_competitive_boost.go: ×1.2 in Competitive when totalKills > 39 in
// ≤30 rounds. The round cap prevents overtime stat-padding from triggering it.
func applyCompetitiveBoost(st *scoreState) Applied {
	gameMode, _ := psGetString(st.ps, gameInfoCategory, Key("game_mode"))
	if gameMode != "Competitive" {
		return Applied{}
	}
	totalKills, _ := psGetInt(st.ps, killsCategory, Key("total_kills"))
	roundCount, _ := psGetInt(st.ps, gameInfoCategory, Key("round_count"))
	if totalKills <= 39 || roundCount > 30 {
		return Applied{}
	}
	st.score *= 1.2
	return Applied{Fired: true}
}

var modCompetitive = &Modifier{
	ID:    "competitive_boost",
	Doc:   "docs/detections/competitive_boost.md",
	Apply: applyCompetitiveBoost,
	Publish: func(ps *PlayerStats, a Applied) {
		if !a.Fired {
			return
		}
		ps.AddMetric(antiCheatCategory, Key("competitive_boost"), Metric{
			Type:        MetricString,
			StringValue: "Yes",
			Description: "Player has more than 39 kills in regulation time (20% boost applied)",
		})
	},
	Adjustments: func(ps *PlayerStats) []BattlelogAdjustment {
		if !yesMetric(ps, Key("competitive_boost")) {
			return nil
		}
		return []BattlelogAdjustment{{Label: "Competitive score boost", Detail: ">39 kills in regulation time"}}
	},
}
