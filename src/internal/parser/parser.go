// Package parser converts a CS2 demo into the battlelog domain model.
//
// It uses one demoinfocs pass for the match timeline and one aim tracker. The
// vendored anti-cheat detector runs as a separate pass in internal/anticheat,
// so its handlers can never interfere with the battlelog timeline.
package parser

import (
	"compress/bzip2"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	dem "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/msg"

	"github.com/Kandru/cs2-demo-analyzer/internal/aim"
	"github.com/Kandru/cs2-demo-analyzer/internal/model"
)

// ErrUnsupportedDemo is returned for CS:GO/Source 1 demos, which demoinfocs v5
// cannot parse.
var ErrUnsupportedDemo = errors.New("unsupported demo format (CS:GO Source 1 demos need demoinfocs v3)")

// ErrNoPlayers is returned when a demo contains no playable participants.
var ErrNoPlayers = errors.New("demo contains no playable players")

const parserVersion = "demoinfocs-golang/v5 v5.3.0"

// Options controls one parse.
type Options struct {
	Logger *slog.Logger
	// SkipAim disables aim stat collection (useful for fast smoke runs).
	SkipAim bool
}

// Parse reads one demo and returns a match model. A partial model plus an error
// is returned when the demo is truncated or corrupt after useful data has been
// read.
func Parse(ctx context.Context, path string, opts Options) (m *model.Match, err error) {
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	abs, aerr := filepath.Abs(path)
	if aerr != nil {
		abs = path
	}
	info, serr := os.Stat(path)
	if serr != nil {
		return nil, fmt.Errorf("stat demo: %w", serr)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory", path)
	}

	m = &model.Match{
		SchemaVersion: model.SchemaVersion,
		DemoPath:      abs,
		DemoName:      filepath.Base(path),
		DemoSize:      info.Size(),
		FileTime:      info.ModTime(),
		AnalyzedAt:    time.Now().UTC(),
		ParserVersion: parserVersion,
		Teams:         []*model.Team{},
		Players:       []*model.Player{},
		Rounds:        []*model.Round{},
		Warmup:        &model.Warmup{Events: []*model.Event{}},
	}

	var st *state
	defer func() {
		if r := recover(); r != nil {
			m.Partial = true
			m.ParseError = fmt.Sprintf("parser panic: %v", r)
			m.Warnings = append(m.Warnings, m.ParseError)
			opts.Logger.Error("demo parser panicked", "demo", path, "panic", r, "stack", string(debug.Stack()))
			if st != nil {
				func() {
					defer func() { _ = recover() }()
					st.finishPartial()
					st.finalize()
				}()
			}
			err = fmt.Errorf("parser panic: %v", r)
		}
	}()

	f, err := os.Open(path)
	if err != nil {
		return m, fmt.Errorf("open demo: %w", err)
	}
	defer f.Close()

	rd, closeReader, err := decompress(f, path)
	if err != nil {
		return m, err
	}
	if closeReader != nil {
		defer closeReader()
	}

	// Hash the raw file bytes in parallel with parsing is unnecessary; the
	// store may hash separately. Keep this field empty unless supplied.
	p := dem.NewParser(rd)
	defer p.Close()

	st = newState(ctx, m, p, opts)
	st.installHandlers()
	if err := st.parse(); err != nil {
		st.finishPartial()
		if len(m.Players) == 0 {
			return m, err
		}
		m.Partial = true
		m.ParseError = err.Error()
		m.Warnings = append(m.Warnings, err.Error())
		opts.Logger.Warn("demo parsed partially", "demo", path, "error", err, "rounds", len(m.Rounds), "players", len(m.Players))
		err = nil
	}
	st.finalize()
	if len(m.Players) == 0 {
		return m, ErrNoPlayers
	}
	return m, nil
}

// HashFile returns the SHA-256 of a demo, used by the manifest cache.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func decompress(f *os.File, path string) (io.Reader, func() error, error) {
	name := strings.ToLower(path)
	switch {
	case strings.HasSuffix(name, ".gz"):
		zr, err := gzip.NewReader(f)
		if err != nil {
			return nil, nil, fmt.Errorf("open gzip demo: %w", err)
		}
		return zr, zr.Close, nil
	case strings.HasSuffix(name, ".bz2"):
		return bzip2.NewReader(f), nil, nil
	}
	// Sniff the magic to catch compressed demos with a .dem extension.
	var magic [8]byte
	n, _ := io.ReadFull(f, magic[:])
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, nil, err
	}
	if n >= 2 && magic[0] == 0x1f && magic[1] == 0x8b {
		zr, err := gzip.NewReader(f)
		if err != nil {
			return nil, nil, fmt.Errorf("open gzip demo: %w", err)
		}
		return zr, zr.Close, nil
	}
	if n >= 3 && magic[0] == 'B' && magic[1] == 'Z' && magic[2] == 'h' {
		return bzip2.NewReader(f), nil, nil
	}
	if n >= 7 {
		head := string(magic[:7])
		if head == "PBDEMS2" || head == "PCDEMS2" {
			return f, nil, nil
		}
		if strings.HasPrefix(string(magic[:3]), "HL2") {
			return nil, nil, ErrUnsupportedDemo
		}
	}
	return f, nil, nil
}

