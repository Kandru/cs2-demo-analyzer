# Report UI

The report is one self-contained HTML file (inline CSS/JS/SVG and inlined brand
images, no network requests, no external fonts). The visual language is a dark
CS 1.6-inspired interface: beveled panels, hard corners, gold HUD accents, CS2
team identity colours and dense numeric tables.

The colour token contract lives in [theme.md](theme.md). The component contract
lives in [ui.md](ui.md). The machine sources of truth are
`src/internal/report/html/assets/theme.css` for tokens and `assets/style.css`
for components.

## Page structure

```
top bar (brand, section tabs, search dropdown)
match hero (map, mode, duration, score, teams, winner, demo, metadata)
Overview
  match badges
  stat tiles
  round history
  scoreboard (both teams, aggregates and totals)
  parser warnings
Watchlist (flagged/review players: signals, adjustments, review cues)
Timeline (one continuous demo time scale, filters and event details)
Players (one card per row: stats, sides, aim, cheat analysis, weapons, rounds)
footer (project link, version, generated time, parser version, byline)
```

Both pages share the same topbar, footer, icon sprite and component templates in
`templates/components.gohtml`. The report brand links back to `index.html` when
an index is generated. There is no “All reports” link inside the header.

Anchors: `#overview`, `#scoreboard`, `#cheaters`, `#rounds`, `#players`,
`#player-<slug>`, `#cheater-<slug>`, `#round-<startTick>`, `#warmup`.

## Header and search

- Header is a single compact row at every width: brand, section tabs, search.
- No map/mode tags and no back link.
- Search is an accessible combobox. Results come from the server-rendered
  `#search-index` and appear in an absolutely positioned dropdown, so the
  header never moves while typing.
- Results cover sections, players, rounds and watchlist entries.
- Keyboard: `/`, arrow keys, Home/End, Enter, Escape.
- Without JavaScript the search stays hidden; all sections remain reachable.

## Hero

- Map, mode, duration and round count on the left.
- Team names around the score in the centre. If a team has no useful clan
  name, the start side (`CT` or `T`) is shown instead of `Team A`/`Team B`.
- The winning score is green, the losing score is red, and a draw uses the
  neutral text colour. A draw is not announced as text in the hero.
- Either a demo download link (when enabled) or the demo file name, never both.
- Date, player count, tickrate and report version sit in a compact metadata
  strip.

## Overview

- Match badges summarise map, mode, duration, rounds, players, tickrate, date,
  server and warnings.
- Stat tiles cover winner, top fragger, best aim, watchlist count and rounds.
- Round history is full width. Cells are large enough to read and show at most
  eight rounds per row on desktop (six on tablet, four/three on phones).
- The scoreboard lives inside Overview, full width. The winning score and
  `WIN` marker use the success colour, the losing score uses the danger colour.
  A team aggregate row shows K/D/A, ADR, HS% and aim rating; the table footer
  keeps the totals.
- Team comparison was removed; its useful values now live in the scoreboard
  aggregate row.
- Parser warnings stay visible at the end of the section.

## Watchlist

One card per flagged or review-level player.

- Verdict, likelihood and meter in the header.
- Ranked signal list with label, evidence bar, raw value, confidence, samples
  and zone.
- Score adjustments are always visible when present. There is no toggle.
- Top review cues with round links and `demo_gototick <tick>` copy buttons.
- “Open full analysis” is bottom-left in the card footer.
- “Watchlist” is the only term used for this feature.
- The statistical disclaimer stays visible.

## Timeline

The Timeline section is one vertical time scale, not a stack of expandable
round panels.

- Pre-match and warmup blocks come first.
- Every round starts with an absolute demo timestamp, duration, score, reason,
  economy chips and per-player purchase badges.
- Round rows show freeze-time end, kills, bomb events, utility, flash blinds,
  grenade damage, hostage events and MVP.
- Every round ends with a labelled round-end marker.
- Intermission blocks contain connects, disconnects, team changes and the
  halftime side swap.
- Every row carries an absolute demo timestamp. Round rows also show a small
  round-relative time where space allows.
- Adjacent rounds alternate structural gold/steel markers. Winner side is a
  separate labelled chip.
- Filters: event type and player. The filter toolbar sticks below the section
  header. JS only enhances; without it every row stays visible.
- The section header is sticky, so the title and filter bar remain visible
  until the next section takes over.

## Players

- One player card per row at every viewport width.
- Header shows avatar, team/side badges, Steam link, rounds, aim grade and
  verdict.
- Stat strip, T/CT split, aim tiles with readable sample nouns, full ranked
  channel list and weapon table.
- Cheat analysis is a compact verdict, likelihood meter, grade badges and a
  bulleted signal list. Score adjustments are always visible when present.
- Per-round table links each row to the timeline round anchor.

## Footer

- Shared by index and detail pages.
- Always sits at the bottom of the viewport on short pages.
- Compact height, below the header height on desktop, with vertically centred
  content.
- Project link and `kandru.net` open in a new tab with
  `rel="noopener noreferrer"`.
- Shows version, generation time, parser version and byline.

## Brand assets

The project logo and favicon are inlined as data URIs:

- `brand-mark.png` - 64x64 icon crop, topbar and footer
- `logo.jpg` - 560x313 wide logo, index hero
- `favicon-32.png` - browser tab icon
- `apple-touch-icon.png` - 152x152 touch icon

The optimised copies live in `src/internal/report/html/assets/`. Source files
remain in `docs/assets/`. The embedded payload budget is enforced by a test.

## Accessibility

- Skip link, semantic `main/section/article/nav`, heading order.
- Search is a labelled combobox with a listbox and keyboard support.
- Timeline filters and tick buttons are real controls with labels.
- Tables use captions, `scope` and horizontal scrolling instead of tiny text.
- Timeline is an ordered list and remains readable without JavaScript.
- Colour is always paired with text (side labels, verdict text, zone text).
- Focus rings use `--focus` and remain visible on every control.
- `prefers-reduced-motion` is respected.
- Anchor links account for the sticky header with `scroll-margin-top`.

## Responsiveness

- Fluid root type scale and content width; no fixed 1280 px cap.
- The content cap is 2400 px so 4K screens are used without microscopic text.
- 320-480 px: one column, three round cells, compact footer.
- <=720 px: compact header, four round cells, wrapped filters.
- <=1024 px: stacked scoreboards, six round cells.
- <=1280 px: hero verdict wraps; player card body becomes one column.
- 1920/2560/3840 px: readable fluid type, eight round cells, capped line length.
- Print stylesheet: white background, hidden navigation/filters/search/copy
  controls, panels avoid page breaks.

## Extending the UI

1. Add fields to `report/html/view.go` (presentation-only) or `model` (contract).
2. Add or reuse a component in `templates/components.gohtml`.
3. Use existing tokens from `theme.css`; add a token and a `docs/theme.md` entry
   if a genuinely new colour role is needed.
4. Follow [ui.md](ui.md) for component markup and typography.
5. Only add JS in `assets/app.js` if the page still works without it.
6. Update this document and `docs/report-schema.md` when the data contract
   changes.
