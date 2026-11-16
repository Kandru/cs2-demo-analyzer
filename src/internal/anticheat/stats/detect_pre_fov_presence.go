package stats

import (
	"fmt"
	"math"
)

const (
	preFOVPresenceID     = "pre_fov_presence"
	preFOVPresenceWeight = 0.10
	preFOVPresenceMode   = ModePositiveOnly
)

// evaluatePreFOVPresence fires only when the player has several tight
// pre-FOV samples and the rest of the lobby does not, which distinguishes a
// wallhack player in a small lobby from a pro in a normal 5v5 where everyone
// accumulates pre-FOV samples.
//
// Conditions:
//   - player has ≥4 pre-FOV samples
//   - player's pre-FOV median ≤ 10° (tight pre-aim, not random)
//   - lobby asymmetry: ≥50% of other players have <2 pre-FOV samples
//
// Score ramps 3 samples → 0.5, 6 samples → 1.0. Confidence is pinned to 1.0
// when the rule fires; otherwise the channel is still appended with
// HasData=false so the published per-channel keys stay stable.
func evaluatePreFOVPresence(ctx *LobbyContext, ps *PlayerStats, sid uint64) Channel {
	ch := Channel{ID: preFOVPresenceID, Weight: preFOVPresenceWeight, Mode: preFOVPresenceMode}
	if sid == 0 || ps == nil {
		return ch
	}
	n := ctx.SamplesBySID[sid]
	med, _ := psGetFloat(ps, channelCategoryBehavioral, Key("pre_fov_aim_median_deg"))
	if n < 4 || med <= 0 || med > 10.0 || !ctx.IsAsymmetric(sid) {
		return ch
	}
	score := clamp01(0.5 + float64(n-3)/3.0*0.5)
	ch.HasData = true
	ch.Score = score
	ch.Confidence = 1.0
	ch.Raw = float64(n)
	ch.SampleN = n
	ch.Zone = zoneFor(score)
	return ch
}

func rawPreFOVPresence(ps *PlayerStats) (float64, int, bool) {
	n, ok := psGetInt(ps, channelCategoryBehavioral, Key("pre_fov_aim_samples"))
	return float64(n), int(n), ok
}

var detPreFOVPresence = &Detection{
	ID:            "pre_fov_presence",
	Label:         "Pre-aim sample asymmetry",
	Unit:          "samples",
	Description:   "Pre-aim samples relative to the rest of the lobby.",
	Doc:           "docs/detections/pre_fov_presence.md",
	Weight:        0.10,
	Mode:          ModePositiveOnly,
	LobbyEvaluate: evaluatePreFOVPresence,
	Raw:           rawPreFOVPresence,
	Narrative: &NarrativeSpec{
		ScoreBased: true,
		Ascending:  true,
		Mild:       math.Inf(1), // presence never produces a mild sentence
		Strong:     0.50,
		Blatant:    0.80,
		Sentence: func(raw float64, n int64, adj string) string {
			return fmt.Sprintf("Pre-FOV pre-aim presence fires on %d samples — lobby asymmetry indicates other players don't produce the same pre-aim pattern.", n)
		},
	},
}