type state struct {
	ctx     context.Context
	m       *model.Match
	p       dem.Parser
	log     *slog.Logger
	aim     *aim.Tracker
	skipAim bool

	teamsByID   map[int]*model.Team
	playersByID map[string]*model.Player
	playerSlugs map[string]bool

	current       *roundBuilder
	official      int
	live          bool
	warmupStart   float64
	warmupStarted bool
	warmupEnded   bool
	lastTick      int
	eventSeq      int

	recentKills []recentKill
	pendingMVP  string
	hadPlayers  bool
	header      *msg.CDemoFileHeader
}

type roundBuilder struct {
	round         *model.Round
	hadDamage     bool
	finalized     bool
	startTick     int
	startScore    [2]int
	mvpAnnounced  string
	teamIDAtStart [2]int
}

type recentKill struct {
	tick     int
	killerID string
	victimID string
	teamID   int
}

func newState(ctx context.Context, m *model.Match, p dem.Parser, opts Options) *state {
	tr := aim.New(p.TickRate())
	if opts.SkipAim {
		tr = nil
	}
	return &state{
		ctx:         ctx,
		m:           m,
		p:           p,
		log:         opts.Logger,
		aim:         tr,
		skipAim:     opts.SkipAim,
		teamsByID:   map[int]*model.Team{},
		playersByID: map[string]*model.Player{},
		playerSlugs: map[string]bool{},
	}
}

func (s *state) installHandlers() {
	p := s.p
	p.RegisterNetMessageHandler(func(h *msg.CDemoFileHeader) {
		s.header = h
		s.m.Map = h.GetMapName()
		s.m.ServerName = h.GetServerName()
		s.m.DemoStamp = h.GetDemoFileStamp()
	})
	p.RegisterEventHandler(func(e events.RoundStart) { s.onRoundStart(e) })
	p.RegisterEventHandler(func(e events.RoundFreezetimeEnd) { s.onFreezeEnd(e) })
	p.RegisterEventHandler(func(e events.RoundEnd) { s.onRoundEnd(e) })
	p.RegisterEventHandler(func(e events.RoundEndOfficial) { s.finalizeRound(nil) })
	p.RegisterEventHandler(func(e events.RoundMVPAnnouncement) { s.onMVP(e) })
	p.RegisterEventHandler(func(e events.Kill) { s.onKill(e) })
	p.RegisterEventHandler(func(e events.PlayerHurt) { s.onHurt(e) })
	p.RegisterEventHandler(func(e events.BombPlanted) { s.onBombPlanted(e) })
	p.RegisterEventHandler(func(e events.BombDefused) { s.onBombDefused(e) })
	p.RegisterEventHandler(func(e events.BombExplode) { s.onBombExplode(e) })
	p.RegisterEventHandler(func(e events.GrenadeProjectileThrow) { s.onGrenade(e) })
	p.RegisterEventHandler(func(e events.IsWarmupPeriodChanged) {
		// The authoritative round transition is handled in onRoundStart.
		if e.NewIsWarmupPeriod {
			s.live = false
			if s.current == nil && !s.warmupStarted {
				s.warmupStart = s.offset()
				s.warmupStarted = true
			}
		}
	})
	if s.aim != nil {
		p.RegisterEventHandler(func(e events.WeaponFire) {
			s.aim.SetTick(s.p.CurrentFrame())
			s.aim.OnWeaponFire(e)
		})
		p.RegisterEventHandler(func(e events.PlayerHurt) {
			s.aim.SetTick(s.p.CurrentFrame())
			s.aim.OnPlayerHurt(e)
		})
		p.RegisterEventHandler(func(e events.Kill) {
			s.aim.SetTick(s.p.CurrentFrame())
			s.aim.OnKill(e)
		})
	}
}

func (s *state) parse() error {
	for {
		if err := s.ctx.Err(); err != nil {
			return err
		}
		ok, err := s.p.ParseNextFrame()
		if err != nil {
			return fmt.Errorf("parse frame: %w", err)
		}
		if !ok {
			return nil
		}
		tick := s.p.CurrentFrame()
		s.lastTick = tick
		if s.aim != nil {
			s.aim.OnFrameState(s.p.GameState().Participants().Playing(), tick)
		}
	}
}

func (s *state) offset() float64 {
	tr := s.p.TickRate()
	if tr <= 0 {
		tr = 64
	}
	return float64(s.lastTick) / tr
}

