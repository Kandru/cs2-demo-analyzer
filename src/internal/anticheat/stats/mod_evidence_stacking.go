package stats

import "fmt"

// mod_evidence_stacking.go: ×1.4 when at least three channels each have
// score × confidence ≥ 0.30. Three independent moderate signals are stronger
// evidence than the raw arithmetic sum suggests.
const (
	evidenceStackingMinStrong  = 3
	evidenceStackingMinChannel = 0.30
	evidenceStackingMultiplier = 1.4
)

func applyEvidenceStacking(st *scoreState) Applied {
	strong := 0
	for _, ch := range st.channels {
		if ch.HasData && ch.Score*ch.Confidence >= evidenceStackingMinChannel {
			strong++
		}
	}
	if strong < evidenceStackingMinStrong {
		return Applied{Count: strong}
	}
	st.score *= evidenceStackingMultiplier
	return Applied{Fired: true, Count: strong}
}

var modEvidenceStacking = &Modifier{
	ID:    "evidence_stacking_boost",
	Doc:   "docs/detections/evidence_stacking.md",
	Apply: applyEvidenceStacking,
	Publish: func(ps *PlayerStats, a Applied) {
		if !a.Fired {
			return
		}
		ps.AddMetric(cheatscoreCategoryAntiCheat, Key("evidence_stacking_boost"), Metric{
			Type:        MetricString,
			StringValue: fmt.Sprintf("Yes (%d strong channels)", a.Count),
			Description: "×1.4 boost — ≥3 channels with score×confidence ≥0.30",
		})
	},
	Adjustments: func(ps *PlayerStats) []BattlelogAdjustment {
		detail, ok := yesStringMetric(ps, Key("evidence_stacking_boost"))
		if !ok {
			return nil
		}
		return []BattlelogAdjustment{{Label: "Evidence stacking boost", Detail: detail}}
	},
}
