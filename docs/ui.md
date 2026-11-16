# UI and component guideline

This page is the single component contract for the generated HTML report and
the report index. It sits between the colour contract in
[theme.md](theme.md) and the implementation in
`src/internal/report/html/assets/style.css`.

If this document, a template and the stylesheet disagree, fix the stylesheet
and this document in the same change. Do not add a one-off visual rule to a
template.

## Authority order

1. `assets/theme.css` - the only file with raw colour values and the only
   source for design tokens.
2. `docs/ui.md` - component names, markup contracts and usage rules.
3. `assets/style.css` - component layout using `var(--token)` only.
4. `templates/*.gohtml`, `view.go`, `index.go`, `app.js` - structure and data.
5. `assets/vendor/new.css` - vendored reset. Never edit.

## Non-negotiables

- Dark mode only.
- No raw `#hex`, `rgb()`, `rgba()`, `hsl()` or named colours outside
  `theme.css`.
- No hardcoded `font-size` values in `style.css`; use `var(--fs-*)`.
- No new badge, heading, table or panel styling when a shared component exists.
- Team colours are identity only. Status uses `--danger`, `--warning`,
  `--success`, `--info`, `--neutral`.
- Gold `--accent` is the interactive/accent colour, never a team or status.
- Status colour is always paired with text.
- Demo content is untrusted: templates escape it; JS uses `textContent`, never
  `innerHTML`.
- Hard corners, bevel borders, dense numeric tables. No gradients, glow, blur
  or decorative perspective.

## Typography

| Role | Token | Notes |
|---|---|---|
| Page title | `--fs-3xl` | `h1`, uppercase |
| Section title | `--fs-2xl` | `h2`, uppercase, gold rail |
| Card title | `--fs-lg` | `h3` |
| Panel/eyebrow label | `--fs-xs` | `h4`, uppercase, muted |
| Body | `--fs-base` | 1rem |
| Table/metadata | `--fs-sm` | 0.9375rem |
| Badge/small label | `--fs-xs` | 0.8125rem |
| Fine print | `--fs-2xs` | 0.75rem |

Body line height is `--lh-base`; headings use `--lh-tight`. Prose blocks use
`max-width: var(--measure)`. Never set a component-specific heading size; use
the heading level.

The root size is fluid:

```css
html { font-size: clamp(1rem, 0.42vw + 0.625rem, 1.5rem); }
```

Targets: 16 px at phone widths, roughly 18 px at 1920, 21 px at 2560 and 24 px
at 4K. Do not reintroduce fixed `html` font sizes in media queries.

## Layout

- `body` is a column flex box with `min-height: 100dvh`.
- `main` grows; the footer stays at the bottom of short pages.
- `.wrap` owns the centered content width (`--content-max`).
- `.section` separates major sections with `--section-gap`.
- `.section-head` is sticky below `--header-h`; it owns the section title and
  optional note. Do not put interactive controls in it.
- A filter toolbar directly under a sticky section head also sticks at
  `calc(var(--header-h) + var(--section-head-h))`.
- Every in-page anchor target uses
  `scroll-margin-top: calc(var(--header-h) + var(--space-4))`.
- `overflow-x: clip` is used on `body` so sticky positioning keeps working.

## Components

### Panel and card

```html
<section class="cs-panel">
  <div class="panel-head"><h3 class="panel-title">Title</h3>...</div>
  ...
</section>
```

Use `.cs-panel` for page-level panels and `.subpanel` for panels nested inside
a card. Both use the shared bevel recipe.

### Section

```html
<section id="overview" class="section">
  <div class="section-head">
    <h2>Overview</h2>
    <span class="section-note">Short note</span>
  </div>
  ...
</section>
```

`id` is required for navigation. Keep the stable anchors documented in
[report-ui.md](report-ui.md).

### Badge

```html
<span class="badge">Neutral</span>
<span class="badge badge--accent">Feature</span>
<span class="badge badge--danger">Flagged</span>
<span class="badge badge--warning">Review</span>
<span class="badge badge--success">Clean</span>
<span class="badge badge--info">Info</span>
<span class="badge badge--neutral">Unknown</span>
<span class="badge badge--ct">CT</span>
<span class="badge badge--t">T</span>
```

Rules:

- One size (`--badge-fs`) and one padding (`--badge-pad`) for every badge.
- Use `.badge-row` when several badges share a line; it wraps by default.
- Add `.badge` after the descriptive text in timeline rows.
- Never create a new badge class with its own font size or padding.
- `verdict`, `grade`, `zone`, `result-chip`, `team-tag`, `pill`, `aim-badge`,
  `cue-tag` and adjustment chips must reuse the badge tokens.

### Stat tile

```html
<div class="stat-tile">
  <span class="stat-label">Label</span>
  <span class="stat-value">Value</span>
  <span class="stat-sub">Detail</span>
</div>
```

Use `.stat-grid` for the overview. A linked tile adds `stat-link`; a warning
state adds `is-alert`.

### Table

- `.score-table` for scoreboards, `.mini-table` for compact data.
- Always wrap in `.table-scroll`.
- Numeric cells use `--font-mono`, tabular numbers and right alignment.
- Header labels are uppercase and muted.
- Provide a `<caption class="sr-only">`.
- Never shrink text to fit; scroll horizontally instead.