func (s *state) onRoundStart(e events.RoundStart) {
	gs := s.p.GameState()
	tick := s.p.CurrentFrame()
	s.lastTick = tick
	s.ensureTeams()
	s.ensurePlayers(gs.Participants().Playing())

	if gs.IsWarmupPeriod() {
		s.live = false
		if s.current != nil {
			s.discardCurrent("warmup started")
		}
		if !s.warmupStarted {
			s.m.Warmup.StartOffset = s.offset()
			s.warmupStarted = true
		}
		if s.aim != nil {
			s.aim.SetLive(false)
		}
		return
	}

	// Restart/duplicate round start without an end: keep the events but mark
	// the previous attempt as a pregame round.
	if s.current != nil && !s.current.finalized {
		s.finalizeRound(&finalizeOpts{Reason: "Round restarted", Pregame: true, NoScore: true})
	}

	s.live = true
	if !s.warmupEnded {
		s.m.Warmup.EndOffset = s.offset()
		s.m.Warmup.Duration = s.m.Warmup.EndOffset - s.m.Warmup.StartOffset
		s.warmupEnded = true
	}
	if s.aim != nil {
		s.aim.SetLive(true)
		s.aim.OnRoundStart(tick)
	}

	startScore := s.currentScore()
	if s.official == 0 && (startScore[0] > 0 || startScore[1] > 0) {
		s.m.Warnings = append(s.m.Warnings, "demo starts mid-match; only rounds observed in the demo are shown and scores begin at the demo's current score")
	}
	r := &model.Round{
		StartTick:   tick,
		StartOffset: s.offset(),
		Number:      s.official + 1,
		StartScore:  startScore,
		Events:      []*model.Event{},
		Players:     map[string]*model.RoundStat{},
	}
	s.current = &roundBuilder{round: r, startTick: tick, startScore: startScore}
	s.recentKills = nil
	s.addEvent(&model.Event{Type: "round_start", Offset: 0})
	s.initRoundPlayers()
	s.updateEconomy()
}

func (s *state) onFreezeEnd(events.RoundFreezetimeEnd) {
	s.syncTick()
	if !s.live || s.current == nil {
		return
	}
	s.addEvent(&model.Event{Type: "freeze_end"})
	s.updateEconomy()
}

func (s *state) onRoundEnd(e events.RoundEnd) {
	s.syncTick()
	if s.current == nil {
		return
	}
	opts := &finalizeOpts{
		Winner:  e.Winner,
		Reason:  reasonString(e.Reason),
		Message: e.Message,
	}
	if t := s.teamState(e.Winner); t != nil {
		opts.WinnerTeamID = s.teamIDForState(t)
	}
	s.addEvent(&model.Event{Type: "round_end", Text: opts.Reason, Score: s.endScore(opts)})
	s.finalizeRound(opts)
}

func (s *state) onMVP(e events.RoundMVPAnnouncement) {
	s.syncTick()
	if e.Player == nil || s.current == nil {
		return
	}
	s.current.mvpAnnounced = playerID(e.Player)
}

type finalizeOpts struct {
	Winner       common.Team
	WinnerTeamID int
	Reason       string
	Message      string
	Pregame      bool
	NoScore      bool
}

func (s *state) endScore(o *finalizeOpts) [2]int {
	score := s.current.startScore
	if o.NoScore {
		return score
	}
	idx := s.teamIndex(o.WinnerTeamID)
	if idx >= 0 {
		score[idx]++
	}
	return score
}

func (s *state) finalizeRound(o *finalizeOpts) {
	s.syncTick()
	rb := s.current
	if rb == nil || rb.finalized {
		return
	}
	rb.finalized = true
	r := rb.round
	r.EndTick = s.lastTick
	r.EndOffset = s.offset()
	if r.EndOffset < r.StartOffset {
		r.EndOffset = r.StartOffset
	}
	r.Duration = r.EndOffset - r.StartOffset

	pregame := o != nil && o.Pregame
	if o == nil {
		// RoundEndOfficial without a RoundEnd: close with whatever score is on
		// the team states, but don't guess a winner.
		r.Reason = "Round ended"
		r.EndScore = s.currentScore()
	} else {
		r.Winner = sideString(o.Winner)
		r.WinnerTeamID = o.WinnerTeamID
		r.Reason = o.Reason
		r.Message = o.Message
		if pregame || o.NoScore {
			r.EndScore = r.StartScore
		} else {
			r.EndScore = s.endScore(o)
		}
	}

	// A candidate round with no score movement and no damage is a knife round
	// or restart, not an official round.
	if !pregame && !rb.hadDamage && r.EndScore == r.StartScore && r.Winner != sideString(common.TeamSpectators) {
		pregame = true
	}
	r.Pregame = pregame
	regulation := 24
	if len(s.m.Players) <= 4 {
		regulation = 8
	}
	r.Overtime = !pregame && s.official >= regulation
	if pregame {
		r.Number = 0
	} else {
		s.official++
	}

	s.finishRoundPlayers(rb)
	if !pregame && s.current != nil {
		s.awardMVP(rb)
	}
	r.Players = rb.round.Players
	s.m.Rounds = append(s.m.Rounds, r)
	if !pregame {
		s.applyRoundToPlayers(r)
		s.updateTeams(r.EndScore)
	}
	s.current = nil
}

func (s *state) discardCurrent(reason string) {
	if s.current == nil {
		return
	}
	s.finalizeRound(&finalizeOpts{Reason: reason, Pregame: true, NoScore: true})
}

func (s *state) finishPartial() {
	if s.current != nil && !s.current.finalized {
		s.finalizeRound(&finalizeOpts{Reason: "Demo ended mid-round", NoScore: true})
	}
}

