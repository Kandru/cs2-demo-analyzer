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
| `stats/snap_collectors.go` | Removed a double degree→radian conversion that inflated snap velocity ~57× |
| `stats/reaction_collectors.go` | Dropped same-tick TTD artifacts, publish `sub_100ms_count` |
| `stats/cheatscore_overrides.go` | Tightened sub-100 ms floor sample/pre-FOV gates |
| `stats/recoil_collectors.go` | Removed debug stdout |
| `analyzer/analyzer.go` | Added `AnalyzeReader` so compressed demos can be analyzed without a temp file |
| `stats/term_*.go`, `stats/html_reporter.go`, `stats/reporters.go`, `stats/report.tmpl.html` | Removed unused upstream terminal/HTML reporters so the battlelog owns rendering and the binary stays slim; `intMetric` kept in `metrics_util.go` |
| `stats/battlelog.go` | **New file**: exposes detector metrics for the report |
| `internal/anticheat/anticheat.go` | **New file**: adapter to `model.CheatReport` |

Scoring logic is otherwise unmodified. See `docs/anticheat.md`.

## Go modules

| Module | License | Use |
|---|---|---|
| `github.com/markus-wa/demoinfocs-golang/v5` | MIT | CS2 demo parsing |
| `github.com/fsnotify/fsnotify` | BSD-3-Clause | filesystem watching |
| `google.golang.org/protobuf` | BSD-3-Clause | transitive (demoinfocs) |
| `github.com/golang/snappy`, `github.com/golang/geo`, etc. | BSD-3-Clause | transitive (demoinfocs) |

Run `cd src && go list -m all` for the complete module graph. License texts are
available from the respective module caches/repositories.

## Design

The Battlelog-inspired look is an original CSS implementation. No EA logos,
fonts, map screenshots, or other assets are included.
