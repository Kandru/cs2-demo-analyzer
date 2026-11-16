# Aim metric documentation

Each descriptive aim metric has one Go spec and one page here. The registry in
`src/internal/aim/metric_specs.go` is the source of truth for IDs, weights,
anchors and sample gates.

Aim metrics are **descriptive**, not part of the cheat score. They use the
battlelog parser's single pass and the `internal/aim.Tracker`, which is
independent of the vendored detector.

## Composite rating

- Each available metric scores `[0, 100]` from its anchors.
- The rating is the weighted average of available scores; weights are
  20 / 20 / 15 / 20 / 10 / 5 / 10 for damage speed, reaction, preaim,
  accuracy, head accuracy, counter-strafe and spray (sum 100).
- Missing components are dropped and the remaining weights renormalized.
- A rating needs at least 3 components; otherwise it is shown as `—`.
- Grade: A+ ≥85, A ≥75, B+ ≥65, B ≥55, C+ ≥45, C ≥35, D ≥25, F below.
- `confidence` is the fraction of total weight that had data.

## Metric pages

| ID | Metric | Weight | Page |
|---|---|---|---|
| `damage_speed` | Damage speed | 20 | [damage_speed.md](damage_speed.md) |
| `reaction` | Reaction | 20 | [reaction.md](reaction.md) |
| `preaim` | Preaim | 15 | [preaim.md](preaim.md) |
| `accuracy` | Accuracy | 20 | [accuracy.md](accuracy.md) |
| `head_accuracy` | Head accuracy | 10 | [head_accuracy.md](head_accuracy.md) |
| `counter_strafe` | Counter-strafe | 5 | [counter_strafe.md](counter_strafe.md) |
| `spray` | Spray | 10 | [spray.md](spray.md) |

## Shared collection rules

- Warmup, pregame rounds and dead players do not contribute (`SetLive`).
- Visibility uses `Player.IsSpottedBy` (server radar spotting). It is an
  approximation of true screen visibility.
- `PlayerHurt` drives TTD, reaction and preaim. Demos without `PlayerHurt`
  show "no data" for shot-based metrics instead of a misleading 0%.
- Accuracy excludes shotguns, knives, grenades and gear. A shot that damages
  multiple victims counts once.
- Counter-strafe uses position deltas; teleports and gaps > 0.5 s are ignored.
- Spray counts bullets 2..N in bursts of at least four automatic bullets, with
  gaps ≤ 220 ms.

## Adding a metric

1. Copy `_TEMPLATE.md` to `<id>.md`.
2. Add `metric_<id>.go` with a `MetricSpec` (anchors, weight, gate, tooltip,
   value/samples/formatters).
3. Register it in `metric_specs.go` at the intended tile order.
4. Add tests to `metrics_test.go`.
5. Run `go test ./internal/aim`; the docs-sync test enforces the page.
