# Counter-strafe

<!-- metric: counter_strafe -->

## Formula

`shots fired while horizontal speed ≤ 34 u/s / shots fired × 100`.
Higher is better.

## Collection

- Source: `Tracker.OnFrameState` and `OnWeaponFire`.
- Horizontal speed is derived from position deltas between frames. CS2 player
  pawns do not expose a usable velocity property in demoinfocs v5.
- Teleports/spawns (speed ≥ 2000 u/s) and gaps > 0.5 s are ignored.
- Every gun shot increments `counterShots`; a shot at or below the speed
  threshold also increments `counterGood`.

## Sample gate

- Minimum 10 shots for a score.
- Low-sample flag below 10.
- No `PlayerHurt` events forces the metric to no-data (the whole shot group is
  cleared), even though counter-strafe itself only needs weapon-fire events.

## Anchors and weight

- 20% → 0, 80% → 100 (higher is better).
- Weight `5`, the smallest. Stopping to shoot is a basic mechanic, and the
  position-delta approximation is noisier than other metrics.

## Caveats

34 u/s is roughly standing accuracy speed. A player who taps slowly, or a
demo with sparse frames, can distort the ratio. Treat it as a style signal.

## Files

- Implementation: `src/internal/aim/metric_counter_strafe.go`
- Tests: `src/internal/aim/metrics_test.go`
- Docs index: `docs/aim/README.md`

## Tuning checklist

1. Change the anchors or weight in `metric_counter_strafe.go`.
2. Keep `counterStrafeSpeed` and the teleport guard in sync.
3. Re-run `go test ./internal/aim`.