func (s *state) currentScore() [2]int {
	var score [2]int
	gs := s.p.GameState()
	for _, side := range []common.Team{common.TeamTerrorists, common.TeamCounterTerrorists} {
		ts := gs.Team(side)
		if ts == nil {
			continue
		}
		idx := s.teamIndexForState(ts)
		if idx < 0 {
			continue
		}
		score[idx] = ts.Score()
	}
	return score
}

func (s *state) teamIndex(teamID int) int {
	for i, t := range s.m.Teams {
		if t.ID == teamID {
			return i
		}
	}
	return -1
}

// teamIndexForState maps a live TeamState to the stable team index captured at
// match start. TeamState.ID is side-based in some CS2 demos (T=2, CT=3), so it
// cannot be used after a side switch; the roster is the stable identity.
func (s *state) teamIndexForState(ts *common.TeamState) int {
	if ts == nil {
		return -1
	}
	counts := map[int]int{}
	for _, p := range ts.Members() {
		if p == nil {
			continue
		}
		mp := s.ensurePlayer(p)
		if mp == nil {
			continue
		}
		counts[mp.TeamID]++
	}
	bestID, bestN := 0, 0
	for id, n := range counts {
		if n > bestN || (n == bestN && id < bestID) {
			bestID, bestN = id, n
		}
	}
	if bestN > 0 {
		if idx := s.teamIndex(bestID); idx >= 0 {
			return idx
		}
	}
	// Fallback before rosters are known: map the current side to the team that
	// started on that side. Only correct in the first half, hence the roster
	// path above is preferred.
	for i, t := range s.m.Teams {
		if t.StartSide == sideString(ts.Team()) {
			return i
		}
	}
	return -1
}

func (s *state) teamIDForState(ts *common.TeamState) int {
	idx := s.teamIndexForState(ts)
	if idx >= 0 {
		return s.m.Teams[idx].ID
	}
	return 0
}

func (s *state) teamState(t common.Team) *common.TeamState {
	if t != common.TeamTerrorists && t != common.TeamCounterTerrorists {
		return nil
	}
	return s.p.GameState().Team(t)
}

func (s *state) updateTeams(score [2]int) {
	for i, t := range s.m.Teams {
		t.Score = score[i]
		t.FinalSide = s.teamFinalSide(t.ID)
	}
}

func (s *state) teamFinalSide(teamID int) model.Side {
	// Score/MVP state at EOF may be gone; prefer the last round's player sides.
	counts := map[model.Side]int{}
	for i := len(s.m.Rounds) - 1; i >= 0; i-- {
		last := s.m.Rounds[i]
		if last.Pregame {
			continue
		}
		for id, rs := range last.Players {
			if rs == nil {
				continue
			}
			if p := s.m.PlayerByID(id); p != nil && p.TeamID == teamID {
				counts[rs.Side]++
			}
		}
		break
	}
	if len(counts) == 0 {
		for _, cp := range s.p.GameState().Participants().Playing() {
			if cp == nil {
				continue
			}
			if mp := s.playersByID[playerID(cp)]; mp != nil && mp.TeamID == teamID {
				counts[sideString(cp.Team)]++
			}
		}
	}
	best, n := model.Side(""), 0
	for side, count := range counts {
		if count > n {
			best, n = side, count
		}
	}
	return best
}

func (s *state) ensureTeams() {
	gs := s.p.GameState()
	for _, side := range []common.Team{common.TeamTerrorists, common.TeamCounterTerrorists} {
		ts := gs.Team(side)
		if ts == nil || ts.ID() == 0 {
			continue
		}
		id := ts.ID()
		if _, ok := s.teamsByID[id]; ok {
			continue
		}
		name := strings.TrimSpace(ts.ClanName())
		if name == "" {
			if len(s.m.Teams) == 0 {
				name = "Team A"
			} else {
				name = "Team B"
			}
		}
		t := &model.Team{
			ID:        id,
			Name:      name,
			ClanName:  name,
			Flag:      ts.Flag(),
			StartSide: sideString(side),
			FinalSide: sideString(side),
			PlayerIDs: []string{},
		}
		s.teamsByID[id] = t
		s.m.Teams = append(s.m.Teams, t)
	}
}

func (s *state) ensurePlayers(players []*common.Player) {
	for _, p := range players {
		s.ensurePlayer(p)
	}
}

func (s *state) ensurePlayer(p *common.Player) *model.Player {
	if p == nil || p.IsUnknown {
		return nil
	}
	if p.Team != common.TeamTerrorists && p.Team != common.TeamCounterTerrorists {
		return nil
	}
	id := playerID(p)
	if id == "" {
		return nil
	}
	if mp, ok := s.playersByID[id]; ok {
		if p.Name != "" && mp.Name != p.Name {
			mp.Aliases = appendUnique(mp.Aliases, p.Name)
			mp.Name = p.Name
		}
		if p.ClanTag() != "" {
			mp.ClanTag = p.ClanTag()
		}
		return mp
	}
	teamID := s.teamIDForNewPlayer(p)
	mp := &model.Player{
		ID:         id,
		SteamID64:  p.SteamID64,
		Name:       p.Name,
		ClanTag:    p.ClanTag(),
		TeamID:     teamID,
		IsBot:      p.IsBot,
		Slug:       s.uniqueSlug(p.Name, id),
		SideStats:  map[model.Side]*model.PlayerStats{},
		RoundStats: map[int]*model.RoundStat{},
		Weapons:    map[string]*model.WeaponStats{},
	}
	if mp.Name == "" {
		mp.Name = fmt.Sprintf("Player %d", p.EntityID)
	}
	s.m.Players = append(s.m.Players, mp)
	s.playersByID[id] = mp
	s.hadPlayers = true
	if t := s.teamsByID[teamID]; t != nil {
		t.PlayerIDs = appendUnique(t.PlayerIDs, mp.ID)
	}
	return mp
}

