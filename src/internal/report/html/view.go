// Package html renders the self-contained battlelog report.
package html

import (
	"fmt"
	"html/template"
	"sort"
	"strings"
	"time"

	"github.com/Kandru/cs2-demo-analyzer/internal/model"
)

// View is the template model for one report.
type View struct {
	Title       string
	Version     string
	GeneratedAt string
	Match       *model.Match
	Teams       []*TeamView
	Players     []*PlayerView
	Rounds      []*RoundView
	Warmup      *TimelineView
	Summary     *model.Summary
	TeamStats   []model.TeamStats
	Warnings    []string
}

// TeamView groups scoreboard rows.
type TeamView struct {
	Team    *model.Team
	Players []*PlayerView
	Stats   model.TeamStats
	Win     bool
	Side    string
}

// PlayerView is the template view of one player.
type PlayerView struct {
	Player              *model.Player
	Team                *model.Team
	Anchor              string
	Initials            string
	TeamName            string
	Side                string
	Rounds              []PlayerRoundView
	Weapons             []WeaponView
	Aim                 []MetricView
	Cheat               *CheatView
	Flagged             bool
	KDDisplay           string
	ADRDisplay          string
	HSDisplay           string
	AimDisplay          string
	AimGrade            string
	ScoreDisplay        string
	AccuracyDisplay     string
	HeadAccuracyDisplay string
	CounterDisplay      string
	SprayDisplay        string
	DamageSpeedDisplay  string
	ReactionDisplay     string
	PreaimDisplay       string
}

// PlayerRoundView is one row in the per-round table.
type PlayerRoundView struct {
	Number      int
	Side        string
	Kills       int
	Deaths      int
	Assists     int
	Damage      int
	EquipValue  int
	Result      string
	ResultClass string
	KAST        bool
	KD          string
}

// WeaponView is one row in the weapon table.
type WeaponView struct {
	Name      string
	Kills     int
	Headshots int
	Damage    int
	Shots     int
	Hits      int
	Accuracy  string
	HSRate    string
}

// MetricView is one aim tile.
type MetricView struct {
	ID        string
	Label     string
	Display   string
	Score     float64
	ScorePct  int
	Samples   int
	LowSample bool
	Tooltip   string
	Class     string
}

// CheatView is the anti-cheat card model.
type CheatView struct {
	Report        *model.CheatReport
	Available     bool
	Verdict       string
	VerdictClass  string
	LikelihoodPct string
	Likelihood    float64
	OverallGrade  string
	Grades        []model.Grade
	Channels      []ChannelView
	Adjustments   []model.Adjustment
	Narrative     string
	Error         string
}

// ChannelView is one detector channel row.
type ChannelView struct {
	Channel   model.CheatChannel
	ScorePct  int
	ConfPct   int
	Raw       string
	BarClass  string
	Zone      string
	ZoneClass string
}

// RoundView is one round panel.
type RoundView struct {
	Round       *model.Round
	Label       string
	Time        string
	Duration    string
	WinnerClass string
	Score       string
	ChipClass   string
	Economy     []EconomyView
	Events      []EventView
}

// EconomyView is a team economy chip.
type EconomyView struct {
	TeamName string
	Side     string
	Equip    int
	BuyType  string
	Class    string
}

// TimelineView is the warmup timeline.
type TimelineView struct {
	Duration string
	Events   []EventView
}

// EventView is a pre-rendered timeline row.
type EventView struct {
	Type        string
	Time        string
	Icon        string
	Parts       []EventPart
	Badges      []string
	DataPlayers string
	Class       string
}

// EventPart is a text fragment, optionally a player link.
type EventPart struct {
	Text string
	Href string
	Bold bool
}

// NewView builds the template model from a match.
func NewView(m *model.Match, version string) *View {
	v := &View{
		Version:     version,
		GeneratedAt: time.Now().Format("2006-01-02 15:04 MST"),
		Match:       m,
		Summary:     m.Summary,
	}
	title := m.Map
	if title == "" {
		title = "CS2"
	}
	if len(m.Teams) >= 2 {
		title += " — " + m.Teams[0].Name + " vs " + m.Teams[1].Name
	}
	v.Title = title
	if m.Summary != nil {
		v.TeamStats = m.Summary.TeamStats
	}
	v.Teams = buildTeams(m, v.TeamStats)
	v.Players = buildPlayers(m)
	v.Rounds = buildRounds(m)
	if m.Warmup != nil && len(m.Warmup.Events) > 0 {
		v.Warmup = &TimelineView{
			Duration: clock(m.Warmup.Duration),
			Events:   buildEvents(m.Warmup.Events, m),
		}
	}
	v.Warnings = m.Warnings
	return v
}

