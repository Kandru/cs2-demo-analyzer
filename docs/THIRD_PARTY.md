# Third-party code and licenses

This project vendors one Go package and depends on several modules.

## Vendored: demo-anticheat

- Source: <https://github.com/TimAnthonyAlexander/demo-anticheat>
- Commit: `635e8c7a152f3026446463cbe8f0ba2e170ec818` (2026-05-18)
- Location: `src/internal/anticheat/analyzer`, `src/internal/anticheat/stats`
- License: **none declared at the pinned commit** (no LICENSE/COPYING file was
  present). Vendoring was done at the project owner's request. If you
  redistribute this project, confirm permission with the upstream author or
  replace the vendored package with your own implementation.

### Local modifications

| File | Change |
|---|---|
| `stats/collect_snap.go` | Removed a double degree→radian conversion that inflated snap velocity ~57×; record timestamped snap cues |
| `stats/collect_reaction.go` | Dropped same-tick TTD artifacts, publish `sub_100ms_count`; record timestamped TTD cues |
| `stats/collect_behavioral.go`, `..._prefov.go`, `..._backkill.go`, `..._attention.go` | Split the upstream behavioral collector by metric; record timestamped pre-FOV cues (scoring unchanged) |
| `stats/mod_ttd_sub100_floor.go` | Tightened sub-100 ms floor sample/pre-FOV gates |
| `stats/collect_recoil.go`, `stats/spray_patterns.go` | Removed debug stdout; record timestamped near-perfect spray cues; split the pattern data from the collector |
| `stats/types.go` | Added `EvidenceMoment` storage with caps (metadata only, never scored) |
| `stats/collect_evidence.go` | **New file**: wallbang/smoke/no-scope review cues |
| `analyzer/analyzer.go` | Added `AnalyzeReader` so compressed demos can be analyzed without a temp file |
| Upstream `term_*.go`, `html_reporter.go`, `reporters.go`, `report.tmpl.html` | Removed unused terminal/HTML reporters so the analyzer owns rendering and the binary stays slim |
| `stats/collector.go` | Trimmed upstream `collectors.go` to the interface and `BaseCollector`; removed the unused per-frame weapon-usage collector |
| `stats/report.go` | **New file**: exposes detector metrics for the report |
| `stats/detect_*.go`, `stats/mod_*.go`, `stats/registry.go`, `stats/run.go`, `stats/combine.go`, `stats/publish.go`, `stats/narrative.go`, `stats/evidence.go` | **New first-party structure**: one file per detection/modifier with an ordered registry. Formulas, weights and thresholds are unchanged from the upstream-equivalent pipeline; this is a structural split, not a retune. |
| `internal/anticheat/anticheat.go` | **New file**: adapter to `model.CheatReport` |

Scoring logic is otherwise unmodified. See `docs/anticheat.md` and
`docs/detections/README.md`.

## Go modules

| Module | Version | License | Use |
|---|---|---|---|
| `github.com/markus-wa/demoinfocs-golang/v5` | v5.3.0 | MIT | CS2 demo parsing |
| `github.com/fsnotify/fsnotify` | v1.10.1 | BSD-3-Clause | filesystem watching |
| `github.com/oklog/ulid/v2` | v2.1.1 | Apache-2.0 | equipment/grenade identifiers |
| `google.golang.org/protobuf` | v1.36.11 | BSD-3-Clause | transitive (demoinfocs) |
| `github.com/golang/snappy` | v1.0.0 | BSD-3-Clause | transitive (demoinfocs) |
| `github.com/golang/geo` | v0.0.0-20230421003525-6adc56603217 | Apache-2.0 | transitive (demoinfocs) |
| `golang.org/x/sys` | v0.50.0 | BSD-3-Clause | transitive (fsnotify) |
| other demoinfocs transitives | see `docs/dependencies.md` | MIT/BSD | parser internals |

See [dependencies.md](dependencies.md) for the compiled-module inventory,
reachability notes, the audit workflow, and the upstream proposals for removing
`pkg/errors`, `golang.org/x/exp`, and optionally `quickhull-go`.

Run `cd src && go list -m all` for the complete module graph. License texts are
available from the respective module caches/repositories.

## Vendored CSS: new.css

- Source: <https://github.com/xz/new.css>
- Commit: `375a937537159abd11c89a07137e53d2411b4487` (2026-11-10)
- Location: `src/internal/report/html/assets/vendor/new.css`
- License: MIT, copyright (c) 2020 Brendan Hersh
- License text: `src/internal/report/html/assets/vendor/new.css.LICENSE`

`new.css` is used only as a small framework/reset layer. Every visible colour
is overridden by the first-party `assets/theme.css` token source. The vendored
file is not edited.

## Design and brand assets

The dark CS 1.6-inspired UI is an original CSS implementation. No EA logos,
game fonts, map screenshots or other game assets are included.

The project logo and favicon were supplied by the repository owner in
`docs/assets/`. Optimised, small derivatives are embedded in generated reports:

- `assets/brand-mark.png`
- `assets/favicon-32.png`
- `assets/apple-touch-icon.png`

The source files are not modified. The report embeds only the optimised copies,
and an automated test enforces the inline asset budget.
