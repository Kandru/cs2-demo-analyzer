// Package html renders the self-contained match report.
package html

import (
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"

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
	TeamStats    []model.TeamStats
	Warnings     []string
	Cheaters     []*CheaterView
	RoundStrip   []RoundStripView
	DemoDownload *DemoLink
	Chrome       *ChromeView
	Footer       *FooterView
	Assets       *AssetView
}

// ChromeView is the shared topbar model for report and index pages.
type ChromeView struct {
	BrandHref  string
	BrandLabel string
	ShowSearch bool
	Nav        []NavLink
	Meta       []string
	BackHref   string
	BackLabel  string
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
	Version       string
	GeneratedAt   string
	ParserVersion string
}

// AssetView holds inlined brand assets as trusted data URIs.
type AssetView struct {
	BrandMark  template.URL
	Logo       template.URL
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
	Rows     []EventView
}

// DemoLink is a direct download link for a source demo.
type DemoLink struct {
	Href string
	Name string
	Size string
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
	Player       *model.Player
	Anchor       string
	Initials     string
	TeamName     string
	Side         string
	Rounds       []PlayerRoundView
	Sides        []SideView
	Weapons      []WeaponView
	Aim          []MetricView
	Cheat        *CheatView
	Moments      []*MomentView
	TopMoments   []*MomentView
	MomentCount  int
	ShowReview   bool
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
	Label     string
	Display   string
	ScorePct  int
	Samples   int
	LowSample bool
	Tooltip   string
	Class     string
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
	Adjustments   []model.Adjustment
	Narrative     string
	Error         string
}

// ChannelView is one detector channel row.
type ChannelView struct {
	Channel   model.CheatChannel
	Rank      int
	ScorePct  int
	ConfPct   int
	Raw       string
	BarClass  string
	Zone      string
	ZoneClass string
}

// CheaterView groups one flagged or review-level player for the Cheat watch
// section. It reuses the player card data so both surfaces stay consistent.
type CheaterView struct {
	Player      *PlayerView
	Cheat       *CheatView
	TopChannels []ChannelView
	TopMoments  []*MomentView
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
	Badges       []string
	DataPlayers  string
	Class        string
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
		title += " - " + m.Teams[0].Name + " vs " + m.Teams[1].Name
	}
	v.Title = title
	if m.Summary != nil {
		v.TeamStats = m.Summary.TeamStats
	}
	v.Players = buildPlayers(m)
	v.Teams = buildTeams(m, v.TeamStats, v.Players)
	v.RoundStrip = buildRoundStrip(m)
	v.Cheaters = buildCheaters(v.Players)
	v.Timeline = buildTimeline(m)
	markFlaggedRounds(v, m)
	v.Warnings = m.Warnings
	v.Description = fmt.Sprintf("CS2 match report for %s - %s - %s", v.Title, m.Mode, byline)
	v.Chrome = &ChromeView{
		BrandHref:  chromeHref(opts.IndexHref),
		BrandLabel: brandLabel(opts.IndexHref),
		ShowSearch: true,
		Nav: []NavLink{
			{Href: "#overview", Label: "Overview", Active: true},
			{Href: "#scoreboard", Label: "Scoreboard"},
			{Href: "#cheaters", Label: "Watch list", Count: len(v.Cheaters)},
			{Href: "#rounds", Label: "Timeline"},
			{Href: "#players", Label: "Players"},
		},
		Meta: compact(m.Map, m.Mode),
	}
	if opts.IndexHref != "" {
		v.Chrome.BackHref = opts.IndexHref
		v.Chrome.BackLabel = "All reports"
	}
	v.Footer = &FooterView{
		RepoURL:       repoURL,
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

func buildTeams(m *model.Match, stats []model.TeamStats, players []*PlayerView) []*TeamView {
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
		Flagged:      p.Cheat != nil && p.Cheat.Flagged,
		KDDisplay:    fmt.Sprintf("%.2f", p.Stats.KD),
		ADRDisplay:   fmt.Sprintf("%.1f", p.Stats.ADR),
		HSDisplay:    fmt.Sprintf("%.0f%%", p.Stats.HSPercent),
		ScoreDisplay: fmt.Sprintf("%d", p.Stats.Score),
	}
	if team := m.TeamByID(p.TeamID); team != nil {
		pv.TeamName = team.Name
		pv.Side = string(team.StartSide)
	}
	terms := []string{p.ID, p.Name, p.Slug, p.ClanTag, pv.TeamName}
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
		pv.MomentCount = len(pv.Moments)
		pv.TopMoments = topMoments(pv.Moments, 8)
	}
	if pv.Cheat != nil && pv.Cheat.Available && (pv.Flagged || pv.Cheat.Likelihood >= 25) {
		pv.ShowReview = true
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
		Label:     mt.Label,
		Display:   plainDashes(mt.Display),
		ScorePct:  clampPct(mt.Score),
		Samples:   mt.Samples,
		LowSample: mt.LowSample,
		Tooltip:   plainDashes(mt.Tooltip),
		Class:     class,
	}
}

