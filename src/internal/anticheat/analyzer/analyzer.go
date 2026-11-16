package analyzer

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/Kandru/cs2-demo-analyzer/internal/anticheat/stats"
	dem "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/msg"
)

// Analyzer represents a CS2 demo analyzer.
type Analyzer struct {
	demoPath   string
	collectors []stats.Collector
}

// Results represents the analysis results.
type Results struct {
	DemoStats *stats.DemoStats
}

// NewAnalyzer creates a new analyzer for the given demo file.
//
// Collector order is significant:
//  1. measurement collectors fill the raw per-player metrics,
//  2. the cheat detector scores those metrics,
//  3. the grading collector reads the final values.
func NewAnalyzer(demoPath string) *Analyzer {
	analyzer := &Analyzer{
		demoPath: demoPath,
	}

	analyzer.RegisterCollector(stats.NewHeadshotCollector())
	analyzer.RegisterCollector(stats.NewSnapAngleCollector())
	analyzer.RegisterCollector(stats.NewReactionTimeCollector())
	analyzer.RegisterCollector(stats.NewRecoilControlCollector())
	analyzer.RegisterCollector(stats.NewGameModeCollector())
	analyzer.RegisterCollector(stats.NewScoreboardCollector())
	analyzer.RegisterCollector(stats.NewGrenadeCollector())
	analyzer.RegisterCollector(stats.NewSniperCollector())     // must run before CheatDetector
	analyzer.RegisterCollector(stats.NewBehavioralCollector()) // wallhack-targeted signals
	analyzer.RegisterCollector(stats.NewEvidenceCollector())   // timestamped review cues
	analyzer.RegisterCollector(stats.NewCheatDetector())       // uses the measurements above
	analyzer.RegisterCollector(stats.NewGradingCollector())    // runs last

	return analyzer
}

// RegisterCollector adds a statistics collector to the analyzer.
func (a *Analyzer) RegisterCollector(collector stats.Collector) {
	a.collectors = append(a.collectors, collector)
}

// AnalyzeReader performs the analysis from an arbitrary reader. It exists so
// compressed demos can be analyzed without a temporary file.
func (a *Analyzer) AnalyzeReader(rd io.Reader) (Results, error) {
	parser := dem.NewParser(rd)
	defer parser.Close()

	demoStats := stats.NewDemoStats()
	demoStats.DemoName = filepath.Base(a.demoPath)

	// v5 removed ParseHeader(); subscribe to the demo file header net message.
	parser.RegisterNetMessageHandler(func(m *msg.CDemoFileHeader) {
		demoStats.MapName = m.GetMapName()
	})

	for _, collector := range a.collectors {
		collector.Setup(parser, demoStats)
	}

	frameCount := 0
	for {
		ok, err := parser.ParseNextFrame()
		if err != nil {
			return Results{}, fmt.Errorf("error parsing frame: %w", err)
		}
		if !ok {
			break
		}
		for _, collector := range a.collectors {
			collector.CollectFrame(parser, demoStats)
		}
		frameCount++
	}

	demoStats.TickCount = frameCount
	demoStats.TickRate = parser.TickRate()

	for _, collector := range a.collectors {
		collector.CollectFinalStats(demoStats)
	}

	return Results{DemoStats: demoStats}, nil
}
