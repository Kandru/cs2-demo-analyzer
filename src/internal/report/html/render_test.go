package html

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/Kandru/cs2-demo-analyzer/internal/model"
)

func testMatch() *model.Match {
	m := &model.Match{
		SchemaVersion: model.SchemaVersion,
		Map:           "de_mirage",
		Mode:          "Competitive",
		Duration:      1800,
		TickRate:      64,
		FileTime:      time.Date(2026, 11, 13, 20, 0, 0, 0, time.UTC),
		ParserVersion: "test",
		Teams: []*model.Team{
			{ID: 2, Name: "Alpha", StartSide: model.SideT, FinalSide: model.SideCT, Score: 13, PlayerIDs: []string{"1"}},
			{ID: 3, Name: "Bravo", StartSide: model.SideCT, FinalSide: model.SideT, Score: 7, PlayerIDs: []string{"2"}},
		},
		Players: []*model.Player{
			{
				ID: "1", SteamID64: 76561198000000001, Name: `<b>kalle</b>`, TeamID: 2, Slug: "kalle-000001", RoundsPlayed: 20,
				Stats: model.PlayerStats{Kills: 25, Deaths: 10, Assists: 5, Damage: 2000, Rounds: 20, HeadshotKills: 12, MVPs: 4,
					ADR: 100, KD: 2.5, HSPercent: 48, KASTPercent: 75, Shots: 200, Hits: 60},
				Aim: &model.AimStats{Rating: 72, Grade: "B+", Metrics: []model.AimMetric{
					{ID: "damage_speed", Label: "Damage speed", Display: "420 ms", Samples: 12, Score: 70, Weight: 20, Tooltip: "x"},
					{ID: "reaction", Label: "Reaction", Display: "380 ms", Samples: 12, Score: 60, Weight: 20, Tooltip: "x"},
					{ID: "preaim", Label: "Preaim", Display: "8.0°", Samples: 12, Score: 80, Weight: 15, Tooltip: "x"},
				}},
				Cheat: &model.CheatReport{Available: true, Likelihood: 12, Flagged: false,
					Channels: []model.CheatChannel{{ID: "hs", Label: "Headshot rate", Score: 0, Confidence: 0.9, Zone: "clean", Raw: 48, Unit: "%", Samples: 20, HasData: true}}},
				Weapons:    map[string]*model.WeaponStats{"AK-47": {Name: "AK-47", Kills: 15, Shots: 150, Hits: 45, Accuracy: 30}},
				RoundStats: map[int]*model.RoundStat{1: {Side: model.SideT, Kills: 2, Damage: 200, EquipValue: 4000, Survived: true}},
				SideStats:  map[model.Side]*model.PlayerStats{model.SideT: {Rounds: 10, Kills: 10, Deaths: 5, Assists: 2, Damage: 900, ADR: 90, HSPercent: 40}},
			},
		},
		Rounds: []*model.Round{{
			Number: 1, StartTick: 100, StartOffset: 10, Duration: 90, Winner: model.SideT, WinnerTeamID: 2,
			Reason: "Bomb exploded", StartScore: [2]int{0, 0}, EndScore: [2]int{1, 0},
			Events:  []*model.Event{{Type: "kill", Offset: 20, ActorID: "1", ActorName: "kalle", VictimName: "victim", Weapon: "AK-47", Headshot: true}},
			Players: map[string]*model.RoundStat{"1": {Side: model.SideT, Kills: 1}},
		}},
		Warmup: &model.Warmup{Duration: 30, Events: []*model.Event{{Type: "kill", Offset: 5, ActorName: "kalle", VictimName: "victim", Weapon: "USP-S"}}},
	}
	return m
}

func TestRenderContainsCoreSectionsAndEscapesNames(t *testing.T) {
	m := testMatch()
	m.Players[0].Name = `<script>alert(1)</script>`
	v := NewView(m, "26.11.1")
	var buf bytes.Buffer
	if err := Render(&buf, v); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	for _, want := range []string{"BATTLELOG", "Scoreboard", "Rounds", "Players", "Aim", "Cheat analysis", "Warmup", "de_mirage", "26.11.1"} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in rendered report", want)
		}
	}
	if strings.Contains(page, "<script>alert(1)</script>") {
		t.Fatal("player name was not escaped")
	}
	if !strings.Contains(page, "&lt;script&gt;") {
		t.Fatal("escaped player name missing")
	}
	if strings.Contains(page, "{{") {
		t.Fatal("unrendered template action")
	}
}