### Timeline

- `.timeline` contains `.tl-section` blocks.
- A round header is `.tl-head` with `.tl-round-badge`, `.tl-head-title`,
  `.tl-head-result` and `.tl-reason`.
- Rows are `.tl-row` with `.tl-time`, `.tl-node` and `.tl-text`.
- Badges follow the text inside `.tl-text`.
- Do not use `overflow: hidden` on `.tl-section`; it clips result chips.
- Use `--tl-time-w` and `--tl-node-w` for the timeline rail so rows and the
  vertical line stay aligned.
- Round result chips use team identity colours; round alternation uses
  `--round-a` / `--round-b`.
- Purchases render in `.tl-buys` with `.buy-team`, `.buy-list` and badges.

### Watchlist card

```html
<article class="cheater-card cs-panel flagged">
  <header class="cheater-head">...</header>
  <div class="meter">...</div>
  <div class="cheater-body">
    <div><h4>Signals</h4>...<div class="adjustment-panel">...</div></div>
    <div><h4>When to look</h4>...</div>
  </div>
  <footer class="cheater-foot"><a href="#player-x">Open full analysis</a></footer>
</article>
```

- The verdict block must never be clipped; keep the header grid at
  `minmax(0, 1fr) auto`.
- Score adjustments are always visible when present. Never hide them in a
  `<details>` element.
- “Open full analysis” lives in the footer, bottom-left.
- The statistical disclaimer stays visible.

### Player card

- One card per row at every viewport width. Do not add a multi-column player
  grid.
- Header badges: team, side, Steam profile, rounds, bot, aim, verdict.
- Aim and channel samples use readable nouns (shots, hits, engagements,
  samples), never a bare `n=`.
- Cheat analysis is a bulleted signal list, not prose. One compact disclaimer
  line remains.
- Score adjustments are always visible when present.

### Search dropdown

- The input is a `role="combobox"` with `aria-expanded`, `aria-controls` and
  `aria-autocomplete="list"`.
- Results come from the server-rendered `#search-index`, never from JS-built
  strings with HTML.
- The popup is absolutely positioned so typing never moves the header.
- Keyboard support: `/` focus, `ArrowUp/Down`, `Home/End`, `Enter`, `Escape`.
- Without JS the search input stays hidden and navigation still works.

### Footer

- Shared by index and report.
- Always at the bottom of the viewport on short pages.
- Height is `--footer-h` or slightly more when wrapping; never taller than the
  header on desktop.
- Content is vertically centered.
- External links use `target="_blank" rel="noopener noreferrer"`.

### Index card

```html
<article class="index-card cs-panel">
  <a class="index-report" href="...">...</a>
  <a class="index-demo" href="..." download="...">Download demo</a>
</article>
```

- Cards are equal height in the grid.
- The download link is optional and bottom-aligned with `margin-top: auto`.
- Never render a “demo not linked” placeholder.
- Do not add an “Open report” line; the whole card links to the report.

## Colour usage matrix

| Meaning | Tokens |
|---|---|
| Page/panel background | `--bg-*`, `--panel*` |
| Body text | `--text` |
| Secondary text | `--muted`, `--dim` |
| Links and focus | `--accent*`, `--focus*` |
| CT identity | `--team-ct-*` |
| T identity | `--team-t-*` |
| Success/win | `--success*` |
| Warning/review | `--warning*` |
| Danger/flag/loss | `--danger*` |
| Neutral/unavailable | `--neutral*` |
| Round alternation | `--round-a*`, `--round-b*` |

Team colours never mean success or danger. Status colours never identify a
team.

## Responsive contract

| Width | Required behaviour |
|---|---|
| 320-480 | one column, three round cells, no horizontal page overflow |
| <= 720 | compact header, four round cells, wrapped controls |
| <= 1024 | stacked scoreboards, six round cells |
| 1920 | readable desktop layout, eight round cells |
| 2560 | larger fluid type, same layout |
| 3840 | capped content width, larger type, no tiny text |

Test 200% browser zoom, JS disabled and print preview after any layout change.

## Adding or changing a component

1. Reuse an existing component and token first.
2. Add a token in `theme.css` and document it in `theme.md` if a new visual
   value is genuinely needed.
3. Add component CSS in `style.css` using `var(--token)` only.
4. Add reusable markup in `templates/components.gohtml` when it is shared.
5. Build data in `view.go`; keep logic out of templates.
6. Add a render test for new sections, anchors, escaping or contracts.
7. Update `docs/report-ui.md` and this page.
8. Run `go test ./...`, `go vet ./...` and check 320/720/1920/2560/3840 px.

## Deprecated patterns

Do not reintroduce:

- `.topmeta`, `.nav-back`, `.tag` header metadata.
- `.search-status`, `.search-hidden`, live DOM filtering.
- `.overview-grid`, `.team-compare`, overview cheat-watch panel.
- Multi-column `.player-grid` or three-column player layouts.
- “Open report”, “Demo not linked”, “All reports” back link in the report.
- Hardcoded badge fonts/padding, hardcoded `font-size` in components.
- `n=` sample counters without a unit.
- `<details>` for score adjustments.
