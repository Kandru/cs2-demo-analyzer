package aim

// metric_reaction.go: median time from first visibility to the first shot that
// hits that enemy.

var metricReaction = &MetricSpec{
	ID:           "reaction",
	Label:        "Reaction",
	Unit:         "ms",
	Tooltip:      "Median time from first visibility to the first shot that hits that enemy. Lower is better.",
	Doc:          "docs/aim/reaction.md",
	Weight:       20,
	HigherBetter: false,
	Lo:           600,
	Hi:           180,
	LowSample:    3,
	Value:        func(st *playerState) float64 { return median(st.reaction) },
	Samples:      func(st *playerState) int { return len(st.reaction) },
	Format:       formatMs,
}
