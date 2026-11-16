package aim

// metric_accuracy.go: bullets that hit an enemy divided by bullets fired.

var metricAccuracy = &MetricSpec{
	ID:           "accuracy",
	Label:        "Accuracy",
	Unit:         "%",
	Tooltip:      "Bullets that hit an enemy divided by bullets fired. Shotguns, knives and utility excluded.",
	Doc:          "docs/aim/accuracy.md",
	Weight:       20,
	HigherBetter: true,
	Lo:           10,
	Hi:           28,
	LowSample:    20,
	Value:        func(st *playerState) float64 { return percent(st.hits, st.shots) },
	Samples:      func(st *playerState) int { return st.shots },
	Format:       formatPct,
}
