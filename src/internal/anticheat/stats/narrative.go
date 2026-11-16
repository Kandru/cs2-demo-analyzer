package stats

import (
	"sort"
	"strings"
)

// narrative.go produces a short prose paragraph summarising each player's
// detection profile. One sentence per detection whose raw value crosses its
// documented threshold, ordered strongest first. A closing sentence notes any
// modifiers that fired.
//
// Anchoring adjectives on raw values (not on the lobby-normalized score) keeps
// clean players' narratives honest: in a pro 5v5 lobby every channel may score
// high after normalization, but raw values still tell the truth.
//
// Per-player, adjectives cycle within each tier so the paragraph does not read
// "strongly X, strongly Y, strongly Z."

var (
	narrativeBlatantAdjs = []string{"exceptionally", "blatantly", "unmistakably"}
	narrativeStrongAdjs  = []string{"strongly", "markedly", "clearly", "notably"}
	narrativeMildAdjs    = []string{"moderately", "appreciably", "slightly"}
)

// NarrativeSpec is the per-detection prose configuration.
type NarrativeSpec struct {
	MinSamples int64
	// Ascending is true when higher raw values are more suspicious.
	Ascending bool
	Mild      float64
	Strong    float64
	Blatant   float64
	// ScoreBased compares score × confidence instead of a raw measurement.
	ScoreBased bool
	// Raw overrides the detection's report raw value for prose purposes.
	Raw func(*PlayerStats) (float64, int64, bool)
	// Sentence renders one tier sentence. adj is the cycling adjective.
	Sentence func(raw float64, n int64, adj string) string
}

type narrativeEntry struct {
	det  *Detection
	raw  float64
	n    int64
	tier int
}

// narrativeTier returns the tier index (3=blatant, 2=strong, 1=mild, 0=skip)
// for a raw value compared to three thresholds.
func narrativeTier(raw, mild, strong, blatant float64, ascendingSuspicion bool) int {
	if ascendingSuspicion {
		switch {
		case raw >= blatant:
			return 3
		case raw >= strong:
			return 2
		case raw >= mild:
			return 1
		default:
			return 0
		}
	}
	switch {
	case raw <= blatant:
		return 3
	case raw <= strong:
		return 2
	case raw <= mild:
		return 1
	default:
		return 0
	}
}

// evaluate returns the narrative tier for one detection, if it has data.
func (spec *NarrativeSpec) evaluate(ps *PlayerStats, det *Detection) (float64, int64, int) {
	if spec.ScoreBased {
		score, ok := psGetFloat(ps, cheatscoreCategoryAntiCheat, Key(det.ID+"_score"))
		if !ok {
			return 0, 0, 0
		}
		conf, _ := psGetFloat(ps, cheatscoreCategoryAntiCheat, Key(det.ID+"_confidence"))
		raw := score * conf
		n := int64(0)
		if det.Raw != nil {
			_, samples, _ := det.Raw(ps)
			n = int64(samples)
		}
		return raw, n, narrativeTier(raw, spec.Mild, spec.Strong, spec.Blatant, spec.Ascending)
	}

	rawFn := spec.Raw
	if rawFn == nil && det.Raw != nil {
		rawFn = func(ps *PlayerStats) (float64, int64, bool) {
			value, samples, ok := det.Raw(ps)
			return value, int64(samples), ok
		}
	}
	if rawFn == nil {
		return 0, 0, 0
	}
	raw, n, ok := rawFn(ps)
	if !ok || n < spec.MinSamples {
		return 0, 0, 0
	}
	return raw, n, narrativeTier(raw, spec.Mild, spec.Strong, spec.Blatant, spec.Ascending)
}

// buildCheatscoreNarrative reads a player's published anti_cheat metrics and
// returns a multi-sentence paragraph. Returns "" if the player has no
// anti_cheat data (e.g., parser failed).
func buildCheatscoreNarrative(ps *PlayerStats) string {
	if ps == nil {
		return ""
	}

	filtered := make([]narrativeEntry, 0, len(detections))
	for _, det := range detections {
		if det.Narrative == nil {
			continue
		}
		raw, n, tier := det.Narrative.evaluate(ps, det)
		if tier > 0 {
			filtered = append(filtered, narrativeEntry{det: det, raw: raw, n: n, tier: tier})
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].tier > filtered[j].tier })

	sentences := make([]string, 0, len(filtered)+2)
	blatantIdx, strongIdx, mildIdx := 0, 0, 0
	for _, e := range filtered {
		var adj string
		switch e.tier {
		case 3:
			adj = narrativeBlatantAdjs[blatantIdx%len(narrativeBlatantAdjs)]
			blatantIdx++
		case 2:
			adj = narrativeStrongAdjs[strongIdx%len(narrativeStrongAdjs)]
			strongIdx++
		default:
			adj = narrativeMildAdjs[mildIdx%len(narrativeMildAdjs)]
			mildIdx++
		}
		if s := e.det.Narrative.Sentence(e.raw, e.n, adj); s != "" {
			sentences = append(sentences, s)
		}
	}

	// Closing sentences for modifiers that fired. Co-occurrence subsumes
	// evidence stacking narratively, so don't repeat both.
	switch {
	case psHasYes(ps, Key("sniper_wallbang_override")) || psHasYes(ps, Key("scout_precision_override")):
		sentences = append(sentences, "A sniper-anomaly override pinned likelihood to 100%.")
	case psHasYes(ps, Key("wallhack_co_occurrence_boost")):
		sentences = append(sentences, "The wallhack co-occurrence pattern triggered — both pre-FOV pre-aim AND elevated back-kill-given rate together, the wallhack-via-info signature.")
	case psHasYes(ps, Key("evidence_stacking_boost")):
		sentences = append(sentences, "Multiple strong channels co-occur, triggering the evidence-stacking boost.")
	}
	if psHasYes(ps, Key("ttd_sub100_high_floor")) {
		sentences = append(sentences, "The sub-100 ms time-to-damage floor enforced a 55% minimum likelihood.")
	}
	if psHasYes(ps, Key("wingman_boost")) {
		sentences = append(sentences, "A Wingman-match KPR boost was applied to reflect the short-format pace.")
	}

	if len(sentences) == 0 {
		return "No suspicious signals registered across the evaluated channels."
	}
	return strings.Join(sentences, " ")
}

func psHasYes(ps *PlayerStats, k Key) bool {
	m, ok := ps.GetMetric(cheatscoreCategoryAntiCheat, k)
	return ok && m.StringValue == "Yes"
}
