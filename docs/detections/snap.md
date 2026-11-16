# Snap velocity P95

<!-- detection: snap -->
<!-- kind: channel -->

## Signal

Snap velocity measures how fast a player's view angle travels immediately
before a kill. A human flick has a wide, heavy-tailed distribution; an
aim-assisted flick can be both larger and faster. The 95th percentile captures
the top of that distribution without depending on a single lucky flick.

## Collection

- Collector: `SnapAngleCollector` (`collect_snap.go`).
- Keeps a 40-tick ring buffer of yaw/pitch per player.
- On a kill, walks backward to the last tick where the aim was stable
  (`< 0.2°` change), then measures angle delta over the elapsed ticks.
- Raw keys: `aiming/p95_snap_velocity`, `aiming/median_snap_velocity`,
  `aiming/avg_snap_velocity`, `aiming/snap_count`.
- Ignores team kills, suicides and kills without a killer.

## Scoring

- Ramp: 2.0 → 3.5 °/ms (higher is worse), linear.
- Confidence: `n / 10` snaps, capped at 1.
- Mode: positive-only. A low P95 only means no fast flicks were sampled.
- Weight: `0.10`. In pro lobbies every aggressive rifler occasionally crosses
  2 °/ms, so a larger weight turns into lobby noise.
- Lobby normalisation applies.

## Evidence

- Kind `snap`, threshold `2.0 °/ms`, rank 2 (mild).
- The cue is recorded at the pre-kill settle tick and shows the velocity.
- Value format: `%.2f °/ms`.

## Calibration

**Unit bug fixed during vendoring.** Upstream `angleDiff` already returns
degrees, but the collector multiplied by `180/π` again, inflating velocity
~57× and turning ordinary flicks into blatant readings. The vendored fix keeps
`angleDiff` in degrees and removes the second conversion. The test
`TestSnapAngleUnits` guards this regression.

Narrative anchors are intentionally higher than the scoring ramp
(4 / 6 / 10 °/ms) because the lobby-normalised score is not intuitive in
prose.

## Interactions

Feeds the Bayesian combiner and evidence stacking. Its cue is a navigation
aid, never a verdict.

## Files

- Implementation: `src/internal/anticheat/stats/detect_snap.go`
- Collector: `src/internal/anticheat/stats/collect_snap.go`
- Tests: `src/internal/anticheat/stats/detect_snap_test.go`,
  `calibration_test.go`
- Report row: registry metadata in `battlelog.go`

## Tuning checklist

1. Change the ramp or weight in `evaluateSnap` only.
2. Never reintroduce a rad→deg conversion in the collector.
3. Update the cue threshold in `detSnap.Evidence.Threshold`.
4. Re-run `TestEvaluateSnap`, `TestSnapAngleUnits` and a clean demo.
