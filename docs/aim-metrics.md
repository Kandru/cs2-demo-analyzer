# Aim metrics

The aim profile is computed by `internal/aim` during the battlelog parse. It is
**descriptive**, not part of the cheat score. The vendored detector has its own
`median_ttd` / `pre_fov` signals, so numbers can differ slightly; that is
expected and documented in the player card.

No rank or external comparison data is used. The composite rating is only a
readable summary of the raw values.

## Definitions

| Tile | Formula | Sample gate | Score anchors |
|---|---|---|---|
| **Damage speed** | median `first damage tick − first visibility tick` (ms), 1000 ms cap | 3 samples | 800 ms → 0, 250 ms → 100 |
| **Reaction** | median `first shot that hits that victim − first visibility tick` (ms) | 3 | 600 → 0, 180 → 100 |
| **Preaim** | median angle between crosshair and victim position **200 ms before the first damage** | 4 | 25° → 0, 5° → 100 |
| **Accuracy** | shots that hit / shots fired × 100 | 20 shots | 10% → 0, 28% → 100 |
| **Head accuracy** | head hits / all hits × 100 | 10 hits | 10% → 0, 45% → 100 |
| **Counter-strafe** | shots fired at horizontal speed ≤ 34 u/s / shots × 100 | 10 shots | 20% → 0, 80% → 100 |
| **Spray** | hits on bullets 2..N in sprays of ≥4 bullets / those bullets × 100 | 6 bullets | 15% → 0, 45% → 100 |

**Aim rating** is the weighted average of available metric scores, with weights
20 / 20 / 15 / 20 / 10 / 5 / 10 (sum 100). Missing components are dropped and
the remaining weights renormalized. A rating needs at least 3 components;
otherwise it is shown as `—`. Grade bands: A+ ≥85, A ≥75, B+ ≥65, B ≥55,
C+ ≥45, C ≥35, D ≥25, F below.

`Confidence` in the JSON is the fraction of total weight that had data.

## Implementation notes

- **Visibility** uses demoinfocs' `Player.IsSpottedBy` (server `m_bSpottedByMask`).
  This is the same line-of-sight signal the vendored detector uses. It can be
  radar spotting rather than exact screen visibility; preaim therefore looks
  back from first damage instead of sampling at the spot event.
- **Reaction** only counts a shot that subsequently hits the same victim, so a
  shot at another target is not misattributed. The hit/shot pairing uses the
  newest unclaimed shot within 250 ms and the same weapon.
- **Damage speed** accepts the first damage in an engagement; a 0 ms sample
  (visibility and damage in the same tick) is not meaningful and is discarded.
- **Accuracy** excludes shotguns (pellet spread), knives, grenades, and gear.
  A single shot that damages multiple victims counts as one hit.
- **Head accuracy** is head **hits** / all hits, not headshot kills.
- **Counter-strafe** uses position deltas between frames (CS2 player pawns do
  not expose a usable velocity property in demoinfocs v5). Teleports/spawns and
  gaps > 0.5 s are ignored. 34 u/s is roughly standing accuracy speed.
- **Spray** only considers automatic weapons (rifle/SMG/LMG, excluding snipers);
  bursts are broken by >220 ms gaps, weapon changes, death, or round end.
- Warmup, pregame rounds, and dead players do not contribute.
- Bots receive metrics but are excluded from best-aim summary selection.
- Some older CS2 demos do not emit `PlayerHurt` at all. In that case
  accuracy/head accuracy/spray are forced to “no data” instead of showing a
  misleading 0%.

## Per-weapon table

`WeaponStats` carries kills, headshot kills, damage, shots, hits, head hits, and
accuracy. Shots/hits come from the aim tracker; kills/damage from match events.
When the demo has no damage events, shots/accuracy show `—`.

## Changing the formula

1. Add/modify the `def` entry in `aim.buildAimStats`.
2. Update this document and `docs/report-schema.md` if fields change.
3. Add a unit test in `internal/aim/metrics_test.go` covering a no-data case and
   a typical case.
4. Re-run a real 2026 demo and sanity-check that no clean player collapses to
   `F` solely because a metric is noisy.
