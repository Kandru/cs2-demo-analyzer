# Detection documentation

Every cheat detection and score modifier has exactly one page in this folder.
The Go registry is the source of truth for IDs, weights, modes and thresholds;
these pages explain the reasoning, the measurement, and what to change safely.

## Pipeline

```
collectors            raw per-player metrics + timestamped evidence moments
   │
   ▼
detections            one Channel per player (score, confidence, zone)
   │
   ├─ pre_fov_presence and the sub-100 floor use lobby context
   ▼
lobby normalisation   trimmed-mean shrinkage per detection
   │
   ▼
Bayesian combine      weighted log-odds with a 10% prior
   │
   ▼
modifiers             boosts, discounts, floors and overrides
   │
   ▼
publish               cheat_likelihood, <id>_score/_confidence/_zone, flag
```

The two parse passes are intentional: the battlelog parser owns the timeline,
and the detector runs separately so its handlers cannot interfere.

## Files

| ID | File | Kind |
|---|---|---|
| [`hs`](hs.md) | `detect_hs.go` | channel |
| [`snap`](snap.md) | `detect_snap.go` | channel |
| [`reaction`](reaction.md) | `detect_reaction.go` | channel |
| [`ttd_sub100`](ttd_sub100.md) | `detect_ttd_sub100.go` | channel |
| [`recoil`](recoil.md) | `detect_recoil.go` | channel |
| [`pre_fov`](pre_fov.md) | `detect_pre_fov.go` | channel |
| [`pre_fov_presence`](pre_fov_presence.md) | `detect_pre_fov_presence.go` | channel (lobby) |
| [`attention`](attention.md) | `detect_attention.go` | channel |
| [`back_killed`](back_killed.md) | `detect_back_killed.go` | channel |
| [`decoupling`](decoupling.md) | `detect_decoupling.go` | channel |
| [`wingman_boost`](wingman_boost.md) | `mod_wingman_boost.go` | modifier |
| [`competitive_boost`](competitive_boost.md) | `mod_competitive_boost.go` | modifier |
| [`position_discount`](position_discount.md) | `mod_position_discount.go` | modifier |
| [`evidence_stacking`](evidence_stacking.md) | `mod_evidence_stacking.go` | modifier |
| [`wallhack_co_occurrence_boost`](wallhack_co_occurrence_boost.md) | `mod_wallhack_cooccurrence.go` | modifier |
| [`ttd_sub100_high_floor`](ttd_sub100_high_floor.md) | `mod_ttd_sub100_floor.go` | modifier |
| [`sniper_overrides`](sniper_overrides.md) | `mod_sniper_overrides.go` | modifier |
| [`evidence`](evidence.md) | `evidence.go` | review cues |

## Adding a detection

1. Copy `_TEMPLATE.md` to `<id>.md` and fill every section.
2. Create `src/internal/anticheat/stats/detect_<id>.go` with a `Detection`
   value: ID, label, unit, description, doc path, weight, mode, `Evaluate`
   (or `LobbyEvaluate`), `Raw`, optional narrative spec, optional evidence
   spec. Use local constants for the ID/weight/mode so the evaluator does not
   depend on initialization order.
3. Register it in `registry.go` in the intended report/narrative order.
4. Add `detect_<id>_test.go` covering no-data, a clean reading, a suspicious
   reading, and the sample-gate boundary.
5. If it emits a review cue, add the cue to the collector and set
   `Evidence.Kind`/`Threshold`; never score from the cue.
6. Run `go test ./internal/anticheat/stats`. The docs-sync test fails until the
   page exists with the required headings.
7. If the channel is user-visible, check the report card on a real demo.

## Adding a modifier

1. Copy `_TEMPLATE.md` to `<id>.md`.
2. Create `mod_<id>.go` with `Apply`, `Publish` and `Adjustments`.
3. Register it in `registry.go` at the correct stage. Order is the scoring
   contract: boosts first, floor after boosts, overrides last.
4. Add a test for fires/does-not-fire and the exact multiplier or floor.
5. Keep the published key and description stable; the report reads them.

## Changing thresholds

Thresholds are calibration decisions, not refactors. Change one thing at a
time, record the observed false-positive/negative pattern in the page, keep
the sample gate honest, and re-run the calibration tests plus a clean demo.

## Removing a detection

Remove the registry entry, the `detect_*.go`, its tests and its page. Keep the
published `<id>_*` keys out of reports only if the change is intentional; the
schema is additive, so removing keys is a consumer-visible change.
