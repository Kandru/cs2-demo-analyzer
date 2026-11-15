# AGENTS.md — internal/config

Runtime configuration shared by analyze/watch.

- Defaults: recursive, 4 workers (analyze), 15 m timeout, 5 s interval, 2 s
  debounce, cheat+aim on, JSON off.
- `Validate` normalizes input/output paths, creates the output directory, and
  rejects empty/invalid values.
- `ManifestDir` is the output dir, else the first input directory, else CWD.
- Watch overrides workers to 1 in `app` (memory), not here.
- Do not read flags/env in this package; `app` owns CLI parsing.

See `docs/cli.md` for the user-facing meaning.
