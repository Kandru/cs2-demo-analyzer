# Releasing

## Version scheme

`YY.MM.BuildNr`, e.g. `26.11.1`.

- `YY.MM` is the UTC month of the release.
- `BuildNr` starts at `1` and increments for every release in that month.
- It resets to `1` on the first day of the next month.
- `VERSION` at the repository root is the single source of truth. It is
  compiled into the binary via ldflags and shown in the report footer/CLI.

Bump with:

```sh
make version-bump   # 26.11.1 -> 26.11.2, or 26.12.1 after a month change
```

Commit `VERSION` separately:

```sh
git add VERSION
git commit -m "chore(release): $(cat VERSION)"
git push origin main
```

## Local build

```sh
make build      # linux/amd64, dist/cs2-demo-analyzer
make release    # linux/windows/darwin archives + dist/SHA256SUMS
make test       # go test ./... in Docker
```

All targets use `golang:1.27` in Docker; no local Go installation is needed.
`make build-local` is available for contributors who do have Go.

Release artifacts are `.tar.gz` (including the Windows `.exe` inside) plus
`SHA256SUMS`, `README.md`, and `VERSION` in each archive.

## Updater and installer contract

`cs2-demo-analyzer update` and `install.sh` depend on the release layout:

- tag name is the plain `YY.MM.BuildNr` version (an optional `v` prefix is
  accepted by the updater);
- archive names are
  `cs2-demo-analyzer-<version>-<os>-<arch>.tar.gz`;
- the executable inside is `cs2-demo-analyzer` (`cs2-demo-analyzer.exe` on
  Windows) under one top-level directory;
- `SHA256SUMS` is attached to the same release and lists every archive by
  basename.

Do not rename assets or drop `SHA256SUMS` without updating
`src/internal/update` and `install.sh`. The workflow keeps the newest five
releases because the updater only reads `/releases/latest`.

## GitHub Actions

`.github/workflows/release.yaml`:

- Triggers on pushes to `main` that touch `VERSION`, and on manual dispatch.
- Validates the `YY.MM.N` format and that the month is the current UTC month.
- Skips if the tag already exists.
- Runs `make release` (Docker cross-compilation).
- Creates a GitHub release with commit notes since the previous tag and uploads
  all archives plus `SHA256SUMS`.
- Deletes older releases, keeping the newest five.

The workflow needs the default `GITHUB_TOKEN` with `contents: write`; no extra
secrets are required.

## Docker image

```sh
make docker-build
docker run --rm -v "$PWD/demos":/demos:ro -v "$PWD/reports":/out \
  cs2-demo-analyzer:latest watch --input /demos --output /out
```

The runtime image is Alpine with CA certificates and tzdata. The binary is
static (`CGO_ENABLED=0`).
