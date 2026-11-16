# Competitive score boost

<!-- modifier: competitive_boost -->
<!-- kind: modifier -->

## Signal

In Competitive mode a player with more than 39 kills in at most 30 rounds is
performing far above a normal match pace. The boost is modest because kill
volume alone is weak evidence.

## Apply

- Condition: `game_info/game_mode == "Competitive"`,
  `total_kills > 39`, `round_count ≤ 30`.
- Multiplier: `×1.2`.
- Applied after the Wingman boost.

The round cap prevents overtime stat-padding from triggering the boost.

## Published metrics

- `anti_cheat/competitive_boost = "Yes"`.

## Report

Adds a `Competitive score boost` adjustment with the detail
`>39 kills in regulation time`.

## Calibration

The threshold is deliberately high and the multiplier small. Raising either
would make a single pop-off game look like a verdict. Keep the regulation cap
in sync with the parser's mode detection if overtime rules change.

## Files

- Implementation: `src/internal/anticheat/stats/mod_competitive_boost.go`
- Registry position: second modifier.

## Tuning checklist

1. Change the condition or multiplier in `applyCompetitiveBoost`.
2. Keep the published description accurate.
3. Re-run a long Competitive demo with overtime.
