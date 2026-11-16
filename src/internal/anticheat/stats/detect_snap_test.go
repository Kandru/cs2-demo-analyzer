package stats

import "testing"

func TestEvaluateSnap(t *testing.T) {
	ps := newTestPlayerStats()
	if ch := evaluateSnap(ps); ch.HasData {
		t.Fatal("empty stats should have no data")
	}
	metricInt(t, ps, channelCategoryAiming, Key("snap_count"), 20)
	metricFloat(t, ps, channelCategoryAiming, Key("p95_snap_velocity"), 3.5)
	ch := evaluateSnap(ps)
	if !ch.HasData || ch.Score != 1 || ch.Confidence != 1 {
		t.Fatalf("unexpected channel: %+v", ch)
	}
}
