# AGENTS.md — internal/report/jsondata

Optional machine-readable sidecar (`--json`).

- Must encode `model.Match` exactly; do not add a second view model.
- Keep `SetEscapeHTML(false)`; JSON is not HTML.
- Schema changes are documented in `docs/report-schema.md` and require a
  `model.SchemaVersion` bump only for breaking consumers.
- Writes go through `store.WriteAtomic` in the caller.
