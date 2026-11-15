# AGENTS.md — internal/aim

Leetify-style descriptive aim metrics, independent from cheat scoring.

## Files

- `tracker.go` — per-frame state machine, shots, LoS, bursts, preaim, finalize.

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

## Changing metrics

Update `docs/aim-metrics.md` and `metrics_test.go`. Keep anchors fixed; do not
introduce rank/external data. Check a real 2026 demo for sane ratings.
