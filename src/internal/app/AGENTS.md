# AGENTS.md — internal/app

CLI parsing, discovery, and per-demo pipeline.

## Invariants

- `Processor.Process` order: hash → parse → anti-cheat → analysis → render →
  atomic write → manifest → optional index refresh.
- Always write the HTML even if the detector fails.
- Per-demo timeout is applied inside `Process`.
- `analyze` never uses the cache to skip; `watch` does via `Processor.Skip`.
- Failures are isolated: analyze continues and exits 1 if any failed; watch logs
  and continues.
- Outputs default next to the demo; `--output` is a single shared directory.
- Demo download links are automatic only when report and demo share a directory;
  `--demo-download` opts in to relative links across directories.
- `index.html` is written when `--index` is set or when reports sit next to their
  demos. Index entries come from the directory scan plus the store mapping.
- `update` is the only command that writes the binary. `--update` on
  analyze/watch is best-effort: check/apply failures log a warning and never
  block processing. Normal runs make no network calls.
- `update --check` must not download the release archive or modify the binary.
- Do not log player names as structured fields.

## Flags

Keep `printUsage` and `docs/cli.md` in sync. Env defaults: `CS2_DEMO_ANALYZER_*`.
README examples should stay runnable.

## Tests

`discover_test.go` covers file filters and env path lists. Add tests for new CLI
semantics where possible without parsing demos.
