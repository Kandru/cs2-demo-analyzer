package stats

import "testing"

func newTestPlayerStats() *PlayerStats {
	return &PlayerStats{Categories: map[Category]map[Key]Metric{}}
}

func TestAddMomentCapsPerKindAndPlayer(t *testing.T) {
	ps := newTestPlayerStats()
	for i := 0; i < evidenceMomentCapPerKind+25; i++ {
		ps.AddMoment(EvidenceMoment{Kind: "snap", Tick: i})
	}
	if got := len(ps.Moments); got != evidenceMomentCapPerKind {
		t.Fatalf("per-kind cap: got %d moments, want %d", got, evidenceMomentCapPerKind)
	}

	for _, kind := range []string{"ttd", "preaim", "recoil", "wallbang"} {
		for i := 0; i < evidenceMomentCapPerKind; i++ {
			ps.AddMoment(EvidenceMoment{Kind: kind, Tick: 1000 + i})
		}
	}
	if got := len(ps.Moments); got != evidenceMomentCapPerPlayer {
		t.Fatalf("per-player cap: got %d moments, want %d", got, evidenceMomentCapPerPlayer)
	}

	ps.AddMoment(EvidenceMoment{Kind: "", Tick: 1})
	ps.AddMoment(EvidenceMoment{Kind: "snap", Tick: -1})
	if got := len(ps.Moments); got != evidenceMomentCapPerPlayer {
		t.Fatalf("invalid moments changed the payload: got %d", got)
	}
}

func TestReportMomentsArePublishedAndSorted(t *testing.T) {
	ps := newTestPlayerStats()
	ps.Player = PlayerIdentifier{SteamID64: 42, Name: "tester"}
	ps.AddMoment(EvidenceMoment{Kind: "preaim", Tick: 200, Value: 4.2, Unit: "°"})
	ps.AddMoment(EvidenceMoment{Kind: "ttd", Tick: 100, Value: 88, Unit: "ms"})

	reports := BuildPlayerReports(&DemoStats{Players: map[uint64]*PlayerStats{42: ps}})
	rep := reports[42]
	if rep == nil {
		t.Fatal("missing player report")
	}
	if len(rep.Moments) != 2 {
		t.Fatalf("got %d moments, want 2", len(rep.Moments))
	}
	if rep.Moments[0].Kind != "ttd" || rep.Moments[0].Tick != 100 {
		t.Fatalf("moments not sorted by tick: %+v", rep.Moments)
	}
	if rep.Moments[0].Label != "Reaction" || rep.Moments[0].Rank != 2 || rep.Moments[0].ValueText != "88 ms" {
		t.Fatalf("moment metadata not populated: %+v", rep.Moments[0])
	}
	if rep.Moments[1].Kind != "preaim" {
		t.Fatalf("unexpected second moment: %+v", rep.Moments[1])
	}
	if rep.Moments[1].Label != "Pre-aim" || rep.Moments[1].Rank != 3 || rep.Moments[1].ValueText != "4.2°" {
		t.Fatalf("preaim metadata not populated: %+v", rep.Moments[1])
	}
}
