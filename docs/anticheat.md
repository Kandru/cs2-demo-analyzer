# Anti-cheat detection

The report embeds the verdicts of a vendored copy of
[`TimAnthonyAlexander/demo-anticheat`](https://github.com/TimAnthonyAlexander/demo-anticheat)
at commit `635e8c7a152f3026446463cbe8f0ba2e170ec818` (2026-05-18).

- Code lives in `src/internal/anticheat/{analyzer,stats}` and is compiled into
  the binary.
- It runs as a **second parse pass** so it cannot interfere with the battlelog
  timeline.
- `stats/battlelog.go` (our file) exposes the published `anti_cheat` metrics in a
  presentation-neutral form. `internal/anticheat/anticheat.go` maps them to
  `model.CheatReport`.
- Upstream has **no license file** at this commit. Confirm redistribution
  permission before publishing binaries/source. See `THIRD_PARTY.md`.

## What the player card shows

- Final `cheat_likelihood` (0–100), threshold 50 %.
- Flag verdict from upstream (`cheater == Yes`).
- Per-channel: human-readable raw value, evidence score bar, confidence,
  zone (`clean/mild/strong/blatant/no data`).
- The detector's narrative paragraph and skill grades.
- Boosts/discounts/overrides that shaped the final score.
- Timestamped review cues (match clock, round, tick) that link into the round
  timeline and can be copied as `demo_gototick <tick>`.
- An explicit “statistical signal, not proof” disclaimer.

Cues are recorded only in the suspicious direction and only when a channel
crosses its ramp start. They are reviewer navigation aids, never a verdict.

## Upstream method (summary)

Ten channels are combined with a Bayesian log-odds model:

| Channel | Signal |
|---|---|
| `hs` | headshot rate |
| `snap` | P95 view-angle snap velocity |
| `reaction` | median sight→damage time |
| `ttd_sub100` | share of sub-100 ms engagements |
| `recoil` | spray deviation vs known patterns |
| `pre_fov` | crosshair-to-enemy angle 200 ms before visibility |
| `pre_fov_presence` | pre-aim sample asymmetry vs the lobby |
| `attention` | crosshair-to-nearest-enemy when idle |
| `back_killed` | share of deaths while looking away |
| `decoupling` | attention − pre-FOV delta |

Scores are lobby-normalized (trimmed mean, 40 % shrinkage), combined, then
boosted/discounted (Wingman/Competitive/position/evidence stacking/co-occurrence)
and finally floored/overridden. Flag threshold: `likelihood ≥ 50`.

## Integration patches (deviations from upstream)

These changes are intentional and covered by tests/docs. Upstream logic is
otherwise unchanged.

1. **Snap velocity unit bug fixed** (`snap_collectors.go`).
   Upstream `angleDiff` returns **degrees**, but the collector multiplied the
   result by `180/π` again, inflating snap velocity ~57×. A clean player with
   normal flicks showed ~7.7 °/ms (blatant) instead of ~0.13 °/ms. Removing the
   double conversion eliminates a major false-positive source.

2. **Sub-100 ms floor requires a real sample** (`cheatscore_overrides.go`).
   Upstream floored any player at 55 % with ≥3 TTD samples, ≥25 % sub-100 ms
   and a pre-FOV median ≤10° in an asymmetric lobby. On short/custom lobbies
   this flagged a clean player (4/17 sub-100, 9.2° preaim). The floor now needs:
   - ≥20 TTD samples,
   - ≥5 sub-100 ms engagements,
   - pre-FOV median ≤9° on ≥8 pre-FOV samples,
   - lobby asymmetry.

3. **Same-tick TTD discarded** (`reaction_collectors.go`).
   A sighting and damage in the same tick is a radar-spot artifact, not a
   reaction. Samples below half a tick are dropped, and a `sub_100ms_count`
   metric is published for the floor gate.

4. **Debug stdout removed** (`recoil_collectors.go`).
   Upstream printed a debug block on every run.

5. **Adapter additions** (`battlelog.go`) only; no scoring changes.

6. **Timestamped evidence cues** (`types.go`, `snap_collectors.go`,
   `reaction_collectors.go`, `behavioral_collectors.go`,
   `recoil_collectors.go`, new `evidence_collector.go`). Each collector appends
   an `EvidenceMoment` when the signal crosses the corresponding channel ramp:

   | Cue | Threshold |
   |---|---|
   | `snap` | ≥ 2.0 °/ms |
   | `ttd` / `ttd_sub100` | ≤ 250 ms / ≤ 100 ms |
   | `preaim` | ≤ 12° pre-FOV angle |
   | `recoil` | ≤ 0.30° mean spray error, ≥ 3 bullets |
   | `wallbang`, `smoke`, `noscope` | any qualifying kill |

   Moments carry an additive per-player cap (50/kind, 200 total), never feed
   the score, and are exposed through `BattlelogReports` → `model.CheatReport`.
   `internal/analysis` maps their ticks to round numbers and seconds. Scoring
   output is unchanged; calibration tests must stay green.

The vendored code should be updated as a unit when upstream changes. Re-apply
the patches and re-run `docs/testing.md` checks plus the anti-cheat calibration
tests.

## Calibration tests

`internal/anticheat/stats/calibration_test.go` covers:

- thin sub-100 sample does **not** pin a clean player at 55 %,
- a large, unambiguous triggerbot profile **does** floor at ≥55 %,
- `angleDiff` remains degrees (snap-unit regression).

## Limitations

- Probabilistic, tuned on a small reference corpus (Wingman + pro 5v5). It is
  not proof and not a ban system.
- Radar-spot visibility (`m_bSpottedByMask`) is an approximation of true screen
  visibility, especially for preaim/TTD channels.
- Short matches, custom lobbies, and unusual game modes reduce confidence.
  Read the sample counts and zones.
- No cross-demo history: each report is one match. A single-game sample can be
  noisy; future work can aggregate per SteamID.
- Warmup is included by the upstream detector (it does not filter warmup),
  which can add noise. The battlelog timeline excludes it.

## Responsible use

Show raw metrics, confidence, and sample sizes. Never present a flag as a
conviction. If you publish reports, include the disclaimer and avoid naming
players as cheaters in headlines.
