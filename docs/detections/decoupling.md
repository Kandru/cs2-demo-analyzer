# Fight/idle decoupling

<!-- detection: decoupling -->
<!-- kind: channel -->

## Signal

`attention_median − pre_fov_median`. A wallhacker concentrates during
engagements but their crosshair drifts toward unseen enemies during quiet
moments, so the gap between fight-time and idle-time crosshair placement is
larger than for a consistent legitimate player.

This is a derived channel: it owns no collector state.

## Collection

- Reads `behavioral/pre_fov_aim_median_deg`,
  `behavioral/pre_fov_aim_samples`,
  `behavioral/nearest_enemy_angle_median_deg` and
  `behavioral/nearest_enemy_angle_samples`.
- Silent if either half is missing.

## Scoring

- Raw delta = attention median − pre-FOV median.
- Ramp: 8° → 22° (higher is worse), linear.
- Confidence: `min(sqrt(pre_fov_samples / 15), attention_samples / 200)`.
  Both halves must be trustworthy.
- Mode: positive-only.
- Weight: `0.10`.
- Lobby normalisation applies.

## Evidence

No evidence cue.

## Calibration

Narrative uses the published score × confidence rather than the raw delta,
because a delta is not intuitive in prose: mild 0.30, strong 0.45, blatant
0.60. The derived nature means it should move only when both source channels
move; do not tune it to compensate for one noisy half.

## Interactions

Depends entirely on [`pre_fov`](pre_fov.md) and [`attention`](attention.md).
It does not feed any modifier.

## Files

- Implementation: `src/internal/anticheat/stats/detect_decoupling.go`
- Tests: `src/internal/anticheat/stats/detect_decoupling_test.go`
- Report row: registry metadata in `battlelog.go`

## Tuning checklist

1. Change the ramp or confidence in `evaluateDecoupling`.
2. Keep the both-halves sample requirement.
3. Update the score-based narrative thresholds if the ramp changes.
4. Re-run `TestEvaluateDecoupling` after changing either source channel.
