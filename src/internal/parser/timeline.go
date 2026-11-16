package parser

import (
	"sort"
	"strings"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"

	"github.com/Kandru/cs2-demo-analyzer/internal/model"
)

// This file owns the optional, high-detail timeline events. Adding a new event
// means: add a model.EventX identifier, register its handler in
// installHandlers, emit the event through state.addEvent, and teach the view
// (view.go) how to label and badge it. No round-accounting code belongs here
// unless the event also changes player or team statistics.

func (s *state) onFlashed(e events.PlayerFlashed) {
	if e.Player == nil || e.Attacker == nil {
		return
	}
	s.syncTick()
	flashed := s.ensurePlayer(e.Player)
	attacker := s.ensurePlayer(e.Attacker)
	if flashed == nil || attacker == nil || flashed.ID == attacker.ID {
		return
	}
	duration := e.FlashDuration().Seconds()
	if duration > 0 && duration < 0.3 {
		return
	}
	s.addEvent(&model.Event{
		Type:       model.EventFlash,
		ActorID:    attacker.ID,
		ActorName:  attacker.Name,
		VictimID:   flashed.ID,
		VictimName: flashed.Name,
		Amount:     duration,
	})
}

func (s *state) onGrenadeDamage(e events.PlayerHurt) {
	if e.Attacker == nil || e.Player == nil || e.Weapon == nil {
		return
	}
	dmg := e.HealthDamageTaken
	if dmg <= 0 {
		dmg = e.HealthDamage
	}
	if dmg <= 0 || e.Weapon.Type.Class() != common.EqClassGrenade {
		return
	}
	// A molotov damages a player over several ticks. One row per second is
	// enough to explain the hit without flooding the timeline.
	key := playerID(e.Attacker) + "|" + playerID(e.Player) + "|" + e.Weapon.String()
	tickRate := int(s.p.TickRate())
	if tickRate <= 0 {
		tickRate = 64
	}
	if last, ok := s.lastGrenadeHit[key]; ok && s.lastTick-last < tickRate {
		return
	}
	if s.lastGrenadeHit == nil {
		s.lastGrenadeHit = map[string]int{}
	}
	s.lastGrenadeHit[key] = s.lastTick

	attacker := s.ensurePlayer(e.Attacker)
	victim := s.ensurePlayer(e.Player)
	if attacker == nil || victim == nil || attacker.ID == victim.ID {
		return
	}
	s.addEvent(&model.Event{
		Type:       model.EventGrenadeDamage,
		ActorID:    attacker.ID,
		ActorName:  attacker.Name,
		VictimID:   victim.ID,
		VictimName: victim.Name,
		Weapon:     e.Weapon.String(),
		Amount:     float64(dmg),
	})
}

func (s *state) onBombPickup(e events.BombPickup) {
	if e.Player == nil {
		return
	}
	p := s.ensurePlayer(e.Player)
	if p == nil {
		return
	}
	s.addEvent(&model.Event{
		Type:      model.EventBombPickup,
		ActorID:   p.ID,
		ActorName: p.Name,
	})
}

func (s *state) onBombDropped(e events.BombDropped) {
	if e.Player == nil {
		return
	}
	s.syncTick()
	// Planting drops the bomb internally. Do not add a misleading "dropped"
	// row next to the plant marker for the same tick.
	if s.bombPlantedTick != 0 && s.lastTick == s.bombPlantedTick {
		return
	}
	p := s.ensurePlayer(e.Player)
	if p == nil {
		return
	}
	s.addEvent(&model.Event{
		Type:      model.EventBombDrop,
		ActorID:   p.ID,
		ActorName: p.Name,
	})
}

func (s *state) onHostageStateChanged(e events.HostageStateChanged) {
	if s.current == nil || !s.live || e.Hostage == nil || e.Hostage.Entity == nil {
		return
	}
	id := e.Hostage.Entity.ID()
	leader := e.Hostage.Leader()
	switch e.NewState {
	case common.HostageStateGettingPickedUp, common.HostageStateBeingCarried:
		if leader == nil {
			return
		}
		p := s.ensurePlayer(leader)
		if p == nil || s.hostageCarrier[id] == p.ID {
			return
		}
		s.hostageCarrier[id] = p.ID
		s.addEvent(&model.Event{
			Type:      model.EventHostagePickup,
			ActorID:   p.ID,
			ActorName: p.Name,
		})
	case common.HostageStateGettingDropped:
		carrier := s.hostageCarrier[id]
		if carrier == "" {
			return
		}
		delete(s.hostageCarrier, id)
		name := ""
		if p := s.m.PlayerByID(carrier); p != nil {
			name = p.Name
		}
		s.addEvent(&model.Event{
			Type:      model.EventHostageDrop,
			ActorID:   carrier,
			ActorName: name,
		})
	}
}

func (s *state) onHostageRescued(e events.HostageRescued) {
	if e.Player == nil {
		return
	}
	p := s.ensurePlayer(e.Player)
	if p == nil {
		return
	}
	s.addEvent(&model.Event{
		Type:      model.EventHostageRescue,
		ActorID:   p.ID,
		ActorName: p.Name,
	})
}

func (s *state) onHostageKilled(e events.HostageKilled) {
	if e.Killer == nil {
		return
	}
	p := s.ensurePlayer(e.Killer)
	if p == nil {
		return
	}
	s.addEvent(&model.Event{
		Type:      model.EventHostageKilled,
		ActorID:   p.ID,
		ActorName: p.Name,
	})
}

// recordPurchases snapshots each player's freeze-time-end inventory. Old
// reports simply have no Bought slice; nothing else depends on it.
func (s *state) recordPurchases() {
	if s.current == nil || !s.live {
		return
	}
	for _, p := range s.p.GameState().Participants().Playing() {
		if p == nil {
			continue
		}
		mp := s.ensurePlayer(p)
		if mp == nil {
			continue
		}
		rs := s.roundStat(mp.ID)
		if rs == nil {
			continue
		}
		rs.Bought = inventoryNames(p)
	}
}

func inventoryNames(p *common.Player) []string {
	if p == nil {
		return nil
	}
	seen := map[string]bool{}
	var names []string
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		names = append(names, name)
	}
	for _, eq := range p.Weapons() {
		if eq == nil {
			continue
		}
		switch eq.Type {
		case common.EqKnife, common.EqBomb:
			continue
		}
		add(eq.String())
	}
	if p.Armor() > 0 {
		if p.HasHelmet() {
			add("Kevlar + Helmet")
		} else {
			add("Kevlar Vest")
		}
	}
	if p.HasDefuseKit() {
		add("Defuse Kit")
	}
	sort.SliceStable(names, func(i, j int) bool {
		return strings.ToLower(names[i]) < strings.ToLower(names[j])
	})
	return names
}