func buildTeams(m *model.Match, stats []model.TeamStats) []*TeamView {
	byID := map[int]model.TeamStats{}
	for _, s := range stats {
		byID[s.TeamID] = s
	}
	winner := 0
	if m.Summary != nil {
		winner = m.Summary.WinnerTeamID
	}
	out := make([]*TeamView, 0, len(m.Teams))
	for _, t := range m.Teams {
		tv := &TeamView{Team: t, Stats: byID[t.ID], Win: t.ID == winner, Side: string(t.StartSide)}
		for _, p := range m.Players {
			if p.TeamID != t.ID {
				continue
			}
			pv := buildPlayer(m, p)
			tv.Players = append(tv.Players, pv)
		}
		sort.SliceStable(tv.Players, func(i, j int) bool {
			if tv.Players[i].Player.Stats.Kills != tv.Players[j].Player.Stats.Kills {
				return tv.Players[i].Player.Stats.Kills > tv.Players[j].Player.Stats.Kills
			}
			return tv.Players[i].Player.Name < tv.Players[j].Player.Name
		})
		out = append(out, tv)
	}
	return out
}

func buildPlayers(m *model.Match) []*PlayerView {
	out := make([]*PlayerView, 0, len(m.Players))
	for _, p := range m.Players {
		out = append(out, buildPlayer(m, p))
	}
	return out
}

func buildPlayer(m *model.Match, p *model.Player) *PlayerView {
	pv := &PlayerView{
		Player:       p,
		Team:         m.TeamByID(p.TeamID),
		Anchor:       "player-" + p.Slug,
		Initials:     initials(p.Name),
		Flagged:      p.Cheat != nil && p.Cheat.Flagged,
		KDDisplay:    fmt.Sprintf("%.2f", p.Stats.KD),
		ADRDisplay:   fmt.Sprintf("%.1f", p.Stats.ADR),
		HSDisplay:    fmt.Sprintf("%.0f%%", p.Stats.HSPercent),
		ScoreDisplay: fmt.Sprintf("%d", p.Stats.Score),
	}
	if pv.Team != nil {
		pv.TeamName = pv.Team.Name
		pv.Side = string(pv.Team.StartSide)
	}
	for _, r := range m.Rounds {
		if r.Pregame {
			continue
		}
		rs := p.RoundStats[r.Number]
		if rs == nil {
			continue
		}
		res := "Loss"
		resClass := "loss"
		if r.Winner == rs.Side {
			res = "Win"
			resClass = "win"
		}
		pv.Rounds = append(pv.Rounds, PlayerRoundView{
			Number:      r.Number,
			Side:        string(rs.Side),
			Kills:       rs.Kills,
			Deaths:      rs.Deaths,
			Assists:     rs.Assists,
			Damage:      rs.Damage,
			EquipValue:  rs.EquipValue,
			Result:      res,
			ResultClass: resClass,
			KAST:        rs.KAST,
			KD:          ratio(rs.Kills, rs.Deaths),
		})
	}
	if p.Aim != nil {
		pv.AimDisplay = fmt.Sprintf("%.0f", p.Aim.Rating)
		pv.AimGrade = p.Aim.Grade
		for _, mt := range p.Aim.Metrics {
			pv.Aim = append(pv.Aim, metricView(mt))
			switch mt.ID {
			case "accuracy":
				pv.AccuracyDisplay = mt.Display
			case "head_accuracy":
				pv.HeadAccuracyDisplay = mt.Display
			case "counter_strafe":
				pv.CounterDisplay = mt.Display
			case "spray":
				pv.SprayDisplay = mt.Display
			case "damage_speed":
				pv.DamageSpeedDisplay = mt.Display
			case "reaction":
				pv.ReactionDisplay = mt.Display
			case "preaim":
				pv.PreaimDisplay = mt.Display
			}
		}
	}
	for _, ws := range p.Weapons {
		wv := WeaponView{
			Name:      ws.Name,
			Kills:     ws.Kills,
			Headshots: ws.Headshots,
			Damage:    ws.Damage,
			Shots:     ws.Shots,
			Hits:      ws.Hits,
		}
		if ws.Shots > 0 {
			wv.Accuracy = fmt.Sprintf("%.1f%%", ws.Accuracy)
		} else {
			wv.Accuracy = "—"
		}
		if ws.Kills > 0 {
			wv.HSRate = fmt.Sprintf("%.0f%%", float64(ws.Headshots)/float64(ws.Kills)*100)
		} else {
			wv.HSRate = "—"
		}
		pv.Weapons = append(pv.Weapons, wv)
	}
	sort.SliceStable(pv.Weapons, func(i, j int) bool {
		if pv.Weapons[i].Kills != pv.Weapons[j].Kills {
			return pv.Weapons[i].Kills > pv.Weapons[j].Kills
		}
		return pv.Weapons[i].Name < pv.Weapons[j].Name
	})
	pv.Cheat = buildCheat(p.Cheat)
	return pv
}

