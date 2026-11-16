// Package html renders the self-contained match report.
package html

import (
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/Kandru/cs2-demo-analyzer/internal/model"
	"github.com/Kandru/cs2-demo-analyzer/internal/report"
)

// View is the template model for one report.
type View struct {
	Title        string
	Description  string
	Version      string
	CSS          template.CSS
	JS           template.JS
	Match        *model.Match
	Teams        []*TeamView
	Players      []*PlayerView
	Timeline     []*TimelineSection
	Summary      *model.Summary
	Warnings     []string
	Cheaters     []*CheaterView
	RoundStrip   []RoundStripView
	DemoDownload *DemoLink
	SearchIndex  []SearchEntry
	Chrome       *ChromeView
	Footer       *FooterView
	Assets       *AssetView
}

// SearchEntry is one server-rendered target for the header search dropdown.
// Names stay plain strings so the template escapes them; app.js only reads
// text content and never builds HTML.
type SearchEntry struct {
	Kind   string
	Label  string
	Detail string
	Href   string
	Search string
}

// ChromeView is the shared topbar model for report and index pages.
type ChromeView struct {
	BrandHref  string
	BrandLabel string
	ShowSearch bool
	Nav        []NavLink
}

// NavLink is one topbar tab.
type NavLink struct {
	Href   string
	Label  string
	Count  int
	Active bool
}

// FooterView is the shared footer model. Both pages render the same component.
type FooterView struct {
	RepoURL       string
	SiteURL       string
	Version       string
	GeneratedAt   string
	ParserVersion string
}

// AssetView holds inlined brand assets as trusted data URIs.
type AssetView struct {
	BrandMark  template.URL
	Favicon    template.URL
	AppleTouch template.URL
}

// TimelineSection is one visually distinct block of the continuous timeline.
type TimelineSection struct {
	Kind     string
	Label    string
	Search   string
	Anchor   string
	Time     string
	EndTime  string
	Duration string
	Score    string
	Reason   string
	Side     string
	Class    string
	Number   int
	HasFlag  bool
	IsRound  bool
	Economy  []EconomyView
	Buys     []TeamBuysView
	Rows     []EventView
}

// TeamBuysView groups the freeze-time-end inventory of one team.
type TeamBuysView struct {
	Name   string
	Side   string
	Class  string
	Buyers []PlayerBuyView
}

// PlayerBuyView is one player's purchase list inside a round header.
type PlayerBuyView struct {
	Name  string
	Href  string
	Items []string
}

// DemoLink is a direct download link for a source demo.
type DemoLink struct {
	Href string
	Name string
	Size string
}

// TeamView groups scoreboard rows.
type TeamView struct {
	Team        *model.Team
	DisplayName string
	Players     []*PlayerView
	Stats       model.TeamStats
	Win         bool
	ScoreClass  string
	Side        string
}

// PlayerView is the template view of one player.
type PlayerView struct {
	Player       *model.Player
	Anchor       string
	Initials     string
	ClanTag      string
	TeamName     string
	Side         string
	Rounds       []PlayerRoundView
	Sides        []SideView
	Weapons      []WeaponView
	Aim          []MetricView
	Cheat        *CheatView
	Moments      []*MomentView
	Flagged      bool
	KDDisplay    string
	ADRDisplay   string
	HSDisplay    string
	AimDisplay   string
	AimGrade     string
	ScoreDisplay string
	Search       string
}

// PlayerRoundView is one row in the per-round table.
type PlayerRoundView struct {
	Number      int
	Anchor      string
	Side        string
	Kills       int
	Deaths      int
	Assists     int
	Damage      int
	EquipValue  int
	Result      string
	ResultClass string
	KAST        bool
}

// SideView is a T/CT split row.
type SideView struct {
	Side    string
	Rounds  int
	Kills   int
	Deaths  int
	Assists int
	ADR     string
	HS      string
}

// WeaponView is one row in the weapon table.
type WeaponView struct {
	Name     string
	Kills    int
	Damage   int
	Shots    int
	Hits     int
	Accuracy string
	HSRate   string
}

// MetricView is one aim tile.
type MetricView struct {
	Label       string
	Display     string
	ScorePct    int
	SampleLabel string
	LowSample   bool
	Tooltip     string
	Class       string
}

// CheatView is the anti-cheat card model.
type CheatView struct {
	Available     bool
	Verdict       string
	VerdictClass  string
	LikelihoodPct string
	Likelihood    float64
	OverallGrade  string
	Grades        []model.Grade
	Channels      []ChannelView
	Adjustments   []AdjustmentView
	Error         string
}

// AdjustmentView is a detector adjustment with a plain-language explanation.
type AdjustmentView struct {
	Label  string
	Detail string
}

// ChannelView is one detector channel row.
type ChannelView struct {
	Channel     model.CheatChannel
	Rank        int
	ScorePct    int
	ConfPct     int
	Raw         string
	SampleLabel string
	BarClass    string
	Zone        string
	ZoneClass   string
}

// CheaterView groups one flagged or review-level player for the Watchlist
// section. It reuses the player card data so both surfaces stay consistent.
type CheaterView struct {
	Player      *PlayerView
	Cheat       *CheatView
	TopChannels []ChannelView
	Moments     []*MomentView
	Search      string
}

// MomentView is one merged review cue (all detector reasons at one tick).
type MomentView struct {
	Tick        int
	Time        string
	RoundLabel  string
	RoundAnchor string
	RoundTime   string
	Weapon      string
	Opponent    string
	Reasons     []MomentReason
	Class       string
	DemoCommand string
	rank        int
}

// MomentReason is one detector signal inside a review cue.
type MomentReason struct {
	Kind   string
	Label  string
	Value  string
	Detail string
}

// RoundStripView is one cell of the flat round-history overview.
type RoundStripView struct {
	Number   int
	Label    string
	Anchor   string
	Class    string
	BuyClass string
	Score    string
	Title    string
	HasFlag  bool
	Overtime bool
}

// EconomyView is a team economy chip.
type EconomyView struct {
	TeamName string
	Side     string
	Equip    int
	BuyType  string
	Class    string
}

// EventView is a pre-rendered timeline row.
type EventView struct {
	Type         string
	AbsoluteTime string
	RoundTime    string
	Icon         string
	Parts        []EventPart
	Badges       []BadgeView
	DataPlayers  string
	Class        string
}

