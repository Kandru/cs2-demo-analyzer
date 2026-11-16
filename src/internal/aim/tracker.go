// Package aim computes Leetify-style per-player aim metrics from a CS2 demo.
//
// The tracker is intentionally independent of the vendored anti-cheat
// detector: aim metrics are descriptive, while the detector's channels are
// calibrated for suspicion scoring. Definitions and anchors live in
// docs/aim-metrics.md. Keep both in sync when changing formulas.
package aim

import (
	"math"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"

	"github.com/Kandru/cs2-demo-analyzer/internal/model"
)

const (
	// sightGraceMs keeps an engagement alive across short LoS flickers.
	sightGraceMs = 200.0
	// maxEngagementMs caps sight-to-damage attribution.
	maxEngagementMs = 1000.0
	// hitWindowMs is how long after a shot a damage event may be attributed.
	hitWindowMs = 250.0
	// sprayGapMs is the maximum gap between shots in one spray.
	sprayGapMs = 220.0
	// counterStrafeSpeed is the horizontal speed (u/s) considered stopped.
	counterStrafeSpeed = 34.0
)

type pendingShot struct {
	tick   int
	weapon string
	burst  *burstShot
	hit    bool
}

type burst struct {
	weapon string
	last   int
	shots  []*burstShot
}

type burstShot struct {
	sprayIndex int // 0 = first bullet; only >=1 counts toward spray accuracy
	hit        bool
}

type engagement struct {
	entryTick  int
	seenTick   int
	damaged    bool
	reacted    bool
	entryAngle float64
}

type posSample struct {
	tick       int
	x, y, z    float64
	yaw, pitch float64
}

type playerState struct {
	shots        int
	hits         int
	headHits     int
	ttd          []float64
	reaction     []float64
	preaim       []float64
	counterShots int
	counterGood  int
	sprayAtt     int
	sprayHits    int
	sprayBursts  int
	speed        float64
	weapons      map[string]*weaponState
	pending      []*pendingShot
	burst        *burst
}

type weaponState struct {
	shots    int
	hits     int
	headHits int
}

// Tracker consumes parser callbacks and accumulates aim metrics.
type Tracker struct {
	tickRate float64
	tick     int
	live     bool
	players  map[uint64]*playerState

	// engagements[attacker][victim]
	engagements map[uint64]map[uint64]*engagement

	lastPos map[uint64]posSample
	history map[uint64][]posSample
	anyHurt bool
}

// New creates a tracker for a demo with the given tick rate.
func New(tickRate float64) *Tracker {
	if tickRate <= 0 {
		tickRate = 64
	}
	return &Tracker{
		tickRate:    tickRate,
		players:     map[uint64]*playerState{},
		engagements: map[uint64]map[uint64]*engagement{},
		lastPos:     map[uint64]posSample{},
		history:     map[uint64][]posSample{},
	}
}

// SetTick updates the current frame tick. The parser calls this before
// dispatching event handlers, which fire before OnFrameState for the frame.
func (t *Tracker) SetTick(tick int) { t.tick = tick }

// SetLive enables/disables accumulation. Warmup and pregame activity is not
// part of the aim profile.
func (t *Tracker) SetLive(live bool) {
	t.live = live
	if !live {
		t.resetRound()
	}
}

// OnRoundStart clears per-round engagement state but keeps match aggregates.
func (t *Tracker) OnRoundStart(tick int) {
	t.tick = tick
	t.resetRound()
}

func (t *Tracker) resetRound() {
	t.closeAllBursts()
	t.engagements = map[uint64]map[uint64]*engagement{}
	for _, st := range t.players {
		st.pending = nil
	}
}

func (t *Tracker) state(id uint64) *playerState {
	st := t.players[id]
	if st == nil {
		st = &playerState{weapons: map[string]*weaponState{}}
		t.players[id] = st
	}
	return st
}

