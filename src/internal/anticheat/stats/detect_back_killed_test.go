package stats

import "testing"

func TestEvaluateBackKilled(t *testing.T) {
	ps := newTestPlayerStats()
	if ch := evaluateBackKilled(ps); ch.HasData {
		t.Fatal("empty stats should have no data")
	}
	metricInt(t, ps, behavioralCategory, Key("back_killed_total_deaths"), 16)
	metricFloat(t, ps, behavioralCategory, Key("back_killed_pct"), 3)
	ch := evaluateBackKilled(ps)
	if !ch.HasData || ch.Score != 1 || ch.Confidence != 1 {
		t.Fatalf("unexpected channel: %+v", ch)
	}
}
