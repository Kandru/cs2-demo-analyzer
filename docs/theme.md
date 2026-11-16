# Theme and color guideline - AI and human contributors

This page is the design contract for every generated HTML page. The machine
source of truth is
[`src/internal/report/html/assets/theme.css`](../src/internal/report/html/assets/theme.css).
If this document and the stylesheet disagree, the stylesheet is authoritative
and this document must be updated before the change lands.

## Authority order

1. `assets/theme.css` - the only file allowed to contain raw color values.
2. `docs/theme.md` - explains roles, contrast targets and component rules.
3. `assets/style.css` - implements components using `var(--token)` only.
4. `templates/*.gohtml`, `view.go`, `index.go`, `app.js` - structure and
   behaviour only; no raw colors.

The vendored `assets/vendor/new.css` is a framework/reset layer. It may contain
its own variables internally, but every visible decision is overridden by
`theme.css`. Do not edit vendor files.

## Load order

```text
assets/vendor/new.css -> assets/theme.css -> assets/style.css
```

## Non-negotiable rules

- Dark mode only. Never add a light theme or `prefers-color-scheme: light`.
- Use tokens, never literal `#hex`, `rgb()`, `rgba()`, `hsl()` or named colors
  in component CSS, templates or scripts.
- Team colors identify teams only. They never mean warning, success, danger or
  interactive state.
- Gold identifies the CS 1.6-style accent: links, active tabs, primary actions
  and focus. Gold is not a team color.
- Round alternation uses the structural gold/steel pair, not CT/T colors.
- Status colors always appear with text or an icon. Never rely on color alone.
- Body text, muted text and team labels must keep at least 4.5:1 contrast on
  their surface.
- Hard corners, 2 px bevel borders and dense numeric tables. No gradients,
  glow, blur or decorative perspective.
- Every colour change is a token change plus a `docs/theme.md` change.

## Palette

### Surfaces and text

| Token | Value | Role |
|---|---|---|
| `--bg-0` | `#090a0b` | page background |
| `--bg-1` | `#0f1113` | topbar, footer, table header |
| `--bg-2` | `#15181b` | inset surfaces |
| `--panel` | `#181c1f` | cards and panels |
| `--panel-2` | `#1e2327` | raised card sections |
| `--panel-3` | `#252b30` | controls and bar tracks |
| `--line-hi` | `#4a5259` | bevel light edge |
| `--line` | `#343a40` | standard border |
| `--line-soft` | `#262c31` | row dividers |
| `--line-lo` | `#050607` | bevel dark edge |
| `--text` | `#eceff1` | primary text |
| `--muted` | `#b7bec4` | secondary text |
| `--dim` | `#8f979e` | tertiary text and labels |
| `--inverse` | `#0b0d0e` | text on light fills |

### CS 1.6 accent

| Token | Value | Role |
|---|---|---|
| `--accent` | `#d7c25a` | links, active tab, primary button |
| `--accent-strong` | `#f0dd86` | hover and emphasis |
| `--accent-ink` | `#151203` | text on accent fill |
| `--accent-soft` | `rgba(215, 194, 90, 0.14)` | selected/active tint |
| `--accent-border` | `rgba(215, 194, 90, 0.46)` | accent outline |

### CS2 team identity

Counter-Terrorists use the familiar CS2 blue family. Terrorists use the
familiar CS2 amber/yellow family. These are identity colors, tuned for contrast
on the dark surfaces.

| Token | Value | Role |
|---|---|---|
| `--team-ct-fill` | `#5b9bd5` | CT dots, bars, avatar fill, 4 px rail |
| `--team-ct-text` | `#8ec5f2` | CT labels on dark |
| `--team-ct-ink` | `#081018` | text on CT fill |
| `--team-ct-soft` | `rgba(91, 155, 213, 0.16)` | CT panel/header tint |
| `--team-ct-border` | `rgba(91, 155, 213, 0.55)` | CT outlines |
| `--team-t-fill` | `#e0a73c` | T dots, bars, avatar fill, 4 px rail |
| `--team-t-text` | `#f2c66d` | T labels on dark |
| `--team-t-ink` | `#171006` | text on T fill |
| `--team-t-soft` | `rgba(224, 167, 60, 0.16)` | T panel/header tint |
| `--team-t-border` | `rgba(224, 167, 60, 0.55)` | T outlines |
| `--team-spec-fill` | `#7d858c` | spectator/unassigned fill |

Team colour application:

