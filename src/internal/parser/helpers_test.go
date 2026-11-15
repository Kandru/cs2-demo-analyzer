package parser

import (
	"testing"
	"time"

	"github.com/Kandru/cs2-demo-analyzer/internal/model"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"kalle":           "kalle",
		"Max /A/":         "max-a",
		"  Hello  World ": "hello-world",
		"??":              "",
		"Player_One.2":    "player-one-2",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q)=%q want %q", in, got, want)
		}
	}
}

func TestBuyType(t *testing.T) {
	if buyType(500, 1) != "Pistol" || buyType(2500, 3) != "Force buy" || buyType(4500, 3) != "Full buy" || buyType(100, 3) != "Eco" {
		t.Fatal("unexpected buy classification")
	}
}

func TestFinalizePlayerStats(t *testing.T) {
	p := &model.Player{
		Stats: model.PlayerStats{
			Kills: 10, Deaths: 5, Assists: 5, Damage: 800, Rounds: 10,
			HeadshotKills: 4, KASTRounds: 7, Shots: 100, Hits: 30, HeadHits: 12,
		},
		SideStats: map[model.Side]*model.PlayerStats{},
		Weapons:   map[string]*model.WeaponStats{"AK-47": {Shots: 50, Hits: 15}},
	}
	finalizePlayerStats(p)
	if p.Stats.ADR != 80 || p.Stats.KD != 2 || p.Stats.HSPercent != 40 {
		t.Fatalf("unexpected stats: %+v", p.Stats)
	}
	if p.Stats.KASTPercent != 70 || p.Stats.AccuracyPct != 30 || p.Stats.HeadAccuracyPct != 40 {
		t.Fatalf("unexpected percentages: %+v", p.Stats)
	}
	if p.Weapons["AK-47"].Accuracy != 30 {
		t.Fatalf("weapon accuracy: %v", p.Weapons["AK-47"].Accuracy)
	}
}

func TestMatchIDStable(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	a := matchID("demo.dem", "de_mirage", ts)
	b := matchID("demo.dem", "de_mirage", ts)
	if a != b || len(a) != 12 {
		t.Fatalf("unexpected ids %q %q", a, b)
	}
}