// BadgeView is one semantic badge.
type BadgeView struct {
	Text  string
	Class string
}

// EventPart is a text fragment, optionally a player link.
type EventPart struct {
	Text string
	Href string
}

// newView builds the template model from the canonical document.
func newView(doc *report.Document, opts Options) *View {
	m := doc.Match
	generatedAt := doc.Header.GeneratedAt.Local().Format("2006-01-02 15:04 MST")
	v := &View{
		Version:      doc.Header.ToolVersion,
		Match:        m,
		Summary:      m.Summary,
		DemoDownload: opts.DemoDownload,
	}
	title := m.Map
	if title == "" {
		title = "CS2"
	}
	if len(m.Teams) >= 2 {
		title += " - " + displayTeamName(m.Teams[0]) + " vs " + displayTeamName(m.Teams[1])
	}
	v.Title = title
	v.Players = buildPlayers(m)
	v.Teams = buildTeams(m, v.Players)
	v.RoundStrip = buildRoundStrip(m)
	v.Cheaters = buildCheaters(v.Players)
	v.Timeline = buildTimeline(m)
	markFlaggedRounds(v, m)
	v.SearchIndex = buildSearchIndex(v)
	v.Warnings = m.Warnings
	v.Description = fmt.Sprintf("CS2 match report for %s - %s - %s", v.Title, m.Mode, byline)
	v.Chrome = &ChromeView{
		BrandHref:  chromeHref(opts.IndexHref),
		BrandLabel: brandLabel(opts.IndexHref),
		ShowSearch: true,
		Nav: []NavLink{
			{Href: "#overview", Label: "Overview", Active: true},
			{Href: "#cheaters", Label: "Watchlist", Count: len(v.Cheaters)},
			{Href: "#rounds", Label: "Timeline"},
			{Href: "#players", Label: "Players"},
		},
	}
	v.Footer = &FooterView{
		RepoURL:       repoURL,
		SiteURL:       siteURL,
		Version:       v.Version,
		GeneratedAt:   generatedAt,
		ParserVersion: m.ParserVersion,
	}
	return v
}

func chromeHref(indexHref string) string {
	if indexHref != "" {
		return indexHref
	}
	return "#overview"
}

func brandLabel(indexHref string) string {
	if indexHref != "" {
		return "Back to report index"
	}
	return "CS2 Demo Analyzer home"
}

// displayTeamName normalises legacy and missing clan names. "Team A"/"Team B"
// came from older parser versions and are not useful in a report; use the side
// the team started on instead. Real clan names pass through unchanged.
func displayTeamName(t *model.Team) string {
	if t == nil {
		return "Team"
	}
	name := strings.TrimSpace(t.Name)
	legacy := strings.EqualFold(name, "Team A") || strings.EqualFold(name, "Team B")
	if name == "" || legacy {
		switch t.StartSide {
		case model.SideCT:
			return "CT"
		case model.SideT:
			return "T"
		}
		if legacy {
			return name
		}
		return "Team"
	}
	return name
}

func buildTeams(m *model.Match, players []*PlayerView) []*TeamView {
	byID := map[int]model.TeamStats{}
	if m.Summary != nil {
		for _, s := range m.Summary.TeamStats {
			byID[s.TeamID] = s
		}
	}
	winner := 0
	if m.Summary != nil {
		winner = m.Summary.WinnerTeamID
	}
	out := make([]*TeamView, 0, len(m.Teams))
	for _, t := range m.Teams {
		scoreClass := "draw"
		if winner != 0 {
			if t.ID == winner {
				scoreClass = "win"
			} else {
				scoreClass = "loss"
			}
		}
		tv := &TeamView{
			Team:        t,
			DisplayName: displayTeamName(t),
			Stats:       byID[t.ID],
			Win:         t.ID == winner,
			ScoreClass:  scoreClass,
			Side:        string(t.StartSide),
		}
		for _, pv := range players {
			if pv.Player.TeamID == t.ID {
				tv.Players = append(tv.Players, pv)
			}
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
		Anchor:       "player-" + p.Slug,
		Initials:     initials(p.Name),
		ClanTag:      displayClanTag(p.ClanTag),
		Flagged:      p.Cheat != nil && p.Cheat.Flagged,
		KDDisplay:    fmt.Sprintf("%.2f", p.Stats.KD),
		ADRDisplay:   fmt.Sprintf("%.1f", p.Stats.ADR),
		HSDisplay:    fmt.Sprintf("%.0f%%", p.Stats.HSPercent),
		ScoreDisplay: fmt.Sprintf("%d", p.Stats.Score),
	}
	if team := m.TeamByID(p.TeamID); team != nil {
		pv.TeamName = displayTeamName(team)
		pv.Side = string(team.StartSide)
	}
	terms := []string{p.ID, p.Name, p.Slug, pv.ClanTag, pv.TeamName}
	terms = append(terms, p.Aliases...)
	if p.SteamID64 != 0 {
		terms = append(terms, strconv.FormatUint(p.SteamID64, 10))
	}
	pv.Search = searchKey(terms...)
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
			Anchor:      fmt.Sprintf("round-%d", r.StartTick),
			Side:        string(rs.Side),
			Kills:       rs.Kills,
			Deaths:      rs.Deaths,
			Assists:     rs.Assists,
			Damage:      rs.Damage,
			EquipValue:  rs.EquipValue,
			Result:      res,
			ResultClass: resClass,
			KAST:        rs.KAST,
		})
	}
	for _, side := range []model.Side{model.SideT, model.SideCT} {
		ss := p.SideStats[side]
		if ss == nil || ss.Rounds == 0 {
			continue
		}
		pv.Sides = append(pv.Sides, SideView{
			Side:    string(side),
			Rounds:  ss.Rounds,
			Kills:   ss.Kills,
			Deaths:  ss.Deaths,
			Assists: ss.Assists,
			ADR:     fmt.Sprintf("%.1f", ss.ADR),
			HS:      fmt.Sprintf("%.0f%%", ss.HSPercent),
		})
	}
	if p.Aim != nil {
		pv.AimDisplay = fmt.Sprintf("%.0f", p.Aim.Rating)
		pv.AimGrade = p.Aim.Grade
		for _, mt := range p.Aim.Metrics {
			pv.Aim = append(pv.Aim, metricView(mt))
		}
	}
	for _, ws := range p.Weapons {
		wv := WeaponView{
			Name:   ws.Name,
			Kills:  ws.Kills,
			Damage: ws.Damage,
			Shots:  ws.Shots,
			Hits:   ws.Hits,
		}
		if ws.Shots > 0 {
			wv.Accuracy = fmt.Sprintf("%.1f%%", ws.Accuracy)
		} else {
			wv.Accuracy = "-"
		}
		if ws.Kills > 0 {
			wv.HSRate = fmt.Sprintf("%.0f%%", float64(ws.Headshots)/float64(ws.Kills)*100)
		} else {
			wv.HSRate = "-"
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
	if p.Cheat != nil {
		cheatTerms := []string{plainDashes(p.Cheat.Narrative), p.Cheat.OverallGrade}
		if pv.Cheat != nil {
			cheatTerms = append(cheatTerms, pv.Cheat.Verdict)
		}
		for _, g := range p.Cheat.Grades {
			cheatTerms = append(cheatTerms, g.Category, g.Label, g.Grade)
		}
		for _, ch := range p.Cheat.Channels {
			cheatTerms = append(cheatTerms, ch.ID, ch.Label, ch.Zone, plainDashes(ch.Description))
		}
		for _, mo := range p.Cheat.Moments {
			cheatTerms = append(cheatTerms, mo.Kind, plainDashes(mo.Label), plainDashes(mo.ValueText), mo.Opponent, mo.Weapon)
		}
		pv.Search = searchKey(append([]string{pv.Search}, cheatTerms...)...)
	}
	if p.Cheat != nil && len(p.Cheat.Moments) > 0 {
		pv.Moments = buildMoments(m, p.Cheat.Moments)
	}
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
		Label:       mt.Label,
		Display:     plainDashes(mt.Display),
		ScorePct:    clampPct(mt.Score),
		SampleLabel: metricSampleLabel(mt.ID, mt.Samples),
		LowSample:   mt.LowSample,
		Tooltip:     plainDashes(mt.Tooltip),
		Class:       class,
	}
}

