# Battlelog UI

The report is one self-contained HTML file (inline CSS/JS/SVG, no network
requests, no external fonts). It recreates the BF3/BF4 Battlelog feel without
copying EA assets: dark navy panels, team blue/amber, orange accents, condensed
uppercase headings, dense numeric tables.

## Page structure

```
top bar (brand, section nav, map/mode)
hero (map, score, winner, date, players, tickrate, version)
Overview (winner, top fragger, best aim, flags, rounds, team comparison, warnings)
Scoreboard (two team tables + totals)
Rounds (chips, warmup, round panels with economy + timeline, filters)
Players (pills, cards: stats, sides, aim, cheat, weapons, per-round table)
footer (generated/version/parser, untrusted-input + probabilistic notice)
```

Anchors: `#overview`, `#scoreboard`, `#rounds`, `#players`,
`#player-<slug>`, `#round-<startTick>`, `#warmup`.

## Design tokens

| Token | Value | Use |
|---|---|---|
| `--bg-0` | `#06090e` | page |
| `--panel` | `#0e1621` | cards |
| `--line` | `#223248` | borders |
| `--text` | `#e9f0f8` | body |
| `--muted` | `#8fa2b8` | secondary |
| `--accent` | `#ff7a1a` | highlights, active tab |
| `--ct` / `--t` | `#3f9add` / `#d99a35` | sides |
| `--red` | `#ff4d55` | flags |
| `--green` | `#46c28b` | clean/good |
| `--amber` | `#e2a73c` | review/warning |

Typography: system sans (`Segoe UI`, Roboto, …), `ui-monospace` for numbers.
No web fonts.

## Components

- **Hero**: map-gradient panel (generic, map-hued), big map name, start-side
  team names, score, winner badge, warnings pill.
- **Tiles**: label/value/sub, `alert` variant for cheat flags.
- **Scoreboard**: sticky-ish team headers, right-aligned numeric columns,
  `flagged` left border, AIM column with grade badge, totals footer.
- **Round panel**: `<details>`; summary shows round label, reason, duration,
  score, event count. Economy chips use buy class. Timeline is an ordered list
  with a vertical line, type-colored icon, time, linked names, and badges.
- **Filters**: event-type segmented buttons, player select, expand/collapse.
  JS only enhances; without it all events remain visible.
- **Player card**: avatar initials, team/side, Steam link, aim badge, verdict;
  stat strip; T/CT split; aim tiles with score bars + tooltips + sample counts;
  cheat meter, grades, channel table; weapons; per-round table.
- **Cheat channels**: label, raw value, evidence bar **with numeric %**,
  confidence bar, zone badge. Hidden zone column on narrow screens.
- **Disclaimer**: every cheat card and the footer; never hide it.

## Accessibility

- Skip link, semantic `main/section/article`, heading order.
- Details/summary is keyboard operable; filters are real buttons/selects.
- Focus styles are browser defaults plus visible hover states; do not remove
  outlines without replacing them.
- Color is always paired with text (side labels, verdict text, zone text).
- `prefers-reduced-motion` is respected by avoiding non-essential animation.
- Tables scroll horizontally instead of shrinking text below 12 px.

## Responsiveness

- 1280 px content max.
- ≤980 px: scoreboard columns stack, channel zone hidden.
- ≤640 px: compact top bar, 2-column aim grid, score hides in round summary.
- Print stylesheet: white background, hidden navigation/filters, cards avoid
  page breaks.

## Extending the UI

1. Add fields to `report/html/view.go` (presentation-only) or `model` (contract).
2. Render in `templates/report.gohtml`; escape everything by default.
3. Add styles to `assets/style.css` using tokens.
4. Only add JS in `assets/app.js` if the page still works without it.
5. Update this document and `docs/report-schema.md` when the data contract
   changes.
