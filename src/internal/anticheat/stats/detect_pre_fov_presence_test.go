package stats

import "testing"

func TestEvaluatePreFOVPresence(t *testing.T) {
	ps := newTestPlayerStats()
	ctx := &LobbyContext{
		SamplesBySID:    map[uint64]int64{1: 6, 2: 0, 3: 0},
		AsymmetricBySID: map[uint64]bool{1: true, 2: false, 3: false},
	}
	if ch := evaluatePreFOVPresence(ctx, ps, 1); ch.HasData {
		t.Fatal("missing median should have no data")
	}
	metricFloat(t, ps, channelCategoryBehavioral, Key("pre_fov_aim_median_deg"), 5)
	ch := evaluatePreFOVPresence(ctx, ps, 1)
	if !ch.HasData || ch.Score != 1 || ch.Confidence != 1 {
		t.Fatalf("unexpected channel: %+v", ch)
	}
	// A symmetric lobby never fires.
	ctx.AsymmetricBySID[1] = false
	if ch := evaluatePreFOVPresence(ctx, ps, 1); ch.HasData {
		t.Fatalf("symmetric lobby should not fire: %+v", ch)
	}
	// The demo-wide placeholder never fires.
	if ch := evaluatePreFOVPresence(ctx, ps, 0); ch.HasData {
		t.Fatalf("SteamID 0 should not fire: %+v", ch)
	}
}
