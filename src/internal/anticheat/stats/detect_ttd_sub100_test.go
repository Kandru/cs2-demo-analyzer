package stats

import "testing"

func TestEvaluateTTDSub100(t *testing.T) {
	ps := newTestPlayerStats()
	if ch := evaluateTTDSub100(ps); ch.HasData {
		t.Fatal("empty stats should have no data")
	}
	metricInt(t, ps, reactionCategory, Key("ttd_samples"), 40)
	metricFloat(t, ps, reactionCategory, Key("sub_100ms_ttd"), 30)
	ch := evaluateTTDSub100(ps)
	if !ch.HasData || ch.Score != 1 || ch.Confidence != 1 {
		t.Fatalf("unexpected channel: %+v", ch)
	}
}
