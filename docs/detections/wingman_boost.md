# Wingman score boost

<!-- modifier: wingman_boost -->
<!-- kind: modifier -->

## Signal

Wingman is a short 2v2 format where a small number of rounds makes raw kill
volume less informative and the lobby is small. A player with ≥10 kills, or a
KPR of at least 0.7, receives an ×1.8 boost to the combined likelihood.

## Apply

- Condition: `game_info/game_mode == "Wingman"` and either
  `total_kills ≥ 10` or `kills / round_count ≥ 0.7`.
- Multiplier: `×1.8`.
- Applied immediately after the Bayesian combination, before every other
  modifier.

## Published metrics

- `anti_cheat/wingman_boost = "Yes"`.
- `anti_cheat/wingman_kpr_boost_reason` is `kills=N` or `KPR=x.xx`.

## Report

Adds a `Wingman score boost` adjustment row with the reason. The narrative
closes with a sentence noting the short-format boost.

## Calibration

The kills backstop exists for demos where the round count is missing or
unreliable. Do not raise the multiplier without checking short Wingman demos;
the format already has fewer rounds and a noisier baseline. The reason string
is intentionally included in both JSON and UI so a reviewer can see which
branch fired.

## Files

- Implementation: `src/internal/anticheat/stats/mod_wingman_boost.go`
- Tests: registry contract and evaluation through the stats tests.
- Registry position: first modifier.

## Tuning checklist

1. Change the threshold or multiplier in `applyWingmanBoost`.
2. Update this page and the published description.
3. Re-run a Wingman demo and confirm no clean player gains a flag.
