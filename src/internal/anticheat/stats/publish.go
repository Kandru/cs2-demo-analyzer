package stats

import "fmt"

// publish.go writes the detector's public metrics. Detection rows are emitted
// generically from the registry and each modifier owns its own metric keys.

// flagThreshold is the cheat_likelihood at or above which a player
// is flagged.
const flagThreshold = 50.0

// publishScores writes all anti_cheat metrics for one player.
func publishScores(st *scoreState, applied map[string]Applied) {
	ps := st.ps
	ps.AddMetric(antiCheatCategory, Key("cheat_likelihood"), Metric{
		Type:        MetricPercentage,
		FloatValue:  st.score,
		Description: "Estimated likelihood of player cheating",
	})

	// Per-channel transparency: <id>_score, <id>_confidence, <id>_zone.
	// Channels with HasData=false still emit zero values for the score key so
	// the table layout stays consistent.
	for _, ch := range st.channels {
		ps.AddMetric(antiCheatCategory, Key(ch.ID+"_score"), Metric{
			Type:        MetricFloat,
			FloatValue:  ch.Score,
			Description: fmt.Sprintf("%s cheat score component (0-1)", ch.ID),
		})
		ps.AddMetric(antiCheatCategory, Key(ch.ID+"_confidence"), Metric{
			Type:        MetricFloat,
			FloatValue:  ch.Confidence,
			Description: fmt.Sprintf("Confidence in %s reading (0-1)", ch.ID),
		})
		ps.AddMetric(antiCheatCategory, Key(ch.ID+"_zone"), Metric{
			Type:        MetricString,
			StringValue: ch.Zone.String(),
			Description: fmt.Sprintf("Interpretation band for %s", ch.ID),
		})
	}

	ps.AddMetric(antiCheatCategory, Key("total_cheat_score"), Metric{
		Type:        MetricFloat,
		FloatValue:  st.combined / 100.0,
		Description: "Pre-boost combined Bayesian likelihood (0-1)",
	})

	for _, m := range modifiers {
		if m.Publish != nil {
			m.Publish(ps, applied[m.ID])
		}
	}

	flag := "No"
	if st.score >= flagThreshold {
		flag = "Yes"
	}
	ps.AddMetric(antiCheatCategory, Key("cheater"), Metric{
		Type:        MetricString,
		StringValue: flag,
		Description: "Flag — Yes if cheat_likelihood ≥ flagThreshold",
	})
}
