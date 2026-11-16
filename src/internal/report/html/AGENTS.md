# AGENTS.md — internal/report/html

Self-contained, flat report renderer.

## Files

- `view.go` — presentation view model and builders.
- `index.go` — output-directory index view and renderer.
- `renderer.go` — embedded template/assets, `Render`, `RenderIndex`.
- `templates/report.gohtml` — one HTML document + event/moment partials.
- `templates/index.gohtml` — self-contained report index.
- `assets/style.css`, `assets/app.js` — inlined at render time.

## Invariants

- No external URLs for assets; no CDN fonts/images/scripts.
- Player/clan/demo strings are rendered via template actions only. Never build
  HTML from `fmt.Sprintf` and never use `template.HTML` for demo content.
- The page must be readable with JS disabled; JS only adds nav highlighting,
  filters, expand/collapse, and tick copying.
- Every cheat card carries the “statistical signal, not proof” disclaimer.
  Timestamped cues are review aids, not proof of cheating.
- Render cue labels, ranks and values from `CheatMoment.Label/Rank/ValueText`.
  Do not add a per-kind switch; unknown kinds fall back to the raw kind.
- Preserve anchors: `#overview`, `#scoreboard`, `#cheaters`, `#rounds`,
  `#players`, `#player-<slug>`, `#cheater-<slug>`, `#round-<tick>`, `#warmup`.
- Keep the dark palette tokens; no light mode.
- Keep the visual language flat (1 px borders, hard corners, no decorative
  perspective/glow), and preserve the orange accent.
- Score/round tables need horizontal scroll, not tiny text.
- The layout must work from ~320 px to 4K without a fixed content-width cap.

## Extending

1. Build data in `view.go`; avoid logic in templates.
2. Add an event icon to the sprite and `iconByType` if needed.
3. Style with existing tokens; support ≤640 px and very wide screens.
4. Add a render test assertion in `render_test.go` for new sections or
   contracts.
5. Update `docs/report-ui.md` and `docs/report-schema.md`.
