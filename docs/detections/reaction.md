# Damage speed (median TTD)

<!-- detection: reaction -->
<!-- kind: channel -->

## Signal

Time-to-damage (TTD) is the delay from the moment an enemy first becomes
visible to the attacker (engine line-of-sight) to the first damage dealt by
that attacker. Median TTD captures a consistent reaction pattern; a single
pre-fired tick does not. Published clean ranges cluster around 500–600 ms,
with fast professionals at 400–500 ms and sub-250 ms medians suspicious.

## Collection

- Collector: `ReactionTimeCollector` (`reaction_collectors.go`).
- Visibility uses `opponent.IsSpottedBy(attacker)` (`m_bSpottedByMask`), not a
  geometric cone, so walls and smoke are respected.
- An engagement starts on first visibility, survives brief LoS flickers up to
  200 ms, and is capped at 1000 ms.
- First damage during the engagement records one sample.
- A same-tick sighting and damage is discarded as a radar-spot artifact
  (samples below half a tick).
- Raw keys: `reaction/median_ttd`, `reaction/p10_ttd`,
  `reaction/ttd_samples`, plus the sub-100 ms keys used by the other channel.

## Scoring

- Ramp: 500 ms → 150 ms (lower is worse), linear.
- Confidence: `sqrt(n / 10)`, capped at 1, because TTD is heavy-tailed.
- Mode: bidirectional. A 500 ms median on many samples is real evidence of
  human-paced reactions.
- Weight: `0.10`.
- Lobby normalisation applies.

## Evidence

- Kind `ttd`, threshold `250 ms`, rank 2 (mild).
- Tick is the engagement's first-visibility tick so the reviewer can watch the
  sighting, not just the shot.
- Value format: `%.0f ms`.

## Calibration

The channel originally used P10 TTD, which was too noisy: one legitimate
pre-fire (audio cue, pre-armed peek, callout) produced a sub-100 ms P10. The
median requires a sustained pattern. The same-tick drop prevents radar-spot
artifacts from becoming reaction samples.

## Interactions

The companion [`ttd_sub100`](ttd_sub100.md) channel and the
[`ttd_sub100_floor`](ttd_sub100_floor.md) modifier read the same sample set.
The floor requires the median/pre-FOV context documented there.

## Files

- Implementation: `src/internal/anticheat/stats/detect_reaction.go`
- Collector: `src/internal/anticheat/stats/reaction_collectors.go`
- Tests: `src/internal/anticheat/stats/detect_reaction_test.go`
- Report row: registry metadata in `battlelog.go`

## Tuning checklist

1. Change the ramp or confidence in `evaluateReactionMedianTTD`.
2. Keep the 1000 ms engagement cap and the same-tick drop.
3. Update the cue threshold in `detReaction.Evidence.Threshold`.
4. Re-run `TestEvaluateReaction` and a real demo.
