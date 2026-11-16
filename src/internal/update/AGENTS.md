# AGENTS.md — internal/update

GitHub release discovery and atomic self-update for the installed binary.

## Contract

- Repository: `Kandru/cs2-demo-analyzer`.
- Latest release:
  `GET https://api.github.com/repos/Kandru/cs2-demo-analyzer/releases/latest`.
- Asset: `cs2-demo-analyzer-<version>-<os>-<arch>.tar.gz`.
- Checksums: `SHA256SUMS` in the same release, sha256sum format.
- Archive member: `cs2-demo-analyzer` (`.exe` on Windows) under one top-level
  directory. Do not assume the directory name.
- Versions are `YY.MM.BuildNr`; compare numerically, never as strings.

## Invariants

- No network call from this package unless `Check` or `Apply` is called.
- `Check` never writes. `Apply` verifies the archive SHA-256 before extracting.
- Extraction only reads the executable member; never write tar paths directly.
- Replace the *real* executable, resolving symlinks, so an installed
  `/usr/local/bin/cs2-demo-analyzer` symlink keeps pointing at
  `/opt/cs2-demo-analyzer/cs2-demo-analyzer`.
- Write a temp file beside the target, `fsync`, then `rename`. Clean up the temp
  file on every failure.
- Preserve the target's executable mode. Surface permission failures with a
  `sudo`/`install.sh` hint.
- Windows cannot replace a running `.exe`; `Apply` must return a clear manual
  download message.
- Tests must not use the network; use `httptest` and `t.TempDir`.
- Changing the asset naming contract requires matching changes in
  `install.sh`, the release workflow, and `docs/install.md`.
