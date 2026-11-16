package stats

import "testing"

func TestEvaluateRecoil(t *testing.T) {
	ps := newTestPlayerStats()
	if ch := evaluateRecoil(ps); ch.HasData {
		t.Fatal("empty stats should have no data")
	}
	metricInt(t, ps, channelCategoryRecoil, Key("total_counted_bullets"), 40)
	metricFloat(t, ps, channelCategoryRecoil, Key("recoil_score"), 0.8)
	ch := evaluateRecoil(ps)
	if !ch.HasData || ch.Score != 0.8 || ch.Confidence != 1 {
		t.Fatalf("unexpected channel: %+v", ch)
	}
}