// metricSampleLabel explains what an aim sample actually is. Detector channels
// use their own labels; aim metrics have a small, stable set of nouns.
func metricSampleLabel(id string, n int) string {
	noun := "samples"
	switch id {
	case "accuracy", "counter_strafe", "spray":
		noun = "shots"
	case "head_accuracy":
		noun = "hits"
	case "damage_speed", "reaction", "preaim":
		noun = "engagements"
	}
	return sampleCount(noun, n)
}

func sampleCount(noun string, n int) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", strings.TrimSuffix(noun, "s"))
	}
	return fmt.Sprintf("%d %s", n, noun)
}

func adjustmentExplanation(a model.Adjustment) string {
	detail := plainDashes(a.Detail)
	switch strings.ToLower(a.Label) {
	case "evidence stacking boost":
		if strings.HasPrefix(detail, "Yes (") {
			inner := strings.TrimSuffix(strings.TrimPrefix(detail, "Yes ("), ")")
			if fields := strings.Fields(inner); len(fields) > 0 {
				return fields[0] + " detector channels showed strong agreement, so the combined score was raised."
			}
		}
		return "Several strong detector channels agree, so the combined score was raised."
	case "wingman score boost":
		if strings.HasPrefix(detail, "kills=") {
			return "Wingman is a shorter format, so " + strings.TrimPrefix(detail, "kills=") + " kills were normalised before scoring."
		}
		return "Wingman is a shorter format, so kill volume is normalised before scoring."
	case "scoreboard-position discount":
		if detail != "" {
			return "A strong scoreboard position lowers the result (" + detail + ") so rank alone cannot inflate the score."
		}
		return "A strong scoreboard position lowers the result slightly so rank alone cannot inflate the score."
	}
	if detail != "" {
		return detail
	}
	return "The detector adjusted the final likelihood."
}

func buildCheat(cr *model.CheatReport) *CheatView {
	if cr == nil {
		return &CheatView{Available: false, Verdict: "Not analyzed", VerdictClass: "unavailable", Error: "anti-cheat detector was not run"}
	}
	cv := &CheatView{
		Available:     cr.Available,
		OverallGrade:  cr.OverallGrade,
		Grades:        cr.Grades,
		Error:         cr.Error,
		Likelihood:    cr.Likelihood,
		LikelihoodPct: fmt.Sprintf("%.0f%%", cr.Likelihood),
	}
	for _, a := range cr.Adjustments {
		cv.Adjustments = append(cv.Adjustments, AdjustmentView{
			Label:  plainDashes(a.Label),
			Detail: adjustmentExplanation(a),
		})
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
		ch.Description = plainDashes(ch.Description)
		cvm := ChannelView{
			Channel:     ch,
			Rank:        len(cv.Channels) + 1,
			ScorePct:    clampPct(ch.Score * 100),
			ConfPct:     clampPct(ch.Confidence * 100),
			SampleLabel: sampleCount("samples", ch.Samples),
			BarClass:    zoneClass(ch.Zone),
			Zone:        ch.Zone,
			ZoneClass:   zoneClass(ch.Zone),
		}
		if ch.HasData {
			cvm.Raw = rawMetric(ch)
		} else {
			cvm.Raw = "-"
		}
		cv.Channels = append(cv.Channels, cvm)
	}
	sort.SliceStable(cv.Channels, func(i, j int) bool {
		si := cv.Channels[i].Channel.Score * cv.Channels[i].Channel.Confidence * cv.Channels[i].Channel.Weight
		sj := cv.Channels[j].Channel.Score * cv.Channels[j].Channel.Confidence * cv.Channels[j].Channel.Weight
		if si != sj {
			return si > sj
		}
		return cv.Channels[i].Channel.Score > cv.Channels[j].Channel.Score
	})
	for i := range cv.Channels {
		cv.Channels[i].Rank = i + 1
	}
	return cv
}

