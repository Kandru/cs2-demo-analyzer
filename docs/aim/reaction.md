# Reaction

<!-- metric: reaction -->

## Formula

Median of `shot tick − first visibility tick`, in milliseconds, where the shot
is the first shot that subsequently hits that victim. Lower is better.

## Collection

- Source: `Tracker.OnPlayerHurt` in `tracker.go`.
- A damage event is attributed to the newest unclaimed shot within 250 ms and
  the same weapon.
- The attributed shot's tick is compared with the engagement entry tick.
- One reaction sample per engagement (`engagement.reacted`).
- Engagements longer than 1000 ms are not attributed.

## Sample gate

- Minimum 3 samples for a score.
- Below 3 samples the tile shows `—`.
- No `PlayerHurt` events means no data.

## Anchors and weight

- 600 ms → 0, 180 ms → 100 (lower is better).
- Weight `20`.
- The lower bound is closer to the human floor than damage speed, because this
  metric measures the shot itself rather than the first damage.

## Caveats

The hit/shot pairing is heuristic. A shot at another target is not
misattributed, but a very fast re-engagement can be rejected by the 250 ms
window. The metric is descriptive, not the detector's `reaction` channel.

## Files

- Implementation: `src/internal/aim/metric_reaction.go`
- Tests: `src/internal/aim/metrics_test.go`
- Docs index: `docs/aim/README.md`

## Tuning checklist

1. Change the anchors or weight in `metric_reaction.go`.
2. Keep the 250 ms hit window aligned with `hitWindowMs`.
3. Re-run `go test ./internal/aim`.
