# Headshot rate

<!-- detection: hs -->
<!-- kind: channel -->

## Signal

Headshot percentage is the share of a player's kills that were headshots.
Sustained high headshot rates can indicate aim assistance that snaps to the
head hitbox. A low rate is **not** exculpatory: supports, AWPers, wallbang
players and wallhack-only cheaters can all have ordinary headshot rates.

## Collection

- Collector: `HeadshotCollector` (`collect_headshot.go`), `events.Kill`.
- Raw keys: `kills/headshot_kills`, `kills/total_kills`,
  `kills/headshot_percentage`.
- Team kills and suicides are ignored.
- The sample gate is `total_kills > 0`.

## Scoring

- Ramp: 55% → 75% (higher is worse), linear.
- Confidence: `n / 20` kills, capped at 1.
- Mode: positive-only.
- Weight: `0.18` — the strongest aim signal after pre-FOV, but only when the
  sample is meaningful.
- Lobby normalisation applies like every other channel.

## Evidence

No evidence cue. A headshot kill is not a reviewable moment on its own; the
card shows the aggregate and sample count.

## Calibration

The ramp deliberately starts above ordinary rifle headshot rates. The old
bidirectional mode contributed strong negative log-odds below 55%, which
drowned out real wallhack signals from `pre_fov`, `decoupling` and
`back_killed`. Narrative anchors are stricter than the scoring ramp
(55 / 65 / 75 with at least 10 kills) so prose only calls out clear outliers.

## Interactions

Feeds the Bayesian combiner and the evidence-stacking counter. No modifier
reads it directly.

## Files

- Implementation: `src/internal/anticheat/stats/detect_hs.go`
- Tests: `src/internal/anticheat/stats/detect_hs_test.go`
- Report row: registry metadata in `battlelog.go`

## Tuning checklist

1. Adjust the ramp in `evaluateHS` only.
2. Update the calibration notes above with the observed corpus.
3. Keep the positive-only mode unless a documented dataset proves otherwise.
4. Re-run `TestEvaluateHS`, `TestDetectionRegistryContract` and a clean demo.
