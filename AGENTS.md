# AGENTS.md — repository guide

This repo turns CS2 demos into a self-contained, dark HTML report with
aim stats and a vendored anti-cheat analysis.

## Read first

1. `docs/README.md` — documentation index and reading order.
2. `docs/architecture.md` — pipeline and package boundaries.
3. `docs/conventions.md` — global rules (errors, escaping, tests, commits).
4. The `AGENTS.md` in the folder you are editing.
5. `docs/THIRD_PARTY.md` before touching `src/internal/anticheat`.

## Ground rules

- Do not modify `VERSION` unless the task is a release.
- Never commit demos, generated reports, `stuff/`, `demos/`, `reports/`, or
  `.cs2-demo-analyzer/` (all ignored).
- Keep the HTML self-contained: no CDN, no external images/fonts/scripts.
- Treat demo content as untrusted. Render through `html/template`; never
  `template.HTML` for player strings.
- Match parser and anti-cheat detector are separate passes. Do not merge
  their event handlers into one parser state.
- Build/test with Docker (`make test`, `make build`). `go` is not assumed on the
  host, though `make build-local` exists.
- Update `docs/` and the relevant `AGENTS.md` when changing behavior, contracts,
  or extension points.

## Layout

```
Makefile, Dockerfile, docker-compose.yml, VERSION, .github/workflows/release.yaml
docs/                     global documentation
src/cmd/cs2-demo-analyzer/    CLI entrypoint
src/internal/app/         CLI + processor orchestration
src/internal/config/      runtime config
src/internal/model/       domain/JSON model
src/internal/parser/      demoinfocs → model
src/internal/aim/         aim metrics
src/internal/anticheat/   vendored detector + adapter
src/internal/analysis/    merge + summary
src/internal/report/html/ template + embedded CSS/JS
src/internal/store/       manifest + atomic output
src/internal/watch/       fsnotify/polling watcher
src/internal/version/     ldflags version
```

## Common tasks

| Task | Command |
|---|---|
| Help | `make` |
| Build | `make build` |
| Test | `make test` |
| Format/vet | `make fmt && make vet` |
| Release archives | `make release` |
| Run local binary | `go run ./cmd/cs2-demo-analyzer analyze stuff/demo.dem` (from `src/`) |

## Definition of done

- `go test ./...` passes.
- `go vet ./...` passes.
- New/changed behavior documented, including the matching
  `docs/detections/<id>.md` or `docs/aim/<id>.md` page.
- Docs-sync tests pass for every registered detection/modifier/metric.
- No unescaped user content reaches HTML.
- Clean demo does not gain a new false-positive flag; calibration tests pass.
- Partial/corrupt demo still produces a report or a cached, actionable error.
