# AGENTS.md — internal/anticheat

Adapter + vendored `demo-anticheat` detector.

## Files

- `anticheat.go` — runs the analyzer as a second parse pass, maps results.
- `analyzer/` — vendored orchestration (collector order lives here).
- `stats/` — collectors + detection scoring. See `stats/AGENTS.md`.
- `stats/report.go` — **ours**: presentation-neutral metric snapshot.

## Invariants

- The detector runs **after** the match parse, in its own parser instance.
- Never let detector errors fail the report; attach an unavailable `CheatReport`.
- Upstream scoring is frozen except for documented, tested integration patches.
  See `docs/anticheat.md` and `docs/THIRD_PARTY.md`.
- No new stdout/stderr prints in vendored code.
- Keep `raw` values and `samples` in the adapter; the UI needs transparency.
- SteamID 0 is the upstream demo-wide placeholder and must not become a player.
- One detection or modifier per file and one matching
  `docs/detections/<id>.md` page. The docs-sync test enforces this.

## Before changing anything

1. Read `docs/anticheat.md`, then `docs/detections/README.md`.
2. Reproduce the behavior with a real demo in `stuff/`.
3. Add a unit test in the matching `detect_*_test.go`.
4. Update the detection page and, if needed, the patch table in
   `docs/THIRD_PARTY.md`.

Do not silently retune thresholds to make one demo pass; document the observed
false-positive pattern and the sample gate that prevents it.
