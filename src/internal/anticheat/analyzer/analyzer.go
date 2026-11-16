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

// NewAnalyzer creates a new analyzer for the given demo file.
//
// Collector order is significant:
//  1. measurement collectors fill the raw per-player metrics,
//  2. the cheat detector scores those metrics,
//  3. the grading collector reads the final values.
func NewAnalyzer(demoPath string) *Analyzer {
	return &Analyzer{
		demoPath: demoPath,
		collectors: []stats.Collector{
			stats.NewHeadshotCollector(),
			stats.NewSnapAngleCollector(),
			stats.NewReactionTimeCollector(),
			stats.NewRecoilControlCollector(),
			stats.NewGameModeCollector(),
			stats.NewScoreboardCollector(),
			stats.NewGrenadeCollector(),
			stats.NewSniperCollector(),     // must run before CheatDetector
			stats.NewBehavioralCollector(), // wallhack-targeted signals
			stats.NewEvidenceCollector(),   // timestamped review cues
			stats.NewCheatDetector(),       // uses the measurements above
			stats.NewGradingCollector(),    // runs last
		},
	}
}

// AnalyzeReader performs the analysis from an arbitrary reader. It exists so
// compressed demos can be analyzed without a temporary file.
func (a *Analyzer) AnalyzeReader(rd io.Reader) (*stats.DemoStats, error) {
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
			return nil, fmt.Errorf("parse demo frame: %w", err)
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
	return demoStats, nil
}
