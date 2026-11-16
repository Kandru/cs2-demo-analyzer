package stats

import "fmt"

const (
	attentionID     = "attention"
	attentionWeight = 0.06
	attentionMode   = ModePositiveOnly
)

// evaluateAttention scores nearest-enemy angle median during off-engagement
// moments. Ramp 33°→18° (clean→blatant — descending). n_full=200 frames.
// Positive-only — a high attention angle only means the crosshair isn't
// tight, which is not exoneration.
func evaluateAttention(ps *PlayerStats) Channel {
	n, hasN := psGetInt(ps, channelCategoryBehavioral, Key("nearest_enemy_angle_samples"))
	if !hasN || n <= 0 {
		return Channel{ID: attentionID, Weight: attentionWeight, Mode: attentionMode}
	}
	med, _ := psGetFloat(ps, channelCategoryBehavioral, Key("nearest_enemy_angle_median_deg"))
	score := linearScore(med, 33.0, 18.0)
	return Channel{
		ID:         attentionID,
		Score:      score,
		Confidence: linearConfidence(n, 200),
		Raw:        med,
		SampleN:    n,
		Weight:     attentionWeight,
		Zone:       zoneFor(score),
		Mode:       attentionMode,
		HasData:    true,
	}
}

func rawAttention(ps *PlayerStats) (float64, int, bool) {
	value, ok := psGetFloat(ps, channelCategoryBehavioral, Key("nearest_enemy_angle_median_deg"))
	n, _ := psGetInt(ps, channelCategoryBehavioral, Key("nearest_enemy_angle_samples"))
	return value, int(n), ok
}

var detAttention = &Detection{
	ID:          "attention",
	Label:       "Off-engagement attention",
	Unit:        "°",
	Description: "Median crosshair-to-nearest-enemy angle outside fights.",
	Doc:         "docs/detections/attention.md",
	Weight:      0.06,
	Mode:        ModePositiveOnly,
	Evaluate:    evaluateAttention,
	Raw:         rawAttention,
	Narrative: &NarrativeSpec{
		MinSamples: 200,
		Ascending:  false,
		Mild:       27.0,
		Strong:     24.0,
		Blatant:    20.0,
		Sentence: func(raw float64, n int64, adj string) string {
			return fmt.Sprintf("Off-engagement crosshair drifts to within a %.1f° median of the nearest enemy — %s aware of unseen opponents.", raw, adj)
		},
	},
}
