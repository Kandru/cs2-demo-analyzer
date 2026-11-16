package stats

import "fmt"

// evidence.go defines how a timestamped review cue is presented and ranked.
// Each detection owns its cue metadata; this file only builds the lookup used
// by AddMoment and the report adapter. Cues are reviewer navigation aids and
// never feed the score.

// EvidenceSpec describes one cue kind.
type EvidenceSpec struct {
	Kind      string
	Label     string
	Rank      int // 1 = info, 2 = mild, 3 = strong
	Unit      string
	Threshold float64
	// Severity orders cues within a kind when a cap is full. Higher is worse.
	Severity func(value float64) float64
	// ValueText formats the value shown in the report.
	ValueText func(value float64) string
}

// extraEvidence covers kill-feed anomalies that do not have their own score
// channel. They are recorded by EvidenceCollector.
var extraEvidence = []*EvidenceSpec{
	{
		Kind:     "wallbang",
		Label:    "Wallbang",
		Rank:     3,
		Unit:     "penetrations",
		Severity: func(v float64) float64 { return 20 + v },
		ValueText: func(v float64) string {
			if v > 1 {
				return fmt.Sprintf("%.0f layers", v)
			}
			return "1 layer"
		},
	},
	{
		Kind:      "smoke",
		Label:     "Through smoke",
		Rank:      2,
		Severity:  func(v float64) float64 { return 15 },
		ValueText: func(v float64) string { return "kill" },
	},
	{
		Kind:      "noscope",
		Label:     "No-scope",
		Rank:      2,
		Severity:  func(v float64) float64 { return 14 },
		ValueText: func(v float64) string { return "kill" },
	},
}

// evidenceByKind is immutable after package initialisation.
var evidenceByKind = buildEvidenceIndex()

func buildEvidenceIndex() map[string]*EvidenceSpec {
	out := make(map[string]*EvidenceSpec, len(detections)+len(extraEvidence))
	for _, d := range detections {
		if d.Evidence != nil {
			out[d.Evidence.Kind] = d.Evidence
		}
	}
	for _, spec := range extraEvidence {
		out[spec.Kind] = spec
	}
	return out
}

// evidenceSpec returns the spec for a cue kind, or nil for an unknown kind.
func evidenceSpec(kind string) *EvidenceSpec {
	return evidenceByKind[kind]
}

// momentSeverity orders cues within a kind so the cap keeps the strongest
// samples. Unknown kinds sort last.
func momentSeverity(m EvidenceMoment) float64 {
	if spec := evidenceSpec(m.Kind); spec != nil && spec.Severity != nil {
		return spec.Severity(m.Value)
	}
	return 1
}
