package stats

import "fmt"

const (
	preFOVID     = "pre_fov"
	preFOVWeight = 0.22
	preFOVMode   = ModeBidirectional
)

// evaluatePreFOV scores pre-FOV pre-aim median angle. Ramp 12°→4° (clean→
// blatant — descending). n_full=15, sqrt confidence. Bidirectional.
//
// Calibration deviates from the published research table (7°→2.5°) because
// observed wingman cheaters land at 6.17° and 7.25°; a 7° clean threshold
// would zero them out. The 12°→4° ramp is anchored on the observed corpus.
//
// Weight 0.22: pre-FOV pre-aim is the most-cited wallhack indicator and the
// only channel that reliably distinguishes a wallhack-only cheater from a
// legit player.
func evaluatePreFOV(ps *PlayerStats) Channel {
	n, hasN := psGetInt(ps, behavioralCategory, Key("pre_fov_aim_samples"))
	if !hasN || n <= 0 {
		return Channel{ID: preFOVID, Weight: preFOVWeight, Mode: preFOVMode}
	}
	med, _ := psGetFloat(ps, behavioralCategory, Key("pre_fov_aim_median_deg"))
	score := linearScore(med, 12.0, 4.0)
	return Channel{
		ID:         preFOVID,
		Score:      score,
		Confidence: sqrtConfidence(n, 15),
		Raw:        med,
		SampleN:    n,
		Weight:     preFOVWeight,
		Zone:       zoneFor(score),
		Mode:       preFOVMode,
		HasData:    true,
	}
}

func rawPreFOV(ps *PlayerStats) (float64, int, bool) {
	value, ok := psGetFloat(ps, behavioralCategory, Key("pre_fov_aim_median_deg"))
	n, _ := psGetInt(ps, behavioralCategory, Key("pre_fov_aim_samples"))
	return value, int(n), ok
}

var detPreFOV = &Detection{
	ID:          "pre_fov",
	Label:       "Pre-FOV pre-aim",
	Unit:        "°",
	Description: "Median crosshair-to-enemy angle 200 ms before visibility.",
	Doc:         "docs/detections/pre_fov.md",
	Weight:      0.22,
	Mode:        ModeBidirectional,
	Evaluate:    evaluatePreFOV,
	Raw:         rawPreFOV,
	Narrative: &NarrativeSpec{
		MinSamples: 4,
		Ascending:  false,
		Mild:       8.0,
		Strong:     7.0,
		Blatant:    6.0,
		Sentence: func(raw float64, n int64, adj string) string {
			return fmt.Sprintf("Pre-FOV pre-aim median of %.2f° on %d kills is %s tight — crosshair near enemy positions before line-of-sight.", raw, n, adj)
		},
	},
	Evidence: &EvidenceSpec{
		Kind:      "preaim",
		Label:     "Pre-aim",
		Rank:      3,
		Unit:      "°",
		Threshold: 12.0,
		Severity:  func(v float64) float64 { return 90 - v },
		ValueText: func(v float64) string { return fmt.Sprintf("%.1f°", v) },
	},
}