// teamIDForNewPlayer assigns a player to the stable team captured at match
// start. A late substitute may first appear after a side switch, where
// TeamState.ID is no longer meaningful; the existing roster is the best
// identity signal.
func (s *state) teamIDForNewPlayer(p *common.Player) int {
	if p == nil {
		return 0
	}
	if p.TeamState != nil {
		counts := map[int]int{}
		for _, mate := range p.TeamState.Members() {
			if mate == nil {
				continue
			}
			if mp := s.playersByID[playerID(mate)]; mp != nil {
				counts[mp.TeamID]++
			}
		}
		best, n := 0, 0
		for id, c := range counts {
			if c > n {
				best, n = id, c
			}
		}
		if n > 0 {
			return best
		}
	}
	// First player of a side: map to the team that started there.
	for _, t := range s.m.Teams {
		if t.StartSide == sideString(p.Team) {
			return t.ID
		}
	}
	if p.TeamState != nil {
		return p.TeamState.ID()
	}
	return 0
}

func (s *state) uniqueSlug(name, id string) string {
	base := slugify(name)
	if base == "" {
		base = "player"
	}
	short := id
	if len(short) > 6 {
		short = short[len(short)-6:]
	}
	cand := base + "-" + short
	for i := 2; s.playerSlugs[cand]; i++ {
		cand = fmt.Sprintf("%s-%s-%d", base, short, i)
	}
	s.playerSlugs[cand] = true
	return cand
}

func playerID(p *common.Player) string {
	if p == nil {
		return ""
	}
	if p.SteamID64 != 0 {
		return fmt.Sprintf("%d", p.SteamID64)
	}
	if p.EntityID != 0 {
		return fmt.Sprintf("bot-%d", p.EntityID)
	}
	return ""
}

func (s *state) syncTick() {
	if s.p != nil {
		s.lastTick = s.p.CurrentFrame()
	}
}

func (s *state) addEvent(e *model.Event) {
	s.syncTick()
	if s.current == nil {
		if !s.live {
			s.addWarmupEvent(e)
		}
		return
	}
	if !s.live {
		s.addWarmupEvent(e)
		return
	}
	s.eventSeq++
	e.Seq = s.eventSeq
	e.Tick = s.lastTick
	if e.Offset == 0 {
		e.Offset = s.offset() - s.current.round.StartOffset
	}
	s.current.round.Events = append(s.current.round.Events, e)
}

func (s *state) addWarmupEvent(e *model.Event) {
	s.eventSeq++
	e.Seq = s.eventSeq
	e.Tick = s.lastTick
	if e.Offset == 0 {
		e.Offset = s.offset()
	}
	s.m.Warmup.Events = append(s.m.Warmup.Events, e)
}

func (s *state) roundStat(id string) *model.RoundStat {
	if s.current == nil {
		return nil
	}
	rs := s.current.round.Players[id]
	if rs == nil {
		rs = &model.RoundStat{}
		s.current.round.Players[id] = rs
	}
	return rs
}

func (s *state) initRoundPlayers() {
	if s.current == nil {
		return
	}
	for _, p := range s.p.GameState().Participants().Playing() {
		if p == nil || p.Team == common.TeamSpectators || p.Team == common.TeamUnassigned {
			continue
		}
		mp := s.ensurePlayer(p)
		if mp == nil {
			continue
		}
		rs := s.roundStat(mp.ID)
		rs.Side = sideString(p.Team)
	}
}

func (s *state) updateEconomy() {
	if s.current == nil || !s.live {
		return
	}
	gs := s.p.GameState()
	var econ []*model.TeamEconomy
	for _, side := range []common.Team{common.TeamTerrorists, common.TeamCounterTerrorists} {
		ts := gs.Team(side)
		if ts == nil {
			continue
		}
		money := 0
		for _, p := range ts.Members() {
			if p == nil {
				continue
			}
			money += p.Money()
			if mp := s.ensurePlayer(p); mp != nil {
				if rs := s.roundStat(mp.ID); rs != nil {
					rs.EquipValue = p.EquipmentValueFreezeTimeEnd()
					rs.Spent = p.MoneySpentThisRound()
					rs.Side = sideString(p.Team)
				}
			}
		}
		equip := ts.FreezeTimeEndEquipmentValue()
		econ = append(econ, &model.TeamEconomy{
			TeamID:         ts.ID(),
			Side:           sideString(side),
			EquipmentValue: equip,
			Money:          money,
			BuyType:        buyType(equip, s.current.round.Number),
		})
	}
	s.current.round.Economy = econ
}