func metricView(mt model.AimMetric) MetricView {
	class := "empty"
	if mt.Samples > 0 {
		switch {
		case mt.Score >= 70:
			class = "good"
		case mt.Score >= 40:
			class = "mid"
		default:
			class = "low"
		}
	}
	return MetricView{
		ID:        mt.ID,
		Label:     mt.Label,
		Display:   mt.Display,
		Score:     mt.Score,
		ScorePct:  clampPct(mt.Score),
		Samples:   mt.Samples,
		LowSample: mt.LowSample,
		Tooltip:   mt.Tooltip,
		Class:     class,
	}
}

func buildCheat(cr *model.CheatReport) *CheatView {
	if cr == nil {
		return &CheatView{Available: false, Verdict: "Not analyzed", VerdictClass: "unavailable", Error: "anti-cheat detector was not run"}
	}
	cv := &CheatView{
		Report:        cr,
		Available:     cr.Available,
		OverallGrade:  cr.OverallGrade,
		Grades:        cr.Grades,
		Adjustments:   cr.Adjustments,
		Narrative:     cr.Narrative,
		Error:         cr.Error,
		Likelihood:    cr.Likelihood,
		LikelihoodPct: fmt.Sprintf("%.0f%%", cr.Likelihood),
	}
	switch {
	case !cr.Available:
		cv.Verdict = "Unavailable"
		cv.VerdictClass = "unavailable"
	case cr.Flagged:
		cv.Verdict = "Flagged"
		cv.VerdictClass = "flagged"
	case cr.Likelihood >= 25:
		cv.Verdict = "Review"
		cv.VerdictClass = "review"
	default:
		cv.Verdict = "Low signal"
		cv.VerdictClass = "clean"
	}
	for _, ch := range cr.Channels {
		cvm := ChannelView{
			Channel:   ch,
			ScorePct:  clampPct(ch.Score * 100),
			ConfPct:   clampPct(ch.Confidence * 100),
			BarClass:  zoneClass(ch.Zone),
			Zone:      ch.Zone,
			ZoneClass: zoneClass(ch.Zone),
		}
		if ch.HasData {
			cvm.Raw = rawMetric(ch)
		} else {
			cvm.Raw = "—"
		}
		cv.Channels = append(cv.Channels, cvm)
	}
	return cv
}

func zoneClass(zone string) string {
	switch strings.ToLower(zone) {
	case "clean":
		return "clean"
	case "mild":
		return "mild"
	case "strong":
		return "strong"
	case "blatant":
		return "blatant"
	default:
		return "empty"
	}
}

func rawMetric(ch model.CheatChannel) string {
	switch ch.Unit {
	case "ms":
		return fmt.Sprintf("%.0f ms", ch.Raw)
	case "°/ms":
		return fmt.Sprintf("%.2f °/ms", ch.Raw)
	case "°":
		return fmt.Sprintf("%.1f°", ch.Raw)
	case "%":
		return fmt.Sprintf("%.1f%%", ch.Raw)
	case "samples":
		return fmt.Sprintf("%.0f", ch.Raw)
	default:
		return fmt.Sprintf("%.1f", ch.Raw)
	}
}

var iconByType = map[string]string{
	"round_start":  "flag",
	"freeze_end":   "buy",
	"kill":         "kill",
	"teamkill":     "warning",
	"suicide":      "warning",
	"bomb_plant":   "bomb",
	"bomb_defuse":  "defuse",
	"bomb_explode": "explosion",
	"grenade":      "grenade",
	"mvp":          "star",
	"round_end":    "trophy",
}

