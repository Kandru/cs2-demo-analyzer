# Preaim

<!-- metric: preaim -->

## Formula

Median angle in degrees between the attacker's crosshair and the victim's
position **200 ms before the first damage** of an engagement. Lower is better.

## Collection

- Source: `Tracker.preaimAngle` in `geometry.go`, called from
  `OnPlayerHurt`.
- Uses the per-player eye-level position + view-angle history.
- Looks back `200 ms` worth of ticks plus one, takes the nearest sample at or
  before that tick for both players, and computes the angle.
- Samples outside `(0°, 90°]` are rejected.

## Sample gate

- Minimum 4 samples for a score.
- Low-sample flag below 4.
- No `PlayerHurt` events means no data.

## Anchors and weight

- 25° → 0, 5° → 100 (lower is better).
- Weight `15`.
- Preaim is meaningful but noisier than pure mechanics, so it sits below the
  two timing metrics.

## Caveats

This measures crosshair placement, not information: a wallhacker and a good
angle-holder can look similar. It is deliberately separate from the detector's
`pre_fov` channel, which measures timing relative to visibility.

## Files

- Implementation: `src/internal/aim/metric_preaim.go`
- Geometry: `src/internal/aim/geometry.go`
- Tests: `src/internal/aim/metrics_test.go`
- Docs index: `docs/aim/README.md`

## Tuning checklist

1. Change the anchors or weight in `metric_preaim.go`.
2. Keep the 200 ms lookback consistent with the detector's definition.
3. Re-run `go test ./internal/aim`.
