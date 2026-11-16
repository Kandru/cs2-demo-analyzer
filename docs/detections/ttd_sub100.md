# Sub-100ms damage

<!-- detection: ttd_sub100 -->
<!-- kind: channel -->

## Signal

The share of engagements where the attacker dealt damage within 100 ms of
first visibility. The human reaction floor is roughly 150 ms, so a sustained
sub-100 ms rate is statistically implausible without information or aim
assistance. Two sub-100 ms events in one match are already surprising; a rate
is meaningful only with enough samples.

## Collection

- Collector: `ReactionTimeCollector`, same engagements and visibility source
  as [`reaction`](reaction.md).
- Raw keys: `reaction/sub_100ms_ttd` (percentage), `reaction/sub_100ms_count`,
  `reaction/ttd_samples`.
- Same-tick sightings are discarded before they reach the sample list.

## Scoring

- Ramp: 2% → 30% (higher is worse), linear.
- Confidence: `sqrt(n / 30)`, but pinned to 1.0 when
  `sub_100ms_count ≥ 2`.
- Mode: positive-only.
- Weight: `0.10`.
- Lobby normalisation applies.

## Evidence

- Kind `ttd_sub100`, threshold `100 ms`, rank 3 (strong).
- Recorded at the first-visibility tick, same as `ttd`.
- Value format: `%.0f ms`.

## Calibration

This channel is deliberately paired with a hard sample gate in the
[`ttd_sub100_floor`](ttd_sub100_floor.md) modifier: at least 20 TTD samples,
at least 5 sub-100 ms events, and a tight pre-FOV median on at least 8
samples in an asymmetric lobby. The earlier upstream floor fired at three
samples and produced false positives on short/custom lobbies.

## Interactions

The floor modifier reads this channel's raw keys, not its score, so lobby
normalisation cannot weaken the gate.

## Files

- Implementation: `src/internal/anticheat/stats/detect_ttd_sub100.go`
- Collector: `src/internal/anticheat/stats/reaction_collectors.go`
- Tests: `src/internal/anticheat/stats/detect_ttd_sub100_test.go`,
  `calibration_test.go`
- Report row: registry metadata in `battlelog.go`

## Tuning checklist

1. Change the ramp or confidence pin in `evaluateTTDSub100`.
2. Keep the count pin in sync with the floor modifier's sample gate.
3. Update the cue threshold in `detTTDSub100.Evidence.Threshold`.
4. Re-run `TestEvaluateTTDSub100` and `TestTTDSub100FloorRequiresStrongSample`.
