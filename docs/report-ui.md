# Report UI

The report is one self-contained HTML file (inline CSS/JS/SVG and inlined brand
images, no network requests, no external fonts). The visual language is a dark
CS 1.6-inspired interface: beveled panels, hard corners, compact uppercase
tabs, gold HUD accents, CS2 team identity colours, and dense numeric tables.

The colour and token contract lives in [theme.md](theme.md). The machine source
of truth is `src/internal/report/html/assets/theme.css`; every other stylesheet,
template and script consumes its custom properties.

## Page structure

```
top bar (brand -> report index when available, section tabs, search, map/mode,
          All reports link)
match hero (map, mode, duration, score, winner, date, players, tickrate, version)
Overview (tiles, round-history strip, team comparison, watch preview, warnings)
Scoreboard (two team tables + totals)
Cheat watch (flagged/review players: ranked signals + review cues)
Rounds and timeline (one continuous demo time scale with filters)
Players (quick jumps, cards: stats, sides, aim, cheat channels, cues, weapons,
         per-round table)
footer (project link, version, generated time, parser version, byline)
```

Both pages share the same topbar, footer, icon sprite and component templates in
`templates/components.gohtml`. The detail page brand links back to `index.html`
when an index is generated; the header also exposes an "All reports" link.

Anchors: `#overview`, `#scoreboard`, `#cheaters`, `#rounds`, `#players`,
`#player-<slug>`, `#cheater-<slug>`, `#round-<startTick>`, `#warmup`.

## Continuous timeline

The Rounds section is one vertical time scale, not a stack of expandable round
panels.

- Pre-match and warmup blocks come first.
- Every round starts with an absolute demo timestamp, duration, score, reason
  and economy chips.
- Round rows show freeze-time end, kills, bomb events, utility and MVP.
- Every round ends with a labelled round-end marker.
- Intermission blocks between rounds contain connects, disconnects, team
  changes and the halftime side swap.
- Lifecycle events have their own icons and neutral styling.
- Rows carry an absolute demo timestamp. Round rows also show a small
  round-relative time.
- Adjacent rounds alternate structural gold/steel markers. Winner side is a
  separate labelled chip, so a round marker never implies CT or T.
- Filters: event type, player. JS only enhances; without it every row stays
  visible.
- Search filters player cards, cheat cards and timeline sections. It is
  progressive enhancement; the page is complete without it.

## Components

- **Match hero**: map/mode/duration on the left, teams around the score in the
  centre, winner and warnings on the right, metadata strip below.
- **Stat tiles**: label/value/sub with a linked cheat-flags tile.
- **Round history**: one cell per round, winner-side colour, buy-type underline,
  flag marker, links to `#round-<startTick>`.
- **Scoreboard**: two dense team tables, right-aligned numeric columns, team
  header tint and rail, `flagged` row marker, AIM column with grade badge,
  totals footer. Tables scroll horizontally on narrow screens.
- **Cheat watch**: compact priority cards for flagged and >=25% review players.
  - ranked signal list with label, evidence bar, raw value, confidence, samples
    and zone
  - top review cues with round links and `demo_gototick <tick>` copy buttons
  - link to the full player card
  - statistical disclaimer stays visible
- **Player card**: avatar initials, team/side tag, Steam link, aim badge,
  verdict; stat strip; T/CT split; aim tiles with score bars, tooltips and
  sample counts; full ranked channel list; all review cues; weapons; per-round
  table.
- **Channels**: ranked ordered list, not a wide grid. Each row shows rank,
  label, evidence bar, raw value, confidence and sample count.
- **Footer**: shared by index and detail pages. Project link, version,
  generation time, parser version where relevant, and
  `by cs2-demo-analyzer from kandru.net`.

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
- Search, filters and tick buttons are real controls with labels.
- Tables use captions, `scope`, and horizontal scrolling instead of tiny text.
- Timeline is an ordered list and remains readable without JavaScript.
- Colour is always paired with text (side labels, verdict text, zone text).
- Focus rings use `--focus` and remain visible on every control.
- `prefers-reduced-motion` is respected.
- The detail page keeps anchor links stable for player and round jumps.

## Responsiveness

- Fluid root type scale and content width; no fixed 1280 px cap.
- The content cap is 3000 px so 4K screens are used without microscopic text.
- <=1280 px: hero verdict wraps; overview becomes two columns; player body
  becomes one column.
- <=1024 px: scoreboards and cheat cards stack; index hero narrows.
- <=720 px: compact topbar, search moves to its own row, hero stacks, round
  timeline compacts, cue rows wrap, index becomes one column.
- <=480 px: single-column stat tiles, full-width filters, compact round badges.
- >=1600 px: two-column player grid with stacked card bodies.
- >=2600 px: three-column player grid on very wide displays.
- >=2200/3000 px: larger root scale and wider content for 4K.
- Print stylesheet: white background, hidden navigation/filters/search/copy
  controls, panels avoid page breaks.

## Extending the UI

1. Add fields to `report/html/view.go` (presentation-only) or `model` (contract).
2. Add or reuse a component in `templates/components.gohtml`.
3. Use existing tokens from `theme.css`; add a token and a `docs/theme.md` entry
   if a genuinely new colour role is needed.
4. Only add JS in `assets/app.js` if the page still works without it.
5. Update this document and `docs/report-schema.md` when the data contract
   changes.