func buildSearchIndex(v *View) []SearchEntry {
	out := make([]SearchEntry, 0, len(v.Players)+len(v.RoundStrip)+8)
	add := func(kind, label, detail, href string, terms ...string) {
		label = strings.TrimSpace(label)
		if label == "" || href == "" {
			return
		}
		out = append(out, SearchEntry{
			Kind:   kind,
			Label:  label,
			Detail: strings.TrimSpace(detail),
			Href:   href,
			Search: searchKey(append([]string{kind, label, detail, href}, terms...)...),
		})
	}
	add("section", "Overview", "Match at a glance", "#overview")
	add("section", "Watchlist", fmt.Sprintf("%d player(s)", len(v.Cheaters)), "#cheaters")
	add("section", "Timeline", "Rounds and events", "#rounds")
	add("section", "Players", "Stats and cheat analysis", "#players")
	for _, c := range v.Cheaters {
		add("watchlist", c.Player.Player.Name, c.Cheat.Verdict+" - "+c.Cheat.LikelihoodPct,
			"#cheater-"+c.Player.Player.Slug, c.Search)
	}
	for _, p := range v.Players {
		detail := strings.Join(compact(p.TeamName, p.ClanTag), " ")
		add("player", p.Player.Name, detail, "#"+p.Anchor, p.Search)
	}
	for _, r := range v.RoundStrip {
		add("round", "Round "+r.Label, r.Score+" - "+r.Title, "#"+r.Anchor)
	}
	return out
}

func buildCheaters(players []*PlayerView) []*CheaterView {
	out := make([]*CheaterView, 0, len(players))
	for _, pv := range players {
		if pv.Cheat == nil || !pv.Cheat.Available {
			continue
		}
		if !pv.Flagged && pv.Cheat.Likelihood < 25 {
			continue
		}
		out = append(out, &CheaterView{
			Player:      pv,
			Cheat:       pv.Cheat,
			TopChannels: topChannels(pv.Cheat.Channels, 3),
			Moments:     pv.Moments,
			Search:      pv.Search,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Player.Flagged != out[j].Player.Flagged {
			return out[i].Player.Flagged
		}
		if out[i].Cheat.Likelihood != out[j].Cheat.Likelihood {
			return out[i].Cheat.Likelihood > out[j].Cheat.Likelihood
		}
		return out[i].Player.Player.Name < out[j].Player.Player.Name
	})
	return out
}

func topChannels(channels []ChannelView, n int) []ChannelView {
	filtered := make([]ChannelView, 0, len(channels))
	for _, ch := range channels {
		if ch.Channel.HasData && ch.Channel.Score > 0 {
			filtered = append(filtered, ch)
		}
	}
	if len(filtered) == 0 {
		for _, ch := range channels {
			if ch.Channel.HasData {
				filtered = append(filtered, ch)
			}
		}
	}
	if len(filtered) > n {
		filtered = filtered[:n]
	}
	for i := range filtered {
		filtered[i].Rank = i + 1
	}
	return filtered
}

func buildMoments(m *model.Match, raw []model.CheatMoment) []*MomentView {
	byTick := map[int]*MomentView{}
	for _, rm := range raw {
		if rm.Tick < 0 {
			continue
		}
		mv := byTick[rm.Tick]
		if mv == nil {
			seconds := rm.Seconds
			if seconds <= 0 && rm.Tick > 0 && m.TickRate > 0 {
				seconds = float64(rm.Tick) / m.TickRate
			}
			mv = &MomentView{
				Tick:        rm.Tick,
				Time:        clock(seconds),
				DemoCommand: fmt.Sprintf("demo_gototick %d", rm.Tick),
			}
			momentLocation(m, mv, seconds)
			byTick[rm.Tick] = mv
		}
		reason, rank := momentReason(rm)
		mv.Reasons = append(mv.Reasons, reason)
		if rank > mv.rank {
			mv.rank = rank
			mv.Class = momentClass(rank)
		}
		if rm.Weapon != "" && mv.Weapon == "" {
			mv.Weapon = rm.Weapon
		}
		if rm.Opponent != "" && mv.Opponent == "" {
			mv.Opponent = rm.Opponent
		}
	}
	out := make([]*MomentView, 0, len(byTick))
	for _, mv := range byTick {
		out = append(out, mv)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Tick < out[j].Tick
	})
	return out
}

func momentLocation(m *model.Match, mv *MomentView, seconds float64) {
	mv.RoundLabel = "Pre-match"
	for _, r := range m.Rounds {
		if r.Pregame || r.Number == 0 {
			continue
		}
		if mv.Tick >= r.StartTick && mv.Tick <= r.EndTick {
			mv.RoundLabel = fmt.Sprintf("Round %d", r.Number)
			mv.RoundAnchor = fmt.Sprintf("round-%d", r.StartTick)
			mv.RoundTime = clock(seconds - r.StartOffset)
			return
		}
	}
	firstStart := -1
	for _, r := range m.Rounds {
		if r.Pregame || r.Number == 0 {
			continue
		}
		firstStart = r.StartTick
		break
	}
	if firstStart >= 0 && mv.Tick < firstStart {
		mv.RoundLabel = "Warmup"
		if m.Warmup != nil {
			mv.RoundAnchor = "warmup"
			mv.RoundTime = clock(seconds)
		} else {
			mv.RoundLabel = "Pre-match"
		}
	}
}

func momentReason(rm model.CheatMoment) (MomentReason, int) {
	label := rm.Label
	if label == "" {
		label = rm.Kind
	}
	if label == "" {
		label = "Cue"
	}
	value := rm.ValueText
	if value == "" {
		value = "signal"
	}
	rank := rm.Rank
	if rank < 1 {
		rank = 1
	}
	return MomentReason{Kind: rm.Kind, Label: plainDashes(label), Value: plainDashes(value), Detail: plainDashes(rm.Detail)}, rank
}

func momentClass(rank int) string {
	switch {
	case rank >= 3:
		return "cue-strong"
	case rank == 2:
		return "cue-mild"
	default:
		return "cue-info"
	}
}

func roundAnchor(r *model.Round) string {
	return fmt.Sprintf("round-%d", r.StartTick)
}

func roundLabel(r *model.Round) string {
	if r.Pregame {
		return "Pregame"
	}
	label := fmt.Sprintf("Round %d", r.Number)
	if r.Overtime {
		label += " (OT)"
	}
	return label
}

func roundWinner(m *model.Match, r *model.Round) string {
	if t := m.TeamByID(r.WinnerTeamID); t != nil {
		name := displayTeamName(t)
		// Fallback names ("CT"/"T") describe the start side, which is not the
		// winning side after a half-time swap. Prefer the round's winner side.
		if name == "CT" || name == "T" {
			if r.Winner == model.SideCT || r.Winner == model.SideT {
				return string(r.Winner)
			}
		}
		return name
	}
	return "Draw"
}

