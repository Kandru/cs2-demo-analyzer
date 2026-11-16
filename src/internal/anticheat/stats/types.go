package stats

import (
	"time"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
)

// PlayerIdentifier contains information to identify a player
type PlayerIdentifier struct {
	SteamID64 uint64
	Name      string
}

// Category represents a category of statistics (e.g., weapons, movement, etc.)
type Category string

// Key represents a specific statistic key within a category
type Key string

// MetricType defines the type of a statistic value (count, percentage, duration, etc.)
type MetricType string

const (
	// MetricCount represents a count metric
	MetricCount MetricType = "count"
	// MetricPercentage represents a percentage metric
	MetricPercentage MetricType = "percentage"
	// MetricDuration represents a duration metric
	MetricDuration MetricType = "duration"
	// MetricFloat represents a floating point value
	MetricFloat MetricType = "float"
	// MetricInteger represents an integer value
	MetricInteger MetricType = "integer"
	// MetricString represents a string value
	MetricString MetricType = "string"
)

// Metric represents a single statistical measure
type Metric struct {
	Type          MetricType
	FloatValue    float64
	IntValue      int64
	DurationValue time.Duration
	StringValue   string
	Description   string
}

// PlayerStats contains all statistics for a player
type PlayerStats struct {
	Player     PlayerIdentifier
	Categories map[Category]map[Key]Metric
	// Moments holds timestamped, suspicious-direction evidence recorded by the
	// collectors. It is additive metadata for the report and never feeds the
	// cheat score. See docs/anticheat.md.
	Moments []EvidenceMoment
}

// EvidenceMoment is one timestamped detector observation that a reviewer can
// jump to in the demo. Values mirror the channel that produced the signal.
// Label, Rank and ValueText are filled from the registered EvidenceSpec by
// AddMoment so the report does not need a per-kind switch.
type EvidenceMoment struct {
	Kind      string
	Label     string
	Rank      int
	ValueText string
	Tick      int
	Value     float64
	Unit      string
	Opponent  string
	Weapon    string
	Detail    string
}

const (
	// evidenceMomentCapPerKind bounds one channel's moments per player. A long
	// match can produce hundreds of snap samples; the report only needs a
	// useful review shortlist.
	evidenceMomentCapPerKind = 50
	// evidenceMomentCapPerPlayer bounds the total payload written to JSON/HTML.
	evidenceMomentCapPerPlayer = 200
)

// AddMoment records one review cue, enforcing per-kind and per-player caps.
// When a cap is full, a more suspicious cue replaces the weakest one so long
// matches keep their strongest review targets.
func (ps *PlayerStats) AddMoment(m EvidenceMoment) {
	if ps == nil || m.Kind == "" || m.Tick < 0 {
		return
	}
	if spec := evidenceSpec(m.Kind); spec != nil {
		if m.Label == "" {
			m.Label = spec.Label
		}
		if m.Rank == 0 {
			m.Rank = spec.Rank
		}
		if m.ValueText == "" && spec.ValueText != nil {
			m.ValueText = spec.ValueText(m.Value)
		}
	}
	if m.Label == "" {
		m.Label = m.Kind
	}
	if m.Rank < 1 {
		m.Rank = 1
	}
	if m.ValueText == "" {
		m.ValueText = "signal"
	}
	kindIdx := make([]int, 0, 8)
	for i, existing := range ps.Moments {
		if existing.Kind == m.Kind {
			kindIdx = append(kindIdx, i)
		}
	}
	if len(kindIdx) >= evidenceMomentCapPerKind {
		weakest := kindIdx[0]
		for _, i := range kindIdx[1:] {
			if momentSeverity(ps.Moments[i]) < momentSeverity(ps.Moments[weakest]) {
				weakest = i
			}
		}
		if momentSeverity(m) <= momentSeverity(ps.Moments[weakest]) {
			return
		}
		ps.Moments[weakest] = m
		return
	}
	if len(ps.Moments) >= evidenceMomentCapPerPlayer {
		weakest := 0
		for i := 1; i < len(ps.Moments); i++ {
			if momentSeverity(ps.Moments[i]) < momentSeverity(ps.Moments[weakest]) {
				weakest = i
			}
		}
		if momentSeverity(m) <= momentSeverity(ps.Moments[weakest]) {
			return
		}
		ps.Moments[weakest] = m
		return
	}
	ps.Moments = append(ps.Moments, m)
}

