package stats

// mod_wallhack_cooccurrence.go: ×1.2 when both halves of the wallhack-via-info
// signature are present:
//
//	pre_fov.Score × pre_fov.Confidence ≥ 0.45 — low pre-aim after lobby norm
//	back_kill_given_pct ≥ 8.0 on ≥ 4 kills    — flanks unaware opponents
//
// Either alone is noisy: pro lobbies have low post-norm pre-FOV, and
// legitimate flankers have high back-kill-given. The conjunction is the
// wallhack signature.
const (
	coOccurrencePreFOVProduct = 0.45
	coOccurrenceBackKillPct   = 8.0
	coOccurrenceBackKillMin   = 4
	coOccurrenceMultiplier    = 1.20
)

func applyWallhackCoOccurrenceBoost(st *scoreState) Applied {
	var preFOVProd float64
	for _, ch := range st.channels {
		if ch.ID == detPreFOV.ID && ch.HasData {
			preFOVProd = ch.Score * ch.Confidence
			break
		}
	}
	if preFOVProd < coOccurrencePreFOVProduct {
		return Applied{}
	}
	pct, hasPct := psGetFloat(st.ps, channelCategoryBehavioral, Key("back_kill_given_pct"))
	if !hasPct || pct < coOccurrenceBackKillPct {
		return Applied{}
	}
	kills, hasKills := psGetInt(st.ps, channelCategoryBehavioral, Key("back_kill_given_total_kills"))
	if !hasKills || kills < coOccurrenceBackKillMin {
		return Applied{}
	}
	st.score *= coOccurrenceMultiplier
	return Applied{Fired: true}
}

var modCooccurrence = &Modifier{
	ID:    "wallhack_co_occurrence_boost",
	Doc:   "docs/detections/wallhack_cooccurrence.md",
	Apply: applyWallhackCoOccurrenceBoost,
	Publish: func(ps *PlayerStats, a Applied) {
		if !a.Fired {
			return
		}
		ps.AddMetric(cheatscoreCategoryAntiCheat, Key("wallhack_co_occurrence_boost"), Metric{
			Type:        MetricString,
			StringValue: "Yes",
			Description: "×1.2 boost — pre-FOV pre-aim AND back-kill-given both elevated",
		})
	},
	Adjustments: func(ps *PlayerStats) []BattlelogAdjustment {
		if !yesMetric(ps, Key("wallhack_co_occurrence_boost")) {
			return nil
		}
		return []BattlelogAdjustment{{
			Label:  "Wallhack co-occurrence boost",
			Detail: "pre-FOV pre-aim and back-kill signal both elevated",
		}}
	},
}
