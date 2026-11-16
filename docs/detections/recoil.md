# Recoil deviation

<!-- detection: recoil -->
<!-- kind: channel -->

## Signal

Recoil control is measured by comparing a player's spray against the known
weapon spray pattern. Humans pull down and counter-strafe with visible
variance; a near-perfect mean angular error across many bullets suggests
programmatic compensation.

## Collection

- Collector: `RecoilControlCollector` (`collect_recoil.go`).
- Tracks automatic weapons with a defined pattern (`AK-47`, `M4A4`, `M4A1-S`,
  `MP9`, `P90`) in bursts of at least three bullets.
- Raw keys: `recoil/mean_angular_error`, `recoil/recoil_score`,
  `recoil/total_counted_bullets`, `recoil/recoil_interpretation`.
- Weapons without a ground-truth pattern are skipped rather than compared to a
  synthetic fallback.
- The score is intentionally omitted when there is no data.

## Scoring

- Score: passthrough `recoil_score` clamped to `[0, 1]`.
- Confidence: `bullets / 20`, capped at 1.
- Mode: positive-only. A clean recoil reading only means no suspicious sprays
  were sampled.
- Weight: `0.10`.
- Lobby normalisation applies.

## Evidence

- Kind `recoil`, threshold `0.30°` mean error on at least 3 bullets, rank 2.
- Value format: `%.2f°`.

## Calibration

The interpretation bands are oriented around the cheat axis, not skill:
`Bot-perfect` ≤ 0.30°, `Very tight` ≤ 0.60°, `Human range` ≤ 1.0°, then
`Wide spread`. A pro with `Wide spread` is not bad — their spray is simply in
the normal human range. Narrative anchors are 0.30 / 0.55 / 0.80 on the score
with at least 20 counted bullets.

The upstream debug stdout block was removed during vendoring.

## Interactions

Feeds the Bayesian combiner and evidence stacking. Gains no direct modifier.

## Files

- Implementation: `src/internal/anticheat/stats/detect_recoil.go`
- Collector: `src/internal/anticheat/stats/collect_recoil.go`
- Patterns: `src/internal/anticheat/stats/spray_patterns.go`
- Tests: `src/internal/anticheat/stats/detect_recoil_test.go`
- Report row: registry metadata in `battlelog.go`

## Tuning checklist

1. Change the ramp or confidence in `evaluateRecoil`.
2. Add a weapon only with a real spray pattern and a documented reason.
3. Update the cue threshold in `detRecoil.Evidence.Threshold`.
4. Re-run `TestEvaluateRecoil` and inspect a clean demo's spray sample count.
