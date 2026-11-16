# AGENTS.md — internal/store

Manifest cache, output naming, atomic writes.

## Invariants

- Manifest lives at `<dir>/.cs2-demo-analyzer/manifest.json`; never write reports
  into that directory.
- Writes are `.tmp` + fsync + rename; readers must never see partial reports.
- Skip only when `ok` + same size/mtime + same tool/schema version + same report
  options + HTML exists. Error entries are skipped while unchanged (retry on
  change/`--force`).
- Output names are sanitized (`safeName`); default is
  `<demo-base>.html` next to the demo.
- Shared-output basename collisions get a short source hash.
- Successful entries cache `ReportMeta` (map, server, end score, played time)
  for index pages. Metadata is additive; older manifests remain readable.
- Keep `schemaVersion` in sync with `model.SchemaVersion`.
- Manifest access is mutex-protected (`watch` workers).

## Tests

`store_test.go` covers default naming, collisions, skip invalidation, cached
failures, atomic writes. Add a case for any naming/cache change.
