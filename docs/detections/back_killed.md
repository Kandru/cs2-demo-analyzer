# Back-killed rate

<!-- detection: back_killed -->
<!-- kind: channel -->

## Signal

The share of a player's deaths where they were looking away from the killer
(more than 100° between their view direction and the killer's position). A
wallhacker is rarely caught from behind because they always know where enemies
are, so an unusually low rate on enough deaths is suspicious.

## Collection

- Collector: `BehavioralCollector`, `events.Kill`.
- Charged to the victim: total deaths and deaths where the victim was looking
  away.
- Also records a diagnostic killer-side rate
  (`behavioral/back_kill_given_pct`) that is **not** scored directly — flanking
  is a legitimate playstyle — but feeds the wallhack co-occurrence modifier.
- Raw keys: `behavioral/back_killed_pct`,
  `behavioral/back_killed_total_deaths`.
- Requires at least 4 deaths to publish; the channel requires at least 8.

## Scoring

- Ramp: 25% → 3% (lower is worse), linear.
- Confidence: `deaths / 8`, capped at 1.
- Mode: positive-only. A high back-killed rate is simply normal play.
- Weight: `0.06`.
- Lobby normalisation applies.

## Evidence

No evidence cue. A death while looking away is not a review target.

## Calibration

Narrative anchors are 8% / 4% / 0.01% with at least 8 deaths. The near-zero
blatant anchor reflects that never being caught from behind on a large sample
is the strongest reading. Very low-death matches must not score.

## Interactions

[`wallhack_cooccurrence`](wallhack_co_occurrence_boost.md) reads
`back_kill_given_pct` and `back_kill_given_total_kills` on the killer side.

## Files

- Implementation: `src/internal/anticheat/stats/detect_back_killed.go`
- Collector: `src/internal/anticheat/stats/collect_behavioral_backkill.go`
- Tests: `src/internal/anticheat/stats/detect_back_killed_test.go`
- Report row: registry metadata in `battlelog.go`

## Tuning checklist

1. Change the ramp in `evaluateBackKilled`.
2. Keep the 4-death publish gate and 8-death score gate aligned.
3. Do not score `back_kill_given_pct` directly; it is a co-occurrence input.
4. Re-run `TestEvaluateBackKilled` and a short demo.
