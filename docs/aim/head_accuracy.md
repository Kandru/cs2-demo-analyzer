# Head accuracy

<!-- metric: head_accuracy -->

## Formula

`head hits / all hits × 100`. Higher is better.

## Collection

- Source: `Tracker.OnPlayerHurt`.
- Increments `headHits` when the attributed shot's hit group is the head.
- Denominator is all hits, not kills.

## Sample gate

- Minimum 10 hits for a score.
- Low-sample flag below 10.
- No `PlayerHurt` events means no data.

## Anchors and weight

- 10% → 0, 45% → 100 (higher is better).
- Weight `10`.
- Head accuracy is informative but easier to inflate with a single lucky
  engagement, so it carries half the weight of overall accuracy.

## Caveats

Aim-punch, wallbangs and spamming through smoke all affect the ratio. Compare
within a similar weapon class. This is not the same as the detector's headshot
kill rate.

## Files

- Implementation: `src/internal/aim/metric_head_accuracy.go`
- Tests: `src/internal/aim/metrics_test.go`
- Docs index: `docs/aim/README.md`

## Tuning checklist

1. Change the anchors or weight in `metric_head_accuracy.go`.
2. Keep the denominator as hits, not kills.
3. Re-run `go test ./internal/aim`.
