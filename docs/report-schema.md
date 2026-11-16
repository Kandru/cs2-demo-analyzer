# Report data schema (version 1)

The HTML report and the optional `--json` sidecar share one canonical
`report.Document`: a `header` object followed by `model.Match` fields. The data
schema version is `model.SchemaVersion` (currently `1`).

Every artifact is self-describing. The header is always the first JSON key and
an HTML comment immediately after the doctype:

```html
<!doctype html>
<!--cs2-demo-analyzer:{"format":1,"toolVersion":"26.11.1",...}-->
```

Readers can therefore decide whether a file is current by reading a few
kilobytes. See “Artifact header” below.

## Artifact header

| Field | Type | Notes |
|---|---|---|
| `format` | int | Header format (`report.HeaderFormat`, currently `1`) |
| `toolVersion` | string | `version.CacheKey()`; `dev+<commit>` for dev builds |
| `schemaVersion` | int | Mirrors `model.SchemaVersion` |
| `options` | string | `cheat=…;aim=…;demo-download=…;index=…` |
| `demoSize` | int64 | Source file size |
| `demoModUnixNano` | int64 | Source mtime |
| `generatedAt` | RFC3339 | Render time |
| `map`, `serverName`, `score`, `hasScore`, `playedAt` | — | Index metadata; not part of freshness |

Freshness ignores presentation fields (`generatedAt`, map/server/score/played)
and compares format, tool version, schema version, options, demo size and demo
mtime. A missing or corrupt header marks the artifact stale.

## Root: `Match`

| Field | Type | Notes |
|---|---|---|
| `header` | object | Artifact identity; always first (see above) |
| `schemaVersion` | int | 1 |
| `id` | string | short hash of demo name/map/mtime |
| `demoPath`, `demoName`, `demoSize`, `demoHash` | string/int64 | source metadata |
| `map`, `mode`, `tickRate`, `ticks`, `durationSeconds` | — | match metadata |
| `fileTime`, `analyzedAt` | RFC3339 | file mtime / analysis time |
| `teams` | `Team[]` | stable order, indices used in scores |
| `players` | `Player[]` | sorted by team then kills |
| `rounds` | `Round[]` | pregame rounds have `number: 0`, `pregame: true` |
| `warmup` | `Warmup?` | omitted if no warmup activity |
| `events` | `Event[]?` | match-level lifecycle events; offsets are absolute demo seconds |
| `summary` | `Summary?` | derived; also recomputed after cheat merge |
| `warnings`, `partial`, `parseError` | — | partial-data reporting |
| `parserVersion`, `toolVersion` | string | provenance |

## Match lifecycle events

`Match.events` is a separate stream from round and warmup event lists. It is
optional and backward compatible. Every entry uses the normal `Event` shape and
an absolute `offsetSeconds`.

| Type | Meaning |
|---|---|
| `player_connect` | player or bot connected |
| `player_disconnect` | player disconnected |
| `team_change` | player moved between spectator/unassigned and a team, or changed sides outside a halftime swap |
| `team_swap` | both teams exchanged sides at halftime or in overtime |

`team_change` entries may carry `oldSide` and `newSide`. `team_swap` entries use
`text` with the new side assignment. Round and warmup `Event` offsets remain
relative to their section; only `Match.events` offsets are absolute.

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
`wallbang`, `smoke`, `noscope`), `label`, `rank` (1 info, 2 mild, 3 strong),
`valueText` (display string), `tick`, `round` (0 = warmup/pre-match),
`seconds` (demo clock), `value`, `unit`, `opponent`, `weapon`, `detail`.
`label`, `rank` and `valueText` are additive presentation metadata supplied by
the detector's evidence spec; reporters can render a cue without a per-kind
switch. These are review cues, not scoring inputs. `round`/`seconds` are
filled by `internal/analysis` from the match round model.

## `Round`

`number`, `startTick`, `endTick`, `startOffsetSeconds`, `endOffsetSeconds`,
`durationSeconds`, `winner`, `winnerTeamId`, `reason`, `message`,
`startScore`/`endScore` (array indexed by team order), `overtime`, `pregame`,
`economy[]`, `events[]`, `players` (map player ID → `RoundStat`), `warnings`.

`RoundStat`: side, kills, deaths, assists, damage, headshotKills, firstKill,
firstDeath, tradeKill, traded, kast, survived, clutch, equipValue, spent,
`bought[]` (freeze-time-end inventory for that player), plants, defuses,
utilityDamage, flashAssists, result. `bought` is additive in schema v1; older
artifacts simply omit it.

## `Event`

`seq`, `tick`, `offsetSeconds`, `type`, actor/victim/assister IDs and names,
`weapon`, `headshot`, `wallbang`, `noscope`, `throughSmoke`, `assistedFlash`,
`site`, `grenade`, `teamKill`, `suicide`, `text`, `score`, `amount`.

`amount` is additive: flash duration in seconds for `flash`, damage for
`grenade_damage`.

Types currently emitted: `round_start`, `freeze_end`, `kill`, `teamkill`,
`suicide`, `bomb_plant`, `bomb_defuse`, `bomb_explode`, `bomb_pickup`,
`bomb_drop`, `grenade`, `grenade_damage`, `flash`, `mvp`, `round_end`,
`hostage_pickup`, `hostage_drop`, `hostage_rescue`, `hostage_killed`.
Match lifecycle events use the same shape in `Match.events`:
`player_connect`, `player_disconnect`, `team_change`, `team_swap`.

## `Summary`

Winner, score, rounds, regulation rounds, flagged count, top fragger, best aim,
team stats (kills/deaths/assists/damage/ADR/HS%/aim rating).

## Compatibility rules

- The `header` object is additive: existing Match consumers can ignore it.
- Additive fields are non-breaking; keep `schemaVersion` unless a consumer would
  misinterpret old data.
- Renaming/removing fields requires a schema bump and invalidation of cached
  artifacts (the header comparison handles this through `schemaVersion`).
- Keep raw detector metrics in `CheatChannel` even when score/zone are present;
  the UI relies on them for transparency.
