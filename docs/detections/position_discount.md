# Scoreboard-position discount

<!-- modifier: position_discount -->
<!-- kind: modifier -->

## Signal

A player who never climbs off the bottom of their team's scoreboard is less
likely to be a cheating carry; low-impact players can still produce noisy
per-channel readings. The discount modestly damps the combined likelihood.

## Apply

- Reads `scoreboard/position_factor`: the average rank fraction within the
  player's side at round 5, halftime and the final round. `0` is top of the
  team, `1` is bottom.
- `discount = 0.20 × position_factor`.
- `score = score × (1 − discount)`.
- No-op when the factor is absent or zero.

The factor is computed by `ScoreboardCollector` from per-round kill snapshots.
Ties are ordered by SteamID for reproducible output.

## Published metrics

- `anti_cheat/position_discount` as a percentage (`discount × 100`).

## Report

Adds a `Scoreboard-position discount` row with `N% reduction`.

## Calibration

The maximum discount is 20%, so position alone can never exonerate a player.
The checkpoints (round 5, midpoint, end) smooth out early-game noise and avoid
depending on the final scoreboard only.

## Files

- Implementation: `src/internal/anticheat/stats/mod_position_discount.go`
- Factor source: `src/internal/anticheat/stats/scoreboard_collector.go`
- Registry position: third modifier.

## Tuning checklist

1. Change the maximum discount in `applyPositionDiscount`.
2. Keep tie-breaking deterministic in the scoreboard collector.
3. Verify a top-fragger clean player and a bottom-fragger clean player.
