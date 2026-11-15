# Testing

## Supported commands

```sh
make test          # go test ./... inside golang:1.27
make test-cover    # coverage summary
make vet
make fmt
make build
```

Local Go 1.27+ can use `cd src && go test ./...`; CI/Docker is the supported
path and the one used by the Makefile.

## Test layers

| Layer | Location | Notes |
|---|---|---|
| Aim formulas | `internal/aim/metrics_test.go` | anchors, medians, no-damage degradation, display strings |
| Parser helpers | `internal/parser/helpers_test.go` | slugs, buy types, ADR/HS/accuracy finalization, stable IDs |
| Anti-cheat calibration | `internal/anticheat/stats/calibration_test.go` | sub-100 floor gates, snap-unit regression |
| HTML report | `internal/report/html/render_test.go` | template compiles, core sections render, names are escaped |
| Cache/store | `internal/store/store_test.go` | output naming, collisions, skip/invalidations, atomic writes |
| Discovery | `internal/app/discover_test.go` | recursive/non-recursive filters, env path lists |
| Watcher | `internal/watch/watch_test.go` | settle-before-queue, extensions |

No test requires network access or a demo file.

## Demo-backed checks (manual/opt-in)

1. Put demos in `stuff/` (git-ignored).
2. `make build` (or `go run ./cmd/cs2-battlelog`).
3. `./dist/cs2-battlelog analyze --json stuff/`.
4. Verify:
   - round count/score equals the in-game scoreboard,
   - warmup appears before round 1,
   - scoreboard totals equal the sum of player rows,
   - a clean player is not flagged; flagged players have plausible raw
     snap/TTD/pre-FOV values and sufficient sample counts,
   - aim samples (`n=`) are non-trivial before trusting a rating,
   - HTML opens offline and filters work,
   - JSON sidecar has `schemaVersion: 1`.

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
- [ ] one clean and one suspicious demo inspected manually
- [ ] no false-positive flags on the clean reference demo
- [ ] `VERSION` bumped with the correct month
- [ ] workflow succeeds and assets include `SHA256SUMS`
