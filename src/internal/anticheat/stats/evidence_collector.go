package stats

import (
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

// This file is NOT part of the upstream demo-anticheat project. It records
// timestamped review cues for the battlelog report. The thresholds match the
// starting point of the corresponding cheat-score channel ramps, so a cue is
// "where the detector started caring" — never proof of cheating.
//
// The other collectors append their own cues at the point where the signal is
// calculated (snap, reaction, pre-FOV, recoil). This collector covers kill-feed
// anomalies that have no dedicated score channel.

const (
	// evidenceSnapVelocity is the snap channel's ramp start (see
	// cheatscore_channels.go evaluateSnap).
	evidenceSnapVelocity = 2.0 // °/ms
	// evidenceReactionMs is the reaction channel's "suspicious" start.
	evidenceReactionMs = 250.0 // ms
	// evidenceSub100Ms is the sub-100 ms floor threshold.
	evidenceSub100Ms = 100.0 // ms
	// evidencePreFOVAngle is the pre-FOV channel's ramp start.
	evidencePreFOVAngle = 12.0 // °
	// evidenceRecoilError is the recoil collector's "perfect" threshold.
	evidenceRecoilError = 0.30 // °
)

// EvidenceCollector records rare kill-feed anomalies (wallbangs, smoke kills,
// no-scopes) as timestamped moments. It never changes the score.
type EvidenceCollector struct {
	*BaseCollector
}

// NewEvidenceCollector creates an EvidenceCollector.
func NewEvidenceCollector() *EvidenceCollector {
	return &EvidenceCollector{
		BaseCollector: NewBaseCollector("Evidence Moments", Category("evidence")),
	}
}

// Setup registers the kill handler.
func (ec *EvidenceCollector) Setup(parser demoinfocs.Parser, demoStats *DemoStats) {
	parser.RegisterEventHandler(func(e events.Kill) {
		if e.Killer == nil || e.Victim == nil || e.Killer == e.Victim {
			return
		}
		if e.Killer.Team == e.Victim.Team {
			return
		}
		ps := demoStats.GetOrCreatePlayerStats(e.Killer)
		if ps == nil {
			return
		}
		tick := parser.CurrentFrame()
		weapon := getWeaponName(e.Weapon)
		victim := e.Victim.Name

		if e.IsWallBang() {
			ps.AddMoment(EvidenceMoment{
				Kind:     "wallbang",
				Tick:     tick,
				Value:    float64(e.PenetratedObjects),
				Unit:     "penetrations",
				Opponent: victim,
				Weapon:   weapon,
				Detail:   "kill through world geometry",
			})
		}
		if e.ThroughSmoke {
			ps.AddMoment(EvidenceMoment{
				Kind:     "smoke",
				Tick:     tick,
				Opponent: victim,
				Weapon:   weapon,
				Detail:   "kill through smoke",
			})
		}
		if e.NoScope {
			ps.AddMoment(EvidenceMoment{
				Kind:     "noscope",
				Tick:     tick,
				Opponent: victim,
				Weapon:   weapon,
				Detail:   "no-scope kill",
			})
		}
	})
}
