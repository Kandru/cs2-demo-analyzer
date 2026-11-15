// Package analysis post-processes a parsed match with data from the
// anti-cheat detector and recomputes match-level summaries.
package analysis

import (
	"fmt"

	"github.com/Kandru/cs2-demo-analyzer/internal/anticheat"
	"github.com/Kandru/cs2-demo-analyzer/internal/model"
)

// Options controls the post-processing step.
type Options struct {
	ToolVersion string
	Cheat       anticheat.Report
}

// Apply merges anti-cheat verdicts into the match and refreshes the summary.
// Anti-cheat failures are recorded on every player report and never abort the
// battlelog report.
func Apply(m *model.Match, opts Options) {
	if m == nil {
		return
	}
	m.ToolVersion = opts.ToolVersion
	if opts.Cheat.Err != nil || len(opts.Cheat.Players) == 0 {
		errText := ""
		if opts.Cheat.Err != nil {
			errText = opts.Cheat.Err.Error()
		} else {
			errText = "anti-cheat detector returned no player data"
		}
		for _, p := range m.Players {
			p.Cheat = &model.CheatReport{Available: false, Error: errText}
		}
	} else {
		for _, p := range m.Players {
			if p.SteamID64 == 0 {
				continue
			}
			if cr, ok := opts.Cheat.Players[p.SteamID64]; ok {
				p.Cheat = cr
			} else {
				p.Cheat = &model.CheatReport{Available: false, Error: "player not present in anti-cheat results"}
			}
		}
	}
	RefreshSummary(m)
}

// RefreshSummary recomputes summary values from the current match state.
func RefreshSummary(m *model.Match) {
	if m == nil {
		return
	}
	regulation := 24
	if m.Mode == "Wingman" {
		regulation = 16
	}
	sum := &model.Summary{
		Rounds:           len(m.Rounds),
		RegulationRounds: regulation,
	}
	official := 0
	for _, r := range m.Rounds {
		if !r.Pregame {
			official++
		}
	}
	sum.Rounds = official
	for i, t := range m.Teams {
		if i < 2 {
			sum.Score[i] = t.Score
		}
	}
	winIdx := 0
	if len(m.Teams) >= 2 && m.Teams[1].Score > m.Teams[0].Score {
		winIdx = 1
	}
	if len(m.Teams) > winIdx && (m.Teams[0].Score != m.Teams[1].Score) {
		sum.WinnerTeamID = m.Teams[winIdx].ID
		sum.WinnerName = m.Teams[winIdx].Name
	}

	for _, p := range m.Players {
		if p.Stats.Kills > sum.TopFraggerKills {
			sum.TopFraggerID = p.ID
			sum.TopFraggerName = p.Name
			sum.TopFraggerKills = p.Stats.Kills
		}
		if p.Aim != nil && p.Aim.Rating > sum.BestAimRating && !p.IsBot {
			sum.BestAimID = p.ID
			sum.BestAimName = p.Name
			sum.BestAimRating = p.Aim.Rating
			sum.BestAimGrade = p.Aim.Grade
		}
		if p.Cheat != nil && p.Cheat.Flagged {
			sum.FlaggedPlayers++
		}
	}

	for _, t := range m.Teams {
		ts := model.TeamStats{TeamID: t.ID, Name: t.Name, Side: string(t.StartSide), Score: t.Score}
		var aimSum float64
		var aimN, hsKills int
		rounds := 0
		for _, p := range m.Players {
			if p.TeamID != t.ID {
				continue
			}
			ts.Kills += p.Stats.Kills
			ts.Deaths += p.Stats.Deaths
			ts.Assists += p.Stats.Assists
			ts.Damage += p.Stats.Damage
			hsKills += p.Stats.HeadshotKills
			if p.Aim != nil && p.Aim.Rating > 0 && !p.IsBot {
				aimSum += p.Aim.Rating
				aimN++
			}
			if p.Stats.Rounds > rounds {
				rounds = p.Stats.Rounds
			}
		}
		if rounds > 0 {
			ts.ADR = float64(ts.Damage) / float64(rounds)
		}
		if ts.Kills > 0 {
			ts.HSPercent = float64(hsKills) / float64(ts.Kills) * 100
		}
		if aimN > 0 {
			ts.AimRating = aimSum / float64(aimN)
		}
		sum.TeamStats = append(sum.TeamStats, ts)
	}
	m.Summary = sum
}

// Validate performs cheap consistency checks and appends warnings.
func Validate(m *model.Match) {
	if m == nil {
		return
	}
	if m.Map == "" {
		m.Warnings = append(m.Warnings, "map name missing from demo header")
		m.Map = "unknown"
	}
	if len(m.Rounds) == 0 {
		m.Warnings = append(m.Warnings, "no official rounds found; only warmup/pregame activity is shown")
	}
	if len(m.Teams) < 2 {
		m.Warnings = append(m.Warnings, "fewer than two teams detected; scoreboard may be incomplete")
	}
	for _, r := range m.Rounds {
		if r.EndTick < r.StartTick {
			m.Warnings = append(m.Warnings, fmt.Sprintf("round %d has an invalid tick range", r.Number))
		}
	}
}