func (s *state) onKill(e events.Kill) {
	s.syncTick()
	if e.Victim == nil {
		return
	}
	vp := s.ensurePlayer(e.Victim)
	if vp == nil {
		return
	}
	vName := vp.Name
	if !s.live {
		ev := &model.Event{Type: "kill", VictimID: vp.ID, VictimName: vName}
		if e.Killer != nil {
			if kp := s.ensurePlayer(e.Killer); kp != nil {
				ev.ActorID, ev.ActorName = kp.ID, kp.Name
			}
		}
		if e.Weapon != nil {
			ev.Weapon = e.Weapon.String()
		}
		ev.Headshot, ev.Wallbang, ev.Noscope, ev.ThroughSmoke = e.IsHeadshot, e.IsWallBang(), e.NoScope, e.ThroughSmoke
		s.addWarmupEvent(ev)
		return
	}
	if s.current == nil {
		return
	}

	ev := &model.Event{
		Type:          "kill",
		VictimID:      vp.ID,
		VictimName:    vName,
		Headshot:      e.IsHeadshot,
		Wallbang:      e.IsWallBang(),
		Noscope:       e.NoScope,
		ThroughSmoke:  e.ThroughSmoke,
		AssistedFlash: e.AssistedFlash,
	}
	if e.Weapon != nil {
		ev.Weapon = e.Weapon.String()
	}
	if e.Assister != nil {
		if ap := s.ensurePlayer(e.Assister); ap != nil {
			ev.AssisterID, ev.AssisterName = ap.ID, ap.Name
		}
	}

	vrs := s.roundStat(vp.ID)
	vrs.Deaths++
	vrs.Survived = false
	s.current.hadDamage = true
	if vrs.FirstDeath == false && !s.roundHasEvent("kill") {
		vrs.FirstDeath = true
	}

	// Teamkill and suicide handling.
	if e.Killer == nil || e.Killer == e.Victim {
		ev.Type = "suicide"
		ev.Suicide = true
		s.addEvent(ev)
		s.updateClutchAfterKill()
		return
	}
	if kp := s.ensurePlayer(e.Killer); kp != nil {
		ev.ActorID, ev.ActorName = kp.ID, kp.Name
	}
	if e.Killer.Team == e.Victim.Team {
		ev.Type = "teamkill"
		ev.TeamKill = true
		s.addEvent(ev)
		s.updateClutchAfterKill()
		return
	}

	kr := s.ensurePlayer(e.Killer)
	if kr == nil {
		s.addEvent(ev)
		return
	}
	krs := s.roundStat(kr.ID)
	krs.Kills++
	if e.IsHeadshot {
		krs.Headshots++
	}
	if !s.roundHasEvent("kill") {
		krs.FirstKill = true
	}
	if e.Assister != nil {
		if ap := s.ensurePlayer(e.Assister); ap != nil && ap.ID != kr.ID && ap.ID != vp.ID {
			ars := s.roundStat(ap.ID)
			ars.Assists++
			if e.AssistedFlash {
				ars.FlashAssists++
			}
		}
	}
	s.recordWeaponKill(kr.ID, ev.Weapon, e.IsHeadshot)
	s.detectTrade(kr.ID, vp.ID)
	s.addEvent(ev)
	s.updateClutchAfterKill()
}

func (s *state) detectTrade(killerID, victimID string) {
	if s.current == nil {
		return
	}
	tick := s.lastTick
	window := int(5 * s.p.TickRate())
	for i := len(s.recentKills) - 1; i >= 0; i-- {
		k := s.recentKills[i]
		if tick-k.tick > window {
			break
		}
		if k.victimID == killerID {
			// The current victim is avenging their teammate.
			if prev := s.current.round.Players[k.victimID]; prev != nil {
				prev.Traded = true
			}
			if rs := s.current.round.Players[killerID]; rs != nil {
				rs.TradeKill = true
			}
			break
		}
	}
	s.recentKills = append(s.recentKills, recentKill{tick: tick, killerID: killerID, victimID: victimID})
	if len(s.recentKills) > 32 {
		s.recentKills = s.recentKills[len(s.recentKills)-32:]
	}
}

func (s *state) updateClutchAfterKill() {
	if s.current == nil {
		return
	}
	gs := s.p.GameState()
	for _, side := range []common.Team{common.TeamTerrorists, common.TeamCounterTerrorists} {
		var alive, enemies int
		var solo *common.Player
		other := common.TeamCounterTerrorists
		if side == common.TeamCounterTerrorists {
			other = common.TeamTerrorists
		}
		for _, p := range gs.Participants().Playing() {
			if p == nil || !p.IsAlive() {
				continue
			}
			switch p.Team {
			case side:
				alive++
				solo = p
			case other:
				enemies++
			}
		}
		if alive == 1 && enemies >= 1 && solo != nil {
			if mp := s.ensurePlayer(solo); mp != nil {
				if rs := s.roundStat(mp.ID); rs != nil && enemies > rs.Clutch {
					rs.Clutch = enemies
				}
			}
		}
	}
}