| Surface | Rule |
|---|---|
| Banner | 4 px side rail; side label in `*-text`; team name stays `--text` |
| Scoreboard | header tint `*-soft`, left rail `*-fill`, names `--text` |
| Player avatar | `*-fill` background, `*-ink` initials |
| Side dot / side tag | `*-fill` or `*-text` plus a visible `CT`/`T` label |
| Economy chip | `*-fill` left rail, neutral body |
| Team comparison bars | `*-fill` |
| Round result | `*-text` label on `*-soft` background |
| Timeline round header | result chip only; the round parity marker is structural |
| Player side split | `*-border` + `*-text` |

Never fill a whole card with a team colour. Never use CT blue for links.

### Status colours

| Token | Value | Role |
|---|---|---|
| `--danger` / `--danger-text` | `#e06060` / `#ff9a9a` | flags, errors |
| `--danger-soft` / `--danger-border` | `rgba(224, 96, 96, 0.14)` / `rgba(224, 96, 96, 0.5)` | error tint and outline |
| `--warning` / `--warning-text` | `#e08a3c` / `#f3b47a` | review, warnings |
| `--warning-soft` / `--warning-border` | `rgba(224, 138, 60, 0.14)` / `rgba(224, 138, 60, 0.5)` | warning tint and outline |
| `--success` / `--success-text` | `#7fc98a` / `#a7e3b0` | clean, positive |
| `--success-soft` / `--success-border` | `rgba(127, 201, 138, 0.14)` / `rgba(127, 201, 138, 0.5)` | success tint and outline |
| `--info` / `--info-text` | `#6fa8dc` / `#9cc8ef` | informational labels |
| `--neutral` / `--neutral-text` | `#8f979e` / `#b7bec4` | unknown, unavailable |
| `--neutral-soft` / `--neutral-border` | `rgba(143, 151, 158, 0.12)` / `rgba(143, 151, 158, 0.42)` | neutral tint and outline |

Each group also has `*-soft` and `*-border` tokens for tinted surfaces and
outlines. Status colours are independent from team colours.

### Structural round colours

| Token | Value | Role |
|---|---|---|
| `--round-a` | `#d7c25a` | alternating round marker A |
| `--round-a-soft` | `rgba(215, 194, 90, 0.13)` | A tint |
| `--round-b` | `#8095b0` | alternating round marker B |
| `--round-b-soft` | `rgba(128, 149, 176, 0.14)` | B tint |

Round colours distinguish adjacent rounds. They must not imply a winning side.
The winner is shown separately as `CT win`, `T win` or `Draw` with the matching
team colour.

### Bevel and focus

| Token | Value | Role |
|---|---|---|
| `--bevel-light` | `#4a5259` | top/left border |
| `--bevel-dark` | `#050607` | bottom/right border |
| `--bevel-inset` | `rgba(255, 255, 255, 0.07)` | inner highlight |
| `--bevel-shadow` | `rgba(0, 0, 0, 0.55)` | outer shadow |
| `--focus` | `#f0dd86` | focus outline |
| `--focus-ring` | `rgba(240, 221, 134, 0.55)` | focus halo |
| `--print-bg` | `#ffffff` | print page and panel background |
| `--print-text` | `#111111` | print body text |
| `--print-muted` | `#555555` | print secondary text |
| `--print-line` | `#cccccc` | print rules and borders |
| `--print-link` | `#0645ad` | print links |

### Typography, spacing and layout

| Token | Role |
|---|---|
| `--font-ui` | system UI stack, no web fonts |
| `--font-mono` | timestamps and numeric tables |
| `--fs-2xs` | 0.75rem fine print and tick labels |
| `--fs-xs` | 0.8125rem small labels and badges |
| `--fs-sm` | 0.9375rem table and metadata text |
| `--fs-base` | 1rem body text |
| `--fs-md` | 1.125rem emphasis |
| `--fs-lg` | 1.375rem card headings |
| `--fs-xl` | 1.75rem large card values |
| `--fs-2xl` | fluid section title |
| `--fs-3xl` | fluid page title |
| `--fs-display` | fluid hero score |
| `--fs-display-sm` | fluid hero score separator |
| `--lh-tight` / `--lh-base` | heading and body line height |
| `--space-1` | 0.25rem |
| `--space-2` | 0.5rem |
| `--space-3` | 0.75rem |
| `--space-4` | 1rem |
| `--space-5` | 1.5rem |
| `--space-6` | 2rem |
| `--radius` | always `0` for hard corners |
| `--border-w` | bevel border width |
| `--tap` | minimum interactive target |
| `--touch-gap` | minimum gap between touch targets |
| `--content-max` | maximum content width for 4K screens |
| `--measure` | maximum prose line length |
| `--header-h` | header height and sticky offset |
| `--footer-h` | compact footer minimum height |
| `--section-gap` | spacing between major sections |
| `--section-head-h` | sticky section-head height used by the filter bar |
| `--badge-pad` | shared badge padding |
| `--badge-fs` | shared badge font size |
| `--tl-time-w` | timeline timestamp column |
| `--tl-node-w` | timeline event node column |

