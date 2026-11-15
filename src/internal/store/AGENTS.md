# AGENTS.md — internal/store

Manifest cache, output naming, atomic writes.

## Invariants

- Manifest lives at `<dir>/.cs2battlelog/manifest.json`; never write reports
  into that directory.
- Writes are `.tmp` + fsync + rename; readers must never see partial reports.
- Skip only when `ok` + same size/mtime + same tool/schema version + HTML exists.
  Error entries are skipped while unchanged (retry on change/`--force`).
- Output names are sanitized (`safeName`); default is
  `<demo-base>.battlelog.html` next to the demo.
- Shared-output basename collisions get a short source hash.
- Keep `schemaVersion` in sync with `model.SchemaVersion`.
- Manifest access is mutex-protected (`watch` workers).

## Tests

`store_test.go` covers default naming, collisions, skip invalidation, cached
failures, atomic writes. Add a case for any naming/cache change.
