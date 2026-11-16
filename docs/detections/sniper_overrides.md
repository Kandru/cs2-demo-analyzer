# Sniper overrides

<!-- modifier: sniper_overrides -->
<!-- kind: modifier -->

## Signal

Two rare, high-specificity sniper patterns pin the likelihood to 100%:

- more than 10 sniper wallbang kills across the match, or
- more than 10 Scout (SSG-08) kills with a headshot rate of at least 80%.

A handful of wallbangs is normal pre-fire play; ten or more implies consistent
knowledge of enemies behind geometry. The Scout's slow rate of fire, lack of
aim punch and pixel-tight head hitbox make 10+ kills at 80%+ headshots
statistically implausible without assistance.

## Apply

- Reads `sniper/sniper_wallbang_kills`,
  `sniper/scout_kills`, `sniper/scout_hs_rate`.
- Sets `score = 100` when either condition holds.
- Applied last, after the clamp, so an override can only raise the score.
- Can trigger both override keys at once.

## Published metrics

- `anti_cheat/sniper_wallbang_override = "Yes"`.
- `anti_cheat/scout_precision_override = "Yes"`.

## Report

Adds one `Sniper anomaly override` row per triggered key. The narrative says
a sniper-anomaly override pinned likelihood to 100%.

## Calibration

These are the only hard overrides. Treat them as exceptional: changing the
threshold or adding a third rule requires multiple confirmed demos. The
existing rules were supplied by the upstream author and are intentionally
pinned rather than additive.

## Files

- Implementation: `src/internal/anticheat/stats/mod_sniper_overrides.go`
- Collector: `src/internal/anticheat/stats/sniper_collector.go`
- Registry position: last modifier.

## Tuning checklist

1. Change thresholds in `applySniperOverrides`.
2. Keep the override last in the registry.
3. Re-run a sniper-heavy clean demo before keeping any change.
