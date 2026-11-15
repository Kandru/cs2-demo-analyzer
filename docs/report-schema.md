# Report data schema (version 1)

The HTML report and the optional `--json` sidecar share `model.Match`. The
schema version is `model.SchemaVersion` (currently `1`) and is stored in the
manifest so template/schema changes invalidate cached reports.

## Root: `Match`

| Field | Type | Notes |
|---|---|---|
| `schemaVersion` | int | 1 |
| `id` | string | short hash of demo name/map/mtime |
| `demoPath`, `demoName`, `demoSize`, `demoHash` | string/int64 | source metadata |
| `map`, `mode`, `tickRate`, `ticks`, `durationSeconds` | — | match metadata |
| `fileTime`, `analyzedAt` | RFC3339 | file mtime / analysis time |
| `teams` | `Team[]` | stable order, indices used in scores |
| `players` | `Player[]` | sorted by team then kills |
| `rounds` | `Round[]` | pregame rounds have `number: 0`, `pregame: true` |
| `warmup` | `Warmup?` | omitted if no warmup activity |
| `summary` | `Summary?` | derived; also recomputed after cheat merge |
| `warnings`, `partial`, `parseError` | — | partial-data reporting |
| `parserVersion`, `toolVersion` | string | provenance |

## `Team`

`id` (stable synthetic team ID captured at match start; the roster, not the
live `TeamState.ID`, is authoritative after a side switch), `name`, `clanName`,
`flag`, `startSide`, `finalSide`, `score`, `playerIds`.

## `Player`

`id` (SteamID64 decimal or `bot-<entityID>`), `steamId64`, `name`, `aliases`,
`clanTag`, `teamId`, `isBot`, `slug`, `roundsPlayed`, `stats`, `sideStats`,
`roundStats`, `weapons`, `aim`, `cheat`.

### `PlayerStats`

Kills, deaths, assists, headshot kills, damage, MVPs, score, first kills/deaths,
trade kills/traded deaths, KAST rounds, plants, defuses, clutches, utility
damage, flash assists, rounds, shots/hits/head hits, damage taken, equipment
value sum, money spent, plus derived ADR, K/D, KDR, HS%, KAST%, accuracy%,
head-accuracy%.

### `AimStats`

`rating` (0–100), `grade`, `confidence` (weight coverage 0–1), `metrics[]`.
Each metric: `id`, `label`, `value`, `display`, `unit`, `score` (0–100),
`weight`, `samples`, `lowSample`, `tooltip`.

IDs: `damage_speed`, `reaction`, `preaim`, `accuracy`, `head_accuracy`,
`counter_strafe`, `spray`.

### `CheatReport`

`available`, `error`, `likelihood`, `flagged`, `flagThreshold`, `overallGrade`,
`grades[]`, `narrative`, `channels[]`, `adjustments[]`, `moments[]`.

`CheatChannel`: `id`, `label`, `score` (0–1), `confidence` (0–1), `zone`,
`raw`, `unit`, `samples`, `weight`, `hasData`, `description`.

Adjustments: `label`, `detail`.

`CheatMoment`: `kind` (`snap`, `ttd`, `ttd_sub100`, `preaim`, `recoil`,
`wallbang`, `smoke`, `noscope`), `tick`, `round` (0 = warmup/pre-match),
`seconds` (demo clock), `value`, `unit`, `opponent`, `weapon`, `detail`.
These are additive review cues, not scoring inputs. `round`/`seconds` are
filled by `internal/analysis` from the battlelog round model.

## `Round`

`number`, `startTick`, `endTick`, `startOffsetSeconds`, `endOffsetSeconds`,
`durationSeconds`, `winner`, `winnerTeamId`, `reason`, `message`,
`startScore`/`endScore` (array indexed by team order), `overtime`, `pregame`,
`economy[]`, `events[]`, `players` (map player ID → `RoundStat`), `warnings`.

`RoundStat`: side, kills, deaths, assists, damage, headshotKills, firstKill,
firstDeath, tradeKill, traded, kast, survived, clutch, equipValue, spent,
plants, defuses, utilityDamage, flashAssists, result.

## `Event`

`seq`, `tick`, `offsetSeconds`, `type`, actor/victim/assister IDs and names,
`weapon`, `headshot`, `wallbang`, `noscope`, `throughSmoke`, `assistedFlash`,
`site`, `grenade`, `teamKill`, `suicide`, `text`, `score`.

Types currently emitted: `round_start`, `freeze_end`, `kill`, `teamkill`,
`suicide`, `bomb_plant`, `bomb_defuse`, `bomb_explode`, `grenade`, `mvp`,
`round_end`.

## `Summary`

Winner, score, rounds, regulation rounds, flagged count, top fragger, best aim,
team stats (kills/deaths/assists/damage/ADR/HS%/aim rating).

## Compatibility rules

- Additive fields are non-breaking; keep `schemaVersion` unless a consumer would
  misinterpret old data.
- Renaming/removing fields requires a schema bump and a manifest invalidation.
- Keep raw detector metrics in `CheatChannel` even when score/zone are present;
  the UI relies on them for transparency.
