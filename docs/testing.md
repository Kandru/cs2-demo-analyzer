# Testing

## Supported commands

```sh
make test          # go test ./... inside golang:1.27.4
make test-cover    # coverage summary
make vet
make fmt
make build
make tidy-check    # fail if go.mod/go.sum need tidying
make audit         # go mod verify + govulncheck
make check-install # validate install.sh syntax
```

Local Go 1.27.4+ can use `cd src && go test ./...`; CI/Docker is the supported
path and the one used by the Makefile. All commands run with the digest-pinned
toolchain described in [dependencies.md](dependencies.md).

## Test layers

| Layer | Location | Notes |
|---|---|---|
| Aim formulas | `internal/aim/metrics_test.go` | anchors, medians, no-damage degradation, display strings |
| Parser helpers | `internal/parser/helpers_test.go` | slugs, buy types, ADR/HS/accuracy finalization, stable IDs |
| Anti-cheat calibration | `internal/anticheat/stats/calibration_test.go` | sub-100 floor gates, snap-unit regression |
| Cheat detections | `internal/anticheat/stats/detect_*_test.go`, `registry_test.go` | one test per signal, registry contract |
| Detection docs | `internal/anticheat/stats/docs_sync_test.go` | every registered feature has a page |
| Aim metrics | `internal/aim/metrics_test.go`, `docs_sync_test.go` | anchors, medians, no-damage degradation, doc sync |
| HTML report | `internal/report/html/render_test.go`, `index_test.go` | template compiles, metadata header, core sections render, search data, names are escaped, index links |
| Report document | `internal/report/report_test.go` | header-first JSON, HTML/JSON header probes, round trip, freshness matrix, comment safety |
| Artifact cache | `internal/app/processor_test.go` | version/option/size invalidation, missing JSON, cached failures, collision-safe planning, index metadata from headers |
| Index/demo links | `internal/app/index_test.go` | same-directory and opt-in links, index entries |
| Demo filenames | `internal/demofile/demofile_test.go` | supported extensions, base names, sibling lookup |
| Sizes | `internal/units/units_test.go` | IEC byte formatting |
| Cache/store | `internal/store/store_test.go` | output naming, collisions, skip/invalidations, atomic writes |
| Self-update | `internal/update/update_test.go` | version compare, asset selection, checksums, archive extraction, atomic replace |
| Discovery | `internal/app/discover_test.go` | recursive/non-recursive filters, env path lists |
| Watcher | `internal/watch/watch_test.go` | settle-before-queue |

No test requires network access or a demo file.

## Behavior-preserving refactors

Two opt-in harnesses make structural refactors auditable:

```sh
# Full report + JSON baseline
./dist/cs2-demo-analyzer analyze --json --output /tmp/before stuff/

# Internal detector metrics baseline
CS2_DEMO_ANALYZER_GOLDEN_DEMO=stuff/match.dem \
CS2_DEMO_ANALYZER_GOLDEN_OUT=/tmp/before.metrics \
go test ./internal/anticheat/stats -run TestGoldenDemoMetrics
```

Diff the JSON sidecars (ignore the `header.generatedAt`/`analyzedAt`/`fileTime`
fields) and the metric dumps after the refactor. Additive JSON fields are
allowed; scores, channels, flags, weights and HTML should be unchanged unless
the change is explicitly a calibration or schema change.

## Demo-backed checks (manual/opt-in)

1. Put demos in `stuff/` (git-ignored).
2. `make build` (or `go run ./cmd/cs2-demo-analyzer`).
3. `./dist/cs2-demo-analyzer analyze --json stuff/`.
4. Verify:
   - round count/score equals the in-game scoreboard,
   - warmup appears before round 1,
   - scoreboard totals equal the sum of player rows,
   - a clean player is not flagged; flagged players have plausible raw
     snap/TTD/pre-FOV values and sufficient sample counts,
   - timestamped cheat cues land inside the round shown, and
     `demo_gototick <tick>` reaches the corresponding moment,
   - aim samples (`n=`) are non-trivial before trusting a rating,
   - HTML opens offline and filters work,
   - JSON sidecar starts with a `header` object and `schemaVersion: 1`; the HTML
     starts with the matching metadata comment.

Regression demos worth keeping locally:

- a normal 5v5 GOTV CS2 demo (2025+),
- a short/Wingman demo (small sample behavior),
- a demo with no `PlayerHurt` events (older CS2 format),
- a truncated/corrupt demo (partial report + cached failure).

## Adding a demo fixture

Do **not** commit demos. They are large and often personal. Keep them in
`stuff/` (ignored by git and Docker) and document the expected results in the
PR rather than checking the file in.

## Release checklist

- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go mod tidy -diff` is clean
- [ ] `govulncheck ./...` reports no vulnerabilities
- [ ] one clean and one suspicious demo inspected manually
- [ ] no false-positive flags on the clean reference demo
- [ ] `VERSION` bumped with the correct month
- [ ] workflow succeeds and assets include `SHA256SUMS`
- [ ] `sh -n install.sh` passes
- [ ] `cs2-demo-analyzer update --check` sees the new release and the expected archive name
