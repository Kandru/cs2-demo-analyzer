# Global conventions

These rules apply to the whole repository.

## Go

- Format with `gofmt`; keep packages small and single-purpose.
- Errors: wrap with `%w` and context (`fmt.Errorf("parse round: %w", err)`).
  Return actionable messages; never `panic` for input problems.
- Logging: `log/slog` only, to stderr. Include `demo`, `error`, and useful
  counts. Never log player names as structured fields (escape/render only).
- No global mutable state except build-time version vars and the parsed
  template. Package-level templates are immutable after `init`.
- Use `context.Context` for anything that can block (parse loop, watcher,
  processor). Check cancellation at least once per frame loop.
- Prefer the standard library. New dependencies require a documented reason in
  the PR and a line in `docs/THIRD_PARTY.md`.
- Exported identifiers need doc comments. Unexported helpers should be obvious.

## Security / untrusted demo data

- Player names, clan tags, server names, and demo strings are **untrusted**.
  Render them only through `html/template`; never build HTML with `fmt.Sprintf`.
- The current report embeds no remote content and no chat. If chat is added,
  treat links as text, not anchors, and keep the CSP-simple inline strategy.
- Output filenames are sanitized in `store.safeName`; never use raw demo names
  as paths.
- Path traversal, symlink loops, and hidden directories are skipped by the
  watcher. Do not add recursive following of symlinked directories.
- The self-contained HTML must remain functional offline; no CDN fonts, images,
  or scripts.

## Errors and partial data

- Partial reports are better than nothing: set `Match.Partial`, `ParseError`,
  and append human-readable `Warnings`.
- Never emit a report with zero players; return `ErrNoPlayers` instead.
- Cheat-detection failure is non-fatal; record it on each `CheatReport`.

## Feature units

- One cheat detection per `detect_<id>.go`, one modifier per `mod_<id>.go`,
  and one aim metric per `metric_<id>.go`. Each needs a matching page:
  `docs/detections/<id>.md` or `docs/aim/<id>.md`.
- Metadata belongs in the feature spec, not in the pipeline. Do not add
  per-feature switches to `run.go`, `publish.go`, the report builder or the
  template.
- The registry order is part of the behavior contract; changing it requires a
  test update and a docs note.
- The docs-sync tests fail when a registered feature has no page or an orphan
  page exists. Run them before committing.

## Tests

- Every feature folder needs tests for pure logic and at least one integration
  boundary.
- Tests must not require network access or checked-in demos. Use `t.TempDir()`
  and synthetic models; demo-backed tests are opt-in via an env var.
- Run `go test ./...` (or `make test`). The Docker path is the supported one.

## Commits / releases

- Conventional-ish subjects: `feat:`, `fix:`, `docs:`, `chore:`, `test:`.
- Bump `VERSION` in its own commit (`chore(release): 26.11.2`) so the release
  workflow only fires on that change.
- Do not commit demos, generated reports, or `.cs2-demo-analyzer/`.

## UI

- Dark only, no light theme. Use the CSS tokens in
  `report/html/assets/style.css`.
- Works with JS disabled: every section and every event must render.
- Minimum 4.5:1 contrast for body text; do not rely on color alone (badges and
  text accompany team/flag colors).
- Use semantic headings in order (`h1` → `h2` → `h3` → `h4`).
