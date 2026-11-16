package stats

// modifier.go defines the score-modifier contract. A modifier reads the
// combined score plus player metrics and returns whether it fired. Publishing
// and report-adjustment text live with the modifier so the pipeline has no
// per-modifier switches.

// Applied is the outcome of one modifier.
type Applied struct {
	Fired  bool
	Detail string
	Value  float64
	Count  int
	Names  []string
}

// scoreState carries the mutable scoring context through one player's
// modifiers.
type scoreState struct {
	score           float64
	combined        float64 // pre-modifier Bayesian likelihood, [0, 100]
	ps              *PlayerStats
	channels        []Channel
	lobbyAsymmetric bool
}

// Modifier is one documented score adjustment.
type Modifier struct {
	ID  string
	Doc string
	// Apply may update st.score.
	Apply func(st *scoreState) Applied
	// Publish writes the modifier's anti_cheat metrics when it fired.
	Publish func(ps *PlayerStats, a Applied)
	// Adjustments returns the human-readable report rows for this modifier.
	Adjustments func(ps *PlayerStats) []BattlelogAdjustment
}

func yesMetric(ps *PlayerStats, key Key) bool {
	m, ok := ps.GetMetric(antiCheatCategory, key)
	return ok && m.StringValue == "Yes"
}

func yesStringMetric(ps *PlayerStats, key Key) (string, bool) {
	m, ok := ps.GetMetric(antiCheatCategory, key)
	if !ok || m.StringValue == "" {
		return "", false
	}
	return m.StringValue, true
}