// NewPlayerStats creates a new PlayerStats instance
func NewPlayerStats(player *common.Player) *PlayerStats {
	return &PlayerStats{
		Player: PlayerIdentifier{
			SteamID64: player.SteamID64,
			Name:      player.Name,
		},
		Categories: make(map[Category]map[Key]Metric),
	}
}

// AddMetric adds or updates a metric for a player
func (ps *PlayerStats) AddMetric(category Category, key Key, metric Metric) {
	if _, exists := ps.Categories[category]; !exists {
		ps.Categories[category] = make(map[Key]Metric)
	}
	ps.Categories[category][key] = metric
}

// GetMetric retrieves a metric for a player
func (ps *PlayerStats) GetMetric(category Category, key Key) (Metric, bool) {
	if categoryMap, exists := ps.Categories[category]; exists {
		if metric, found := categoryMap[key]; found {
			return metric, true
		}
	}
	return Metric{}, false
}

// IncrementIntMetric increments an integer metric
func (ps *PlayerStats) IncrementIntMetric(category Category, key Key) {
	if _, exists := ps.Categories[category]; !exists {
		ps.Categories[category] = make(map[Key]Metric)
		ps.Categories[category][key] = Metric{
			Type:     MetricInteger,
			IntValue: 1,
		}
		return
	}

	if metric, found := ps.Categories[category][key]; found {
		metric.IntValue++
		ps.Categories[category][key] = metric
	} else {
		ps.Categories[category][key] = Metric{
			Type:     MetricInteger,
			IntValue: 1,
		}
	}
}

// IncrementFloatMetric adds a value to a float metric
func (ps *PlayerStats) IncrementFloatMetric(category Category, key Key, value float64) {
	if _, exists := ps.Categories[category]; !exists {
		ps.Categories[category] = make(map[Key]Metric)
		ps.Categories[category][key] = Metric{
			Type:       MetricFloat,
			FloatValue: value,
		}
		return
	}

	if metric, found := ps.Categories[category][key]; found {
		metric.FloatValue += value
		ps.Categories[category][key] = metric
	} else {
		ps.Categories[category][key] = Metric{
			Type:       MetricFloat,
			FloatValue: value,
		}
	}
}

// DemoStats contains statistics for all players in a demo
type DemoStats struct {
	Players   map[uint64]*PlayerStats
	TickRate  float64
	TickCount int
	DemoName  string
	MapName   string
}

// NewDemoStats creates a new DemoStats instance
func NewDemoStats() *DemoStats {
	return &DemoStats{
		Players: make(map[uint64]*PlayerStats),
	}
}

// GetOrCreatePlayerStats gets existing player stats or creates new ones if they don't exist
func (ds *DemoStats) GetOrCreatePlayerStats(player *common.Player) *PlayerStats {
	if player == nil {
		return nil
	}

	if _, exists := ds.Players[player.SteamID64]; !exists {
		ds.Players[player.SteamID64] = NewPlayerStats(player)
	}
	return ds.Players[player.SteamID64]
}

// GetOrCreatePlayerStatsBySteamID gets existing player stats or creates new ones by SteamID
func (ds *DemoStats) GetOrCreatePlayerStatsBySteamID(steamID uint64) *PlayerStats {
	if _, exists := ds.Players[steamID]; !exists {
		// Create a placeholder player
		ds.Players[steamID] = &PlayerStats{
			Player: PlayerIdentifier{
				SteamID64: steamID,
				Name:      "Unknown",
			},
			Categories: make(map[Category]map[Key]Metric),
		}
	}
	return ds.Players[steamID]
}
