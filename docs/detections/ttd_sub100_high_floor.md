# Sub-100ms floor

<!-- modifier: ttd_sub100_high_floor -->
<!-- kind: modifier -->

## Signal

Floors the likelihood at 55% when a player shows an unambiguous triggerbot
profile. The floor prevents a strong but isolated signal from being diluted by
an otherwise clean lobby baseline.

## Apply

All of the following must hold:

- `reaction/ttd_samples ≥ 20`;
- `reaction/sub_100ms_ttd ≥ 25%`;
- `reaction/sub_100ms_count ≥ 5`;
- `behavioral/pre_fov_aim_samples ≥ 8` with a median `> 0°` and `≤ 9°`;
- lobby asymmetry is true for the player (at least half of the other players
  have fewer than 2 pre-FOV samples).

When they hold, `score = max(score, 55)`. Applied after boosts, so it caps
regression rather than amplifying it.

## Published metrics

- `anti_cheat/ttd_sub100_high_floor = "Yes"`.

## Report

Adds a `Sub-100ms floor` row and a closing narrative sentence.

## Calibration

**Documented deviation from upstream.** Upstream floored at 55% with only 3
TTD samples, a 25% sub-100 rate and a 10° pre-FOV median. On short/custom
lobbies that combination fired on ordinary prefires: a 4/17 sub-100 rate with
a 9.2° pre-aim pinned a clean player at 55%. The tightened gates keep the
triggerbot signal while avoiding that false positive.

The asymmetry requirement avoids pro 5v5 lobbies where every player has
pre-FOV samples by virtue of match length. A separate upstream patch drops
same-tick radar-spot artifacts from the TTD samples.

## Files

- Implementation: `src/internal/anticheat/stats/mod_ttd_sub100_floor.go`
- Tests: `src/internal/anticheat/stats/calibration_test.go`
- Registry position: sixth modifier, after boosts.

## Tuning checklist

1. Change the gates in `applyTTDSub100Floor` only.
2. Never lower all gates at once; record the observed false positive.
3. Keep `buildLobbyContext` as the asymmetry source.
4. Re-run `TestTTDSub100FloorRequiresStrongSample` and a short demo.
