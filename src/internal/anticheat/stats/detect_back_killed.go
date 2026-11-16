package stats

import "fmt"

const (
	backKilledID     = "back_killed"
	backKilledWeight = 0.06
	backKilledMode   = ModePositiveOnly
)

// evaluateBackKilled scores back-kill avoidance. Ramp 25%→3% (clean→blatant —
// descending; a low back-killed rate is suspicious). n_full=8 deaths.
func evaluateBackKilled(ps *PlayerStats) Channel {
	n, hasN := psGetInt(ps, channelCategoryBehavioral, Key("back_killed_total_deaths"))
	if !hasN || n <= 0 {
		return Channel{ID: backKilledID, Weight: backKilledWeight, Mode: backKilledMode}
	}
	rate, _ := psGetFloat(ps, channelCategoryBehavioral, Key("back_killed_pct"))
	score := linearScore(rate, 25.0, 3.0)
	return Channel{
		ID:         backKilledID,
		Score:      score,
		Confidence: linearConfidence(n, 8),
		Raw:        rate,
		SampleN:    n,
		Weight:     backKilledWeight,
		Zone:       zoneFor(score),
		Mode:       backKilledMode,
		HasData:    true,
	}
}

func rawBackKilled(ps *PlayerStats) (float64, int, bool) {
	value, ok := psGetFloat(ps, channelCategoryBehavioral, Key("back_killed_pct"))
	n, _ := psGetInt(ps, channelCategoryBehavioral, Key("back_killed_total_deaths"))
	return value, int(n), ok
}

var detBackKilled = &Detection{
	ID:          "back_killed",
	Label:       "Back-killed rate",
	Unit:        "%",
	Description: "Share of deaths where the player was looking away from the killer.",
	Doc:         "docs/detections/back_killed.md",
	Weight:      0.06,
	Mode:        ModePositiveOnly,
	Evaluate:    evaluateBackKilled,
	Raw:         rawBackKilled,
	Narrative: &NarrativeSpec{
		MinSamples: 8,
		Ascending:  false,
		Mild:       8.0,
		Strong:     4.0,
		Blatant:    0.01,
		Sentence: func(raw float64, n int64, adj string) string {
			return fmt.Sprintf("Back-killed rate of %.1f%% across %d deaths is %s low — rarely caught from behind.", raw, n, adj)
		},
	},
}
