package stats

import "testing"

func TestEvaluateDecoupling(t *testing.T) {
	ps := newTestPlayerStats()
	if ch := evaluateDecoupling(ps); ch.HasData {
		t.Fatal("empty stats should have no data")
	}
	metricInt(t, ps, behavioralCategory, Key("pre_fov_aim_samples"), 20)
	metricFloat(t, ps, behavioralCategory, Key("pre_fov_aim_median_deg"), 4)
	metricInt(t, ps, behavioralCategory, Key("nearest_enemy_angle_samples"), 400)
	metricFloat(t, ps, behavioralCategory, Key("nearest_enemy_angle_median_deg"), 26)
	ch := evaluateDecoupling(ps)
	if !ch.HasData || ch.Score != 1 || ch.Confidence != 1 {
		t.Fatalf("unexpected channel: %+v", ch)
	}
	// Missing one half means no reading.
	delete(ps.Categories[behavioralCategory], Key("nearest_enemy_angle_samples"))
	if ch := evaluateDecoupling(ps); ch.HasData {
		t.Fatalf("half data should not score: %+v", ch)
	}
}
