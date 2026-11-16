package stats

import "testing"

func TestEvaluateHS(t *testing.T) {
	ps := newTestPlayerStats()
	if ch := evaluateHS(ps); ch.HasData {
		t.Fatal("empty stats should have no data")
	}
	metricInt(t, ps, killsCategory, Key("total_kills"), 25)
	metricFloat(t, ps, killsCategory, Key("headshot_percentage"), 80)
	ch := evaluateHS(ps)
	if !ch.HasData || ch.Score != 1 || ch.Confidence != 1 {
		t.Fatalf("unexpected channel: %+v", ch)
	}
}
