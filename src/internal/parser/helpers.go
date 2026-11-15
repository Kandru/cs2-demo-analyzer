package parser

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"

	"github.com/Kandru/cs2-demo-analyzer/internal/model"
)

func sideString(t common.Team) model.Side {
	switch t {
	case common.TeamTerrorists:
		return model.SideT
	case common.TeamCounterTerrorists:
		return model.SideCT
	case common.TeamSpectators:
		return model.SideSpectator
	case common.TeamUnassigned:
		return model.SideUnassigned
	default:
		return ""
	}
}

func reasonString(r events.RoundEndReason) string {
	switch r {
	case events.RoundEndReasonTargetBombed:
		return "Bomb exploded"
	case events.RoundEndReasonVIPEscaped:
		return "VIP escaped"
	case events.RoundEndReasonVIPKilled:
		return "VIP killed"
	case events.RoundEndReasonTerroristsEscaped:
		return "Terrorists escaped"
	case events.RoundEndReasonCTStoppedEscape:
		return "CTs stopped escape"
	case events.RoundEndReasonTerroristsStopped:
		return "Terrorists stopped"
	case events.RoundEndReasonBombDefused:
		return "Bomb defused"
	case events.RoundEndReasonCTWin:
		return "CTs eliminated the T side"
	case events.RoundEndReasonTerroristsWin:
		return "Terrorists eliminated the CT side"
	case events.RoundEndReasonDraw:
		return "Draw"
	case events.RoundEndReasonHostagesRescued:
		return "Hostages rescued"
	case events.RoundEndReasonTargetSaved:
		return "Bomb saved"
	case events.RoundEndReasonHostagesNotRescued:
		return "Hostages not rescued"
	case events.RoundEndReasonTerroristsNotEscaped:
		return "Terrorists did not escape"
	case events.RoundEndReasonVIPNotEscaped:
		return "VIP did not escape"
	case events.RoundEndReasonGameStart:
		return "Game start"
	case events.RoundEndReasonTerroristsSurrender:
		return "Terrorists surrendered"
	case events.RoundEndReasonCTSurrender:
		return "CTs surrendered"
	case events.RoundEndReasonTerroristsPlanted:
		return "Terrorists planted the bomb"
	case events.RoundEndReasonCTsReachedHostage:
		return "CTs reached the hostage"
	default:
		return "Round ended"
	}
}

func buyType(equip, round int) string {
	if round <= 1 {
		return "Pistol"
	}
	switch {
	case equip >= 4000:
		return "Full buy"
	case equip >= 2000:
		return "Force buy"
	default:
		return "Eco"
	}
}

func detectMode(players int) string {
	switch {
	case players == 0:
		return "Unknown"
	case players <= 4:
		return "Wingman"
	case players == 10:
		return "Competitive"
	default:
		return "Custom"
	}
}

func matchID(name, mapName string, fileTime time.Time) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%d", name, mapName, fileTime.UnixNano())))
	return fmt.Sprintf("%x", h[:6])
}

func slugify(s string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_' || r == '.':
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
		if b.Len() >= 40 {
			break
		}
	}
	return strings.Trim(b.String(), "-")
}

