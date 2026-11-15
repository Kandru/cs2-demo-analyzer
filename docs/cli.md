# CLI reference

```
cs2-battlelog analyze [flags] <demo|directory>...
cs2-battlelog watch   [flags] <directory|demo>...
cs2-battlelog version
cs2-battlelog help
```

No command prints the help. Unknown commands exit `2`.

## Flags

| Flag | Default | Meaning |
|---|---|---|
| `-i, --input PATH` | — | Repeatable input directory/file. Also accepted as positional arguments. |
| `-o, --output DIR` | demo directory | Where reports are written. Created if missing. |
| `--recursive` | `true` | Walk input directories recursively. |
| `--workers N` | `min(4, CPU)` (analyze), `1` (watch) | Parallel processors. Each demo uses substantial RAM while the anti-cheat pass runs. |
| `--json` | `false` | Write `<demo>.battlelog.json` next to the HTML. |
| `--cheat` | `true` | Run the vendored anti-cheat detector (second parse pass). |
| `--aim` | `true` | Collect aim statistics. |
| `--force` | `false` | Reprocess unchanged demos (watch cache bypass). |
| `--timeout` | `15m` | Per-demo timeout, including both parse passes. |
| `--interval` | `5s` | Watch fallback rescan interval. |
| `--debounce` | `2s` | Watch settle/debounce window. Also requires two identical scans before queueing. |
| `--quiet` | `false` | Errors only. |
| `--verbose` | `false` | Debug logging. |

## Environment

| Variable | Meaning |
|---|---|
| `CS2BATTLELOG_INPUT` | OS path-list separated inputs (`:` on Linux/macOS, `;` on Windows). |
| `CS2BATTLELOG_OUTPUT` | Output directory. |
| `CS2BATTLELOG_JSON` | `true`/`false`. |
| `CS2BATTLELOG_CHEAT` | `true`/`false`. |
| `CS2BATTLELOG_AIM` | `true`/`false`. |

Flags override environment values.

## Output

- `<base>.battlelog.html` — self-contained report. `<base>` strips `.dem`,
  `.dem.gz`, `.dem.bz2`.
- `<base>.battlelog.json` — optional sidecar, schema documented in
  `docs/report-schema.md`.
- `<manifest-dir>/.cs2battlelog/manifest.json` — watch cache. Manifest dir is
  `--output` if set, otherwise the first input directory (or CWD for a single
  file input).

When several demos share a basename and write into the same output directory,
the later one gets a short source-hash suffix: `match-a1b2c3.battlelog.html`.

## Analyze behavior

- Always processes every discovered demo; the manifest is used for output-name
  collision tracking, not for skipping.
- Continues after individual failures and prints `analyzed N demo(s), M failed`.
- Exit `1` if anything failed or the context was cancelled.

## Watch behavior

- Scans on start, on fsnotify events (debounced), and every `--interval`.
- A file must have the same size and mtime on two consecutive scans before it
  is queued. This prevents parsing downloads in progress.
- Successful demos are skipped while size, mtime, tool version, schema version,
  and the report file are unchanged.
- Failed demos are cached as well; they are retried only when the file changes
  or with `--force`.
- Files matching `.dem`, `.dem.gz`, `.dem.bz2` (case-insensitive) are eligible.
- Hidden directories (`.git`, `.cs2battlelog`, …) are skipped.
- SIGINT/SIGTERM stops scanning, drains in-flight work, and saves the manifest.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Processing failure / cancelled |
| `2` | Usage or configuration error |

## Examples

```sh
# one demo, report alongside it
cs2-battlelog analyze stuff/match.dem

# whole tree into a shared report directory, with JSON
cs2-battlelog analyze --output reports --json demos/

# headless container watch
docker run --rm \
  -v "$PWD/demos":/demos:ro -v "$PWD/reports":/out \
  cs2-battlelog:latest watch --input /demos --output /out

# force a re-analysis after changing the template
cs2-battlelog watch --input /demos --output /reports --force
```
