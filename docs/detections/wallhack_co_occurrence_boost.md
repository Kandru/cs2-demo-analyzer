# Wallhack co-occurrence boost

<!-- modifier: wallhack_co_occurrence_boost -->
<!-- kind: modifier -->

## Signal

Combines the two halves of the wallhack-via-information signature:

1. tight pre-FOV pre-aim after lobby normalisation
   (`pre_fov.score × pre_fov.confidence ≥ 0.45`), and
2. a high killer-side back-kill-given rate
   (`back_kill_given_pct ≥ 8.0` on at least 4 kills).

Either alone is noisy: pro lobbies have low post-normalisation pre-FOV scores,
and legitimate flankers have high back-kill-given rates. The conjunction is
much more specific.

## Apply

- Multiplier `×1.2` when both conditions hold.
- No change otherwise.
- Applied after evidence stacking and before the sub-100 ms floor.

## Published metrics

- `anti_cheat/wallhack_co_occurrence_boost = "Yes"`.

## Report

Adds a `Wallhack co-occurrence boost` row. The narrative prefers this sentence
over the generic evidence-stacking sentence.

## Calibration

The 0.45 product threshold has a wide insensitive zone (0.40–0.55 produced the
same result on the reference corpus). It is not fitted to one player's exact
value. Wingman cheaters usually fail because opponents cannot be flanked;
pros fail because their post-normalisation pre-FOV product is lower. Keep the
back-kill sample gate at 4.

## Files

- Implementation: `src/internal/anticheat/stats/mod_wallhack_cooccurrence.go`
- Inputs: `detect_pre_fov.go`, `detect_back_killed.go`
- Registry position: fifth modifier.

## Tuning checklist

1. Change the product or back-kill gate in
   `applyWallhackCoOccurrenceBoost`.
2. Keep the rule a conjunction.
3. Re-run a pro 5v5 demo and a small-lobby demo.
