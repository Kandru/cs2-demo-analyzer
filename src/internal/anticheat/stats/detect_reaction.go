package stats

import "fmt"

const (
	reactionID     = "reaction"
	reactionWeight = 0.10
	reactionMode   = ModeBidirectional
)

// evaluateReactionMedianTTD scores median time-to-damage. Ramp 500→150 ms,
// n_full=10, sqrt confidence. Bidirectional: a 500 ms median on many samples
// is real evidence of human-paced reactions.
//
// Median is preferable to P10 because a single pre-fired tick (legitimate via
// audio cues, pre-armed peeks, or callouts) produces a sub-100 ms P10 without
// indicating cheat behavior. The median only registers a consistent pattern
// of fast reactions.
func evaluateReactionMedianTTD(ps *PlayerStats) Channel {
	n, hasN := psGetInt(ps, reactionCategory, Key("ttd_samples"))
	if !hasN || n <= 0 {
		return Channel{ID: reactionID, Weight: reactionWeight, Mode: reactionMode}
	}
	median, _ := psGetFloat(ps, reactionCategory, Key("median_ttd"))
	score := linearScore(median, 500.0, 150.0) // descending: low ms → high score
	return Channel{
		ID:         reactionID,
		Score:      score,
		Confidence: sqrtConfidence(n, 10),
		Raw:        median,
		SampleN:    n,
		Weight:     reactionWeight,
		Zone:       zoneFor(score),
		Mode:       reactionMode,
		HasData:    true,
	}
}

func rawReaction(ps *PlayerStats) (float64, int, bool) {
	value, ok := psGetFloat(ps, reactionCategory, Key("median_ttd"))
	n, _ := psGetInt(ps, reactionCategory, Key("ttd_samples"))
	return value, int(n), ok
}

var detReaction = &Detection{
	ID:          "reaction",
	Label:       "Damage speed (TTD)",
	Unit:        "ms",
	Description: "Median time from visibility to first damage.",
	Doc:         "docs/detections/reaction.md",
	Weight:      0.10,
	Mode:        ModeBidirectional,
	Evaluate:    evaluateReactionMedianTTD,
	Raw:         rawReaction,
	Narrative: &NarrativeSpec{
		MinSamples: 10,
		Ascending:  false,
		Mild:       400.0,
		Strong:     250.0,
		Blatant:    150.0,
		Sentence: func(raw float64, n int64, adj string) string {
			return fmt.Sprintf("Median time-to-damage of %.0f ms across %d engagements is %s fast — consistently quick reactions rather than a single prefired tick.", raw, n, adj)
		},
	},
	Evidence: &EvidenceSpec{
		Kind:      "ttd",
		Label:     "Reaction",
		Rank:      2,
		Unit:      "ms",
		Threshold: 250.0,
		Severity:  func(v float64) float64 { return 1000 - v },
		ValueText: func(v float64) string { return fmt.Sprintf("%.0f ms", v) },
	},
}
