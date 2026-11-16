# <label>

<!-- metric: <id> -->

## Formula

Exact formula and unit.

## Collection

- Which tracker callback records the sample.
- Which `playerState` slice/counter stores it.
- Visibility, engagement or weapon filtering rules.

## Sample gate

- Minimum samples for a score.
- Low-sample threshold.
- How missing `PlayerHurt` or weapon exclusions degrade the tile.

## Anchors and weight

- `Lo` → 0, `Hi` → 100 (state direction).
- Weight in the composite rating.
- Why this weight.

## Caveats

Known noise, role effects, or approximation limits. Be explicit about what the
metric does not prove.

## Files

- Implementation: `src/internal/aim/metric_<id>.go`
- Tests: `src/internal/aim/metrics_test.go`
- Docs index: `docs/aim/README.md`

## Tuning checklist

1. Change anchors/weight in `metric_<id>.go` only.
2. Update this page and `docs/aim-metrics.md` if the rating changes.
3. Re-run the metric test and check a real 2026 demo.
