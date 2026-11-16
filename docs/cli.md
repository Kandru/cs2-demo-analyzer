# CLI reference

```
cs2-demo-analyzer analyze [flags] <demo|directory>...
cs2-demo-analyzer watch   [flags] <directory|demo>...
cs2-demo-analyzer update  [--check] [--force]
cs2-demo-analyzer version
cs2-demo-analyzer help
```

`help` prints usage. Unknown commands exit `2`.

## Flags

| Flag | Default | Meaning |
|---|---|---|
| `-i, --input PATH` | — | Repeatable input directory/file. Also accepted as positional arguments. |
| `-o, --output DIR` | demo directory | Where reports are written. Created if missing. |
| `--recursive` | `true` | Walk input directories recursively. |
| `--workers N` | `min(4, CPU)` (analyze), `1` (watch) | Parallel processors. Each demo uses substantial RAM while the anti-cheat pass runs. |
| `--json` | `false` | Write `<demo>.json` next to the HTML. The canonical report document is always built in memory; this only controls the disk sidecar. |
| `--index` | `false` | Write an `index.html` listing reports in each report directory. Also automatic when reports are written next to their demos. |
| `--demo-download` | `false` | Link the source demo for download even when it lives outside the report directory. Same-directory links are always enabled. |
| `--cheat` | `true` | Run the vendored anti-cheat detector (second parse pass). |
| `--aim` | `true` | Collect aim statistics. |
| `--force` | `false` | Reprocess unchanged demos (bypasses the artifact header check and watch cache). On `update`, reinstall the latest release even when versions match. |
| `--update` | `false` | `analyze`, `watch`: check GitHub and self-update before running. A failed update warns but does not stop processing. |
| `--check` | `false` | `update`: only print current/latest versions; never write the binary. |
| `--timeout` | `15m` | Per-demo timeout, including both parse passes. |
| `--interval` | `5s` | Watch fallback rescan interval. |
| `--debounce` | `2s` | Watch settle/debounce window. Also requires two identical scans before queueing. |
| `--quiet` | `false` | Errors only. |
| `--verbose` | `false` | Debug logging. |

## Environment

| Variable | Meaning |
|---|---|
| `CS2_DEMO_ANALYZER_INPUT` | OS path-list separated inputs (`:` on Linux/macOS, `;` on Windows). |
| `CS2_DEMO_ANALYZER_OUTPUT` | Output directory. |
| `CS2_DEMO_ANALYZER_JSON` | `true`/`false`. |
| `CS2_DEMO_ANALYZER_INDEX` | `true`/`false`. |
| `CS2_DEMO_ANALYZER_DEMO_DOWNLOAD` | `true`/`false`. |
| `CS2_DEMO_ANALYZER_CHEAT` | `true`/`false`. |
| `CS2_DEMO_ANALYZER_AIM` | `true`/`false`. |
| `CS2_DEMO_ANALYZER_UPDATE` | `true`/`false`; update before `analyze`/`watch`. |

Flags override environment values.

## Output

- `<base>.html` — self-contained report. `<base>` strips `.dem`,
  `.dem.gz`, `.dem.bz2`. It starts with `<!doctype html>` and a
  `<!--cs2-demo-analyzer:{...}-->` header comment carrying the tool version,
  schema, report options and demo fingerprint.
- `<base>.json` — optional sidecar, schema documented in
  `docs/report-schema.md`. Its first key is the same `header` object.
- `index.html` — self-contained list of reports in a report directory. Written
  when `--index` is set or when a report is written next to its demo.
- `<manifest-dir>/.cs2-demo-analyzer/manifest.json` — watch cache. Manifest dir is
  `--output` if set, otherwise the first input directory (or CWD for a single
  file input).

Both report artifacts are reused only while the header matches the current tool
version, schema, report options and demo file (size+mtime+fingerprint). Delete a
report or pass `--force` to rebuild it.

Each report links its source `.dem` as a download when the demo is in the same
directory as the report (the default output layout). Use `--demo-download` to
opt in to relative links when `--output` points elsewhere. The link points at
the original demo path; move or delete demos accordingly.

When several demos share a basename and write into the same output directory,
the later one gets a short source-hash suffix: `match-a1b2c3.html`.

## Analyze behavior

- Scans every discovered demo but reuses up-to-date reports. A report is current
  when its HTML (and JSON when `--json` is enabled) carries a matching header.
- `--force` reprocesses everything; changing the tool version, schema, or
  `--cheat`/`--aim`/`--demo-download`/`--index` also invalidates.
- Continues after individual failures and prints
  `analyzed N demo(s), S skipped, M failed`.
- Exit `1` if anything failed or the context was cancelled.
- Index pages are rewritten at the end of the run, including directories whose
  demos were all skipped.

## Watch behavior

- Scans on start, on fsnotify events (debounced), and every `--interval`.
- A file must have the same size and mtime on two consecutive scans before it
  is queued. This prevents parsing downloads in progress.
- Successful demos are skipped while the artifact headers match the current
  tool version, schema, report options, demo size/mtime and demo fingerprint.
  The JSON sidecar is required when `--json` is enabled.
- Report indexes are refreshed on startup for cached reports, then after each
  successful analysis.
- Failed demos are cached in the manifest as well; they are retried only when
  the file changes or with `--force`.
- Files matching `.dem`, `.dem.gz`, `.dem.bz2` (case-insensitive) are eligible.
- Hidden directories (`.git`, `.cs2-demo-analyzer`, …) are skipped.
- SIGINT/SIGTERM stops scanning, drains in-flight work, and saves the manifest.

## Update behavior

`update` queries the latest GitHub release for
`Kandru/cs2-demo-analyzer`, compares `YY.MM.BuildNr` numerically, and installs
the matching `cs2-demo-analyzer-<version>-<os>-<arch>.tar.gz` after verifying
`SHA256SUMS`.

- `update --check` prints `current`, `latest`, and whether an update is
  available. It never downloads the archive.
- `update` installs only when the latest version is newer.
- `update --force` reinstalls the latest release and is also required to move a
  `dev` build onto a release.
- The downloaded executable is written beside the current binary and atomically
  renamed over it. Updating `/opt/cs2-demo-analyzer/cs2-demo-analyzer` normally
  requires `sudo`.
- Linux and macOS support self-update. Windows reports that self-update is
  unsupported and points at the release download.
- No command contacts GitHub unless `update` is run or `--update` is passed to
  `analyze`/`watch`.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success / already up to date |
| `1` | Processing failure / cancelled / update failure |
| `2` | Usage or configuration error |

## Examples

```sh
# one demo, report alongside it
cs2-demo-analyzer analyze stuff/match.dem

# whole tree into a shared report directory, with JSON
cs2-demo-analyzer analyze --output reports --json demos/

# headless container watch
docker run --rm --user "$(id -u):$(id -g)" \
  -v "$PWD/demos":/demos:ro -v "$PWD/reports":/out \
  cs2-demo-analyzer:latest watch --input /demos --output /out

# force a re-analysis after changing the template
cs2-demo-analyzer watch --input /demos --output /reports --force

# check for a newer release without changing the binary
cs2-demo-analyzer update --check

# update the installed binary (root-owned installs need sudo)
sudo cs2-demo-analyzer update

# update before processing a batch
cs2-demo-analyzer analyze --update demos/
```
