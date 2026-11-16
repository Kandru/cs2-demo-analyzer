# AGENTS.md — internal/store

Manifest cache, output naming, atomic writes.

## Invariants

- Manifest lives at `<dir>/.cs2-demo-analyzer/manifest.json`; never write reports
  into that directory.
- Writes are `.tmp` + fsync + rename; readers must never see partial reports.
- Skip decisions live in `internal/app` and read report headers. The store only
  caches failures (`CachedFailure`) and records manifest metadata.
- Output names are chosen and reserved atomically with `PlanOutput`; the
  optional `occupied` callback lets the caller reject a candidate owned by a
  different demo even when no manifest mapping exists.
- Default is `<demo-base>.html` next to the demo. A demo whose base would
  produce `index.html` is disambiguated so it cannot overwrite the report index.
- Shared-output basename collisions get a short source hash.
- Successful entries cache `ReportMeta` (map, server, end score, played time)
  for index pages. Metadata is additive; older manifests remain readable.
- Keep `schemaVersion` in sync with `model.SchemaVersion`.
- Manifest access is mutex-protected (`watch` workers).

## Tests

`store_test.go` covers default naming, atomic planning/collisions, cached
failures, and atomic writes. Add a case for any naming or cache change.