func appendUnique(xs []string, v string) []string {
	if v == "" {
		return xs
	}
	for _, x := range xs {
		if x == v {
			return xs
		}
	}
	return append(xs, v)
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func sortPlayers(ps []*model.Player) []*model.Player {
	sort.SliceStable(ps, func(i, j int) bool {
		if ps[i].TeamID != ps[j].TeamID {
			return ps[i].TeamID < ps[j].TeamID
		}
		if ps[i].Stats.Kills != ps[j].Stats.Kills {
			return ps[i].Stats.Kills > ps[j].Stats.Kills
		}
		return ps[i].Name < ps[j].Name
	})
	return ps
}

func finalizePlayerStats(p *model.Player) {
	if p == nil {
		return
	}
	st := &p.Stats
	if st.Rounds <= 0 {
		st.ADR = 0
	} else {
		st.ADR = float64(st.Damage) / float64(st.Rounds)
	}
	if st.Deaths > 0 {
		st.KD = float64(st.Kills) / float64(st.Deaths)
		st.KDR = float64(st.Kills+st.Assists) / float64(st.Deaths)
	} else if st.Kills > 0 {
		st.KD = float64(st.Kills)
		st.KDR = float64(st.Kills)
	}
	if st.Kills > 0 {
		st.HSPercent = float64(st.HeadshotKills) / float64(st.Kills) * 100
	}
	if st.Rounds > 0 {
		st.KASTPercent = float64(st.KASTRounds) / float64(st.Rounds) * 100
	}
	if st.Shots > 0 {
		st.AccuracyPct = float64(st.Hits) / float64(st.Shots) * 100
	}
	if st.Hits > 0 {
		st.HeadAccuracyPct = float64(st.HeadHits) / float64(st.Hits) * 100
	}
	for _, ws := range p.Weapons {
		if ws.Shots > 0 {
			ws.Accuracy = float64(ws.Hits) / float64(ws.Shots) * 100
		}
	}
	for _, ss := range p.SideStats {
		if ss.Rounds > 0 {
			ss.ADR = float64(ss.Damage) / float64(ss.Rounds)
			ss.HSPercent = percentOf(ss.HeadshotKills, ss.Kills)
			ss.KASTPercent = 0
		}
	}
	p.RoundsPlayed = st.Rounds
}

func percentOf(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b) * 100
}

func buildSummary(m *model.Match) *model.Summary {
	sum := &model.Summary{
		Rounds:           len(m.Rounds),
		RegulationRounds: regulationRounds(len(m.Players)),
	}
	for _, r := range m.Rounds {
		if !r.Pregame {
			sum.Rounds = countOfficial(m.Rounds)
			break
		}
	}
	sum.Score = [2]int{}
	for i, t := range m.Teams {
		if i < 2 {
			sum.Score[i] = t.Score
		}
	}
	winIdx := 0
	if len(m.Teams) >= 2 && m.Teams[1].Score > m.Teams[0].Score {
		winIdx = 1
	}
	if len(m.Teams) > winIdx {
		sum.WinnerTeamID = m.Teams[winIdx].ID
		sum.WinnerName = m.Teams[winIdx].Name
	}

	for _, p := range m.Players {
		if p.Stats.Kills > sum.TopFraggerKills {
			sum.TopFraggerID = p.ID
			sum.TopFraggerName = p.Name
			sum.TopFraggerKills = p.Stats.Kills
		}
		if p.Aim != nil && p.Aim.Rating > sum.BestAimRating {
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
		var aimN int
		for _, p := range m.Players {
			if p.TeamID != t.ID {
				continue
			}
			ts.Kills += p.Stats.Kills
			ts.Deaths += p.Stats.Deaths
			ts.Assists += p.Stats.Assists
			ts.Damage += p.Stats.Damage
			if p.Aim != nil && p.Aim.Rating > 0 {
				aimSum += p.Aim.Rating
				aimN++
			}
		}
		rounds := 0
		for _, p := range m.Players {
			if p.TeamID == t.ID && p.Stats.Rounds > rounds {
				rounds = p.Stats.Rounds
			}
		}
		if rounds > 0 {
			ts.ADR = float64(ts.Damage) / float64(rounds)
		}
		if aimN > 0 {
			ts.AimRating = aimSum / float64(aimN)
		}
		sum.TeamStats = append(sum.TeamStats, ts)
	}
	return sum
}

func regulationRounds(players int) int {
	if players > 0 && players <= 4 {
		return 8
	}
	return 24
}

func countOfficial(rounds []*model.Round) int {
	n := 0
	for _, r := range rounds {
		if !r.Pregame {
			n++
		}
	}
	return n
}
