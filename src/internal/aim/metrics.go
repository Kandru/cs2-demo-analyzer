package aim

import "github.com/Kandru/cs2-demo-analyzer/internal/model"

// metrics.go builds the per-player aim profile from the registry. It is a pure
// transformation of playerState so it can be tested without a parser.

func buildAimStats(st *playerState, anyHurt bool) *model.AimStats {
	if !anyHurt {
		// Older demos without PlayerHurt must show "no data" for shot-based
		// metrics rather than a misleading 0%.
		st.shots = 0
		st.hits = 0
		st.sprayAtt = 0
		st.sprayHits = 0
	}

	stats := &model.AimStats{}
	var weighted, weightSum float64
	valid := 0
	for _, spec := range metricSpecs {
		value := spec.Value(st)
		samples := spec.Samples(st)
		m := model.AimMetric{
			ID:        spec.ID,
			Label:     spec.Label,
			Unit:      spec.Unit,
			Value:     value,
			Weight:    spec.Weight,
			Samples:   samples,
			LowSample: samples > 0 && samples < spec.LowSample,
			Tooltip:   spec.Tooltip,
		}
		if samples > 0 {
			m.Score = anchorScore(value, spec.Lo, spec.Hi, spec.HigherBetter)
			weighted += m.Score * spec.Weight
			weightSum += spec.Weight
			valid++
			m.Display = spec.Format(value)
		} else {
			m.Display = "—"
		}
		stats.Metrics = append(stats.Metrics, m)
	}
	if valid >= 3 && weightSum > 0 {
		stats.Rating = weighted / weightSum
		stats.Grade = grade(stats.Rating)
		stats.Confidence = weightSum / 100.0
	}
	return stats
}
