package stats

import "fmt"

const (
	hsID     = "hs"
	hsWeight = 0.18
	hsMode   = ModePositiveOnly
)

// evaluateHS scores headshot percentage. Ramp 55%→75%, n_full=20.
//
// Positive-only: a high HS% on many kills is real cheat signal, but a low
// HS% is not exculpatory — supports, awpers, and post-plant playstyles all
// produce low HS rates without any cheat assistance, and wallhack-only
// cheaters (info but no aim assistance) have normal HS rates. A bidirectional
// mode would drown out legitimate wallhack signals from pre_fov/decoupling/
// back_killed.
func evaluateHS(ps *PlayerStats) Channel {
	totalKills, hasKills := psGetInt(ps, channelCategoryKills, Key("total_kills"))
	if !hasKills || totalKills <= 0 {
		return Channel{ID: hsID, Weight: hsWeight, Mode: hsMode}
	}
	hsPct, _ := psGetFloat(ps, channelCategoryKills, Key("headshot_percentage"))
	score := linearScore(hsPct, 55.0, 75.0)
	return Channel{
		ID:         hsID,
		Score:      score,
		Confidence: linearConfidence(totalKills, 20),
		Raw:        hsPct,
		SampleN:    totalKills,
		Weight:     hsWeight,
		Zone:       zoneFor(score),
		Mode:       hsMode,
		HasData:    true,
	}
}

func rawHS(ps *PlayerStats) (float64, int, bool) {
	value, ok := psGetFloat(ps, channelCategoryKills, Key("headshot_percentage"))
	n, _ := psGetInt(ps, channelCategoryKills, Key("total_kills"))
	return value, int(n), ok
}

var detHS = &Detection{
	ID:          "hs",
	Label:       "Headshot rate",
	Unit:        "%",
	Description: "Share of kills that were headshots.",
	Doc:         "docs/detections/hs.md",
	Weight:      0.18,
	Mode:        ModePositiveOnly,
	Evaluate:    evaluateHS,
	Raw:         rawHS,
	Narrative: &NarrativeSpec{
		MinSamples: 10,
		Ascending:  true,
		Mild:       55.0,
		Strong:     65.0,
		Blatant:    75.0,
		Sentence: func(raw float64, n int64, adj string) string {
			return fmt.Sprintf("Headshot rate of %.0f%% over %d kills is %s elevated.", raw, n, adj)
		},
	},
}