func roundSearch(m *model.Match, r *model.Round, label string) string {
	search := searchKey("round", strconv.Itoa(r.Number), label, r.Reason, r.Message, roundWinner(m, r), teamSideClass(r.Winner))
	if r.Overtime {
		search += " overtime ot"
	}
	return search
}

func buildRoundStrip(m *model.Match) []RoundStripView {
	out := make([]RoundStripView, 0, len(m.Rounds))
	for _, r := range m.Rounds {
		if r.Pregame {
			continue
		}
		rv := RoundStripView{
			Number:   r.Number,
			Label:    strconv.Itoa(r.Number),
			Anchor:   roundAnchor(r),
			Class:    teamSideClass(r.Winner),
			Score:    fmt.Sprintf("%d:%d", r.EndScore[0], r.EndScore[1]),
			Overtime: r.Overtime,
		}
		if len(r.Economy) > 0 {
			rv.BuyClass = buyClass(r.Economy[0].BuyType)
		}
		rv.Title = fmt.Sprintf("%s - %s %d:%d - %s", roundLabel(r), roundWinner(m, r), r.EndScore[0], r.EndScore[1], r.Reason)
		out = append(out, rv)
	}
	return out
}

type timelineEntry struct {
	view   EventView
	offset float64
	seq    int
}

func buildTimeline(m *model.Match) []*TimelineSection {
	global := globalEventsWithFallback(m)
	globalEntries := make([]timelineEntry, 0, len(global))
	for _, e := range global {
		globalEntries = append(globalEntries, timelineEntry{
			view:   buildEvent(e, m, 0),
			offset: e.Offset,
			seq:    e.Seq,
		})
	}
	usedGlobal := make([]bool, len(globalEntries))

	var sections []*TimelineSection
	firstStart := -1.0
	for _, r := range m.Rounds {
		firstStart = r.StartOffset
		break
	}
	warmupStart, warmupEnd := 0.0, 0.0
	hasWarmup := m.Warmup != nil && len(m.Warmup.Events) > 0
	if hasWarmup {
		warmupStart = m.Warmup.StartOffset
		warmupEnd = m.Warmup.EndOffset
		if warmupEnd <= warmupStart {
			if firstStart >= 0 {
				warmupEnd = firstStart
			} else if m.Warmup.Duration > 0 {
				warmupEnd = warmupStart + m.Warmup.Duration
			} else {
				for _, e := range m.Warmup.Events {
					if e.Offset > warmupEnd {
						warmupEnd = e.Offset
					}
				}
			}
		}
	}

	preLimit := firstStart
	if hasWarmup {
		preLimit = warmupStart
	}
	var pre []timelineEntry
	if preLimit >= 0 {
		for i, entry := range globalEntries {
			beforeWarmup := entry.offset < preLimit
			afterWarmup := hasWarmup && entry.offset > warmupEnd && firstStart >= 0 && entry.offset < firstStart
			if beforeWarmup || afterWarmup {
				pre = append(pre, entry)
				usedGlobal[i] = true
			}
		}
	}
	if len(pre) > 0 {
		sortTimelineEntries(pre)
		sections = append(sections, &TimelineSection{
			Kind:   "pre",
			Label:  "Before the first round",
			Search: searchKey("pre-match", "before the first round"),
			Time:   pre[0].view.AbsoluteTime,
			Rows:   timelineViews(pre),
			Class:  "alt-a",
		})
	}

	if hasWarmup {
		rows := make([]timelineEntry, 0, len(m.Warmup.Events))
		for _, e := range m.Warmup.Events {
			rows = append(rows, timelineEntry{view: buildEvent(e, m, 0), offset: e.Offset, seq: e.Seq})
		}
		for i, entry := range globalEntries {
			if usedGlobal[i] {
				continue
			}
			if entry.offset >= warmupStart && entry.offset <= warmupEnd {
				rows = append(rows, entry)
				usedGlobal[i] = true
			}
		}
		sortTimelineEntries(rows)
		end := clock(warmupEnd)
		if warmupEnd <= 0 {
			end = clock(m.Warmup.Duration)
		}
		sections = append(sections, &TimelineSection{
			Kind:     "warmup",
			Label:    "Warmup",
			Search:   searchKey("warmup practice before the first round"),
			Anchor:   "warmup",
			Time:     clock(warmupStart),
			EndTime:  end,
			Duration: clock(m.Warmup.Duration),
			Rows:     timelineViews(rows),
			Class:    "alt-b",
		})
	}

	for i, r := range m.Rounds {
		var rows []timelineEntry
		hasRoundEnd := false
		for _, e := range r.Events {
			if e.Type == model.EventRoundStart {
				continue
			}
			if e.Type == model.EventRoundEnd {
				hasRoundEnd = true
			}
			rows = append(rows, timelineEntry{
				view:   buildEvent(e, m, r.StartOffset),
				offset: r.StartOffset + e.Offset,
				seq:    e.Seq,
			})
		}
		for j, entry := range globalEntries {
			if usedGlobal[j] {
				continue
			}
			if entry.offset >= r.StartOffset && entry.offset <= r.EndOffset {
				rows = append(rows, entry)
				usedGlobal[j] = true
			}
		}
		if !hasRoundEnd {
			endEvent := &model.Event{
				Type:   model.EventRoundEnd,
				Offset: r.EndOffset - r.StartOffset,
				Text:   r.Reason,
				Score:  r.EndScore,
			}
			if endEvent.Text == "" {
				endEvent.Text = "Round ended"
			}
			rows = append(rows, timelineEntry{
				view:   buildEvent(endEvent, m, r.StartOffset),
				offset: r.EndOffset,
			})
		}
		sortTimelineEntries(rows)
		label := roundLabel(r)
		search := roundSearch(m, r, label)
		sections = append(sections, &TimelineSection{
			Kind:     "round",
			Label:    label,
			Search:   search,
			Anchor:   roundAnchor(r),
			Time:     clock(r.StartOffset),
			EndTime:  clock(r.EndOffset),
			Duration: clock(r.Duration),
			Score:    fmt.Sprintf("%d : %d", r.EndScore[0], r.EndScore[1]),
			Reason:   r.Reason,
			Side:     string(r.Winner),
			Class:    roundClass(r),
			Number:   r.Number,
			IsRound:  true,
			Economy:  buildEconomy(r, m),
			Buys:     buildBuys(r, m),
			Rows:     timelineViews(rows),
		})

		// Lifecycle events between this round and the next one. They belong to
		// the gap, not to the round that just ended.
		var gap []timelineEntry
		for j, entry := range globalEntries {
			if usedGlobal[j] || entry.offset <= r.EndOffset {
				continue
			}
			if i+1 < len(m.Rounds) && entry.offset >= m.Rounds[i+1].StartOffset {
				continue
			}
			gap = append(gap, entry)
			usedGlobal[j] = true
		}
		if len(gap) > 0 {
			sortTimelineEntries(gap)
			label := fmt.Sprintf("Between round %d and %d", r.Number, r.Number+1)
			if i+1 >= len(m.Rounds) {
				label = fmt.Sprintf("After round %d", r.Number)
			}
			sections = append(sections, &TimelineSection{
				Kind:   "gap",
				Label:  label,
				Search: searchKey(label, "between rounds"),
				Time:   gap[0].view.AbsoluteTime,
				Rows:   timelineViews(gap),
				Class:  "gap",
			})
		}
	}

	var rest []timelineEntry
	for i, entry := range globalEntries {
		if !usedGlobal[i] {
			rest = append(rest, entry)
		}
	}
	if len(rest) > 0 {
		sortTimelineEntries(rest)
		label := "After the last round"
		search := searchKey("after the last round", "post-match")
		if len(m.Rounds) == 0 {
			label = "Match events"
			search = searchKey("match events")
		}
		sections = append(sections, &TimelineSection{
			Kind:   "post",
			Label:  label,
			Search: search,
			Time:   rest[0].view.AbsoluteTime,
			Rows:   timelineViews(rest),
			Class:  "gap",
		})
	}
	return sections
}

