package aim

// metric_counter_strafe.go: shots fired while horizontally stopped.

var metricCounterStrafe = &MetricSpec{
	ID:           "counter_strafe",
	Label:        "Counter-strafe",
	Unit:         "%",
	Tooltip:      "Shots fired while horizontal speed was 34 u/s or less. Lower sample counts are unreliable.",
	Doc:          "docs/aim/counter_strafe.md",
	Weight:       5,
	HigherBetter: true,
	Lo:           20,
	Hi:           80,
	LowSample:    10,
	Value:        func(st *playerState) float64 { return percent(st.counterGood, st.counterShots) },
	Samples:      func(st *playerState) int { return st.counterShots },
	Format:       formatPct,
}
