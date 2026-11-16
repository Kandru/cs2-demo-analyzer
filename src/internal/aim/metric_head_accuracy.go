package aim

// metric_head_accuracy.go: head hits divided by all hits (not headshot kills).

var metricHeadAccuracy = &MetricSpec{
	ID:           "head_accuracy",
	Label:        "Head accuracy",
	Unit:         "%",
	Tooltip:      "Head hits divided by all hits (not kills).",
	Doc:          "docs/aim/head_accuracy.md",
	Weight:       10,
	HigherBetter: true,
	Lo:           10,
	Hi:           45,
	LowSample:    10,
	Value:        func(st *playerState) float64 { return percent(st.headHits, st.hits) },
	Samples:      func(st *playerState) int { return st.hits },
	Format:       formatPct,
}
