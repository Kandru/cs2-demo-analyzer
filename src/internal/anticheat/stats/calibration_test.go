package stats

import "testing"

func metricInt(t *testing.T, ps *PlayerStats, cat Category, key Key, v int64) {
	t.Helper()
	ps.AddMetric(cat, key, Metric{Type: MetricInteger, IntValue: v})
}

func metricFloat(t *testing.T, ps *PlayerStats, cat Category, key Key, v float64) {
	t.Helper()
	ps.AddMetric(cat, key, Metric{Type: MetricFloat, FloatValue: v})
}

func TestTTDSub100FloorRequiresStrongSample(t *testing.T) {
	ps := &PlayerStats{Categories: map[Category]map[Key]Metric{}}
	metricInt(t, ps, Category("reaction"), Key("ttd_samples"), 17)
	metricFloat(t, ps, Category("reaction"), Key("sub_100ms_ttd"), 23.5)
	metricInt(t, ps, Category("reaction"), Key("sub_100ms_count"), 4)
	metricInt(t, ps, Category("behavioral"), Key("pre_fov_aim_samples"), 4)
	metricFloat(t, ps, Category("behavioral"), Key("pre_fov_aim_median_deg"), 9.2)

	// The kalle-style 4/17 sample with a 9.2° pre-aim must not pin the verdict.
	if score, fired := applyTTDSub100Floor(12, ps, true); fired || score != 12 {
		t.Fatalf("thin sample should not trigger floor: score=%v fired=%v", score, fired)
	}

	// A larger, unambiguous triggerbot profile still floors.
	ps2 := &PlayerStats{Categories: map[Category]map[Key]Metric{}}
	metricInt(t, ps2, Category("reaction"), Key("ttd_samples"), 30)
	metricFloat(t, ps2, Category("reaction"), Key("sub_100ms_ttd"), 40)
	metricInt(t, ps2, Category("reaction"), Key("sub_100ms_count"), 12)
	metricInt(t, ps2, Category("behavioral"), Key("pre_fov_aim_samples"), 10)
	metricFloat(t, ps2, Category("behavioral"), Key("pre_fov_aim_median_deg"), 5)
	if score, fired := applyTTDSub100Floor(12, ps2, true); !fired || score < 55 {
		t.Fatalf("strong profile should floor: score=%v fired=%v", score, fired)
	}
}

func TestSnapAngleUnits(t *testing.T) {
	// angleDiff returns degrees. The upstream snap collector multiplied the
	// result by 180/pi a second time, inflating velocities ~57x.
	if got := angleDiff(0, 90); got != 90 {
		t.Fatalf("angleDiff should be degrees, got %v", got)
	}
}
