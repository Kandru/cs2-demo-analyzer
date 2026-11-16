# Accuracy

<!-- metric: accuracy -->

## Formula

`hit shots / fired shots × 100`. Higher is better.

## Collection

- Source: `Tracker.OnWeaponFire` and `OnPlayerHurt`.
- A shot counts when the weapon passes `countsForAccuracy`: guns except
  shotguns (pellets make the ratio meaningless), and never knives or utility.
- Damage attributes to the newest unclaimed shot within 250 ms and the same
  weapon. Each shot counts as one hit at most, even if it damages several
  victims.
- Counters live in `playerState.shots/hits` and per-weapon state.

## Sample gate

- Minimum 20 shots for a score.
- Low-sample flag below 20.
- No `PlayerHurt` events forces the metric to no-data.

## Anchors and weight

- 10% → 0, 28% → 100 (higher is better).
- Weight `20`.
- Accuracy is one of the strongest descriptive skill signals, but it includes
  spam and wallbangs, so it is not a precision measure.

## Caveats

A player who fires few shots but hits most of them can score very high on a
small sample; read `n=`. Suppressing fire and wallbangs lower the ratio
without reflecting aim quality.

## Files

- Implementation: `src/internal/aim/metric_accuracy.go`
- Tests: `src/internal/aim/metrics_test.go`
- Docs index: `docs/aim/README.md`

## Tuning checklist

1. Change the anchors or weight in `metric_accuracy.go`.
2. Keep the shotgun exclusion unless the collection model changes.
3. Re-run `go test ./internal/aim`.
