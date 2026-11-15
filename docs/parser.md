# Parser

`internal/parser` converts a CS2 demo into `model.Match` using
`demoinfocs-golang/v5`. It is a single pass; aim collection is fed from the same
pass, and the anti-cheat detector is a separate pass in `internal/anticheat`.

## Lifecycle

1. Open the file; transparently decompress `.gz`/`.bz2` and sniff compressed
   magic bytes even when the extension is `.dem`.
2. Reject Source 1/CS:GO demos (`ErrUnsupportedDemo`). v5 only parses CS2.
3. Read `CDemoFileHeader` (`map`, `server_name`, `demo_file_stamp`).
4. Register handlers, then loop `ParseNextFrame()` until EOF, an error, or
   context cancellation. After every frame, feed the aim tracker with
   `GameState().Participants().Playing()`.
5. Finalize the open round, player aggregates, weapons, aim, summary.
6. On panic or mid-demo error, keep the partial model, add a warning, and
   return it with `Partial=true` when at least one player was seen.

## Event → model mapping

| Event | Model effect |
|---|---|
| `RoundStart` | Ensure teams/players; warmup vs official decision; new `Round` |
| `RoundFreezetimeEnd` | Economy snapshot: freeze-time-end equipment, team money, buy type |
| `Kill` | Round/aggregate kills/deaths/assists/HS, weapon kills, first kill/death, trades, teamkill/suicide classification |
| `PlayerHurt` | Damage (using `HealthDamageTaken`), utility damage, aim shots/hits. **Not present in some early CS2 demos**; metrics degrade gracefully |
| `WeaponFire` | Aim accuracy, counter-strafe, spray bursts |
| `BombPlanted` / `BombDefused` / `BombExplode` | Timeline + plant/defuse counts |
| `GrenadeProjectileThrow` | Timeline utility entry (type/name) |
| `RoundMVPAnnouncement` | MVP; falls back to round top fragger when absent |
| `RoundEnd` | Winner, reason, score, duration, end event |
| `RoundEndOfficial` | Safety net if `RoundEnd` was missed |
| `IsWarmupPeriodChanged` | Switches aim/aggregate collection off during warmup |
| `TeamSideSwitch` / half changes | Not stored directly; sides are read per player at round start |

## Round counting and scores

- Team order is captured at the first round: `teams[0]` = starting T,
  `teams[1]` = starting CT. In some CS2 demos `TeamState.ID()` is side-based
  and is **not** stable across a side switch, so scores/winner states are
  mapped to the stable team via the current roster (majority of players by
  their original team). Never index scores by `TeamState.ID()` after halftime.
- Scores are calculated as `startScore + 1` for the winning team (TeamState
  scores are not yet updated inside `RoundEnd`).
- A candidate round with no damage and no score movement is marked
  `Pregame` (knife round/restart), numbered `0`, and excluded from aggregates.
- Restarts: an unfinished round is closed as pregame before the new one starts.
- Overtime: rounds after regulation (24 competitive, 16 Wingman) get
  `Overtime=true`. Mode is inferred from human player count and the server name
  (`wingman`), falling back to Custom.
- Demos that start mid-match show only observed rounds; scores start from the
  demo's current TeamState values.

## Warmup

All kills/utility before the first official round go to `Match.Warmup`, not to
player aggregates or aim stats. The report always renders warmup first when
present. Warmup end is the first official round start.

## Stats

- ADR = damage / official rounds.
- HS% = headshot kills / kills.
- KAST = kill, assist, survive, or traded in the round.
- Trades: a kill within 5 s of the victim's teammate killing the original killer.
- First kills/deaths: first non-teamkill kill in a round.
- Clutches: player was the last alive on their team and won; size = enemies
  alive at that moment (1v1 = 1).
- MVPs: announcement, else top round fragger.
- Economy: freeze-time-end equipment value; `Pistol` on round 1, otherwise
  `Full buy` ≥ 4000, `Force buy` ≥ 2000, `Eco` below.

## Edge cases handled

- **Missing `PlayerHurt`** (some 2023-era CS2 demos): damage/accuracy/reaction
  show `—`; the report still renders.
- **Bots / zero SteamID**: stable synthetic ID `bot-<entityID>`, shown as BOT.
- **Name changes**: latest name + alias list (aliases not currently rendered).
- **Spectators/coaches/unassigned**: excluded from scoreboard and stats.
- **Teamkills/suicides/world damage**: deaths recorded, killer kills not.
- **Corrupt/truncated demo**: partial report + warnings; parse error surfaced.
- **Compressed demos**: gzip/bzip2 by extension or magic.
- **Empty match**: returns `ErrNoPlayers`.
- **Warmup-only / deathmatch**: rendered as warmup activity with a warning.

## Adding an event

1. Extend `model.Event` fields and `EventType` documentation.
2. Add a handler in `parser.installHandlers`; update the round/stat aggregates.
3. Add parts/badges/icon in `report/html/view.go` and `assets/style.css`.
4. Add a parser test if the event affects round accounting.
