# Damage speed

<!-- metric: damage_speed -->

## Formula

Median of `first damage tick − first visibility tick`, in milliseconds, capped
at 1000 ms. Lower is better.

## Collection

- Source: `Tracker.OnPlayerHurt` and `OnFrameState` in `tracker.go`.
- Visibility starts an engagement when `victim.IsSpottedBy(attacker)` first
  becomes true; brief LoS flickers up to 200 ms keep it alive.
- The first damage in the engagement appends one sample to `playerState.ttd`.
- A 0 ms sample (sighting and damage in the same tick) is discarded.
- Engagements longer than 1000 ms are not attributed.

## Sample gate

- Minimum 3 samples for a score.
- Below 3 samples the tile shows `—`.
- No `PlayerHurt` events means no data.

## Anchors and weight

- 800 ms → 0, 250 ms → 100 (lower is better).
- Weight `20` in the composite rating.
- The anchors span a slow-but-normal reaction to a very fast consistent one.

## Caveats

Visibility is radar spotting, so a defender holding an angle may be "visible"
before the peek. The median absorbs occasional outliers, but a small sample is
still noisy. This is descriptive and has no relation to the detector's
`reaction` channel.

## Files

- Implementation: `src/internal/aim/metric_damage_speed.go`
- Tests: `src/internal/aim/metrics_test.go`
- Docs index: `docs/aim/README.md`

## Tuning checklist

1. Change the anchors or weight in `metric_damage_speed.go`.
2. Update this page and the rating table in `docs/aim-metrics.md`.
3. Re-run `go test ./internal/aim` and inspect a real demo.
