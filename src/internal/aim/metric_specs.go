package aim

import "strconv"

// metric_specs.go defines the aim metric contract and the ordered registry.
// Each metric lives in its own metric_<id>.go file with its anchors, weight,
// sample gate and page in docs/aim/<id>.md. The registry is immutable after
// package initialisation.

// MetricSpec is one tile of the descriptive aim profile.
type MetricSpec struct {
	ID      string
	Label   string
	Unit    string
	Tooltip string
	Doc     string // repo-relative documentation path

	Weight       float64
	HigherBetter bool
	// Lo/Hi are the value anchors for the score. For higher-is-better metrics
	// Lo is the bad end and Hi the good end; for lower-is-better metrics Lo is
	// the good end and Hi the bad end.
	Lo, Hi float64
	// LowSample is the sample count below which the tile is marked as
	// low-sample. Data still renders; the flag is advisory.
	LowSample int

	Value   func(*playerState) float64
	Samples func(*playerState) int
	Format  func(float64) string
}

var metricSpecs = []*MetricSpec{
	metricDamageSpeed,
	metricReaction,
	metricPreaim,
	metricAccuracy,
	metricHeadAccuracy,
	metricCounterStrafe,
	metricSpray,
}

func formatMs(v float64) string  { return strconv.FormatFloat(v, 'f', 0, 64) + " ms" }
func formatDeg(v float64) string { return strconv.FormatFloat(v, 'f', 1, 64) + "°" }
func formatPct(v float64) string { return strconv.FormatFloat(v, 'f', 1, 64) + "%" }
