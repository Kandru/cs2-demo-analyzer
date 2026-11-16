# Off-engagement attention

<!-- detection: attention -->
<!-- kind: channel -->

## Signal

The median angle from a player's crosshair to the nearest visible-or-not
enemy while no enemy is in the field of view. Legitimate players clear angles
and look where they are going; a wallhacker's crosshair drifts toward enemies
they cannot legally see, producing a smaller median angle.

## Collection

- Collector: `BehavioralCollector`, every frame.
- For each alive player, computes the smallest angle to any alive enemy.
- Only records frames where the nearest enemy is at least 5° away (no active
  engagement).
- Raw keys: `behavioral/nearest_enemy_angle_median_deg`,
  `behavioral/nearest_enemy_angle_samples`.
- Requires at least 200 frames; per-frame samples are plentiful but noisy.

## Scoring

- Ramp: 33° → 18° (lower is worse), linear.
- Confidence: `n / 200`, capped at 1.
- Mode: positive-only. A large angle only means the crosshair is not tight;
  it is not exoneration.
- Weight: `0.06`, deliberately small because idle crosshair placement varies
  enormously by role and playstyle.
- Lobby normalisation applies.

## Evidence

No evidence cue. Per-frame attention is an aggregate signal; individual
frames are not meaningful review targets.

## Calibration

Narrative anchors are 27° / 24° / 20° with at least 200 frames. The channel is
noisy in isolation and is mainly useful in combination with pre-FOV and
decoupling. Do not raise its weight without a larger corpus.

## Interactions

[`decoupling`](decoupling.md) subtracts the pre-FOV median from this median.
The difference is the fight-vs-idle signal.

## Files

- Implementation: `src/internal/anticheat/stats/detect_attention.go`
- Collector: `src/internal/anticheat/stats/collect_behavioral_attention.go`
- Tests: `src/internal/anticheat/stats/detect_attention_test.go`
- Report row: registry metadata in `report.go`

## Tuning checklist

1. Change the ramp or weight in `evaluateAttention`.
2. Keep the 5° FOV exclusion and 200-frame gate.
3. Update decoupling if the median definition changes.
4. Re-run `TestEvaluateAttention` and a demo with a range of playstyles.
