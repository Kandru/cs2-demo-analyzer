# AGENTS.md - internal/report/html

Self-contained, flat report renderer with a shared CS 1.6-inspired dark theme.

## Files

- `view.go` - presentation view model, continuous timeline, builders.
- `index.go` - output-directory index view and renderer.
- `renderer.go` - embedded templates/assets, CSS assembly, `Render`.
- `templates/components.gohtml` - shared topbar, footer, search index, icons,
  timeline rows, cue lists, channel lists and adjustment panels. Parsed into
  both pages.
- `templates/report.gohtml` - one HTML document.
- `templates/index.gohtml` - self-contained report index.
- `assets/vendor/new.css` - vendored MIT framework/reset. Do not edit.
- `assets/theme.css` - the only source of raw colours and theme tokens.
- `assets/style.css` - component layout using theme tokens only.
- `assets/app.js` - progressive enhancement.
- `assets/brand-mark.png`, `logo.jpg`, `favicon-32.png`,
  `apple-touch-icon.png` - optimised inlined brand assets.

## Invariants

- No external URLs for assets; no CDN fonts/images/scripts. Brand images are
  embedded as data URIs from `assets/`.
- `assets/theme.css` is the single colour source. Never add a raw `#hex`,
  `rgb()`, `rgba()` or `hsl()` to `style.css`, templates or `app.js`.
- Read `docs/theme.md` before changing colours and `docs/ui.md` before
  changing typography, badges, headings, search or layout components.
- One badge primitive: `.badge` plus a semantic modifier. Do not add a new
  badge class with its own font size or padding.
- The header is a single row: brand, nav, search. No map/mode tags and no
  "All reports" link.
- Search is a server-rendered `#search-index` plus a JS combobox dropdown.
  Never filter by hiding page content and never move the header while typing.
- The footer is pinned to the bottom, compact and vertically centred. External
  links use `target="_blank" rel="noopener noreferrer"`.
- Section headers and a timeline filter bar stick below `--header-h`.
- Player cards are one per row at every width.
- Score adjustments are always visible; never put them behind `<details>`.
- Match score colours follow win/loss; draw is neutral and is not announced in
  the hero.
- CS2 CT blue and T yellow are identity colours only. Gold is the CS 1.6 accent
  and interactive colour. Round alternation uses structural gold/steel.
- Keep the dark palette, hard corners, 2 px bevel language and dense numeric
  tables. No light mode, gradients, glow or decorative perspective.
- Player/clan/demo strings are rendered via template actions only. Never build
  HTML with `fmt.Sprintf` and never use `template.HTML` for demo content.
- The page must be readable with JS disabled; JS only adds nav highlighting,
  search, filters and tick copying.
- Every cheat card carries the "statistical signal, not proof" disclaimer.
  Timestamped cues are review aids, not proof of cheating.
- Render cue labels, ranks and values from `CheatMoment.Label/Rank/ValueText`.
  Do not add a per-kind switch; unknown kinds fall back to the raw kind.
- Preserve anchors: `#overview`, `#scoreboard`, `#cheaters`, `#rounds`,
  `#players`, `#player-<slug>`, `#cheater-<slug>`, `#round-<tick>`, `#warmup`.
- Score/round tables need horizontal scroll, not tiny text.
- The layout must work from ~320 px to 4K without a fixed content-width cap.
- The index and report share `assets/style.css`, `theme.css`,
  `templates/components.gohtml` and the same footer.
- Index cards show map + end score, server + played time, and optional demo
  download; group by server and keep newest-first ordering.
- The report brand links to `index.html` when `View.IndexHref` is set; fall back
  to `#overview` otherwise.

## Continuous timeline contract

- `View.Timeline` is ordered: pre-match, warmup, round, intermission, next
  round, post-match.
- Every row has an absolute demo timestamp.
- Round headers are the round-start markers. Do not emit a duplicate
  `round_start` row.
- Every round gets a round-end marker, synthesised when the parser did not emit
  one.
- Detailed round rows include flash blinds, grenade damage, bomb pickup/drop,
  hostage pickup/drop/rescue/kill and per-player purchases when the data model
  provides them.
- Lifecycle rows (connect, disconnect, team change, team swap) live in
  intermission, warmup or pre-match blocks, not in player aggregates.
- Winner side is a labelled chip; alternating round markers must not imply a
  team colour.
- `markFlaggedRounds` must mark both the overview strip and timeline sections.

## Extending

1. Build data in `view.go`; avoid logic in templates.
2. Reuse a component in `templates/components.gohtml`; add one only when it is
   shared or the markup is non-trivial.
3. Add an event icon to the sprite and `iconByType` if needed.
4. Style with existing tokens; support <=480 px and very wide screens.
5. Add a render test assertion in `render_test.go` for new sections or
   contracts. Theme tests enforce raw-colour, font-size and docs sync.
6. Update `docs/report-ui.md`, `docs/ui.md`, `docs/theme.md` and
   `docs/report-schema.md`.
