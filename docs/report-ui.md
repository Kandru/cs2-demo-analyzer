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

The index page has no hero panel; it starts directly with the grouped report
list.

Anchors: `#overview`, `#scoreboard`, `#cheaters`, `#rounds`, `#players`,
`#player-<slug>`, `#cheater-<slug>`, `#round-<startTick>`, `#warmup`.

## Header and search

- Header is a single compact row at every width: brand, section tabs, search.
  Tabs and the search field fill the header height.
- No map/mode tags and no back link.
- Search is an accessible combobox. Results come from the server-rendered
  `#search-index` and appear in an absolutely positioned dropdown, so the
  header never moves while typing.
- Results cover sections, players, rounds and watchlist entries.
- Keyboard: `/`, arrow keys, Home/End, Enter, Escape.
- Without JavaScript the search stays hidden; all sections remain reachable.

## Hero

- Map, mode, duration and round count on the left.
- Team names and the score are right-aligned. If a team has no useful clan
  name, the start side (`CT` or `T`) is shown instead of `Team A`/`Team B`.
- The winning score is green, the losing score is red, and a draw uses the
  neutral text colour. There is no separate verdict block.
- The metadata strip contains only the optional demo download and the report
  version, linked to the project page.

## Overview

- Match badges summarise map, mode, duration, rounds, players, tickrate, date,
  server and warnings. They sit in the shared sticky `.section-tools` bar.
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

- Verdict, likelihood and meter in the header; team, rounds and K/D/A are
  badges.
- Ranked signal cards with label, evidence bar, raw value, labelled evidence,
  confidence and sample badges, and zone.
- Score adjustments are cards with a readable title and a plain-language
  explanation.
- All review cues with round links and `demo_gototick <tick>` copy buttons,
  laid out as cards with a header, badge body and footer.
- “Open full analysis” is a `.button` in the card footer, bottom-left.
- “Watchlist” is the only term used for this feature.
- The short “Statistic calculation only. No Proof.” subtext stays visible.

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

- One player card per row at every viewport width, with a sticky player filter
  using the shared `.section-tools` design.
- Header shows avatar, team/side badges, Steam link, rounds, sanitised clan
  tag, aim grade and verdict.
- The card body is one full-width column: stat strip, T/CT split, aim tiles,
  weapons, cheat analysis, cues and per-round table.
- Aim and channel samples use readable nouns.
- Channel rows flow in two columns when space allows.
- Cheat analysis is a compact verdict, likelihood meter, grade badges and a
  bulleted signal list. Each adjustment has a title and explanation. The short
  “Statistic calculation only. No Proof.” subtext stays visible.
- “When to look” shows every cue as a card in a three-column grid.
- The per-round table uses fixed column widths; the Round column stays narrow.
- Per-round rows link to the timeline round anchor.

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
