# CS2 Battlelog

Analyze Counter-Strike 2 demos into a **self-contained, dark, Battlelog-style HTML report** — scoreboard, every round's event timeline, per-player stats, Leetify-style aim metrics, and a channel-by-channel cheat analysis.

- One demo in → one `<demo>.html` out (no server, no external assets, works offline).
- `watch` mode keeps an eye on folders and processes new demos automatically.
- Built in Go with [`demoinfocs-golang/v5`](https://github.com/markus-wa/demoinfocs-golang) and a vendored copy of [`demo-anticheat`](https://github.com/TimAnthonyAlexander/demo-anticheat).
- All Go work (build/test) runs in Docker; no local toolchain is required.

> **Assumption:** the UI is Battlelog-inspired; the data is CS2. demoinfocs v5 parses CS2 (Source 2) demos only. CS:GO (Source 1) demos are rejected with a clear error.

## Quick start (Docker only)

```sh
make                     # help
make build               # dist/cs2-battlelog (linux/amd64)
make test                # go test ./... inside golang:1.27

# one-shot
./dist/cs2-battlelog analyze --output ./reports ./demos

# or watch
./dist/cs2-battlelog watch --input ./demos --output ./reports
```

Reports default to the **same directory as the demo**:

```sh
./dist/cs2-battlelog analyze ./demos/match.dem
# ./demos/match.html (plus a downloadable link to match.dem)
```

When demos and reports share a directory, the tool also writes an
`index.html` there and links each report's `.dem` as a download. Use
`--index` / `--demo-download` to enable indexes and demo links when `--output`
points somewhere else.

With Docker Compose:

```sh
mkdir -p demos reports
docker compose up -d --build
```

## What the report contains

1. **Overview** — winner, top fragger, best aim, cheat flags, team comparison.
2. **Scoreboard** — both teams, K/D/A, ADR, HS%, aim rating, MVPs; player names link to their cards.
3. **Rounds** — warmup timeline first, then every round: economy, kills with HS/WB/no-scope/smoke badges, bomb plants/defuses, utility, MVP, round result. Filter by event type or player.
4. **Players** — aggregate and side splits, aim tiles, per-round table, weapon accuracy, and the cheat-analysis card.

Aim profile (fixed anchors, no rank lookup): damage speed, reaction, preaim, accuracy, head accuracy, counter-strafe, spray, composite aim rating and grade.

Cheat analysis: 10-channel Bayesian detector output with raw metric, score, confidence, zone, narrative, boosts/overrides, and an explicit **“statistical signal, not proof”** disclaimer.

## CLI

```text
cs2-battlelog analyze [flags] <demo|directory>...
cs2-battlelog watch   [flags] <directory>...
cs2-battlelog version
```

Key flags: `--input/-i`, `--output/-o`, `--recursive`, `--workers`, `--json`, `--index`, `--demo-download`, `--cheat`, `--aim`, `--force`, `--timeout`, `--interval`, `--debounce`, `--quiet`, `--verbose`.

See [docs/cli.md](docs/cli.md) for all flags, env vars, exit codes, and watch semantics.

## Versioning & releases

`VERSION` uses `YY.MM.BuildNr` (`26.11.1`). The build number increments within a month and resets to `1` on the first of the next month. `make version-bump` updates it. Pushing a `VERSION` change to `main` triggers `.github/workflows/release.yaml`, which cross-compiles Linux/Windows/macOS binaries, attaches `SHA256SUMS`, and keeps the last five releases.

See [docs/releasing.md](docs/releasing.md).

## Documentation

| Doc | Contents |
|---|---|
| [docs/architecture.md](docs/architecture.md) | Pipeline, packages, data flow, extension points |
| [docs/conventions.md](docs/conventions.md) | Global coding, error, logging, testing rules |
| [docs/cli.md](docs/cli.md) | Flags, env vars, watch/cache behavior |
| [docs/parser.md](docs/parser.md) | demoinfocs mapping, round model, edge cases |
| [docs/battlelog-ui.md](docs/battlelog-ui.md) | Sections, design tokens, components, accessibility |
| [docs/aim-metrics.md](docs/aim-metrics.md) | Exact aim formulas, anchors, sample gates |
| [docs/anticheat.md](docs/anticheat.md) | Vendored detector, calibrations, limitations |
| [docs/watching.md](docs/watching.md) | Watch lifecycle, manifest, Docker mounts |
| [docs/report-schema.md](docs/report-schema.md) | JSON/HTML data contract |
| [docs/releasing.md](docs/releasing.md) | Version scheme and release workflow |
| [docs/testing.md](docs/testing.md) | Running tests, demo fixtures |
| [docs/THIRD_PARTY.md](docs/THIRD_PARTY.md) | Vendored code and licenses |

Every feature package also has an `AGENTS.md` with local invariants and extension instructions.

## Anti-cheat caveat

The detector is probabilistic and tuned on a small reference corpus. It is **not** proof of cheating and must not be used as the sole basis for accusations or bans. We vendored the upstream logic, fixed a snap-velocity unit bug, and tightened the sub-100 ms floor to avoid small-sample false positives. Read [docs/anticheat.md](docs/anticheat.md).

The upstream `demo-anticheat` repository has no license file at the pinned commit; if you redistribute this project, confirm permission with the upstream author.