func buildCheat(cr *model.CheatReport) *CheatView {
	if cr == nil {
		return &CheatView{Available: false, Verdict: "Not analyzed", VerdictClass: "unavailable", Error: "anti-cheat detector was not run"}
	}
	cv := &CheatView{
		Available:     cr.Available,
		OverallGrade:  cr.OverallGrade,
		Grades:        cr.Grades,
		Adjustments:   cr.Adjustments,
		Narrative:     plainDashes(cr.Narrative),
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
		ch.Description = plainDashes(ch.Description)
		cvm := ChannelView{
			Channel:   ch,
			Rank:      len(cv.Channels) + 1,
			ScorePct:  clampPct(ch.Score * 100),
			ConfPct:   clampPct(ch.Confidence * 100),
			BarClass:  zoneClass(ch.Zone),
			Zone:      ch.Zone,
			ZoneClass: zoneClass(ch.Zone),
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
			TopMoments:  topMoments(pv.Moments, 10),
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

func topMoments(in []*MomentView, n int) []*MomentView {
	if len(in) == 0 {
		return nil
	}
	sorted := append([]*MomentView(nil), in...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].rank != sorted[j].rank {
			return sorted[i].rank > sorted[j].rank
		}
		return sorted[i].Tick < sorted[j].Tick
	})
	if len(sorted) > n {
		sorted = sorted[:n]
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Tick < sorted[j].Tick
	})
	return sorted
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
		return t.Name
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
			if e.Type == "round_start" {
				continue
			}
			if e.Type == "round_end" {
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
				Type:   "round_end",
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
			ev.TeamName = t.Name
		}
		ev.Side = string(e.Side)
		out = append(out, ev)
	}
	return out
}

// globalEventsWithFallback keeps older cached reports useful: when the parser
// did not record a team_swap event, detect the side flip from round economy.
func globalEventsWithFallback(m *model.Match) []*model.Event {
	events := append([]*model.Event(nil), m.Events...)
	for _, e := range events {
		if e.Type == "team_swap" {
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
				if team := m.TeamByID(id); team != nil && team.Name != "" {
					if side == model.SideCT {
						ctName = team.Name
					} else if side == model.SideT {
						tName = team.Name
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
			events = append(events, &model.Event{Type: "team_swap", Offset: offset, Text: text})
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
	"round_start":       "flag",
	"freeze_end":        "buy",
	"kill":              "kill",
	"teamkill":          "warning",
	"suicide":           "warning",
	"bomb_plant":        "bomb",
	"bomb_defuse":       "defuse",
	"bomb_explode":      "explosion",
	"grenade":           "grenade",
	"mvp":               "star",
	"round_end":         "trophy",
	"player_connect":    "connect",
	"player_disconnect": "disconnect",
	"team_change":       "swap",
	"team_swap":         "swap",
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
	case "round_start":
		ev.Parts = []EventPart{{Text: "Round started"}}
	case "freeze_end":
		ev.Parts = []EventPart{{Text: "Freeze time ended - weapons are live"}}
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
			ev.Parts = append(ev.Parts, EventPart{Text: fmt.Sprintf(" - %d : %d", e.Score[0], e.Score[1])})
		}
	case "player_connect":
		ev.Parts = append(ev.Parts, link(e.ActorID, actorName(e)))
		ev.Parts = append(ev.Parts, EventPart{Text: " connected"})
	case "player_disconnect":
		ev.Parts = append(ev.Parts, link(e.ActorID, actorName(e)))
		ev.Parts = append(ev.Parts, EventPart{Text: " disconnected"})
	case "team_change":
		ev.Parts = append(ev.Parts, link(e.ActorID, actorName(e)))
		ev.Parts = append(ev.Parts, EventPart{Text: teamChangeText(e)})
		if e.NewSide != model.SideUnknown {
			ev.Badges = append(ev.Badges, string(e.NewSide))
		}
	case "team_swap":
		text := e.Text
		if text == "" {
			text = "Teams switched sides"
		}
		ev.Parts = []EventPart{{Text: text}}
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
	case "player_connect", "player_disconnect", "team_change":
		return "evt-life"
	case "team_swap":
		return "evt-swap"
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
