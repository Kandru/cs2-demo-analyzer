# Aim metrics

The aim profile is computed by `internal/aim` during the battlelog parse. It is
**descriptive**, not part of the cheat score. The vendored detector has its own
timing/pre-aim channels, so numbers can differ slightly; that is expected and
noted in the player card.

No rank or external comparison data is used. The composite rating is only a
readable summary of the raw values.

The per-metric definitions, collection rules, anchors and caveats now live in
[`docs/aim/`](aim/README.md), one page per metric:

| Tile | Metric page |
|---|---|
| **Damage speed** | [damage_speed.md](aim/damage_speed.md) |
| **Reaction** | [reaction.md](aim/reaction.md) |
| **Preaim** | [preaim.md](aim/preaim.md) |
| **Accuracy** | [accuracy.md](aim/accuracy.md) |
| **Head accuracy** | [head_accuracy.md](aim/head_accuracy.md) |
| **Counter-strafe** | [counter_strafe.md](aim/counter_strafe.md) |
| **Spray** | [spray.md](aim/spray.md) |

## Rating

Each available metric scores `[0, 100]` from its anchors. The rating is the
weighted average of available scores, with weights 20 / 20 / 15 / 20 / 10 / 5 /
10 (sum 100). Missing components are dropped and the remaining weights
renormalized. A rating needs at least 3 components; otherwise it is shown as
`—`. Grade bands: A+ ≥85, A ≥75, B+ ≥65, B ≥55, C+ ≥45, C ≥35, D ≥25, F below.

`Confidence` in the JSON is the fraction of total weight that had data.

## Shared implementation notes

- **Visibility** uses demoinfocs' `Player.IsSpottedBy` (server
  `m_bSpottedByMask`). It can be radar spotting rather than exact screen
  visibility; preaim therefore looks back from first damage.
- **Reaction** only counts a shot that subsequently hits the same victim. The
  hit/shot pairing uses the newest unclaimed shot within 250 ms and the same
  weapon.
- **Damage speed** accepts the first damage in an engagement; a 0 ms sample is
  discarded.
- **Accuracy** excludes shotguns, knives, grenades, and gear. One shot that
  damages multiple victims counts once.
- **Head accuracy** is head **hits** / all hits, not headshot kills.
- **Counter-strafe** uses position deltas; teleports and gaps > 0.5 s are
  ignored.
- **Spray** only considers automatic weapons, excludes snipers, and breaks
  bursts on >220 ms gaps, weapon changes, death, or round end.
- Warmup, pregame rounds, and dead players do not contribute.
- Bots receive metrics but are excluded from best-aim summary selection.
- Older CS2 demos without `PlayerHurt` show “no data” for shot-based metrics.

## Per-weapon table

`WeaponStats` carries kills, headshot kills, damage, shots, hits, head hits, and
accuracy. Shots/hits come from the aim tracker; kills/damage from match events.
When the demo has no damage events, shots/accuracy show `—`.

## Changing a formula

1. Edit the owning `src/internal/aim/metric_<id>.go` spec.
2. Update `docs/aim/<id>.md`; the docs-sync test enforces the page.
3. Update this page only if rating weights or the composite rule change.
4. Add a unit test in `internal/aim/metrics_test.go`.
5. Re-run a real 2026 demo and sanity-check that no clean player collapses to
   `F` solely because a metric is noisy.
