package stats

// registry.go is the ordered catalogue of detections and modifiers. Order is
// significant:
//
//   - detections: report/publish order and narrative tie-break order.
//   - modifiers: the exact order their multipliers/floors are applied.
//
// The registry is immutable after package initialisation. To add a feature,
// create detect_<id>.go or mod_<id>.go, register it here, add
// docs/detections/<id>.md, and run the tests.

var detections = []*Detection{
	detHS,
	detSnap,
	detReaction,
	detTTDSub100,
	detRecoil,
	detPreFOV,
	detPreFOVPresence,
	detAttention,
	detBackKilled,
	detDecoupling,
}

var modifiers = []*Modifier{
	modWingman,
	modCompetitive,
	modPosition,
	modEvidenceStacking,
	modCooccurrence,
	modFloor,
	modSniper,
}

// evaluateChannelsForPlayer runs every lobby-independent detection.
func evaluateChannelsForPlayer(ps *PlayerStats) []Channel {
	out := make([]Channel, 0, len(detections))
	for _, d := range detections {
		if d.Evaluate != nil {
			out = append(out, d.Evaluate(ps))
		}
	}
	return out
}
