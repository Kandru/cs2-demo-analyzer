package stats

import "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"

// collectAttention records, for each alive player, the smallest angle from
// their view direction to any alive enemy. Only frames where no enemy is
// inside fovEntryDegrees are recorded, so this measures attention drift rather
// than active engagements.
func (bc *BehavioralCollector) collectAttention(playing []*common.Player) {
	for _, attacker := range playing {
		if attacker == nil || attacker.SteamID64 == 0 || !attacker.IsAlive() {
			continue
		}
		viewVec := viewDirectionToVector(float64(attacker.ViewDirectionX()), float64(attacker.ViewDirectionY()))
		attackerPos := attacker.Position()

		minAngle := 180.0
		for _, opponent := range playing {
			if opponent == nil || opponent.SteamID64 == 0 || !opponent.IsAlive() {
				continue
			}
			if opponent.Team == attacker.Team || opponent.SteamID64 == attacker.SteamID64 {
				continue
			}
			oppPos := opponent.Position()
			ang := angleBetweenViewAndTarget(viewVec, attackerPos.X, attackerPos.Y, attackerPos.Z, oppPos.X, oppPos.Y, oppPos.Z)
			if ang < minAngle {
				minAngle = ang
			}
		}

		if minAngle < fovEntryDegrees {
			continue
		}
		bc.attentionMin[attacker.SteamID64] = append(bc.attentionMin[attacker.SteamID64], minAngle)
	}
}

// publishAttention writes the median nearest-enemy angle and sample count.
func (bc *BehavioralCollector) publishAttention(ps *PlayerStats, sid uint64) {
	if angles := bc.attentionMin[sid]; len(angles) >= minAttentionSamples {
		med := median(angles)
		ps.AddMetric(behavioralCategory, Key("nearest_enemy_angle_median_deg"), Metric{
			Type:        MetricFloat,
			FloatValue:  med,
			Description: "Median per-frame angle (deg) from view direction to nearest enemy when not in FOV (low = suspicious)",
		})
		ps.AddMetric(behavioralCategory, Key("nearest_enemy_angle_samples"), Metric{
			Type:        MetricInteger,
			IntValue:    int64(len(angles)),
			Description: "Number of frames contributing to nearest-enemy attention metric",
		})
	}
}
