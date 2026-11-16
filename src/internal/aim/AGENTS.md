# AGENTS.md — internal/aim

Leetify-style descriptive aim metrics, independent from cheat scoring.

## Files

- `tracker.go` — per-frame state machine, shots, LoS, bursts, finalize.
- `metric_specs.go` — `MetricSpec` + ordered registry.
- `metric_<id>.go` — one file per metric (anchors, weight, gate, formatting).
- `metrics.go` — registry-driven profile builder.
- `score.go` — anchor score and grade bands.
- `geometry.go` — view/position math and preaim lookback.
- `helpers.go` — percentages, medians and weapon-class helpers.
- `docs_sync_test.go` — one `docs/aim/<id>.md` page per metric.

## Invariants

- Parser calls the tracker explicitly: `SetTick`, `OnFrameState`,
  `OnWeaponFire`, `OnPlayerHurt`, `OnKill`, `OnRoundStart`, `SetLive`.
- Match aggregates only; warmup and pregame are excluded via `SetLive`.
- Accuracy excludes shotguns/knife/utility. One shot can hit multiple victims
  but counts once.
- Head accuracy is head hits / all hits.
- Counter-strafe is derived from position deltas (no velocity property in v5);
  ignore teleports and gaps > 0.5 s.
- Spray counts bullets 2..N in bursts of ≥4 automatics, ≤220 ms gaps.
- Preaim uses view/position history 200 ms before first damage, not the radar
  spot event.
- If the demo has no `PlayerHurt`, accuracy/head/spray must show “no data”.
- The metric registry is immutable after package initialization.
- Do not put per-metric switches in `metrics.go` or the report.

## Changing metrics

1. Edit `metric_<id>.go` only.
2. Update `docs/aim/<id>.md` and, for rating changes, `docs/aim-metrics.md`.
3. Add or update tests in `metrics_test.go`.
4. Keep anchors fixed; do not introduce rank/external data.
5. Check a real 2026 demo for sane ratings.
