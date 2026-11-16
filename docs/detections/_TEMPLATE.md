# <label>

<!-- detection: <id> -->
<!-- kind: channel | modifier | evidence -->

## Signal

What the measurement is and why it can indicate cheating. One paragraph.
State explicitly whether a clean reading is exculpatory or merely neutral.

## Collection

- Source collector and event handlers.
- Raw metric keys and category.
- Sample gate and why it exists.
- Warmup, death, team-kill or visibility handling.

## Scoring

- Ramp: clean anchor → blatant anchor, direction.
- Confidence: formula and full-confidence sample count.
- Mode: bidirectional or positive-only.
- Weight: registry value and the reason for it.
- Lobby normalisation: whether the channel participates and any special case.

## Evidence

- Cue kind, threshold, rank and value format.
- If there is no cue: "No evidence cue."

## Calibration

- Anchors from the observed corpus and published references.
- Known false-positive patterns and the gate that prevents them.
- Deliberate deviations from upstream, if any.

## Interactions

Which modifiers or other channels consume this signal, and in what direction.

## Files

- Implementation: `src/internal/anticheat/stats/<file>.go`
- Tests: `src/internal/anticheat/stats/<file>_test.go`
- Report row: registry metadata in `battlelog.go`

## Tuning checklist

1. Change the ramp/weight/confidence in the detection file only.
2. Update this page with the reason and the observed evidence.
3. Update `registry_test.go` if weight or mode changed.
4. Re-run the per-detection test and the clean-demo check.
