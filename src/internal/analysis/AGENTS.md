# AGENTS.md — internal/analysis

Post-parse merge and summary recomputation.

## Invariants

- `Apply` always attaches a `CheatReport` to every player, even when the
  detector failed or is disabled.
- Anti-cheat failure is non-fatal.
- Summary must be recomputed after cheating data exists (`FlaggedPlayers`, best
  aim, team stats).
- Do not parse demos here; operate on `model.Match` only.
- `Validate` only appends warnings/fills map name; never drops data.

## Tests

Add tests here for summary/team aggregation changes. No demo files.
