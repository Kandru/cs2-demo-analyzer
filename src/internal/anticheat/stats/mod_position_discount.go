package stats

import "fmt"

// mod_position_discount.go: multiply the score by (1 − 0.2 × position_factor).
// Players who never climb off the bottom of their team's scoreboard get a
// modest discount; the factor is computed by ScoreboardCollector.
func applyPositionDiscount(st *scoreState) Applied {
	factor, ok := psGetFloat(st.ps, scoreboardCategory, Key("position_factor"))
	if !ok || factor <= 0 {
		return Applied{}
	}
	discount := 0.20 * factor
	st.score *= 1.0 - discount
	return Applied{Fired: true, Value: discount}
}

var modPosition = &Modifier{
	ID:    "position_discount",
	Doc:   "docs/detections/position_discount.md",
	Apply: applyPositionDiscount,
	Publish: func(ps *PlayerStats, a Applied) {
		if !a.Fired {
			return
		}
		ps.AddMetric(antiCheatCategory, Key("position_discount"), Metric{
			Type:        MetricPercentage,
			FloatValue:  a.Value * 100,
			Description: "Reduction applied for low scoreboard position vs. teammates",
		})
	},
	Adjustments: func(ps *PlayerStats) []ReportAdjustment {
		m, ok := ps.GetMetric(antiCheatCategory, Key("position_discount"))
		if !ok || m.FloatValue <= 0 {
			return nil
		}
		return []ReportAdjustment{{
			Label:  "Scoreboard-position discount",
			Detail: fmt.Sprintf("%.0f%% reduction", m.FloatValue),
		}}
	},
}
