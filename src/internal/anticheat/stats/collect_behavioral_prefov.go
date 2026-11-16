package stats

import "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"

// recordPreFOV measures the killer's crosshair angle to the victim 200 ms
// before the victim first entered the killer's FOV.
//
// Walk the killer's history backward from the kill tick to find when the
// victim FIRST entered the killer's FOV, then look further back by
// preFOVLookbackMs and measure the angle. The victim's position at that tick
// is taken from the victim's own history when available, otherwise the kill
// position is used (a small error for short engagements).
func (bc *BehavioralCollector) recordPreFOV(e events.Kill, demoStats *DemoStats) {
	killerID := e.Killer.SteamID64
	victimID := e.Victim.SteamID64

	killerHistory := bc.history[killerID]
	if len(killerHistory) < 2 {
		return
	}

	victimHistory := bc.history[victimID]
	victimByTick := make(map[int]playerSnapshot, len(victimHistory))
	for _, s := range victimHistory {
		victimByTick[s.tick] = s
	}

	fovEntryIdx := -1
	for i := len(killerHistory) - 1; i >= 0; i-- {
		ks := killerHistory[i]
		vs, ok := victimByTick[ks.tick]
		if !ok {
			continue
		}
		viewVec := viewDirectionToVector(ks.yawX, ks.pitch)
		ang := angleBetweenViewAndTarget(viewVec, ks.posX, ks.posY, ks.posZ, vs.posX, vs.posY, vs.posZ)
		if ang >= fovEntryDegrees {
			// First tick going backward where the victim is OUT of FOV; the
			// frame after this one is the FOV entry from the killer's POV.
			fovEntryIdx = i + 1
			break
		}
	}
	if fovEntryIdx <= 0 || fovEntryIdx >= len(killerHistory) {
		return
	}

	tickRate := bc.tickRate
	if tickRate <= 0 {
		tickRate = 64.0
	}
	lookbackTicks := int(preFOVLookbackMs * tickRate / 1000.0)
	targetTick := killerHistory[fovEntryIdx].tick - lookbackTicks

	var ks, vs playerSnapshot
	foundK := false
	for i := fovEntryIdx; i >= 0; i-- {
		if killerHistory[i].tick <= targetTick {
			ks = killerHistory[i]
			foundK = true
			break
		}
	}
	if !foundK {
		return
	}
	v, ok := victimByTick[ks.tick]
	if !ok {
		return
	}
	vs = v

	viewVec := viewDirectionToVector(ks.yawX, ks.pitch)
	preFOVAngle := angleBetweenViewAndTarget(viewVec, ks.posX, ks.posY, ks.posZ, vs.posX, vs.posY, vs.posZ)
	bc.preFOVAngles[killerID] = append(bc.preFOVAngles[killerID], preFOVAngle)

	// Report integration: record the pre-aim sample as a timestamped cue.
	// The tick is when the victim entered the killer's FOV, so the reviewer can
	// watch the approach rather than only the shot. Scoring is unchanged.
	if preFOVAngle <= detPreFOV.Evidence.Threshold {
		if ps := demoStats.GetOrCreatePlayerStats(e.Killer); ps != nil {
			ps.AddMoment(EvidenceMoment{
				Kind:     "preaim",
				Tick:     killerHistory[fovEntryIdx].tick,
				Value:    preFOVAngle,
				Unit:     "°",
				Opponent: e.Victim.Name,
				Weapon:   getWeaponName(e.Weapon),
				Detail:   "crosshair angle 200 ms before the enemy became visible",
			})
		}
	}
}

// publishPreFOV writes the median pre-FOV angle and sample count.
func (bc *BehavioralCollector) publishPreFOV(ps *PlayerStats, sid uint64) {
	if angles := bc.preFOVAngles[sid]; len(angles) >= minPreFOVSamples {
		med := median(angles)
		ps.AddMetric(behavioralCategory, Key("pre_fov_aim_median_deg"), Metric{
			Type:        MetricFloat,
			FloatValue:  med,
			Description: "Median angle (deg) between killer view and victim position 200 ms before FOV entry (low = suspicious)",
		})
		ps.AddMetric(behavioralCategory, Key("pre_fov_aim_samples"), Metric{
			Type:        MetricInteger,
			IntValue:    int64(len(angles)),
			Description: "Number of kills contributing to pre-FOV pre-aim metric",
		})
	}
}
