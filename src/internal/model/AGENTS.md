# AGENTS.md — internal/model

Pure domain and JSON contract. No behavior, no parser imports.

## Invariants

- `SchemaVersion` must match `store`'s schema constant. Bump only for breaking
  consumer changes; additive fields do not need a bump.
- Scores are arrays indexed by stable team order (see parser).
- Optional numeric fields that may be absent are `omitempty` and/or documented.
- Do not add demoinfocs types here.
- Keep JSON tags stable; the HTML/JSON sidecar depend on them.
- Methods are lookup helpers only (`PlayerByID`, `TeamByID`, `RoundByNumber`);
  no I/O, no computation.
- `CheatMoment.Label/Rank/ValueText` are additive detector metadata. Reporters
  must use them instead of switching on `Kind`.

When changing the contract, update `docs/report-schema.md` and the store schema.
