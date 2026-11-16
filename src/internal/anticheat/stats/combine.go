package stats

import (
	"math"
	"sort"
)

// combine.go implements the Bayesian log-odds combination and the
// lobby-relative normalisation shared by every detection.

const (
	// scorePrior is the base-rate cheater probability before any evidence
	// is seen. 10% means a complete absence of signal pulls the combined
	// likelihood toward ~10%, not 50%.
	scorePrior = 0.10

	// logitEpsilon clamps scores away from 0/1 before logit() so contributions
	// are bounded and cannot dominate the sum.
	logitEpsilon = 0.02

	// lobbyNormAlpha is the shrinkage factor applied to the lobby trimmed mean.
	lobbyNormAlpha = 0.4

	// lobbyNormMinConf is the minimum confidence required for a player's
	// channel reading to count toward the lobby baseline.
	lobbyNormMinConf = 0.25
)

func logit(p float64) float64 {
	if p < logitEpsilon {
		p = logitEpsilon
	}
	if p > 1.0-logitEpsilon {
		p = 1.0 - logitEpsilon
	}
	return math.Log(p / (1.0 - p))
}

func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// bayesianCombine returns the combined cheat likelihood [0, 100]
// for one player from a slice of channels.
func bayesianCombine(channels []Channel) float64 {
	logOdds := logit(scorePrior)
	for _, ch := range channels {
		if !ch.HasData || ch.Confidence <= 0 || ch.Weight <= 0 {
			continue
		}
		contrib := ch.Weight * ch.Confidence * logit(ch.Score)
		if ch.Mode == ModePositiveOnly && contrib < 0 {
			contrib = 0
		}
		logOdds += contrib
	}
	return sigmoid(logOdds) * 100.0
}

// normalizeLobby applies the lobby-relative trimmed-mean shrinkage
// across all players for every channel with ≥2 confident contributors.
//
// For each registered detection:
//  1. Collect scores from players with HasData AND Confidence ≥ lobbyNormMinConf.
//  2. If at least 2 contributors: drop the highest, take mean → μ_trim.
//  3. Adjust every contributor's score: clamp01(score − α·μ_trim).
//
// The trim prevents a lone cheater dragging the baseline up and depressing
// their own outlier signal.
func normalizeLobby(perPlayer map[uint64][]Channel) {
	if len(perPlayer) == 0 {
		return
	}
	for _, det := range detections {
		var contributors []float64
		for _, channels := range perPlayer {
			for _, ch := range channels {
				if ch.ID != det.ID || !ch.HasData || ch.Confidence < lobbyNormMinConf {
					continue
				}
				contributors = append(contributors, ch.Score)
			}
		}
		if len(contributors) < 2 {
			continue
		}
		sort.Float64s(contributors)
		trimmed := contributors[:len(contributors)-1]
		sum := 0.0
		for _, v := range trimmed {
			sum += v
		}
		muTrim := sum / float64(len(trimmed))

		for _, channels := range perPlayer {
			for i, ch := range channels {
				if ch.ID != det.ID || !ch.HasData {
					continue
				}
				adjusted := clamp01(ch.Score - lobbyNormAlpha*muTrim)
				channels[i].Score = adjusted
				channels[i].Zone = zoneFor(adjusted)
			}
		}
	}
}

// buildLobbyContext returns per-player pre-FOV sample counts and whether the
// rest of the lobby is asymmetric (≥50% of OTHER players have <2 samples).
// Used by the pre_fov_presence detection and the TTD-sub100 floor.
func buildLobbyContext(demoStats *DemoStats) *LobbyContext {
	ctx := &LobbyContext{
		SamplesBySID:    map[uint64]int64{},
		AsymmetricBySID: map[uint64]bool{},
	}
	for sid, ps := range demoStats.Players {
		if sid == 0 || ps == nil {
			continue
		}
		n, _ := psGetInt(ps, behavioralCategory, Key("pre_fov_aim_samples"))
		ctx.SamplesBySID[sid] = n
	}
	for sid := range ctx.SamplesBySID {
		others, lowSampleOthers := 0, 0
		for otherSID, otherN := range ctx.SamplesBySID {
			if otherSID == sid {
				continue
			}
			others++
			if otherN < 2 {
				lowSampleOthers++
			}
		}
		ctx.AsymmetricBySID[sid] = others >= 2 && float64(lowSampleOthers)/float64(others) >= 0.5
	}
	return ctx
}
