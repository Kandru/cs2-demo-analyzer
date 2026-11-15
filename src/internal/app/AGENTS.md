# AGENTS.md — internal/app

CLI parsing, discovery, and per-demo pipeline.

## Invariants

- `Processor.Process` order: hash → parse → anti-cheat → analysis → render →
  atomic write → manifest.
- Always write the HTML even if the detector fails.
- Per-demo timeout is applied inside `Process`.
- `analyze` never uses the cache to skip; `watch` does via `Processor.Skip`.
- Failures are isolated: analyze continues and exits 1 if any failed; watch logs
  and continues.
- Outputs default next to the demo; `--output` is a single shared directory.
- Do not log player names as structured fields.

## Flags

Keep `printUsage` and `docs/cli.md` in sync. Env defaults: `CS2BATTLELOG_*`.
README examples should stay runnable.

## Tests

`discover_test.go` covers file filters and env path lists. Add tests for new CLI
semantics where possible without parsing demos.