// OnFrameState updates LoS engagements from the parser's game state.
func (t *Tracker) OnFrameState(playing []*common.Player, tick int) {
	t.tick = tick
	if !t.live {
		return
	}
	graceTicks := int(sightGraceMs * t.tickRate / 1000.0)
	if graceTicks < 1 {
		graceTicks = 1
	}

	for _, p := range playing {
		if p == nil || p.SteamID64 == 0 || !p.IsAlive() {
			continue
		}
		pos := p.Position()
		if eyes, ok := p.PositionEyes(); ok {
			snap := posSample{tick: tick, x: eyes.X, y: eyes.Y, z: eyes.Z,
				yaw: float64(p.ViewDirectionX()), pitch: float64(p.ViewDirectionY())}
			hist := append(t.history[p.SteamID64], snap)
			cutoffTick := tick - int(t.tickRate) - 2
			for len(hist) > 0 && hist[0].tick < cutoffTick {
				hist = hist[1:]
			}
			t.history[p.SteamID64] = hist
		}
		if prev, ok := t.lastPos[p.SteamID64]; ok && tick > prev.tick {
			dt := float64(tick-prev.tick) / t.tickRate
			if dt > 0 && dt <= 0.5 {
				sp := math.Hypot(pos.X-prev.x, pos.Y-prev.y) / dt
				if sp < 2000 { // ignore teleports/spawns
					t.state(p.SteamID64).speed = sp
				}
			}
		}
		t.lastPos[p.SteamID64] = posSample{tick: tick, x: pos.X, y: pos.Y}
	}

	for _, attacker := range playing {
		if attacker == nil || attacker.SteamID64 == 0 || !attacker.IsAlive() {
			continue
		}
		aid := attacker.SteamID64
		victims := t.engagements[aid]
		if victims == nil {
			victims = map[uint64]*engagement{}
			t.engagements[aid] = victims
		}
		for _, victim := range playing {
			if victim == nil || victim.SteamID64 == 0 || victim.SteamID64 == aid {
				continue
			}
			if victim.Team == attacker.Team || !victim.IsAlive() {
				continue
			}
			if !victim.IsSpottedBy(attacker) {
				continue
			}
			vid := victim.SteamID64
			eng := victims[vid]
			if eng == nil || tick-eng.seenTick > graceTicks {
				eng = &engagement{entryTick: tick, seenTick: tick}
				eng.entryAngle = viewToTargetAngle(attacker, victim)
				victims[vid] = eng
			} else {
				eng.seenTick = tick
			}
		}
	}
}

// OnWeaponFire records a shot, its movement state, and spray membership.
func (t *Tracker) OnWeaponFire(e events.WeaponFire) {
	if !t.live || e.Shooter == nil || e.Weapon == nil {
		return
	}
	id := e.Shooter.SteamID64
	if id == 0 {
		return
	}
	w := e.Weapon
	name := w.String()
	if name == "" {
		name = "Unknown"
	}
	st := t.state(id)
	tick := t.tick

	if isGun(w.Type) {
		speed := st.speed
		st.counterShots++
		if speed <= counterStrafeSpeed {
			st.counterGood++
		}
	}

	if !countsForAccuracy(w.Type) {
		return
	}

	st.shots++
	ws := st.weapons[name]
	if ws == nil {
		ws = &weaponState{}
		st.weapons[name] = ws
	}
	ws.shots++

	var bs *burstShot
	if isAutomatic(w.Type) {
		if st.burst != nil && st.burst.weapon == name && float64(tick-st.burst.last)*1000.0/t.tickRate <= sprayGapMs {
			st.burst.last = tick
			bs = &burstShot{sprayIndex: len(st.burst.shots)}
			st.burst.shots = append(st.burst.shots, bs)
		} else {
			t.closeBurst(id)
			bs = &burstShot{sprayIndex: 0}
			st.burst = &burst{weapon: name, last: tick, shots: []*burstShot{bs}}
		}
	} else {
		t.closeBurst(id)
	}

	st.pending = append(st.pending, &pendingShot{tick: tick, weapon: name, burst: bs})
	// Keep the pending queue bounded. Damage normally arrives immediately.
	cutoff := tick - int(hitWindowMs*t.tickRate/1000.0) - 5
	for len(st.pending) > 0 && st.pending[0].tick < cutoff {
		st.pending = st.pending[1:]
	}
}

