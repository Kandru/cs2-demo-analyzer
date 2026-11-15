# Watching and caching

`cs2-battlelog watch` is designed for a long-running container pointed at a
folder that receives demos from a game server, SFTP sync, or manual copy.

## Lifecycle

1. Load `.cs2battlelog/manifest.json`.
2. Start workers (default 1).
3. Start fsnotify watchers recursively; if unavailable, log and rely on polling.
4. Scan immediately, then every `--interval` and after debounced fsnotify events.
5. On SIGINT/SIGTERM: stop watching, drain the queue, save the manifest.

## Settle logic

A demo must have the same size and mtime on two consecutive scans before it is
queued. This handles downloads/copies still in progress. With the defaults
(5 s interval + 2 s debounce) a new file is normally processed within ~10 s.

## Cache / skip

`manifest.json` stores per absolute path: size, mtime, SHA-256, report paths,
tool version, schema version, status, error, timestamp.

Skip when **all** match:

- entry status is `ok`,
- size and mtime are unchanged,
- tool version and schema version match,
- the HTML report still exists.

Broken demos are cached as `error` and skipped until the file changes or
`--force` is used. This prevents a corrupt file from being reparsed every scan.

`--force` bypasses both successful and failed cache entries.

## Manifest location

| Configuration | Manifest directory |
|---|---|
| `--output DIR` | `DIR` |
| no output, first input is a directory | first input directory |
| no output, input is a file | current working directory |

Reports go to `<output>/<base>.html`, or next to the demo when no
output is set. Basename collisions in a shared output directory get a short
source hash.

## Docker

```yaml
services:
  cs2-battlelog:
    build: .
    volumes:
      - /srv/cs2/demos:/demos:ro
      - /srv/cs2/reports:/out
    command: ["watch", "--input", "/demos", "--output", "/out", "--workers", "1"]
    restart: unless-stopped
```

- The container runs as root by default so bind-mounted output directories work
  without UID mapping. Add `user: "1000:1000"` if you prefer.
- Mount demos read-only.
- Keep `--workers` low: each worker holds two parser states (battlelog + cheat).
- Memory guidance: assume 0.5–1 GB peak per concurrent worker for large GOTV
  demos.

## Debugging

- `--verbose` logs queue decisions and parse timings.
- `--interval 2s --debounce 1s` speeds up feedback while testing.
- Delete `.cs2battlelog/manifest.json` to force a full rescan.
- A file that never gets queued: check extension, hidden directory, stability,
  and `--recursive`.
