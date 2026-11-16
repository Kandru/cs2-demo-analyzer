package stats

import "fmt"

const (
	snapID     = "snap"
	snapWeight = 0.10
	snapMode   = ModePositiveOnly
)

// evaluateSnap scores P95 snap velocity. Ramp 2.0→3.5 °/ms, n_full=10.
//
// Positive-only: a low P95 doesn't exonerate, only flags upward. Weight 0.10
// reflects that in pro lobbies every aggressive rifler occasionally crosses
// 2°/ms; pre-FOV/decoupling/back-killed carry the wallhack signature more
// reliably.
func evaluateSnap(ps *PlayerStats) Channel {
	snapCount, hasN := psGetInt(ps, aimingCategory, Key("snap_count"))
	if !hasN || snapCount <= 0 {
		return Channel{ID: snapID, Weight: snapWeight, Mode: snapMode}
	}
	p95, _ := psGetFloat(ps, aimingCategory, Key("p95_snap_velocity"))
	score := linearScore(p95, 2.0, 3.5)
	return Channel{
		ID:         snapID,
		Score:      score,
		Confidence: linearConfidence(snapCount, 10),
		Raw:        p95,
		SampleN:    snapCount,
		Weight:     snapWeight,
		Zone:       zoneFor(score),
		Mode:       snapMode,
		HasData:    true,
	}
}

func rawSnap(ps *PlayerStats) (float64, int, bool) {
	value, ok := psGetFloat(ps, aimingCategory, Key("p95_snap_velocity"))
	n, _ := psGetInt(ps, aimingCategory, Key("snap_count"))
	return value, int(n), ok
}

var detSnap = &Detection{
	ID:          "snap",
	Label:       "Snap velocity P95",
	Unit:        "°/ms",
	Description: "95th percentile view-angle snap speed.",
	Doc:         "docs/detections/snap.md",
	Weight:      0.10,
	Mode:        ModePositiveOnly,
	Evaluate:    evaluateSnap,
	Raw:         rawSnap,
	Narrative: &NarrativeSpec{
		MinSamples: 5,
		Ascending:  true,
		Mild:       4.0,
		Strong:     6.0,
		Blatant:    10.0,
		Sentence: func(raw float64, n int64, adj string) string {
			return fmt.Sprintf("P95 snap velocity of %.2f °/ms across %d snaps is %s above the lobby baseline.", raw, n, adj)
		},
	},
	Evidence: &EvidenceSpec{
		Kind:      "snap",
		Label:     "Snap",
		Rank:      2,
		Unit:      "°/ms",
		Threshold: 2.0,
		Severity:  func(v float64) float64 { return v },
		ValueText: func(v float64) string { return fmt.Sprintf("%.2f °/ms", v) },
	},
}
