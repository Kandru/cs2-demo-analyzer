# AGENTS.md — internal/watch

Folder watcher and queue.

## Invariants

- fsnotify is an accelerator; periodic rescan is the reliability guarantee
  (network mounts).
- A file is queued only after two identical size/mtime scans (settle window).
- `Processor.Skip` decides cache hits; the watcher must not duplicate that
  logic.
- Queue entries are deduped (`queued` map) and removed after processing; a full
  queue retries on the next scan.
- One bad demo never stops the watcher. SIGINT/SIGTERM drains workers and saves
  the manifest.
- Skip hidden dirs and `.cs2-demo-analyzer`; optionally recurse.
- Keep the `Processor` interface small (`Process`, `Skip`, `Save`) so it can be
  faked in tests.

## Tests

`watch_test.go` checks settle behavior. Demo extension checks live in
`internal/demofile`. Use fake processors; never start fsnotify in tests.
