# AGENTS.md — internal/version

Build-time metadata only.

- Vars `Version`, `Commit`, `Date` are injected via ldflags by the Makefile,
  Dockerfile, and release workflow.
- Defaults are `dev`/empty so `go run` works.
- `Short()` is the manifest/UI version; `Info()` is the CLI text.
- Do not bump versions here; edit root `VERSION` (`make version-bump`).