func (s *state) roundHasEvent(typ string) bool {
	if s.current == nil {
		return false
	}
	for _, e := range s.current.round.Events {
		if e.Type == typ {
			return true
		}
	}
	return false
}

func (s *state) onHurt(e events.PlayerHurt) {
	s.syncTick()
	if !s.live || s.current == nil || e.Attacker == nil || e.Player == nil {
		return
	}
	if e.Attacker == e.Player || e.Attacker.Team == e.Player.Team {
		return
	}
	ap := s.ensurePlayer(e.Attacker)
	vp := s.ensurePlayer(e.Player)
	if ap == nil || vp == nil {
		return
	}
	ars := s.roundStat(ap.ID)
	vrs := s.roundStat(vp.ID)
	dmg := e.HealthDamageTaken
	if dmg <= 0 {
		dmg = e.HealthDamage
	}
	if dmg < 0 {
		dmg = 0
	}
	ars.Damage += dmg
	vp.Stats.DamageTaken += dmg
	_ = vrs
	if e.Weapon != nil && dmg > 0 {
		s.recordWeaponDamage(ap.ID, e.Weapon.String(), dmg)
	}
	if dmg > 0 {
		s.current.hadDamage = true
	}
	if e.Weapon != nil && (e.Weapon.Type.Class() == common.EqClassGrenade ||
		e.Weapon.Type == common.EqHE || e.Weapon.Type == common.EqMolotov || e.Weapon.Type == common.EqIncendiary) {
		ars.UtilityDamage += dmg
	}
}

func (s *state) onBombPlanted(e events.BombPlanted) {
	s.syncTick()
	if e.Player == nil {
		return
	}
	p := s.ensurePlayer(e.Player)
	if p == nil {
		return
	}
	e2 := &model.Event{Type: "bomb_plant", ActorID: p.ID, ActorName: p.Name, Site: string(e.Site)}
	if s.live && s.current != nil {
		if rs := s.roundStat(p.ID); rs != nil {
			rs.Plants++
		}
	}
	s.addEvent(e2)
}

func (s *state) onBombDefused(e events.BombDefused) {
	s.syncTick()
	if e.Player == nil {
		return
	}
	p := s.ensurePlayer(e.Player)
	if p == nil {
		return
	}
	if s.live && s.current != nil {
		if rs := s.roundStat(p.ID); rs != nil {
			rs.Defuses++
		}
	}
	s.addEvent(&model.Event{Type: "bomb_defuse", ActorID: p.ID, ActorName: p.Name, Site: string(e.Site)})
}

func (s *state) onBombExplode(e events.BombExplode) {
	s.syncTick()
	s.addEvent(&model.Event{Type: "bomb_explode", Site: string(e.Site), Text: "The bomb has exploded"})
}

func (s *state) onGrenade(e events.GrenadeProjectileThrow) {
	s.syncTick()
	if e.Projectile == nil || e.Projectile.Thrower == nil || e.Projectile.WeaponInstance == nil {
		return
	}
	if !s.live {
		return
	}
	p := s.ensurePlayer(e.Projectile.Thrower)
	if p == nil {
		return
	}
	name := e.Projectile.WeaponInstance.String()
	s.addEvent(&model.Event{Type: "grenade", ActorID: p.ID, ActorName: p.Name, Grenade: name, Weapon: name})
}

func (s *state) recordWeaponDamage(playerID, weapon string, dmg int) {
	if weapon == "" {
		return
	}
	p := s.m.PlayerByID(playerID)
	if p == nil {
		return
	}
	ws := p.Weapons[weapon]
	if ws == nil {
		ws = &model.WeaponStats{Name: weapon}
		p.Weapons[weapon] = ws
	}
	ws.Damage += dmg
}

func (s *state) recordWeaponKill(playerID, weapon string, hs bool) {
	if weapon == "" {
		weapon = "Unknown"
	}
	p := s.m.PlayerByID(playerID)
	if p == nil {
		return
	}
	ws := p.Weapons[weapon]
	if ws == nil {
		ws = &model.WeaponStats{Name: weapon}
		p.Weapons[weapon] = ws
	}
	ws.Kills++
	if hs {
		ws.Headshots++
	}
}

func (s *state) applyRoundToPlayers(r *model.Round) {
	for id, rs := range r.Players {
		p := s.m.PlayerByID(id)
		if p == nil || rs == nil {
			continue
		}
		p.RoundsPlayed++
		p.RoundStats[r.Number] = rs
		rs.Result = r.Winner
		st := &p.Stats
		st.Rounds++
		st.Kills += rs.Kills
		st.Deaths += rs.Deaths
		st.Assists += rs.Assists
		st.HeadshotKills += rs.Headshots
		st.Damage += rs.Damage
		st.FirstKills += boolInt(rs.FirstKill)
		st.FirstDeaths += boolInt(rs.FirstDeath)
		st.TradeKills += boolInt(rs.TradeKill)
		st.TradedDeaths += boolInt(rs.Traded)
		st.KASTRounds += boolInt(rs.KAST)
		st.Plants += rs.Plants
		st.Defuses += rs.Defuses
		st.UtilityDamage += rs.UtilityDamage
		st.FlashAssists += rs.FlashAssists
		st.EquipValueSum += rs.EquipValue
		st.MoneySpent += rs.Spent
		if rs.Clutch > 0 && rs.Survived && r.Winner == rs.Side {
			st.Clutches++
		}
		if rs.Side != "" {
			ss := p.SideStats[rs.Side]
			if ss == nil {
				ss = &model.PlayerStats{}
				p.SideStats[rs.Side] = ss
			}
			ss.Rounds++
			ss.Kills += rs.Kills
			ss.Deaths += rs.Deaths
			ss.Assists += rs.Assists
			ss.Damage += rs.Damage
			ss.HeadshotKills += rs.Headshots
			ss.FirstKills += boolInt(rs.FirstKill)
		}
	}
}