func sortTimelineEntries(entries []timelineEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].offset != entries[j].offset {
			return entries[i].offset < entries[j].offset
		}
		return entries[i].seq < entries[j].seq
	})
}

func timelineViews(entries []timelineEntry) []EventView {
	out := make([]EventView, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.view)
	}
	return out
}

func roundClass(r *model.Round) string {
	parity := "alt-a"
	if r.Number%2 == 0 {
		parity = "alt-b"
	}
	return strings.Join(compact(teamSideClass(r.Winner), parity), " ")
}

func buildEconomy(r *model.Round, m *model.Match) []EconomyView {
	out := make([]EconomyView, 0, len(r.Economy))
	for _, e := range r.Economy {
		ev := EconomyView{Equip: e.EquipmentValue, BuyType: e.BuyType, Class: buyClass(e.BuyType)}
		if t := m.TeamByID(e.TeamID); t != nil {
			ev.TeamName = displayTeamName(t)
		}
		ev.Side = string(e.Side)
		out = append(out, ev)
	}
	return out
}

// buildBuys groups the freeze-time inventory snapshot by team and player.
// Older reports have no Bought data and simply render nothing.
func buildBuys(r *model.Round, m *model.Match) []TeamBuysView {
	out := make([]TeamBuysView, 0, len(m.Teams))
	for _, t := range m.Teams {
		tv := TeamBuysView{Name: displayTeamName(t), Side: string(t.StartSide), Class: teamSideClass(t.StartSide)}
		for _, p := range m.Players {
			if p.TeamID != t.ID {
				continue
			}
			rs := r.Players[p.ID]
			if rs == nil || len(rs.Bought) == 0 {
				continue
			}
			tv.Buyers = append(tv.Buyers, PlayerBuyView{
				Name:  p.Name,
				Href:  "#player-" + p.Slug,
				Items: append([]string(nil), rs.Bought...),
			})
		}
		sort.SliceStable(tv.Buyers, func(i, j int) bool {
			return tv.Buyers[i].Name < tv.Buyers[j].Name
		})
		if len(tv.Buyers) > 0 {
			out = append(out, tv)
		}
	}
	return out
}

// globalEventsWithFallback keeps older cached reports useful: when the parser
// did not record a team_swap event, detect the side flip from round economy.
func globalEventsWithFallback(m *model.Match) []*model.Event {
	events := append([]*model.Event(nil), m.Events...)
	for _, e := range events {
		if e.Type == model.EventTeamSwap {
			return events
		}
	}
	var prev map[int]model.Side
	for _, r := range m.Rounds {
		if r.Pregame || len(r.Economy) < 2 {
			continue
		}
		cur := map[int]model.Side{}
		for _, e := range r.Economy {
			cur[e.TeamID] = e.Side
		}
		if len(prev) >= 2 && sidesSwapped(prev, cur) {
			text := "Teams switched sides"
			var ctName, tName string
			for id, side := range cur {
				if team := m.TeamByID(id); team != nil && displayTeamName(team) != "" {
					if side == model.SideCT {
						ctName = displayTeamName(team)
					} else if side == model.SideT {
						tName = displayTeamName(team)
					}
				}
			}
			if ctName != "" && tName != "" {
				text = fmt.Sprintf("%s now CT - %s now T", ctName, tName)
			}
			offset := r.StartOffset - 0.01
			if offset < 0 {
				offset = 0
			}
			events = append(events, &model.Event{Type: model.EventTeamSwap, Offset: offset, Text: text})
		}
		prev = cur
	}
	return events
}

func sidesSwapped(a, b map[int]model.Side) bool {
	common := 0
	for id, side := range a {
		other, ok := b[id]
		if !ok {
			continue
		}
		common++
		if other == side {
			return false
		}
	}
	return common >= 2
}

