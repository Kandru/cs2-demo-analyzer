package stats

import "fmt"

const (
	recoilID     = "recoil"
	recoilWeight = 0.10
	recoilMode   = ModePositiveOnly
)

// evaluateRecoil reuses the already-computed recoil_score (0–1, where 1 is
// suspicious). Confidence ramps over 20 counted bullets. Positive-only — a
// clean recoil reading only means no suspicious sprays were seen.
func evaluateRecoil(ps *PlayerStats) Channel {
	raw, ok := psGetFloat(ps, recoilCategory, Key("recoil_score"))
	bullets, _ := psGetInt(ps, recoilCategory, Key("total_counted_bullets"))
	if !ok || bullets <= 0 {
		return Channel{ID: recoilID, Weight: recoilWeight, Mode: recoilMode}
	}
	score := clamp01(raw)
	return Channel{
		ID:         recoilID,
		Score:      score,
		Confidence: linearConfidence(bullets, 20),
		Raw:        raw,
		SampleN:    bullets,
		Weight:     recoilWeight,
		Zone:       zoneFor(score),
		Mode:       recoilMode,
		HasData:    true,
	}
}

func rawRecoil(ps *PlayerStats) (float64, int, bool) {
	value, ok := psGetFloat(ps, recoilCategory, Key("mean_angular_error"))
	bullets, _ := psGetInt(ps, recoilCategory, Key("total_counted_bullets"))
	return value, int(bullets), ok
}

// rawRecoilScore is the narrative/anchor input. The report table shows the
// physical mean angular error, but the prose describes the normalised score.
func rawRecoilScore(ps *PlayerStats) (float64, int64, bool) {
	value, ok := psGetFloat(ps, recoilCategory, Key("recoil_score"))
	bullets, _ := psGetInt(ps, recoilCategory, Key("total_counted_bullets"))
	return value, bullets, ok
}

var detRecoil = &Detection{
	ID:          "recoil",
	Label:       "Recoil deviation",
	Unit:        "°",
	Description: "Mean angular error against known spray patterns.",
	Doc:         "docs/detections/recoil.md",
	Weight:      0.10,
	Mode:        ModePositiveOnly,
	Evaluate:    evaluateRecoil,
	Raw:         rawRecoil,
	Narrative: &NarrativeSpec{
		MinSamples: 20,
		Ascending:  true,
		Mild:       0.30,
		Strong:     0.55,
		Blatant:    0.80,
		Raw:        rawRecoilScore,
		Sentence: func(raw float64, n int64, adj string) string {
			return fmt.Sprintf("Recoil-control anomaly is %s elevated.", adj)
		},
	},
	Evidence: &EvidenceSpec{
		Kind:      "recoil",
		Label:     "Recoil",
		Rank:      2,
		Unit:      "°",
		Threshold: 0.30,
		Severity:  func(v float64) float64 { return 10 - v },
		ValueText: func(v float64) string { return fmt.Sprintf("%.2f°", v) },
	},
}
