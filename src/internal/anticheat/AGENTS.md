# AGENTS.md — internal/anticheat

Adapter + vendored `demo-anticheat` detector.

## Files

- `anticheat.go` — runs the analyzer as a second parse pass, maps results.
- `analyzer/` — vendored upstream (orchestration).
- `stats/` — vendored upstream collectors/scoring.
- `stats/battlelog.go` — **ours**: presentation-neutral metric snapshot.
- `stats/calibration_test.go` — floor/snap regressions.

## Invariants

- The detector runs **after** the battlelog parse, in its own parser instance.
- Never let detector errors fail the report; attach an unavailable `CheatReport`.
- Upstream scoring is frozen except for documented, tested integration patches
  (snap unit, sub-100 floor, same-tick TTD, debug output). See
  `docs/anticheat.md` and `docs/THIRD_PARTY.md`.
- No new stdout/stderr prints in vendored code.
- Keep `raw` values and `samples` in the adapter; the UI needs transparency.
- SteamID 0 is the upstream demo-wide placeholder and must not become a player.

## Before changing anything

1. Read `docs/anticheat.md`.
2. Reproduce the behavior with a real demo in `stuff/`.
3. Add a unit test (prefer pure `PlayerStats` setup over parsing).
4. Update the patch table in `docs/THIRD_PARTY.md` and `docs/anticheat.md`.

Do not silently retune thresholds to make one demo pass; document the observed
false-positive pattern and the sample gate that prevents it.
