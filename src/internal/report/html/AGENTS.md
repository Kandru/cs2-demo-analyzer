# AGENTS.md — internal/report/html

Self-contained Battlelog-style report renderer.

## Files

- `view.go` — presentation view model and builders.
- `renderer.go` — embedded template/assets, `Render`.
- `templates/report.gohtml` — one HTML document + event partial.
- `assets/style.css`, `assets/app.js` — inlined at render time.

## Invariants

- No external URLs for assets; no CDN fonts/images/scripts.
- Player/clan/demo strings are rendered via template actions only. Never build
  HTML from `fmt.Sprintf` and never use `template.HTML` for demo content.
- The page must be readable with JS disabled; JS only adds nav highlighting,
  filters, and expand/collapse.
- Every cheat card carries the “statistical signal, not proof” disclaimer.
- Preserve anchors: `#overview`, `#scoreboard`, `#rounds`, `#players`,
  `#player-<slug>`, `#round-<tick>`, `#warmup`.
- Keep the dark palette tokens; no light mode.
- Score/round tables need horizontal scroll, not tiny text.

## Extending

1. Build data in `view.go`; avoid logic in templates.
2. Add an event icon to the sprite and `iconByType` if needed.
3. Style with existing tokens; support ≤640 px.
4. Add a render test assertion in `render_test.go` for new sections or contracts.
5. Update `docs/battlelog-ui.md` and `docs/report-schema.md`.
