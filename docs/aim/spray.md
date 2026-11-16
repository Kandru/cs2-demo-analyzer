# Spray

<!-- metric: spray -->

## Formula

`hits on bullets 2..N in sprays of ≥4 bullets / those bullets × 100`.
Higher is better.

## Collection

- Source: `Tracker.OnWeaponFire`, `OnPlayerHurt`, `OnKill`, `OnRoundStart`.
- Only automatic weapons count: SMGs, rifles except snipers, and the M249 /
  Negev.
- A burst starts on the first automatic shot, continues while gaps are
  ≤ 220 ms and the weapon does not change, and closes on death, round end or a
  longer gap.
- Bullet 1 is excluded (it is pre-aim, not spray control); bullets 2..N count
  toward `sprayAtt` and `sprayHits`.

## Sample gate

- Minimum 6 counted bullets for a score.
- Low-sample flag below 6.
- No `PlayerHurt` events forces no-data.

## Anchors and weight

- 15% → 0, 45% → 100 (higher is better).
- Weight `10`.
- Spray control is a real mechanic, but only rifle/SMG players produce
  samples, so it cannot dominate the profile.

## Caveats

Spraying through smoke or at long range lowers the hit rate without indicating
bad control. A player with very few bursts is noisy. The detector's recoil
channel uses a different comparison against the weapon pattern.

## Files

- Implementation: `src/internal/aim/metric_spray.go`
- Tests: `src/internal/aim/metrics_test.go`
- Docs index: `docs/aim/README.md`

## Tuning checklist

1. Change the anchors or weight in `metric_spray.go`.
2. Keep the burst rules (`sprayGapMs`, bullet 1 exclusion) in the tracker.
3. Re-run `go test ./internal/aim`.
