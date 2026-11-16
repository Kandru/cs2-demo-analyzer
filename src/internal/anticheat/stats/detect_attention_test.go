package stats

import "testing"

func TestEvaluateAttention(t *testing.T) {
	ps := newTestPlayerStats()
	if ch := evaluateAttention(ps); ch.HasData {
		t.Fatal("empty stats should have no data")
	}
	metricInt(t, ps, channelCategoryBehavioral, Key("nearest_enemy_angle_samples"), 400)
	metricFloat(t, ps, channelCategoryBehavioral, Key("nearest_enemy_angle_median_deg"), 18)
	ch := evaluateAttention(ps)
	if !ch.HasData || ch.Score != 1 || ch.Confidence != 1 {
		t.Fatalf("unexpected channel: %+v", ch)
	}
}
