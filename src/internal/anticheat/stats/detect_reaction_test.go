package stats

import "testing"

func TestEvaluateReaction(t *testing.T) {
	ps := newTestPlayerStats()
	if ch := evaluateReactionMedianTTD(ps); ch.HasData {
		t.Fatal("empty stats should have no data")
	}
	metricInt(t, ps, reactionCategory, Key("ttd_samples"), 20)
	metricFloat(t, ps, reactionCategory, Key("median_ttd"), 150)
	ch := evaluateReactionMedianTTD(ps)
	if !ch.HasData || ch.Score != 1 || ch.Confidence != 1 {
		t.Fatalf("unexpected channel: %+v", ch)
	}
	// Bidirectional: a slow median produces negative (exculpatory) evidence,
	// represented by a zero score but a real HasData reading.
	metricFloat(t, ps, reactionCategory, Key("median_ttd"), 600)
	if ch := evaluateReactionMedianTTD(ps); !ch.HasData || ch.Score != 0 {
		t.Fatalf("slow median should be a clean reading: %+v", ch)
	}
}
