package stats

import "fmt"

// mod_wingman_boost.go: ×1.8 in Wingman when KPR ≥ 0.7 OR kills ≥ 10.
// Short Wingman demos that end at 8–9 rounds still receive the boost when KPR
// is suspicious. The kills backstop covers demos with no round count.

const wingmanKPRThreshold = 0.7

func applyWingmanBoost(st *scoreState) Applied {
	gameMode, _ := psGetString(st.ps, cheatscoreCategoryGameInfo, Key("game_mode"))
	if gameMode != "Wingman" {
		return Applied{}
	}
	totalKills, _ := psGetInt(st.ps, channelCategoryKills, Key("total_kills"))
	roundCount, _ := psGetInt(st.ps, cheatscoreCategoryGameInfo, Key("round_count"))

	var kpr float64
	if roundCount > 0 {
		kpr = float64(totalKills) / float64(roundCount)
	}
	if totalKills >= 10 {
		st.score *= 1.8
		return Applied{Fired: true, Detail: fmt.Sprintf("kills=%d", totalKills)}
	}
	if kpr >= wingmanKPRThreshold {
		st.score *= 1.8
		return Applied{Fired: true, Detail: fmt.Sprintf("KPR=%.2f", kpr)}
	}
	return Applied{}
}

var modWingman = &Modifier{
	ID:    "wingman_boost",
	Doc:   "docs/detections/wingman_boost.md",
	Apply: applyWingmanBoost,
	Publish: func(ps *PlayerStats, a Applied) {
		if !a.Fired {
			return
		}
		ps.AddMetric(cheatscoreCategoryAntiCheat, Key("wingman_boost"), Metric{
			Type:        MetricString,
			StringValue: "Yes",
			Description: "Wingman boost applied (" + a.Detail + ")",
		})
		ps.AddMetric(cheatscoreCategoryAntiCheat, Key("wingman_kpr_boost_reason"), Metric{
			Type:        MetricString,
			StringValue: a.Detail,
			Description: "Reason the Wingman boost fired",
		})
	},
	Adjustments: func(ps *PlayerStats) []BattlelogAdjustment {
		if !yesMetric(ps, Key("wingman_boost")) {
			return nil
		}
		detail, _ := yesStringMetric(ps, Key("wingman_kpr_boost_reason"))
		return []BattlelogAdjustment{{Label: "Wingman score boost", Detail: detail}}
	},
}
