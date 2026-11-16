package aim

// metric_damage_speed.go: median time from first visibility to first damage.

var metricDamageSpeed = &MetricSpec{
	ID:           "damage_speed",
	Label:        "Damage speed",
	Unit:         "ms",
	Tooltip:      "Median time from first visibility to first damage. Lower is better.",
	Doc:          "docs/aim/damage_speed.md",
	Weight:       20,
	HigherBetter: false,
	Lo:           800,
	Hi:           250,
	LowSample:    3,
	Value:        func(st *playerState) float64 { return median(st.ttd) },
	Samples:      func(st *playerState) int { return len(st.ttd) },
	Format:       formatMs,
}