// OnPlayerHurt attributes damage to the most recent unclaimed shot and records
// TTD / reaction samples.
func (t *Tracker) OnPlayerHurt(e events.PlayerHurt) {
	if !t.live || e.Attacker == nil || e.Player == nil {
		return
	}
	attacker, victim := e.Attacker, e.Player
	if attacker.SteamID64 == 0 || victim.SteamID64 == 0 || attacker.Team == victim.Team {
		return
	}
	t.anyHurt = true
	aid, vid := attacker.SteamID64, victim.SteamID64
	st := t.state(aid)

	// Engagement samples are recorded once per sighting.
	if victims := t.engagements[aid]; victims != nil {
		if eng := victims[vid]; eng != nil {
			elapsed := float64(t.tick-eng.entryTick) * 1000.0 / t.tickRate
			if !eng.damaged && elapsed >= 0 && elapsed <= maxEngagementMs {
				st.ttd = append(st.ttd, elapsed)
				// Radar spotting can happen well before an enemy is actually
				// on screen, so angle-at-sighting is noisy. Measure crosshair
				// placement 200 ms before the first damage instead.
				if ang, ok := t.preaimAngle(aid, vid, t.tick); ok {
					st.preaim = append(st.preaim, ang)
				}
				eng.damaged = true
			}
		}
	}

	// Attribute the hit to a recent shot. Prefer the newest matching weapon and
	// only consume it if the shot has not already hit somebody.
	var shot *pendingShot
	window := int(hitWindowMs*t.tickRate/1000.0) + 1
	for i := len(st.pending) - 1; i >= 0; i-- {
		s := st.pending[i]
		if s.hit || t.tick-s.tick > window {
			continue
		}
		if e.Weapon != nil && s.weapon != e.Weapon.String() {
			continue
		}
		shot = s
		break
	}
	if shot == nil {
		return
	}
	shot.hit = true
	st.hits++
	ws := st.weapons[shot.weapon]
	if ws != nil {
		ws.hits++
	}
	if e.HitGroup == events.HitGroupHead {
		st.headHits++
		if ws != nil {
			ws.headHits++
		}
	}
	if shot.burst != nil {
		shot.burst.hit = true
	}

	// Reaction: first shot that hits this victim after the current sighting.
	if victims := t.engagements[aid]; victims != nil {
		if eng := victims[vid]; eng != nil && !eng.reacted {
			elapsed := float64(shot.tick-eng.entryTick) * 1000.0 / t.tickRate
			if elapsed >= 0 && elapsed <= maxEngagementMs {
				st.reaction = append(st.reaction, elapsed)
				eng.reacted = true
			}
		}
	}
}

// OnKill clears engagements for both players and closes any active burst.
func (t *Tracker) OnKill(e events.Kill) {
	if e.Killer != nil && e.Killer.SteamID64 != 0 {
		t.closeBurst(e.Killer.SteamID64)
	}
	if e.Victim != nil && e.Victim.SteamID64 != 0 {
		t.closeBurst(e.Victim.SteamID64)
		id := e.Victim.SteamID64
		delete(t.engagements, id)
		for attackerID, victims := range t.engagements {
			if _, ok := victims[id]; ok {
				delete(victims, id)
				t.engagements[attackerID] = victims
			}
		}
	}
}

func (t *Tracker) closeBurst(id uint64) {
	st := t.players[id]
	if st == nil || st.burst == nil {
		return
	}
	b := st.burst
	st.burst = nil
	// A spray needs at least four bullets; only bullets after the first count.
	if len(b.shots) >= 4 {
		st.sprayBursts++
		for _, shot := range b.shots {
			if shot.sprayIndex == 0 {
				continue
			}
			st.sprayAtt++
			if shot.hit {
				st.sprayHits++
			}
		}
	}
}

func (t *Tracker) closeAllBursts() {
	for id := range t.players {
		t.closeBurst(id)
	}
}

// Finalize computes model.AimStats for every tracked player.
func (t *Tracker) Finalize() map[uint64]*model.AimStats {
	t.closeAllBursts()
	out := make(map[uint64]*model.AimStats, len(t.players))
	for id, st := range t.players {
		out[id] = buildAimStats(st, t.anyHurt)
	}
	return out
}

// WeaponTotals exposes per-weapon shot/hit counts for the report.
func (t *Tracker) WeaponTotals() map[uint64]map[string]*model.WeaponStats {
	out := map[uint64]map[string]*model.WeaponStats{}
	for id, st := range t.players {
		if len(st.weapons) == 0 {
			continue
		}
		ws := make(map[string]*model.WeaponStats, len(st.weapons))
		for name, v := range st.weapons {
			if v.shots == 0 {
				continue
			}
			ws[name] = &model.WeaponStats{Name: name, Shots: v.shots, Hits: v.hits, HeadHits: v.headHits, Accuracy: percent(v.hits, v.shots)}
		}
		out[id] = ws
	}
	return out
}
