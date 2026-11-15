package stats

// intMetric is kept from the upstream HTML reporter, which was removed from
// this vendored copy because the battlelog owns rendering.
func intMetric(ps *PlayerStats, cat Category, k Key) int64 {
	if m, ok := ps.GetMetric(cat, k); ok {
		return m.IntValue
	}
	return 0
}
