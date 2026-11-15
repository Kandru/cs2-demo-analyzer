package analysis

import (
	"testing"

	"github.com/Kandru/cs2-demo-analyzer/internal/anticheat"
	"github.com/Kandru/cs2-demo-analyzer/internal/model"
)

func TestApplyAnnotatesCheatMomentsWithRoundAndSeconds(t *testing.T) {
	m := &model.Match{
		TickRate: 64,
		Teams: []*model.Team{
			{ID: 1, Name: "Alpha"},
			{ID: 2, Name: "Bravo"},
		},
		Players: []*model.Player{
			{ID: "1", SteamID64: 76561198000000001, Name: "tester", TeamID: 1},
		},
		Rounds: []*model.Round{
			{Number: 1, StartTick: 64, EndTick: 640},
			{Number: 2, StartTick: 700, EndTick: 1400},
		},
	}
	Apply(m, Options{Cheat: anticheat.Report{Players: map[uint64]*model.CheatReport{
		m.Players[0].SteamID64: {
			Available: true,
			Moments: []model.CheatMoment{
				{Kind: "preaim", Tick: 800},
				{Kind: "ttd", Tick: 32},
				{Kind: "snap", Tick: 900},
			},
		},
	}}})

	moments := m.Players[0].Cheat.Moments
	if len(moments) != 3 {
		t.Fatalf("got %d moments, want 3", len(moments))
	}
	if moments[0].Tick != 32 || moments[0].Round != 0 {
		t.Fatalf("warmup moment mis-mapped: %+v", moments[0])
	}
	if moments[0].Seconds != 0.5 {
		t.Fatalf("warmup seconds: got %v, want 0.5", moments[0].Seconds)
	}
	if moments[1].Tick != 800 || moments[1].Round != 2 {
		t.Fatalf("round moment mis-mapped: %+v", moments[1])
	}
	if moments[1].Seconds != 12.5 {
		t.Fatalf("round seconds: got %v, want 12.5", moments[1].Seconds)
	}
	if moments[2].Round != 2 {
		t.Fatalf("last moment round: got %d, want 2", moments[2].Round)
	}
}
