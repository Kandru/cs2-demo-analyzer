package stats

// mod_ttd_sub100_floor.go: floor at 55% when a player exhibits ALL of:
//   - sub_100ms_ttd ≥ 25% on ≥20 TTD samples AND ≥5 such engagements
//   - pre-FOV median ≤ 9° on ≥8 pre-FOV samples (wallhack-style pre-aim)
//   - lobby asymmetry: at least half of OTHER players have <2 pre-FOV samples
//
// This is the documented tightening of the upstream 3-sample floor, which
// pinned clean players at 55% on short/custom lobbies. A same-tick sighting is
// dropped upstream as a radar-spot artifact. Applied after boosts so it caps
// regression, not amplification.
const (
	ttdSub100FloorRate          = 25.0
	ttdSub100FloorSamples       = 20
	ttdSub100FloorCount         = 5
	ttdSub100FloorPreFOVSamples = 8
	ttdSub100FloorScore         = 55.0
)

func applyTTDSub100Floor(st *scoreState) Applied {
	n, hasN := psGetInt(st.ps, channelCategoryReaction, Key("ttd_samples"))
	if !hasN || n < ttdSub100FloorSamples {
		return Applied{}
	}
	rate, hasRate := psGetFloat(st.ps, channelCategoryReaction, Key("sub_100ms_ttd"))
	if !hasRate || rate < ttdSub100FloorRate {
		return Applied{}
	}
	count, hasCount := psGetInt(st.ps, channelCategoryReaction, Key("sub_100ms_count"))
	if !hasCount || count < ttdSub100FloorCount {
		return Applied{}
	}
	preFOVN, hasFOVN := psGetInt(st.ps, channelCategoryBehavioral, Key("pre_fov_aim_samples"))
	if !hasFOVN || preFOVN < ttdSub100FloorPreFOVSamples {
		return Applied{}
	}
	preFOVMed, hasFOVMed := psGetFloat(st.ps, channelCategoryBehavioral, Key("pre_fov_aim_median_deg"))
	if !hasFOVMed || preFOVMed <= 0 || preFOVMed > 9.0 {
		return Applied{}
	}
	if !st.lobbyAsymmetric {
		return Applied{}
	}
	if st.score < ttdSub100FloorScore {
		st.score = ttdSub100FloorScore
	}
	return Applied{Fired: true}
}

var modFloor = &Modifier{
	ID:    "ttd_sub100_high_floor",
	Doc:   "docs/detections/ttd_sub100_floor.md",
	Apply: applyTTDSub100Floor,
	Publish: func(ps *PlayerStats, a Applied) {
		if !a.Fired {
			return
		}
		ps.AddMetric(cheatscoreCategoryAntiCheat, Key("ttd_sub100_high_floor"), Metric{
			Type:        MetricString,
			StringValue: "Yes",
			Description: "Score floored at 55% — sub-100ms TTD rate ≥25% on ≥20 samples with ≥5 sub-100ms events",
		})
	},
	Adjustments: func(ps *PlayerStats) []BattlelogAdjustment {
		if !yesMetric(ps, Key("ttd_sub100_high_floor")) {
			return nil
		}
		return []BattlelogAdjustment{{Label: "Sub-100ms floor", Detail: "verdict floored at 55%"}}
	},
}
