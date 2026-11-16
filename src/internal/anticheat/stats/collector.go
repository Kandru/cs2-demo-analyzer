package stats

import (
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
)

// Collector is the interface implemented by every measurement pass in the
// detector. A collector may register parser event handlers in Setup, read
// per-frame game state in CollectFrame, and publish metrics in
// CollectFinalStats. It never scores players; scoring runs afterwards in
// run.go so it can be unit-tested against plain PlayerStats values.
type Collector interface {
	// Name returns the human-readable collector name used in logs/debugging.
	Name() string

	// Setup is called once before parsing starts. Register event handlers here.
	Setup(parser demoinfocs.Parser, demoStats *DemoStats)

	// CollectFrame is called once per parsed frame.
	CollectFrame(parser demoinfocs.Parser, demoStats *DemoStats)

	// CollectFinalStats is called after parsing completes.
	CollectFinalStats(demoStats *DemoStats)
}

// BaseCollector provides the no-op defaults for a Collector. Embed it and
// override only the phases a collector needs.
type BaseCollector struct {
	name string
}

// NewBaseCollector creates a BaseCollector with the given name.
func NewBaseCollector(name string) *BaseCollector {
	return &BaseCollector{name: name}
}

// Name returns the collector name.
func (bc *BaseCollector) Name() string { return bc.name }

// Setup is a no-op default.
func (bc *BaseCollector) Setup(parser demoinfocs.Parser, demoStats *DemoStats) {}

// CollectFrame is a no-op default.
func (bc *BaseCollector) CollectFrame(parser demoinfocs.Parser, demoStats *DemoStats) {}

// CollectFinalStats is a no-op default.
func (bc *BaseCollector) CollectFinalStats(demoStats *DemoStats) {}