func markFlaggedRounds(v *View, m *model.Match) {
	flagged := map[int]bool{}
	for _, pv := range v.Players {
		if !pv.Flagged {
			continue
		}
		for _, moment := range pv.Moments {
			for _, r := range m.Rounds {
				if r.Pregame || r.Number == 0 {
					continue
				}
				if moment.Tick >= r.StartTick && moment.Tick <= r.EndTick {
					flagged[r.Number] = true
					break
				}
			}
		}
	}
	for i := range v.RoundStrip {
		v.RoundStrip[i].HasFlag = flagged[v.RoundStrip[i].Number]
	}
	for _, section := range v.Timeline {
		if section.IsRound {
			section.HasFlag = flagged[section.Number]
		}
	}
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
	model.EventRoundStart:       "flag",
	model.EventFreezeEnd:        "buy",
	model.EventKill:             "kill",
	model.EventTeamKill:         "warning",
	model.EventSuicide:          "warning",
	model.EventBombPlant:        "bomb",
	model.EventBombDefuse:       "defuse",
	model.EventBombExplode:      "explosion",
	model.EventBombPickup:       "bomb",
	model.EventBombDrop:         "bomb-drop",
	model.EventGrenade:          "grenade",
	model.EventGrenadeDamage:    "grenade",
	model.EventFlash:            "flash",
	model.EventMVP:              "star",
	model.EventRoundEnd:         "trophy",
	model.EventPlayerConnect:    "connect",
	model.EventPlayerDisconnect: "disconnect",
	model.EventTeamChange:       "swap",
	model.EventTeamSwap:         "swap",
	model.EventHostagePickup:    "hostage",
	model.EventHostageDrop:      "hostage",
	model.EventHostageRescue:    "hostage",
	model.EventHostageKilled:    "hostage",
}

func buildEvent(e *model.Event, m *model.Match, base float64) EventView {
	slugs := map[string]string{}
	for _, p := range m.Players {
		slugs[p.ID] = p.Slug
	}
	absolute := base + e.Offset
	roundTime := ""
	if base != 0 {
		roundTime = clock(e.Offset)
	}
	ev := EventView{
		Type:         e.Type,
		AbsoluteTime: clock(absolute),
		RoundTime:    roundTime,
		Icon:         iconByType[e.Type],
		DataPlayers:  strings.Join(compact(e.ActorID, e.VictimID, e.AssisterID), " "),
		Class:        eventClass(e),
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
	case model.EventRoundStart:
		ev.Parts = []EventPart{{Text: "Round started"}}
	case model.EventFreezeEnd:
		ev.Parts = []EventPart{{Text: "Freeze time ended"}}
	case model.EventKill:
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
	case model.EventTeamKill:
		ev.Parts = append(ev.Parts, link(e.ActorID, e.ActorName))
		ev.Parts = append(ev.Parts, EventPart{Text: " team-killed "})
		ev.Parts = append(ev.Parts, link(e.VictimID, e.VictimName))
		if e.Weapon != "" {
			ev.Parts = append(ev.Parts, EventPart{Text: " with " + e.Weapon})
		}
		ev.addBadge("Teamkill", "danger")
		addKillBadges(&ev, e)
	case model.EventSuicide:
		ev.Parts = append(ev.Parts, link(e.VictimID, e.VictimName))
		ev.Parts = append(ev.Parts, EventPart{Text: " died"})
		if e.Weapon != "" && e.Weapon != "World" {
			ev.Parts = append(ev.Parts, EventPart{Text: " (" + e.Weapon + ")"})
		}
		ev.addBadge("Suicide", "danger")
	case model.EventBombPlant:
		ev.Parts = append(ev.Parts, link(e.ActorID, e.ActorName))
		ev.Parts = append(ev.Parts, EventPart{Text: " planted the bomb"})
		if e.Site != "" {
			ev.addBadge("Site "+e.Site, "warning")
		}
	case model.EventBombDefuse:
		ev.Parts = append(ev.Parts, link(e.ActorID, e.ActorName))
		ev.Parts = append(ev.Parts, EventPart{Text: " defused the bomb"})
		if e.Site != "" {
			ev.addBadge("Site "+e.Site, "success")
		}
	case model.EventBombExplode:
		ev.Parts = []EventPart{{Text: "The bomb exploded"}}
	case model.EventBombPickup:
		ev.Parts = append(ev.Parts, link(e.ActorID, actorName(e)))
		ev.Parts = append(ev.Parts, EventPart{Text: " picked up the bomb"})
		ev.addBadge("Bomb carrier", "warning")
	case model.EventBombDrop:
		ev.Parts = append(ev.Parts, link(e.ActorID, actorName(e)))
		ev.Parts = append(ev.Parts, EventPart{Text: " dropped the bomb"})
		ev.addBadge("Bomb dropped", "warning")
	case model.EventGrenade:
		ev.Parts = append(ev.Parts, link(e.ActorID, e.ActorName))
		ev.Parts = append(ev.Parts, EventPart{Text: " threw " + article(e.Grenade) + e.Grenade})
		ev.addBadge("Utility", "info")
	case model.EventGrenadeDamage:
		ev.Parts = append(ev.Parts, link(e.ActorID, actorName(e)))
		ev.Parts = append(ev.Parts, EventPart{Text: " hit "})
		ev.Parts = append(ev.Parts, link(e.VictimID, e.VictimName))
		if e.Weapon != "" {
			ev.Parts = append(ev.Parts, EventPart{Text: " with " + e.Weapon})
		}
		ev.addBadge(fmt.Sprintf("%.0f damage", e.Amount), "warning")
	case model.EventFlash:
		ev.Parts = append(ev.Parts, link(e.ActorID, actorName(e)))
		ev.Parts = append(ev.Parts, EventPart{Text: " flashed "})
		ev.Parts = append(ev.Parts, link(e.VictimID, e.VictimName))
		if e.Amount > 0 {
			ev.addBadge(fmt.Sprintf("%.1fs blind", e.Amount), "info")
		} else {
			ev.addBadge("Blind", "info")
		}
	case model.EventMVP:
		ev.Parts = append(ev.Parts, link(e.ActorID, e.ActorName))
		ev.Parts = append(ev.Parts, EventPart{Text: " was awarded the round MVP"})
		ev.addBadge("MVP", "success")
	case model.EventRoundEnd:
		ev.Parts = []EventPart{{Text: e.Text}}
		if e.Score != [2]int{} {
			ev.addBadge(fmt.Sprintf("%d : %d", e.Score[0], e.Score[1]), "neutral")
		}
	case model.EventPlayerConnect:
		ev.Parts = append(ev.Parts, link(e.ActorID, actorName(e)))
		ev.Parts = append(ev.Parts, EventPart{Text: " connected"})
	case model.EventPlayerDisconnect:
		ev.Parts = append(ev.Parts, link(e.ActorID, actorName(e)))
		ev.Parts = append(ev.Parts, EventPart{Text: " disconnected"})
	case model.EventTeamChange:
		ev.Parts = append(ev.Parts, link(e.ActorID, actorName(e)))
		ev.Parts = append(ev.Parts, EventPart{Text: teamChangeText(e)})
		if e.NewSide != model.SideUnknown {
			ev.addBadge(string(e.NewSide), strings.ToLower(string(e.NewSide)))
		}
	case model.EventTeamSwap:
		text := e.Text
		if text == "" {
			text = "Teams switched sides"
		}
		ev.Parts = []EventPart{{Text: text}}
	case model.EventHostagePickup:
		ev.Parts = append(ev.Parts, link(e.ActorID, actorName(e)))
		ev.Parts = append(ev.Parts, EventPart{Text: " picked up a hostage"})
		ev.addBadge("Hostage", "info")
	case model.EventHostageDrop:
		ev.Parts = append(ev.Parts, link(e.ActorID, actorName(e)))
		ev.Parts = append(ev.Parts, EventPart{Text: " dropped a hostage"})
		ev.addBadge("Hostage", "info")
	case model.EventHostageRescue:
		ev.Parts = append(ev.Parts, link(e.ActorID, actorName(e)))
		ev.Parts = append(ev.Parts, EventPart{Text: " rescued a hostage"})
		ev.addBadge("Rescued", "success")
	case model.EventHostageKilled:
		ev.Parts = append(ev.Parts, link(e.ActorID, actorName(e)))
		ev.Parts = append(ev.Parts, EventPart{Text: " killed a hostage"})
		ev.addBadge("Hostage killed", "danger")
	default:
		ev.Parts = []EventPart{{Text: e.Type}}
	}
	return ev
}

