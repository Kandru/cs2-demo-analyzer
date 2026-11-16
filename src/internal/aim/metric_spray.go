package aim

// metric_spray.go: hit rate on bullets 2..N of automatic sprays.

var metricSpray = &MetricSpec{
	ID:           "spray",
	Label:        "Spray",
	Unit:         "%",
	Tooltip:      "Hit rate of bullets 2..N in sprays of at least four bullets.",
	Doc:          "docs/aim/spray.md",
	Weight:       10,
	HigherBetter: true,
	Lo:           15,
	Hi:           45,
	LowSample:    6,
	Value:        func(st *playerState) float64 { return percent(st.sprayHits, st.sprayAtt) },
	Samples:      func(st *playerState) int { return st.sprayAtt },
	Format:       formatPct,
}
