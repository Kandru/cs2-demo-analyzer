# Anti-cheat detection

The report embeds the verdicts of a vendored copy of
[`TimAnthonyAlexander/demo-anticheat`](https://github.com/TimAnthonyAlexander/demo-anticheat)
at commit `635e8c7a152f3026446463cbe8f0ba2e170ec818` (2026-05-18).

- Code lives in `src/internal/anticheat/{analyzer,stats}` and is compiled into
  the binary.
- It runs as a **second parse pass** so it cannot interfere with the match
  timeline.
- `stats/report.go` exposes the published metrics in a presentation-neutral
  form; `internal/anticheat/anticheat.go` maps them to `model.CheatReport`.
- Upstream has **no license file** at this commit. Confirm redistribution
  permission before publishing binaries/source. See `THIRD_PARTY.md`.

## Detection catalogue

Each signal has its own implementation file, test and documentation page. Start
at [`docs/detections/README.md`](detections/README.md) for the pipeline, the
full file map, and the add/remove/change procedure.

The ten scored channels are:

| Channel | Signal | Weight |
|---|---|---|
| `hs` | headshot rate | 0.18 |
| `snap` | P95 snap velocity | 0.10 |
| `reaction` | median sight→damage time | 0.10 |
| `ttd_sub100` | sub-100 ms engagement share | 0.10 |
| `recoil` | spray deviation vs known patterns | 0.10 |
| `pre_fov` | crosshair-to-enemy angle 200 ms before visibility | 0.22 |
| `pre_fov_presence` | pre-aim sample asymmetry vs the lobby | 0.10 |
| `attention` | crosshair-to-nearest-enemy when idle | 0.06 |
| `back_killed` | share of deaths while looking away | 0.06 |
| `decoupling` | attention − pre-FOV delta | 0.10 |

Seven modifiers shape the final score: Wingman/Competitive boosts,
scoreboard-position discount, evidence stacking, wallhack co-occurrence,
the sub-100 ms floor, and sniper overrides. The ordered registry lives in
`stats/registry.go`.

## What the player card shows

- Final `cheat_likelihood` (0–100), flag threshold 50%.
- Flag verdict from the detector (`cheater == Yes`).
- Per-channel: raw value, score bar, confidence, zone.
- The detector's narrative paragraph and skill grades.
- Boosts/discounts/overrides that shaped the final score.
- Timestamped review cues (match clock, round, tick) that link into the round
  timeline and can be copied as `demo_gototick <tick>`.
- An explicit “statistical signal, not proof” disclaimer.

Cues are recorded only in the suspicious direction and only when a channel
crosses its ramp start. They are reviewer navigation aids, never a verdict.

## Scoring model (summary)

1. Each channel produces `score` in `[0, 1]` and `confidence` in `[0, 1]`.
2. Channels with data are lobby-normalised with a trimmed mean and 40%
   shrinkage, so one outlier cannot inflate their own baseline.
3. Scores combine via Bayesian log-odds with a 10% prior. Positive-only
   channels contribute no negative evidence; bidirectional channels do.
4. Modifiers apply in registry order, then the likelihood is clamped to 100.
5. Flag threshold: `likelihood ≥ 50`.

## Integration patches (deviations from upstream)

These changes are intentional and covered by tests/docs. Scoring logic is
otherwise unchanged.

1. **Snap velocity unit bug fixed** (`collect_snap.go`). Upstream
   `angleDiff` returns degrees but multiplied by `180/π` again, inflating snap
   velocity ~57×. Removing the double conversion eliminates a major
   false-positive source.
2. **Sub-100 ms floor tightened** (`mod_ttd_sub100_floor.go`). Upstream
   floored at 55% with 3 samples; the floor now requires ≥20 TTD samples,
   ≥5 sub-100 ms events, a ≤9° pre-FOV median on ≥8 samples, and lobby
   asymmetry. Details in
   [`ttd_sub100_high_floor.md`](detections/ttd_sub100_high_floor.md).
3. **Same-tick TTD discarded** (`collect_reaction.go`). A sighting and
   damage in the same tick is a radar-spot artifact, not a reaction.
4. **Debug stdout removed** (`collect_recoil.go`).
5. **Unused weapon-usage collector removed** and unused reporter code dropped
   during vendoring.
6. **Timestamped evidence cues** added per detection; see
   [`evidence.md`](detections/evidence.md). Cues never feed the score.
7. **Package reorganisation.** Detection scoring was split from the upstream
   monolithic files into one file per detection/modifier with a registry. The
   formulas, weights and thresholds are unchanged; this is structure only.

The vendored code should be updated as a unit when upstream changes. Re-apply
the patches and re-run the calibration plus per-detection tests.

## Calibration tests

- `stats/calibration_test.go` covers the sub-100 floor and snap-unit
  regression.
- `stats/registry_test.go` pins the channel/modifier surface.
- `stats/detect_*_test.go` covers each signal independently.
- `stats/docs_sync_test.go` fails when a registered feature has no page.

## Limitations

- Probabilistic, tuned on a small reference corpus (Wingman + pro 5v5). It is
  not proof and not a ban system.
- Radar-spot visibility (`m_bSpottedByMask`) is an approximation of true screen
  visibility, especially for preaim/TTD channels.
- Short matches, custom lobbies, and unusual game modes reduce confidence.
  Read the sample counts and zones.
- No cross-demo history: each report is one match.
- Warmup is included by the upstream detector; the match timeline excludes
  it.

## Responsible use

Show raw metrics, confidence, and sample sizes. Never present a flag as a
conviction. If you publish reports, include the disclaimer and avoid naming
players as cheaters in headlines.
