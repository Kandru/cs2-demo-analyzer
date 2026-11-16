package parser

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

// TestParseRealDemo is opt-in: set TEST_DEMO to a local CS2 demo path.
// It validates the invariants that are hard to cover with synthetic fixtures.
func TestParseRealDemo(t *testing.T) {
	path := os.Getenv("TEST_DEMO")
	if path == "" {
		t.Skip("set TEST_DEMO to a local CS2 demo to run this test")
	}
	m, err := Parse(context.Background(), path, Options{Logger: slog.Default()})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Players) == 0 || len(m.Teams) != 2 {
		t.Fatalf("players=%d teams=%d", len(m.Players), len(m.Teams))
	}
	if len(m.Rounds) == 0 {
		t.Fatal("no rounds")
	}
	if len(m.Events) == 0 {
		t.Fatal("no match lifecycle events")
	}
	last := m.Rounds[len(m.Rounds)-1]
	if last.EndScore[0] == 0 && last.EndScore[1] == 0 {
		t.Fatalf("scores look unmapped: %v", last.EndScore)
	}
	var kills, deaths int
	for _, p := range m.Players {
		kills += p.Stats.Kills
		deaths += p.Stats.Deaths
	}
	if kills > deaths {
		t.Fatalf("more kills than deaths: %d/%d", kills, deaths)
	}
}
