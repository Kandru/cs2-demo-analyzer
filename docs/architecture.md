# Architecture

## High-level flow

```
filesystem
   │  (analyze: explicit inputs; watch: fsnotify + rescan + settle)
   ▼
internal/app.Processor
   │
   ├─ internal/parser ── demoinfocs v5 ──► model.Match (timeline, players, rounds)
   │       └─ internal/aim tracker       ──► model.AimStats / weapon shots
   │
   ├─ internal/anticheat ── vendored demo-anticheat ──► CheatReport per SteamID
   │
   ├─ internal/analysis   merge cheat data + recompute summary/team totals
   │
   ├─ internal/report/html (template + embedded CSS/JS) ──► self-contained HTML
   └─ internal/report/jsondata (optional) ──► JSON sidecar
   │
   ▼
internal/store  atomic writes + .cs2-demo-analyzer/manifest.json cache
internal/app    optional index.html + demo download links per report directory
```

This is deliberately **two parse passes**: the match parser owns the event
timeline in one pass, and the vendored anti-cheat analyzer runs separately so
its event handlers and global state cannot interfere with the report. A 65 MB
demo takes roughly 9–11 s on a modern CPU (both passes included).

## Package boundaries

| Package | Responsibility | Must not import |
|---|---|---|
| `internal/model` | Pure domain structs, JSON tags, no behavior | parser, report, demoinfocs |
| `internal/config` | Validated runtime flags | parser, report |
| `internal/parser` | demoinfocs → `model.Match`; rounds, events, economy, player stats | report |
| `internal/aim` | Leetify-style aim tracker (one pass, fed by parser); one `metric_<id>.go` spec per tile | parser (receives events/state) |
| `internal/anticheat` | Adapter around vendored detector | parser, report |
| `internal/anticheat/stats` | Collectors + first-party detection registry (`detect_*`, `mod_*`) | app/report |
| `internal/anticheat/stats`, `.../analyzer` | Vendored upstream code; only integration patches allowed | app/report |
| `internal/analysis` | Cheat merge, summary/team totals, validation | parser internals |
| `internal/report/html` | View model + template + assets | store |
| `internal/report/jsondata` | Placeholder for future format-specific transforms | report |
| `internal/store` | Manifest, skip logic, output naming, atomic I/O | parser |
| `internal/watch` | Discovery loop, settle window, worker queue | report |
| `internal/app` | CLI, wiring, processor, output index | report internals |
| `internal/demofile` | Demo extension checks, report base names, sibling lookup | parser, report, app |
| `internal/units` | Human-readable byte sizes shared by logs and reports | anything |
| `internal/version` | Build-time version vars | anything |

## Round model

`Round` is the unit of the timeline:

- `StartScore`/`EndScore` are indexed by the **stable team order** captured at
  match start (`teams[0]` = T side at start, `teams[1]` = CT side at start).
  Halftime does not reorder teams.
- `Number == 0` + `Pregame == true` is a knife round/restart. It is shown but
  excluded from aggregates and round counters.
- `Overtime` is set for rounds beyond the 24th regulation round (MR12).
- Every `Event` has a global sequence and a round-relative offset. Events with
  the same tick keep insertion order.

## Interaction & state

- `parser.state` owns all match-building state; it is single-threaded inside
  one parsing goroutine.
- `aim.Tracker` is called explicitly from the parser (`OnFrameState`,
  `OnWeaponFire`, `OnPlayerHurt`, `OnKill`, `OnRoundStart`), so its clock is
  deterministic and testable.
- `watch.Runner` uses one queue and N workers. The manifest is mutex-protected;
  report writes are atomic and independent.
- CLI cancellation flows from `signal.NotifyContext` → processor timeout →
  parse loop (`ctx.Err()` checked every frame).

## Extension points

1. **New event type**: add a constant/struct field in `model.Event`, populate it
   in `parser` handlers, render it in `report/html/view.go` + template.
2. **New match stat**: aggregate in `parser.applyRoundToPlayers` or
   `finalizePlayerStats`; expose in `model.PlayerStats`; show in player cards.
3. **New aim metric**: add `src/internal/aim/metric_<id>.go` with a
   `MetricSpec` (anchors, weight, sample gate, formatter), register it in
   `metric_specs.go`, add `docs/aim/<id>.md` from the template, and extend
   `metrics_test.go`. The docs-sync test enforces the page.
4. **New cheat channel**: add `src/internal/anticheat/stats/detect_<id>.go`
   with a `Detection` spec (measurement, ramp, weight, mode, narrative,
   optional evidence cue), register it in `registry.go`, add
   `docs/detections/<id>.md`, and add `detect_<id>_test.go`. Keep scoring
   behavior changes separate from structural changes. See
   `docs/detections/README.md`.
5. **New cheat modifier**: add `mod_<id>.go` with `Apply`, `Publish` and
   `Adjustments`, then register it at the correct stage (boosts → discount →
   stacking/co-occurrence → floor → overrides).
6. **New output format**: implement a renderer in `internal/report/<format>` and
   call it from `app.Processor`; keep the `model` contract unchanged.
7. **New watcher source** (S3/SFTP): implement the `watch.Processor` boundary
   and feed it paths; do not put I/O into the parser.

## Detection architecture

The vendored detector supplies raw measurements; the first-party registry in
`internal/anticheat/stats` turns them into scored channels. One file per
detection/modifier and one matching page keeps a feature editable in isolation:

```
collect_*.go ──► detect_<id>.go ──► combine ──► mod_<id>.go ──► publish
                     │
                     └── docs/detections/<id>.md
```

The registry order is a contract; formulas and thresholds live in the feature
file, never in the pipeline. `internal/aim` follows the same pattern with
`MetricSpec` and `docs/aim/`.

## Failure model

- `parse` returns a partial `Match` plus a warning when useful data was already
  read; a nil match is only returned when nothing could be parsed.
- Anti-cheat failure never blocks the match; it becomes an “unavailable”
  card with the error and keeps metrics collected by the match parser.
- Watch errors are cached per file (size+mtime) so a permanently corrupt demo
  does not spin; `--force` or a file change retries.
- A single bad demo never terminates `watch`; `analyze` continues but exits 1 if
  any demo failed.
