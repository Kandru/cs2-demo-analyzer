package stats

import "fmt"

const (
	ttdSub100ID     = "ttd_sub100"
	ttdSub100Weight = 0.10
	ttdSub100Mode   = ModePositiveOnly
)

// evaluateTTDSub100 scores the sub-100 ms TTD rate. Ramp 2%→30%, n_full=30,
// sqrt confidence — pinned to 1.0 when count_sub100 ≥ 2, because two or more
// sub-100 ms damage events in one match is the surprising signal, not the
// rate. Positive-only.
func evaluateTTDSub100(ps *PlayerStats) Channel {
	n, hasN := psGetInt(ps, reactionCategory, Key("ttd_samples"))
	if !hasN || n <= 0 {
		return Channel{ID: ttdSub100ID, Weight: ttdSub100Weight, Mode: ttdSub100Mode}
	}
	rate, _ := psGetFloat(ps, reactionCategory, Key("sub_100ms_ttd"))
	score := linearScore(rate, 2.0, 30.0)
	conf := sqrtConfidence(n, 30)
	countSub100 := int64(rate / 100.0 * float64(n))
	if countSub100 >= 2 {
		conf = 1.0
	}
	return Channel{
		ID:         ttdSub100ID,
		Score:      score,
		Confidence: conf,
		Raw:        rate,
		SampleN:    n,
		Weight:     ttdSub100Weight,
		Zone:       zoneFor(score),
		Mode:       ttdSub100Mode,
		HasData:    true,
	}
}

func rawTTDSub100(ps *PlayerStats) (float64, int, bool) {
	value, ok := psGetFloat(ps, reactionCategory, Key("sub_100ms_ttd"))
	n, _ := psGetInt(ps, reactionCategory, Key("ttd_samples"))
	return value, int(n), ok
}

var detTTDSub100 = &Detection{
	ID:          "ttd_sub100",
	Label:       "Sub-100ms damage",
	Unit:        "%",
	Description: "Share of engagements completed in under 100 ms.",
	Doc:         "docs/detections/ttd_sub100.md",
	Weight:      0.10,
	Mode:        ModePositiveOnly,
	Evaluate:    evaluateTTDSub100,
	Raw:         rawTTDSub100,
	Narrative: &NarrativeSpec{
		MinSamples: 5,
		Ascending:  true,
		Mild:       15.0,
		Strong:     25.0,
		Blatant:    35.0,
		Sentence: func(raw float64, n int64, adj string) string {
			return fmt.Sprintf("Sub-100 ms time-to-damage rate of %.1f%% across %d samples is %s implausible without information assistance.", raw, n, adj)
		},
	},
	Evidence: &EvidenceSpec{
		Kind:      "ttd_sub100",
		Label:     "Sub-100 ms",
		Rank:      3,
		Unit:      "ms",
		Threshold: 100.0,
		Severity:  func(v float64) float64 { return 1000 - v },
		ValueText: func(v float64) string { return fmt.Sprintf("%.0f ms", v) },
	},
}
