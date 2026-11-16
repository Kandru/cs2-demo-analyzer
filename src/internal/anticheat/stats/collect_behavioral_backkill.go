package stats

import "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"

// recordBackKill updates both sides of the back-kill measurement.
//
// Was the victim looking away from the killer at the moment of death? The same
// answer is evidence in opposite directions:
//   - victim side: high rate = unaware = clean; low rate = always-aware =
//     wallhack-suspicious.
//   - killer side: high rate = kills from behind = flanking style or info
//     exploit. Tracked as a diagnostic only; it feeds the wallhack
//     co-occurrence modifier rather than the scored channel.
func (bc *BehavioralCollector) recordBackKill(e events.Kill) {
	killerID := e.Killer.SteamID64
	victimID := e.Victim.SteamID64

	killerPos := e.Killer.Position()
	victimPos := e.Victim.Position()
	victimView := viewDirectionToVector(float64(e.Victim.ViewDirectionX()), float64(e.Victim.ViewDirectionY()))
	ang := angleBetweenViewAndTarget(victimView, victimPos.X, victimPos.Y, victimPos.Z, killerPos.X, killerPos.Y, killerPos.Z)

	bc.backKillTotal[victimID]++
	bc.backKillGivenTotal[killerID]++
	if ang >= backKillThresholdDeg {
		bc.backKillBack[victimID]++
		bc.backKillGivenBack[killerID]++
	}
}

// publishBackKill writes the victim-side rate and the killer-side diagnostic.
func (bc *BehavioralCollector) publishBackKill(ps *PlayerStats, sid uint64) {
	if total := bc.backKillTotal[sid]; total >= minBackKillSamples {
		rate := float64(bc.backKillBack[sid]) / float64(total)
		ps.AddMetric(channelCategoryBehavioral, Key("back_killed_pct"), Metric{
			Type:        MetricPercentage,
			FloatValue:  rate * 100.0,
			Description: "Percent of own deaths where this player was looking away from the killer (low = suspicious)",
		})
		ps.AddMetric(channelCategoryBehavioral, Key("back_killed_total_deaths"), Metric{
			Type:        MetricInteger,
			IntValue:    int64(total),
			Description: "Total deaths used for back-kill rate",
		})
	}

	if total := bc.backKillGivenTotal[sid]; total >= minBackKillSamples {
		back := bc.backKillGivenBack[sid]
		rate := float64(back) / float64(total)
		ps.AddMetric(channelCategoryBehavioral, Key("back_kill_given_pct"), Metric{
			Type:        MetricPercentage,
			FloatValue:  rate * 100.0,
			Description: "Percent of own kills where the victim was looking away from this player (high = flanking or info exploit)",
		})
		ps.AddMetric(channelCategoryBehavioral, Key("back_kill_given_count"), Metric{
			Type:        MetricInteger,
			IntValue:    int64(back),
			Description: "Number of own kills where the victim was looking away",
		})
		ps.AddMetric(channelCategoryBehavioral, Key("back_kill_given_total_kills"), Metric{
			Type:        MetricInteger,
			IntValue:    int64(total),
			Description: "Total kills used for back-kill-given rate",
		})
	}
}
