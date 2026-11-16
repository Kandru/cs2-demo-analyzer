# Evidence stacking boost

<!-- modifier: evidence_stacking_boost -->
<!-- kind: modifier -->

## Signal

Three independent moderate signals are stronger evidence than the arithmetic
sum suggests. When at least three channels each have
`score × confidence ≥ 0.30`, the combined likelihood receives an ×1.4 boost.

## Apply

- Counts channels with `HasData` and `score × confidence ≥ 0.30`.
- If the count is at least 3: `score ×= 1.4`.
- Otherwise no change.

The count is independent of channel order.

## Published metrics

- `anti_cheat/evidence_stacking_boost = "Yes (N strong channels)"`.

## Report

Adds an `Evidence stacking boost` row whose detail is the published string.
The narrative adds a stacking sentence unless a more specific modifier
(co-occurrence or sniper override) already explains the score.

## Calibration

The 0.30 product threshold is low enough for three genuinely independent
readings and high enough that one strong channel plus noise does not fire it.
Raising the multiplier would let several weak channels compound into a flag;
keep it modest.

## Files

- Implementation: `src/internal/anticheat/stats/mod_evidence_stacking.go`
- Registry position: fourth modifier (after boosts/discount, before
  co-occurrence and the floor).

## Tuning checklist

1. Change the count or product threshold in `applyEvidenceStacking`.
2. Keep the rule conjunction-based, not an OR of weak channels.
3. Re-run a clean pro lobby where every post-normalisation score is elevated.
