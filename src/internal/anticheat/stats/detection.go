package stats

// detection.go defines the metadata and value types shared by every cheat
// detection. A detection owns its measurement metadata, scoring ramp, weight,
// narrative thresholds and evidence cue, so adding or changing one signal only
// requires a new detect_*.go file plus its docs/detections/<id>.md page.

import "math"

// Zone is a human-readable band for a channel score, used by the transparency
// reporting layer. The combiner consumes Score, not Zone.
type Zone uint8

const (
	ZoneNoData Zone = iota
	ZoneClean
	ZoneMild
	ZoneStrong
	ZoneBlatant
)

// String returns the lowercase label published in the anti_cheat metrics.
func (z Zone) String() string {
	switch z {
	case ZoneClean:
		return "clean"
	case ZoneMild:
		return "mild"
	case ZoneStrong:
		return "strong"
	case ZoneBlatant:
		return "blatant"
	default:
		return "no_data"
	}
}

// Mode controls how a channel's contribution enters the combiner.
//
//   - ModeBidirectional: a clean reading produces negative log-odds, i.e.
//     genuine evidence the player is not cheating. Use for graded metrics
//     where below-clean values legitimately exonerate.
//   - ModePositiveOnly: a clean reading contributes 0; only suspicious
//     readings shift the score upward. Use for detector-style metrics where a
//     clean value proves nothing.
type Mode uint8

const (
	ModeBidirectional Mode = iota
	ModePositiveOnly
)

// Channel is one evaluated anti-cheat signal for one player.
type Channel struct {
	ID         string
	Score      float64 // [0, 1] — 0 clean, 1 blatant
	Confidence float64 // [0, 1] — 0 no data, 1 fully trustworthy
	Raw        float64 // underlying measurement (for publish transparency)
	SampleN    int64
	Weight     float64
	Zone       Zone
	Mode       Mode
	HasData    bool
}

// RawFunc extracts the raw measurement and sample count shown in the report.
// ok is false when the underlying metric does not exist.
type RawFunc func(*PlayerStats) (value float64, samples int, ok bool)

// EvaluateFunc scores one player independent of the rest of the lobby.
type EvaluateFunc func(*PlayerStats) Channel

// LobbyEvaluateFunc scores one player using lobby context. Detections that
// need lobby-relative behavior (pre_fov_presence) use this instead of Evaluate.
type LobbyEvaluateFunc func(*LobbyContext, *PlayerStats, uint64) Channel

// Detection is one independent cheat signal.
type Detection struct {
	ID          string
	Label       string
	Unit        string
	Description string
	Doc         string // repo-relative documentation path
	Weight      float64
	Mode        Mode

	Evaluate      EvaluateFunc
	LobbyEvaluate LobbyEvaluateFunc
	Raw           RawFunc
	Narrative     *NarrativeSpec
	Evidence      *EvidenceSpec
}

// LobbyContext is the shared lobby data used by lobby-relative detections.
type LobbyContext struct {
	SamplesBySID    map[uint64]int64
	AsymmetricBySID map[uint64]bool
}

// IsAsymmetric reports whether the rest of the lobby is asymmetric in pre-FOV
// samples for this player.
func (c *LobbyContext) IsAsymmetric(sid uint64) bool {
	return c != nil && c.AsymmetricBySID[sid]
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// linearScore maps a raw measurement onto [0, 1] using a linear ramp anchored
// at cleanX (=0) and blatantX (=1). Pass cleanX < blatantX for "higher is
// worse" metrics, cleanX > blatantX for "lower is worse" metrics.
func linearScore(raw, cleanX, blatantX float64) float64 {
	if cleanX == blatantX {
		return 0
	}
	return clamp01((raw - cleanX) / (blatantX - cleanX))
}

func linearConfidence(n int64, nFull int) float64 {
	if nFull <= 0 || n <= 0 {
		return 0
	}
	return clamp01(float64(n) / float64(nFull))
}

// sqrtConfidence returns sqrt(n / n_full) clamped to [0, 1]. Used for
// heavy-tailed signals where a small number of samples already carries
// meaningful evidence.
func sqrtConfidence(n int64, nFull int) float64 {
	if nFull <= 0 || n <= 0 {
		return 0
	}
	return clamp01(math.Sqrt(float64(n) / float64(nFull)))
}

func zoneFor(score float64) Zone {
	switch {
	case score < 0.25:
		return ZoneClean
	case score < 0.50:
		return ZoneMild
	case score < 0.80:
		return ZoneStrong
	default:
		return ZoneBlatant
	}
}

// --- PlayerStats metric helpers ------------------------------------------

func psGetFloat(ps *PlayerStats, cat Category, key Key) (float64, bool) {
	if m, ok := ps.GetMetric(cat, key); ok {
		return m.FloatValue, true
	}
	return 0, false
}

func psGetInt(ps *PlayerStats, cat Category, key Key) (int64, bool) {
	if m, ok := ps.GetMetric(cat, key); ok {
		return m.IntValue, true
	}
	return 0, false
}

func psGetString(ps *PlayerStats, cat Category, key Key) (string, bool) {
	if m, ok := ps.GetMetric(cat, key); ok {
		return m.StringValue, true
	}
	return "", false
}
