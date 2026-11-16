# Documentation index

Start here, then follow the doc for the area you are changing.

## Orientation

| Doc | Contents |
|---|---|
| [architecture.md](architecture.md) | Pipeline, package boundaries, extension points |
| [conventions.md](conventions.md) | Global coding, error, logging, testing rules |
| [cli.md](cli.md) | Flags, env vars, watch/cache behavior |
| [parser.md](parser.md) | demoinfocs mapping, round model, edge cases |
| [report-schema.md](report-schema.md) | JSON/HTML data contract |

## Analysis features

| Doc | Contents |
|---|---|
| [detections/README.md](detections/README.md) | Cheat-detection pipeline, file map, add/remove procedure |
| [detections/](detections/) | One page per cheat channel, modifier and evidence cue |
| [anticheat.md](anticheat.md) | Vendored detector overview, integration patches, limits |
| [aim/README.md](aim/README.md) | Aim profile pipeline and metric index |
| [aim/](aim/) | One page per aim metric |
| [aim-metrics.md](aim-metrics.md) | Rating, grades and shared collection rules |

## Operations

| Doc | Contents |
|---|---|
| [install.md](install.md) | Linux installer, self-update, upgrade and uninstall |
| [watching.md](watching.md) | Watch lifecycle, manifest, Docker mounts |
| [report-ui.md](report-ui.md) | Sections, components, timeline, accessibility |
| [ui.md](ui.md) | Component contract, typography, layout and responsive rules |
| [theme.md](theme.md) | Colour tokens, CS2 team colours, AI/human design rules |
| [testing.md](testing.md) | Test layers, docs-sync checks, demo harness |
| [releasing.md](releasing.md) | Version scheme and release workflow |
| [dependencies.md](dependencies.md) | Compiled modules, audit workflow, upstream cleanup |
| [THIRD_PARTY.md](THIRD_PARTY.md) | Vendored code, licenses and patch table |

Every Go package also has an `AGENTS.md` with local invariants and the
extension procedure for that package.

## Adding a feature

| Feature | Start here |
|---|---|
| Cheat detection | [detections/README.md](detections/README.md) |
| Cheat modifier | [detections/README.md](detections/README.md) |
| Aim metric | [aim/README.md](aim/README.md) |
| New event/stat | [architecture.md](architecture.md) extension points |
| New report section | [report-ui.md](report-ui.md) and `src/internal/report/html/AGENTS.md` |
| New watcher source | [watching.md](watching.md) |