func (s *state) finishRoundPlayers(rb *roundBuilder) {
	if rb == nil || rb.round == nil {
		return
	}
	for _, rs := range rb.round.Players {
		if rs == nil {
			continue
		}
		rs.Survived = rs.Deaths == 0
		rs.KAST = rs.Kills > 0 || rs.Assists > 0 || rs.Survived || rs.Traded
	}
}

func (s *state) awardMVP(rb *roundBuilder) {
	if rb == nil || rb.round == nil {
		return
	}
	if rb.mvpAnnounced != "" {
		if p := s.m.PlayerByID(rb.mvpAnnounced); p != nil {
			p.Stats.MVPs++
			s.addEvent(&model.Event{Type: "mvp", ActorID: p.ID, ActorName: p.Name, Text: "Round MVP"})
		}
		return
	}
	var best *model.RoundStat
	var bestID string
	for id, rs := range rb.round.Players {
		if rs == nil {
			continue
		}
		if best == nil || rs.Kills > best.Kills || (rs.Kills == best.Kills && rs.Deaths < best.Deaths) {
			best, bestID = rs, id
		}
	}
	if bestID != "" && best != nil && best.Kills > 0 {
		if p := s.m.PlayerByID(bestID); p != nil {
			p.Stats.MVPs++
			s.addEvent(&model.Event{Type: "mvp", ActorID: p.ID, ActorName: p.Name, Text: "Round MVP"})
		}
	}
}

func (s *state) finalize() {
	if s.current != nil && !s.current.finalized {
		s.finalizeRound(&finalizeOpts{Reason: "Demo ended", NoScore: true})
	}
	if !s.warmupEnded {
		s.m.Warmup.EndOffset = s.offset()
		s.m.Warmup.Duration = s.m.Warmup.EndOffset - s.m.Warmup.StartOffset
	}
	if s.m.Warmup != nil && len(s.m.Warmup.Events) == 0 {
		s.m.Warmup = nil
	}
	if len(s.m.Teams) == 0 {
		s.ensureTeams()
	}
	s.m.Ticks = s.lastTick
	s.m.TickRate = s.p.TickRate()
	if s.m.TickRate <= 0 {
		s.m.TickRate = 64
	}
	s.m.Duration = float64(s.lastTick) / s.m.TickRate
	s.m.Mode = detectMode(len(s.m.Players))
	s.m.ID = matchID(s.m.DemoName, s.m.Map, s.m.FileTime)

	s.m.Players = sortPlayers(s.m.Players)

	// Ensure score reflects the last official round.
	var score [2]int
	for _, round := range s.m.Rounds {
		if !round.Pregame {
			score = round.EndScore
		}
	}
	for i, t := range s.m.Teams {
		t.Score = score[i]
		if side := s.teamFinalSide(t.ID); side != "" {
			t.FinalSide = side
		}
	}

	aimStats := map[uint64]*model.AimStats(nil)
	weaponAim := map[uint64]map[string]*model.WeaponStats(nil)
	if s.aim != nil {
		aimStats = s.aim.Finalize()
		weaponAim = s.aim.WeaponTotals()
	}
	for _, p := range s.m.Players {
		if p.SteamID64 != 0 {
			if a := aimStats[p.SteamID64]; a != nil {
				p.Aim = a
			}
			if wa := weaponAim[p.SteamID64]; wa != nil {
				for name, ws := range wa {
					cur := p.Weapons[name]
					if cur == nil {
						cur = &model.WeaponStats{Name: name}
						p.Weapons[name] = cur
					}
					cur.Shots = ws.Shots
					cur.Hits = ws.Hits
					cur.HeadHits = ws.HeadHits
					cur.Accuracy = ws.Accuracy
				}
			}
		}
	}
	for _, p := range s.m.Players {
		finalizePlayerStats(p)
	}
	// Read authoritative score/MVP counters from the last game state.
	for _, cp := range s.p.GameState().Participants().Playing() {
		if cp == nil {
			continue
		}
		if mp := s.playersByID[playerID(cp)]; mp != nil {
			if sc := cp.Score(); sc > 0 {
				mp.Stats.Score = sc
			}
			if mv := cp.MVPs(); mv > 0 {
				mp.Stats.MVPs = mv
			}
		}
	}
	s.m.Summary = buildSummary(s.m)
}
