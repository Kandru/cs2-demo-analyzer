package aim

// metric_preaim.go: median crosshair-to-victim angle when an engagement starts.

var metricPreaim = &MetricSpec{
	ID:           "preaim",
	Label:        "Preaim",
	Unit:         "°",
	Tooltip:      "Median angle between crosshair and enemy when an engagement starts. Lower is better.",
	Doc:          "docs/aim/preaim.md",
	Weight:       15,
	HigherBetter: false,
	Lo:           25,
	Hi:           5,
	LowSample:    4,
	Value:        func(st *playerState) float64 { return median(st.preaim) },
	Samples:      func(st *playerState) int { return len(st.preaim) },
	Format:       formatDeg,
}
