package stats

import "testing"

func TestEvaluatePreFOV(t *testing.T) {
	ps := newTestPlayerStats()
	if ch := evaluatePreFOV(ps); ch.HasData {
		t.Fatal("empty stats should have no data")
	}
	metricInt(t, ps, channelCategoryBehavioral, Key("pre_fov_aim_samples"), 20)
	metricFloat(t, ps, channelCategoryBehavioral, Key("pre_fov_aim_median_deg"), 4)
	ch := evaluatePreFOV(ps)
	if !ch.HasData || ch.Score != 1 || ch.Confidence != 1 {
		t.Fatalf("unexpected channel: %+v", ch)
	}
}