func buildRounds(m *model.Match) []*RoundView {
	out := make([]*RoundView, 0, len(m.Rounds))
	for _, r := range m.Rounds {
		rv := &RoundView{
			Round:       r,
			Time:        clock(r.StartOffset),
			Duration:    clock(r.Duration),
			WinnerClass: "loss",
			Score:       fmt.Sprintf("%d : %d", r.EndScore[0], r.EndScore[1]),
		}
		if r.Pregame {
			rv.Label = "Pregame"
			rv.ChipClass = "pregame"
		} else {
			rv.Label = fmt.Sprintf("Round %d", r.Number)
			if r.Overtime {
				rv.Label += " (OT)"
			}
			rv.ChipClass = teamSideClass(r.Winner)
		}
		for _, e := range r.Economy {
			ev := EconomyView{Equip: e.EquipmentValue, BuyType: e.BuyType, Class: buyClass(e.BuyType)}
			if t := m.TeamByID(e.TeamID); t != nil {
				ev.TeamName = t.Name
			}
			ev.Side = string(e.Side)
			rv.Economy = append(rv.Economy, ev)
		}
		rv.Events = buildEvents(r.Events, m)
		out = append(out, rv)
	}
	return out
}

func buildEvents(events []*model.Event, m *model.Match) []EventView {
	slugs := map[string]string{}
	for _, p := range m.Players {
		slugs[p.ID] = p.Slug
	}
	out := make([]EventView, 0, len(events))
	for _, e := range events {
		ev := EventView{
			Type:        e.Type,
			Time:        clock(e.Offset),
			Icon:        iconByType[e.Type],
			DataPlayers: strings.Join(compact(e.ActorID, e.VictimID, e.AssisterID), " "),
			Class:       eventClass(e),
		}
		if ev.Icon == "" {
			ev.Icon = "dot"
		}
		link := func(id, name string) EventPart {
			if id == "" {
				return EventPart{Text: name}
			}
			if slug := slugs[id]; slug != "" {
				return EventPart{Text: name, Href: "#player-" + slug}
			}
			return EventPart{Text: name}
		}
		switch e.Type {
		case "round_start":
			ev.Parts = []EventPart{{Text: "Round started"}}
		case "freeze_end":
			ev.Parts = []EventPart{{Text: "Freeze time ended — weapons are live"}}
		case "kill":
			ev.Parts = append(ev.Parts, link(e.ActorID, e.ActorName))
			ev.Parts = append(ev.Parts, EventPart{Text: " eliminated "})
			ev.Parts = append(ev.Parts, link(e.VictimID, e.VictimName))
			if e.Weapon != "" {
				ev.Parts = append(ev.Parts, EventPart{Text: " with " + e.Weapon})
			}
			if e.AssisterID != "" {
				ev.Parts = append(ev.Parts, EventPart{Text: " (assist: "})
				ev.Parts = append(ev.Parts, link(e.AssisterID, e.AssisterName))
				ev.Parts = append(ev.Parts, EventPart{Text: ")"})
			}
			addKillBadges(&ev, e)
		case "teamkill":
			ev.Parts = append(ev.Parts, link(e.ActorID, e.ActorName))
			ev.Parts = append(ev.Parts, EventPart{Text: " team-killed "})
			ev.Parts = append(ev.Parts, link(e.VictimID, e.VictimName))
			if e.Weapon != "" {
				ev.Parts = append(ev.Parts, EventPart{Text: " with " + e.Weapon})
			}
			addKillBadges(&ev, e)
		case "suicide":
			ev.Parts = append(ev.Parts, link(e.VictimID, e.VictimName))
			ev.Parts = append(ev.Parts, EventPart{Text: " died"})
			if e.Weapon != "" && e.Weapon != "World" {
				ev.Parts = append(ev.Parts, EventPart{Text: " (" + e.Weapon + ")"})
			}
		case "bomb_plant":
			ev.Parts = append(ev.Parts, link(e.ActorID, e.ActorName))
			ev.Parts = append(ev.Parts, EventPart{Text: " planted the bomb"})
			if e.Site != "" {
				ev.Parts = append(ev.Parts, EventPart{Text: " at site " + e.Site})
			}
		case "bomb_defuse":
			ev.Parts = append(ev.Parts, link(e.ActorID, e.ActorName))
			ev.Parts = append(ev.Parts, EventPart{Text: " defused the bomb"})
			if e.Site != "" {
				ev.Parts = append(ev.Parts, EventPart{Text: " at site " + e.Site})
			}
		case "bomb_explode":
			ev.Parts = []EventPart{{Text: "The bomb exploded"}}
		case "grenade":
			ev.Parts = append(ev.Parts, link(e.ActorID, e.ActorName))
			ev.Parts = append(ev.Parts, EventPart{Text: " threw " + article(e.Grenade) + e.Grenade})
		case "mvp":
			ev.Parts = append(ev.Parts, link(e.ActorID, e.ActorName))
			ev.Parts = append(ev.Parts, EventPart{Text: " was awarded the round MVP"})
		case "round_end":
			ev.Parts = []EventPart{{Text: e.Text}}
			if e.Score != [2]int{} {
				ev.Parts = append(ev.Parts, EventPart{Text: fmt.Sprintf(" — %d : %d", e.Score[0], e.Score[1])})
			}
		default:
			ev.Parts = []EventPart{{Text: e.Type}}
		}
		out = append(out, ev)
	}
	return out
}

