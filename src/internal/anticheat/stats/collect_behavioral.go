package stats

import (
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

// BehavioralCollector implements the wallhack-targeted information channels
// that complement the aim-mechanics collectors:
//
//   - back-kill avoidance (collect_behavioral_backkill.go)
//   - pre-FOV pre-aim angle (collect_behavioral_prefov.go)
//   - off-engagement enemy attention (collect_behavioral_attention.go)
//
// All three use only positional and view-angle information from the demo; no
// BSP or line-of-sight data is required. The shared per-player history is
// maintained here so the per-detection files stay independent.
const (
	// fovEntryDegrees is the half-angle (deg) where a target is considered
	// "in FOV" — matches the reaction collector's threshold.
	fovEntryDegrees = 5.0
	// preFOVLookbackMs is how far before FOV-entry the killer's
	// crosshair-to-victim angle is sampled.
	preFOVLookbackMs = 200.0
	// behavioralBufferTicks bounds the per-player view+position history.
	// 5 s at 64 tick.
	behavioralBufferTicks = 320
	// backKillThresholdDeg is the angle from the victim's view direction to
	// the killer beyond which the kill counts as "from behind".
	backKillThresholdDeg = 100.0
	// minBackKillSamples avoids scoring with too few deaths.
	minBackKillSamples = 4
	// minPreFOVSamples avoids scoring with too few kills. Three-sample
	// readings produced noisy false positives; the noise floor is 4.
	minPreFOVSamples = 4
	// minAttentionSamples avoids scoring on tiny per-frame samples.
	minAttentionSamples = 200
)

// playerSnapshot captures view direction + eye-level position at a tick.
type playerSnapshot struct {
	tick  int
	yawX  float64 // ViewDirectionX, deg
	pitch float64 // ViewDirectionY, deg
	posX  float64
	posY  float64
	posZ  float64
}

// BehavioralCollector accumulates the three wallhack-targeted metrics.
type BehavioralCollector struct {
	*BaseCollector

	tickRate    float64
	currentTick int

	// Per-player rolling history of view + position.
	history map[uint64][]playerSnapshot

	// Per-player accumulators.
	backKillTotal      map[uint64]int // deaths charged to this player as victim
	backKillBack       map[uint64]int // deaths where victim was looking away
	backKillGivenTotal map[uint64]int // kills charged to this player as killer
	backKillGivenBack  map[uint64]int // kills where victim was looking away
	preFOVAngles       map[uint64][]float64
	attentionMin       map[uint64][]float64
}

// NewBehavioralCollector creates a new BehavioralCollector.
func NewBehavioralCollector() *BehavioralCollector {
	return &BehavioralCollector{
		BaseCollector:      NewBaseCollector("Behavioral Wallhack Signals"),
		history:            make(map[uint64][]playerSnapshot),
		backKillTotal:      make(map[uint64]int),
		backKillBack:       make(map[uint64]int),
		backKillGivenTotal: make(map[uint64]int),
		backKillGivenBack:  make(map[uint64]int),
		preFOVAngles:       make(map[uint64][]float64),
		attentionMin:       make(map[uint64][]float64),
	}
}

// Setup registers the kill handler and seeds the tick rate.
func (bc *BehavioralCollector) Setup(parser demoinfocs.Parser, demoStats *DemoStats) {
	bc.tickRate = parser.TickRate()
	if bc.tickRate <= 0 {
		bc.tickRate = 64.0
	}
	parser.RegisterEventHandler(func(e events.TickRateInfoAvailable) {
		if e.TickRate > 0 {
			bc.tickRate = e.TickRate
		}
	})
	parser.RegisterEventHandler(func(e events.Kill) {
		bc.handleKill(e, demoStats)
	})
}

// CollectFrame snapshots state and accumulates the attention metric.
func (bc *BehavioralCollector) CollectFrame(parser demoinfocs.Parser, demoStats *DemoStats) {
	bc.currentTick = parser.CurrentFrame()
	playing := parser.GameState().Participants().Playing()

	for _, p := range playing {
		if p == nil || p.SteamID64 == 0 || !p.IsAlive() {
			continue
		}
		pos := p.Position()
		snap := playerSnapshot{
			tick:  bc.currentTick,
			yawX:  float64(p.ViewDirectionX()),
			pitch: float64(p.ViewDirectionY()),
			posX:  pos.X,
			posY:  pos.Y,
			posZ:  pos.Z,
		}
		buf := bc.history[p.SteamID64]
		buf = append(buf, snap)
		if len(buf) > behavioralBufferTicks {
			buf = buf[len(buf)-behavioralBufferTicks:]
		}
		bc.history[p.SteamID64] = buf
	}

	bc.collectAttention(playing)
}

// handleKill dispatches a kill to the back-kill and pre-FOV measurements.
func (bc *BehavioralCollector) handleKill(e events.Kill, demoStats *DemoStats) {
	if e.Killer == nil || e.Victim == nil {
		return
	}
	if e.Killer.Team == e.Victim.Team {
		return
	}
	if e.Killer.SteamID64 == 0 || e.Victim.SteamID64 == 0 {
		return
	}
	bc.recordBackKill(e)
	bc.recordPreFOV(e, demoStats)
}

// CollectFinalStats publishes the per-player aggregates as metrics.
func (bc *BehavioralCollector) CollectFinalStats(demoStats *DemoStats) {
	for sid, ps := range demoStats.Players {
		bc.publishBackKill(ps, sid)
		bc.publishPreFOV(ps, sid)
		bc.publishAttention(ps, sid)
	}
}
