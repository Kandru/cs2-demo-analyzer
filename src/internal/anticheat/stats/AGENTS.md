# AGENTS.md — internal/anticheat/stats

Measurement collectors plus the first-party detection registry.

## Layout

| Area | Files |
|---|---|
| Contract | `detection.go`, `modifier.go`, `keys.go` |
| Registry | `registry.go` |
| Scoring | `run.go`, `combine.go`, `publish.go`, `narrative.go` |
| Detections | `detect_<id>.go` (one per channel) |
| Modifiers | `mod_<id>.go` (one per adjustment) |
| Collectors | `collect_*.go` / `*_collector(s).go` |
| Evidence | `evidence.go`, `types.go` |
| Report snapshot | `battlelog.go` |
| Tests | `detect_<id>_test.go`, `registry_test.go`, `calibration_test.go`, `docs_sync_test.go` |

## Adding a detection

1. Add `detect_<id>.go` with a `Detection` spec: ID/label/unit/description/doc,
   weight, mode, evaluator, raw extractor, narrative and optional evidence.
2. Use local constants for ID/weight/mode so the evaluator does not depend on
   package initialization order.
3. Register it in `registry.go` in the intended order.
4. Add the matching `docs/detections/<id>.md` page (see `_TEMPLATE.md`).
5. Add `detect_<id>_test.go` covering no data, clean, suspicious and the
   sample-gate boundary.
6. If it emits review cues, put the threshold on the spec and append the
   moment from the collector. Cues never score.

Adding a modifier is the same pattern with `mod_<id>.go`; order in
`registry.go` is the scoring contract.

## Invariants

- Published metric keys are `<id>_score`, `<id>_confidence`, `<id>_zone`.
- Do not change weights/modes/thresholds during a structural refactor.
- Lobby-dependent detections use `LobbyEvaluate`; lobby data comes from
  `buildLobbyContext` only.
- The registry is immutable after package initialization.
- No per-detection switches in `run.go`, `publish.go`, `narrative.go`,
  `battlelog.go` or the HTML renderer.

## Verification

`go test ./internal/anticheat/stats` includes the registry contract, the
per-detection tests and the docs-sync check. For a behavior-preserving change,
also run the opt-in demo golden test documented in
`src/internal/anticheat/stats/demo_golden_test.go`.
