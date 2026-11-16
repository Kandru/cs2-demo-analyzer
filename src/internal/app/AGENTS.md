# AGENTS.md — internal/app

CLI parsing, discovery, and per-demo pipeline.

## Invariants

- `Processor.Process` order: stat → plan outputs → hash → parse → anti-cheat →
  analysis → build `report.Document` → render → atomic writes → manifest →
  optional index refresh.
- `Processor.Plan` reserves every discovered output path before parallel work
  in `analyze`; this keeps same-basename collisions deterministic, including
  for skipped demos.
- `analyze` and `watch` both skip when artifact headers match (HTML, plus JSON
  when `--json` is enabled). `--force` bypasses; unversioned `dev` builds never
  reuse output. Cached failures come from the manifest.
- Artifact freshness lives in `internal/report`; do not reimplement header
  parsing here. Skip and process must share one `headerFor` builder.
- Always write the HTML even if the detector fails.
- Per-demo timeout is applied inside `Process`.
- Failures are isolated: analyze continues and exits 1 if any failed; watch logs
  and continues.
- Outputs default next to the demo; `--output` is a single shared directory.
- Demo download links are automatic only when report and demo share a directory;
  `--demo-download` opts in to relative links across directories.
- `index.html` is written when `--index` is set or when reports sit next to their
  demos. Index entries come from the directory scan plus the store mapping and
  match metadata read from each report header (map, end score, server, played
  time); the manifest is a legacy fallback. Cards are grouped by server and
  sorted newest first; metadata-less reports fall back to the report basename
  and file mtime. Analyze refreshes planned directories once after all workers
  finish so the index reflects every successful output; watch refreshes after
  each successful analysis.
- When an index is enabled, the detail report header links back to `index.html`.
  The report renderer receives this through `View.IndexHref`.
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