func addKillBadges(ev *EventView, e *model.Event) {
	if e.Headshot {
		ev.Badges = append(ev.Badges, "HS")
	}
	if e.Wallbang {
		ev.Badges = append(ev.Badges, "WB")
	}
	if e.Noscope {
		ev.Badges = append(ev.Badges, "No-scope")
	}
	if e.ThroughSmoke {
		ev.Badges = append(ev.Badges, "Smoke")
	}
	if e.AssistedFlash {
		ev.Badges = append(ev.Badges, "Flash assist")
	}
}

func eventClass(e *model.Event) string {
	switch e.Type {
	case "kill":
		return "evt-kill"
	case "teamkill", "suicide":
		return "evt-bad"
	case "bomb_plant", "bomb_defuse", "bomb_explode":
		return "evt-bomb"
	case "mvp":
		return "evt-mvp"
	case "round_end":
		return "evt-end"
	default:
		return "evt-" + e.Type
	}
}

func teamSideClass(side model.Side) string {
	switch side {
	case model.SideCT:
		return "ct"
	case model.SideT:
		return "t"
	default:
		return "neutral"
	}
}

func buyClass(buy string) string {
	switch buy {
	case "Full buy":
		return "full"
	case "Force buy":
		return "force"
	case "Pistol":
		return "pistol"
	default:
		return "eco"
	}
}

func initials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "?"
	}
	if len(parts) == 1 {
		r := []rune(parts[0])
		if len(r) > 2 {
			return strings.ToUpper(string(r[:2]))
		}
		return strings.ToUpper(parts[0])
	}
	r0 := []rune(parts[0])
	r1 := []rune(parts[len(parts)-1])
	out := ""
	if len(r0) > 0 {
		out += string(r0[0])
	}
	if len(r1) > 0 {
		out += string(r1[0])
	}
	return strings.ToUpper(out)
}

func compact(in ...string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func ratio(a, b int) string {
	if b == 0 {
		if a == 0 {
			return "0.00"
		}
		return fmt.Sprintf("%d.00", a)
	}
	return fmt.Sprintf("%.2f", float64(a)/float64(b))
}

func clampPct(v float64) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return int(v + 0.5)
}

func article(g string) string {
	if g == "" {
		return ""
	}
	switch strings.ToLower(g)[0] {
	case 'a', 'e', 'i', 'o', 'u':
		return "an "
	default:
		return "a "
	}
}

func clock(seconds float64) string {
	if seconds < 0 || seconds != seconds {
		seconds = 0
	}
	total := int(seconds + 0.5)
	return fmt.Sprintf("%d:%02d", total/60, total%60)
}

// FuncMap returns template helper functions.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"clock": clock,
		"ratio": ratio,
		"pct": func(v float64) string {
			return fmt.Sprintf("%.1f%%", v)
		},
		"one": func(v float64) string {
			return fmt.Sprintf("%.1f", v)
		},
		"zero": func(v float64) string {
			return fmt.Sprintf("%.0f", v)
		},
		"sideClass": func(s model.Side) string { return teamSideClass(s) },
		"lower":     strings.ToLower,
	}
}
