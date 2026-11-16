# Pre-aim sample asymmetry

<!-- detection: pre_fov_presence -->
<!-- kind: channel (lobby) -->

## Signal

Detects a player who produces a meaningful number of tight pre-FOV samples
while the rest of the lobby does not. In a long 5v5 everyone accumulates
pre-FOV samples; in a small or unusual lobby a wallhacker stands out because
teammates and opponents never approach the same pre-aim density.

This is a **lobby-relative** channel and cannot be evaluated per player alone.

## Collection

- Uses `behavioral/pre_fov_aim_samples` and
  `behavioral/pre_fov_aim_median_deg` from `BehavioralCollector`.
- Lobby context is built once in `buildLobbyContext`: for each player, count
  how many *other* players have fewer than 2 pre-FOV samples. Asymmetric means
  at least two other players and at least 50% of them are below the threshold.

## Scoring

Fires only when all three conditions hold:

- at least 4 pre-FOV samples for the player;
- pre-FOV median greater than 0° and at most 10°;
- lobby asymmetry is true for that player.

When it fires, the score ramps from 0.5 at 3 samples to 1.0 at 6 samples and
confidence is pinned to 1.0. Otherwise the channel is still published with
`HasData=false` so the report keeps a stable row.

- Mode: positive-only.
- Weight: `0.10`.
- Lobby normalisation applies after the channel is appended.

## Evidence

No evidence cue. The pre-aim samples themselves produce `preaim` cues.

## Calibration

The asymmetry requirement avoids false positives in pro 5v5 lobbies where
every player has pre-FOV samples by virtue of match length. The 10° median gate
keeps the rule from firing on random crosshair placement. Narrative uses the
published score directly: strong at 0.50, blatant at 0.80, no mild tier.

## Interactions

The same lobby tally feeds [`ttd_sub100_floor`](ttd_sub100_floor.md). Do not
compute asymmetry twice or with different thresholds.

## Files

- Implementation: `src/internal/anticheat/stats/detect_pre_fov_presence.go`
- Lobby context: `src/internal/anticheat/stats/combine.go`
- Tests: `src/internal/anticheat/stats/detect_pre_fov_presence_test.go`
- Report row: registry metadata in `battlelog.go`

## Tuning checklist

1. Change the gates in `evaluatePreFOVPresence` only.
2. Keep `buildLobbyContext` as the single asymmetry implementation.
3. Update the floor modifier if the asymmetry rule changes.
4. Re-run `TestEvaluatePreFOVPresence` and a short Wingman demo.
