# Report UI

The report is one self-contained HTML file (inline CSS/JS/SVG, no network
requests, no external fonts). The file starts with `<!doctype html>` and a
`<!--cs2-demo-analyzer:{...}-->` metadata comment used to decide whether the
report is current. The comment is not shown to readers.

The visual language is inspired by the BF3/BF4 Battlelog, without copying EA
assets: flat dark-navy surfaces, 1 px steel borders, hard corners, orange
accents, condensed uppercase labels, and dense numeric tables.

## Page structure

```
top bar (brand → report index when available, section tabs, search, map/mode, All reports link)
match banner (map, mode, score, winner, date, players, tickrate, version)
Overview (tiles, flat round-history strip, team comparison, cheat watch, warnings)
Scoreboard (two team tables + totals)
Cheat watch (flagged/review players: why + timestamped when)
Rounds (chips, warmup, round panels with economy + timeline, filters)
Players (pills, cards: stats, sides, aim, cheat, review cues, weapons, per-round)
footer (generated/version/parser, untrusted-input + probabilistic notice)
```

The report index footer links **CS2 Demo Analyzer** to the project repository and
shows the version next to it followed by the generation timestamp.

When a report directory also receives an `index.html`, it uses the same shared
stylesheet and dark design tokens as the detailed reports. Reports are grouped
by server, newest first by played date/time, and each card shows the map name,
end score, server name, played date, and an optional “Download demo” link. The
detail page brand links back to `index.html` whenever an index is generated;
the header also exposes an “All reports” link on wider screens.

Anchors: `#overview`, `#scoreboard`, `#cheaters`, `#rounds`, `#players`,
`#player-<slug>`, `#cheater-<slug>`, `#round-<startTick>`, `#warmup`.

## Design tokens

| Token | Value | Use |
|---|---|---|
| `--bg-0` | `#070b10` | page |
| `--panel` | `#0f161f` | cards |
| `--line` | `#24303c` | borders |
| `--text` | `#e7eef6` | body |
| `--muted` | `#95a6b8` | secondary |
| `--accent` | `#ff7a1a` | active tab, highlights |
| `--ct` / `--t` | `#3f9add` / `#d99a35` | sides |
| `--red` | `#ef4b55` | flags |
| `--green` | `#4fc08d` | clean/good |
| `--amber` | `#e0a53b` | review/warning |

Typography: system sans (`Segoe UI`, Tahoma, Roboto, …), `ui-monospace` for
numbers. No web fonts. Headings are uppercase with wide tracking; body text is
dense but never below 12 px in tables.

## Components

- **Match banner**: compact replacement for the old hero. Map/mode/date on the
  left, start-side teams around the score in the centre, winner + warnings on
  the right, metadata strip below. No perspective grid, glow, or oversized art.
- **Search** (JS enhancement): the top bar search filters player cards, player
  pills, cheat-watch cards and round panels by name, alias, team, SteamID,
  map/round text and detector narrative/channel labels. It reports live result
  counts, opens matching round panels, clears with Escape, and focuses with `/`.
  Without JavaScript the field is hidden and all content stays visible.
- **Anchors** account for the sticky top bar with a single `scroll-padding-top`
  on the document, so section and player links land on the heading instead of
  the previous section.
- **Tiles**: label/value/sub with a linked cheat-flags tile.
- **Round history**: flat one-cell-per-round strip, coloured by winning side,
  with a buy-type underline and a flag marker for rounds with flagged activity.
  Cells link to `#round-<startTick>`.
- **Scoreboard**: two dense team tables, right-aligned numeric columns,
  `flagged` left border, AIM column with grade badge, totals footer.
- **Cheat watch**: cards for flagged and ≥25 % review players.
  - *Why*: narrative, top three channels by score × confidence × weight, grades,
    collapsible score adjustments.
  - *When*: detector cues merged by tick, showing match clock, round and round
    clock, reason chips (`Snap 2.4 °/ms`, `Sub-100 ms 88 ms`, `Pre-aim 3.4°`,
    `Wallbang`, …), weapon/opponent, a round link, and a copy button for
    `demo_gototick <tick>`. Top cues are shown first; the full list lives in a
    native `<details>`.
  - The same cue list appears in the matching player card.
- **Round panel**: `<details>`; summary shows round label, reason, duration,
  score, event count, and a flag marker when the round contains flagged cues.
  Economy chips use buy class. Timeline is an ordered list with a vertical line,
  type-coloured icon, time, linked names, and badges.
- **Filters**: event-type segmented buttons, player select, expand/collapse.
  JS only enhances; without it all events remain visible.
- **Player card**: avatar initials, team/side, Steam link, aim badge, verdict;
  stat strip; T/CT split; aim tiles with score bars + tooltips + sample counts;
  cheat meter, grades, channel table; timestamped review cues; weapons;
  per-round table.
- **Cheat channels**: label, raw value, evidence bar with numeric %, confidence
  bar, zone badge.
- **Disclaimer**: every cheat surface and the footer; never hide it.

## Accessibility

- Skip link, semantic `main/section/article`, heading order.
- The optional report search is labelled, announces result counts through
  `role="status"`/`aria-live`, and never hides content without JavaScript.
- Details/summary is keyboard operable; filters are real buttons/selects.
- Tick copy buttons have an accessible label; the tick number stays visible
  and selectable without JavaScript.
- Color is always paired with text (side labels, verdict text, zone text).
- `prefers-reduced-motion` is respected by avoiding non-essential animation.
- Tables scroll horizontally instead of shrinking text below 12 px.

## Responsiveness

- Fluid root type scale (`clamp`) and content width. No fixed 1280 px cap.
- ≤1200 px: banner stacks, dashboard and player body become one column.
- ≤900 px: scoreboard/detail grids stack; channel zone hidden.
- ≤640 px: compact tab bar, 2-column tiles, compact round summary, cue rows wrap.
  The report index becomes a single column with wrapping map/server names.
- ≤420 px: single-column tiles, narrow cue rows.
- ≥1700 px: wider grids and round cells; ≥2200/3000 px: larger root scale and
  wider content so 4K displays are filled without microscopic text.
- Print stylesheet: white background, hidden navigation/filters/copy buttons,
  cards avoid page breaks.

## Extending the UI

1. Add fields to `report/html/view.go` (presentation-only) or `model` (contract,
   then extend `report.Header` if the field should appear in index metadata).
2. Render in `templates/report.gohtml`; escape everything by default.
3. Add styles to `assets/style.css` using tokens.
4. Only add JS in `assets/app.js` if the page still works without it.
5. Update this document and `docs/report-schema.md` when the data contract or
   artifact header changes.
