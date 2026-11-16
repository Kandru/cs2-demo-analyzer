# AGENTS.md — internal/parser

Converts a CS2 demo into `model.Match` in one demoinfocs pass.

## Files

- `parser.go` — lifecycle, handlers, round builder, warmup, scores, stats.
- `helpers.go` — finalization, summary, slugs, reason/buy strings.

## Invariants

- Do not import `internal/report` or `internal/store`.
- Team order is fixed at first round (`teams[0]` = starting T). Scores are
  indexed by that order and must not be re-sorted.
- `RoundEnd` fires before TeamState scores update. Use `startScore + 1`, never
  the event's team score.
- Warmup/pre-game activity must not enter aggregate stats or aim samples.
- A round with no damage and no score change is `Pregame` (`Number=0`).
- Any event added to the timeline must be escaped by the renderer only; names
  are stored raw.
- Round and warmup event offsets are section-relative. `Match.Events` is the
  match lifecycle stream and uses absolute demo offsets.
- Lifecycle events: `player_connect`, `player_disconnect`, `team_change`,
  `team_swap`. Halftime swaps produce one `team_swap`, not one row per player.
- Resolve stable team IDs through the roster; live `TeamState.ID()` is not
  reliable after a side switch.
- `parse` may return a partial match with an error; do not discard it if it has
  players. `ErrNoPlayers` is the only “nothing useful” case.
- Context cancellation is checked every frame.

## Edge cases

See `docs/parser.md`. Key ones: missing `PlayerHurt` in older demos, restarts,
knife rounds, halftime side swaps, overtime, truncated demos, bots/zero SteamID,
spectators, teamkills/suicides.

## Tests

`helpers_test.go` covers slugs, buy types, finalization, IDs. Add tests for any
new round accounting. Demo-backed tests are manual (`stuff/`).