Media-query breakpoints are documented here because CSS custom properties
cannot be used inside `@media`:

| Width | Behaviour |
|---|---|
| `<= 480px` | single-column stat tiles, three round cells, stacked timeline timestamps, scrollable tables |
| `<= 720px` | single-column cards, compact topbar, four round cells, wrapped filters |
| `<= 1024px` | scoreboards stack, six round cells, index hero narrows |
| `>= 1600px` | wider index grid |
| `>= 2200px` | wider content gutters |
| `>= 3000px` | full 4K layout, no cramped columns |

The root font size is fluid (`clamp(1rem, 0.42vw + 0.625rem, 1.5rem)`) so
1080p, 2K and 4K screens all get a readable base size without browser zoom.

## Component recipes

### Panel and card

```css
background: var(--panel);
border: var(--border-w) solid;
border-color: var(--line-hi) var(--line-lo) var(--line-lo) var(--line-hi);
box-shadow: inset 1px 1px 0 var(--bevel-inset);
border-radius: var(--radius);
```

### Buttons, tabs and focus

- Default: `--panel-2` background, `--line` border bevel, `--text`.
- Hover: `--accent-strong` text, `--line-hi` border.
- Active/pressed: reverse bevel, `--accent` text.
- Selected tab: `--accent` text, `--accent` bottom border.
- Focus: `2px solid var(--focus)` with `box-shadow: 0 0 0 3px var(--focus-ring)`.
- Never use team fills for controls.

### Tables

- Header: `--bg-1`, `--dim` uppercase labels.
- Rows: `--line-soft` dividers; hover `--bg-2`.
- Numeric cells: `--font-mono`, tabular numbers, right aligned.
- Team header: team tint/rail, but body text stays neutral.

### Badges and verdicts

- One primitive: `.badge` with a semantic modifier.
- Flagged/danger: `.badge--danger`.
- Review/unusual: `.badge--warning`.
- Clean/positive: `.badge--success`.
- Informational: `.badge--info`.
- Accent/feature: `.badge--accent`.
- Neutral/unknown: `.badge--neutral`.
- Team identity: `.badge--ct` / `.badge--t`.
- Always include the verdict word, not only the colour.
- Component-specific classes (`verdict`, `grade`, `zone`, `result-chip`) must
  reuse the badge padding, font size and weight tokens.

### Bars

- Track: `--panel-3`.
- Team bars: `--team-ct-fill` / `--team-t-fill`.
- Aim quality: success / warning / danger.
- Detector zones: clean/mild/strong/blatant use status tokens.
- Add the numeric value beside every bar.

### Timeline

- Round header rail: alternating `--round-a` / `--round-b`.
- Round result chip: team label and team tint.
- Round start marker: structural round colour plus the absolute demo time.
- Round end marker: `--panel-3` node with a labelled result chip.
- Team swap: `--accent` marker with text.
- Lifecycle events: neutral icon and text; no team colour unless the row
  describes a join or side change, where the destination side gets a small tag.
- Every row shows an absolute demo timestamp in `--font-mono`.

## Accessibility

- Minimum 4.5:1 for body, muted and team-label text.
- Focus ring is always visible on keyboard focus.
- Timeline and tables remain meaningful without JavaScript.
- Use `aria-current`, `scope`, captions/labels and visible text for state.
- Respect `prefers-reduced-motion`; no essential animation.
- Decorative brand images use empty `alt`; adjacent text carries the name.

## Adding or changing a colour

1. Edit `assets/theme.css`.
2. Add or update the token in this document.
3. Add the token to the required-token test if it is foundational.
4. Style the component with `var(--token)` only.
5. Run `go test ./...`; docs-sync and raw-colour tests must pass.
6. Check 320 px, 720 px, 1280 px, 1920 px and 3840 px.

## AI contributor checklist

- [ ] I did not add a raw colour outside `theme.css`.
- [ ] I used a semantic token, not a team token, for status or interaction.
- [ ] Team colours are paired with `CT`, `T` or a team name.
- [ ] Round alternation is structural gold/steel, not team colour.
- [ ] The change keeps 4.5:1 contrast.
- [ ] I updated this document for every token change.
- [ ] I did not add gradients, glow, blur or a light theme.
