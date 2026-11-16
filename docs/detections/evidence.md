# Evidence cues

<!-- modifier: evidence -->
<!-- kind: evidence -->

## Signal

Timestamped moments that a reviewer can jump to in the demo. Cues are recorded
only in the suspicious direction and only when a signal crosses its channel's
ramp start. They are navigation aids, never a verdict, and never feed the
score.

## Kinds

| Kind | Label | Rank | Threshold | Value |
|---|---|---|---|---|
| `snap` | Snap | 2 | ≥ 2.0 °/ms | `%.2f °/ms` |
| `ttd` | Reaction | 2 | ≤ 250 ms | `%.0f ms` |
| `ttd_sub100` | Sub-100 ms | 3 | ≤ 100 ms | `%.0f ms` |
| `preaim` | Pre-aim | 3 | ≤ 12° | `%.1f°` |
| `recoil` | Recoil | 2 | ≤ 0.30° | `%.2f°` |
| `wallbang` | Wallbang | 3 | any qualifying kill | `N layers` |
| `smoke` | Through smoke | 2 | any qualifying kill | `kill` |
| `noscope` | No-scope | 2 | any qualifying kill | `kill` |

Thresholds live on the owning detection (`Evidence.Threshold`) or in
`extraEvidence` for kill-feed anomalies. Do not duplicate them in collectors.

## Collection

- `SnapAngleCollector` appends `snap` on a fast pre-kill flick.
- `ReactionTimeCollector` appends `ttd` or `ttd_sub100` at first visibility.
- `BehavioralCollector` appends `preaim` at the victim's FOV-entry tick.
- `RecoilControlCollector` appends `recoil` on a near-perfect burst.
- `EvidenceCollector` appends `wallbang`, `smoke` and `noscope` from kill
  flags.

Moments carry the tick, value, unit, opponent, weapon and a short detail.

## Caps and ordering

- At most 50 moments per kind and 200 per player.
- When a cap is full, the weakest cue by severity is replaced, so long matches
  keep their strongest review targets.
- Severity is defined with the kind; higher is worse. Unknown kinds sort last.
- Moments are sorted by tick (then kind) before publication.
- `internal/analysis` maps ticks to round numbers and clock seconds; ticks
  outside an official round keep round 0 and are labelled by the report.

## Report

Cues are merged per tick in the HTML and exposed as `moments` in the JSON
sidecar. The UI offers `demo_gototick <tick>` for copying.

## Calibration

Never add a cue for a clean-direction observation, and never let a cue affect
the score. If a new detection needs a cue, set `Evidence` on the detection and
append the moment at the measurement site.

## Files

- Implementation: `src/internal/anticheat/stats/evidence.go`
- Moment type/caps: `src/internal/anticheat/stats/types.go`
- Tests: `src/internal/anticheat/stats/evidence_test.go`,
  `registry_test.go`
- Report mapping: `src/internal/anticheat/stats/report.go`,
  `src/internal/report/html/view.go`

## Tuning checklist

1. Add or change the spec (threshold, rank, formatter) in the owning file.
2. Keep the collector append condition on `Evidence.Threshold`.
3. Update this table and the detection page.
4. Re-run `TestAddMomentCapsPerKindAndPlayer` and check the cue on a demo.