func actorName(e *model.Event) string {
	if e.ActorName != "" {
		return e.ActorName
	}
	return "Unknown player"
}

func teamChangeText(e *model.Event) string {
	oldSide, newSide := e.OldSide, e.NewSide
	switch {
	case newSide == model.SideSpectator || newSide == model.SideUnassigned:
		return " moved to " + sideLabel(newSide)
	case oldSide == model.SideSpectator || oldSide == model.SideUnassigned || oldSide == model.SideUnknown:
		return " joined " + sideLabel(newSide)
	default:
		return " switched from " + sideLabel(oldSide) + " to " + sideLabel(newSide)
	}
}

func sideLabel(s model.Side) string {
	switch s {
	case model.SideCT:
		return "Counter-Terrorists"
	case model.SideT:
		return "Terrorists"
	case model.SideSpectator:
		return "Spectators"
	case model.SideUnassigned:
		return "Unassigned"
	default:
		return "Unknown"
	}
}

func (ev *EventView) addBadge(text, class string) {
	if text == "" {
		return
	}
	ev.Badges = append(ev.Badges, BadgeView{Text: text, Class: class})
}

func addKillBadges(ev *EventView, e *model.Event) {
	if e.Headshot {
		ev.addBadge("Headshot", "accent")
	}
	if e.Wallbang {
		ev.addBadge("Wallbang", "warning")
	}
	if e.Noscope {
		ev.addBadge("No-scope", "warning")
	}
	if e.ThroughSmoke {
		ev.addBadge("Through smoke", "warning")
	}
	if e.AssistedFlash {
		ev.addBadge("Flash assist", "info")
	}
}

func eventClass(e *model.Event) string {
	switch e.Type {
	case model.EventKill:
		return "evt-kill"
	case model.EventTeamKill, model.EventSuicide:
		return "evt-bad"
	case model.EventBombPlant, model.EventBombDefuse, model.EventBombExplode, model.EventBombPickup, model.EventBombDrop:
		return "evt-bomb"
	case model.EventGrenade, model.EventGrenadeDamage, model.EventFlash:
		return "evt-utility"
	case model.EventMVP:
		return "evt-mvp"
	case model.EventRoundEnd:
		return "evt-end"
	case model.EventPlayerConnect, model.EventPlayerDisconnect, model.EventTeamChange:
		return "evt-life"
	case model.EventTeamSwap:
		return "evt-swap"
	case model.EventHostagePickup, model.EventHostageDrop, model.EventHostageRescue, model.EventHostageKilled:
		return "evt-hostage"
	default:
		return "evt-" + e.Type
	}
}

func gradeClass(g string) string {
	switch g {
	case "A+":
		return "grade-ap"
	case "A":
		return "grade-a"
	case "B+":
		return "grade-bp"
	case "B":
		return "grade-b"
	case "C+":
		return "grade-cp"
	case "C":
		return "grade-c"
	case "D":
		return "grade-d"
	case "F":
		return "grade-f"
	default:
		return "grade-na"
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

func displayClanTag(tag string) string {
	tag = strings.Trim(strings.TrimSpace(tag), "[]")
	if tag == "" {
		return ""
	}
	var b strings.Builder
	var last rune
	repeats := 0
	count := 0
	words := 0
	for _, r := range tag {
		if !unicode.IsPrint(r) || unicode.IsMark(r) || unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Co, r) {
			continue
		}
		if r == last {
			repeats++
			if repeats > 2 {
				continue
			}
		} else {
			last = r
			repeats = 1
		}
		if count == 10 {
			b.WriteString("...")
			break
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			words++
		}
		b.WriteRune(r)
		count++
	}
	// Decorative tags made only from symbols, ligatures or punctuation are
	// common griefing payloads. They render as an unreadable mess across fonts;
	// omit them instead of guessing what the game client shows.
	if words == 0 {
		return ""
	}
	return strings.TrimSpace(b.String())
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

func searchKey(parts ...string) string {
	var b strings.Builder
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(part)
	}
	return b.String()
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

// plainDashes keeps generated UI text to the project's plain "-" style. It is
// only applied to analyser/parser text, never to untrusted player names.
var dashReplacer = strings.NewReplacer("\u2014", "-", "\u2013", "-", "\u00b7", "-")

func plainDashes(s string) string {
	return dashReplacer.Replace(s)
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
		"clock":  clock,
		"author": func() string { return author },
		"byline": func() string { return byline },
		"one": func(v float64) string {
			return fmt.Sprintf("%.1f", v)
		},
		"zero": func(v float64) string {
			return fmt.Sprintf("%.0f", v)
		},
		"lower":      strings.ToLower,
		"gradeClass": gradeClass,
	}
}
