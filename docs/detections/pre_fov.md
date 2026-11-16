# Pre-FOV pre-aim

<!-- detection: pre_fov -->
<!-- kind: channel -->

## Signal

The angle between the killer's crosshair and the victim's position 200 ms
before the victim entered the killer's field of view. Legitimate players
pre-aim static map angles; a wallhacker tracks a moving enemy before it is
visible. This is the strongest wallhack-only signal because it measures
information, not aim.

## Collection

- Collector: `BehavioralCollector` (`collect_behavioral.go` + `collect_behavioral_prefov.go`).
- Keeps a 320-tick (5 s at 64 tick) history of view direction and position per
  player.
- On a kill, walks backward to the first tick where the victim left the
  killer's 5° FOV, then looks back 200 ms and measures the angle from the
  killer's view vector to the victim's position at that tick.
- Raw keys: `behavioral/pre_fov_aim_median_deg`,
  `behavioral/pre_fov_aim_samples`.
- Requires at least 4 samples; three-sample readings were noisy (a clean
  teammate landed at 6.94° on exactly three samples).
- No BSP or line-of-sight data is used; positions and view angles are enough.

## Scoring

- Ramp: 12° → 4° (lower is worse), linear.
- Confidence: `sqrt(n / 15)`, capped at 1.
- Mode: bidirectional. A wide pre-aim angle is legitimate evidence of not
  tracking through walls.
- Weight: `0.22`, the largest channel. It is the only reliable wallhack-only
  discriminator.
- Lobby normalisation applies; this is important in pro lobbies where every
  player accumulates samples.

## Evidence

- Kind `preaim`, threshold `12°` (the ramp start), rank 3 (strong).
- Tick is the victim's FOV-entry tick so the reviewer can watch the approach.
- Value format: `%.1f°`.

## Calibration

The ramp deviates from the published 7° → 2.5° table because observed wingman
cheaters land at 6.17° and 7.25°; a 7° clean anchor would zero them out. The
12° → 4° ramp is anchored on the observed corpus. Narrative anchors are
8 / 7 / 6 ° with at least 4 samples.

Visibility is radar-based (`m_bSpottedByMask`), an approximation of true
screen visibility. Treat tight angles as a review cue, not proof.

## Interactions

- [`pre_fov_presence`](pre_fov_presence.md) consumes the sample count and
  lobby asymmetry.
- [`wallhack_cooccurrence`](wallhack_co_occurrence_boost.md) requires the
  post-normalisation score × confidence.
- [`ttd_sub100_floor`](ttd_sub100_high_floor.md) requires a tight pre-FOV median.
- [`decoupling`](decoupling.md) subtracts this median from attention.

## Files

- Implementation: `src/internal/anticheat/stats/detect_pre_fov.go`
- Collector: `src/internal/anticheat/stats/collect_behavioral_prefov.go`
- Tests: `src/internal/anticheat/stats/detect_pre_fov_test.go`
- Report row: registry metadata in `report.go`

## Tuning checklist

1. Change the ramp or weight in `evaluatePreFOV` only.
2. Keep the 200 ms lookback and 4-sample gate unless a new corpus justifies
   otherwise.
3. Update the cue threshold and the floor's pre-FOV gate together.
4. Re-run `TestEvaluatePreFOV` and a clean 5v5 demo.
