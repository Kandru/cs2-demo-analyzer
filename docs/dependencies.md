# Dependency audit

This page records the Go modules and build inputs the binary actually consumes,
why each one is present, and the process for keeping the set small.

Run the audit locally with:

```sh
make tidy-check   # go.mod/go.sum must match the source
make audit        # go mod verify + govulncheck
```

Both targets run in the pinned Docker toolchain. CI runs them on every pull
request.

## Compiled modules

`go list -m all` includes test and example dependencies of demoinfocs. Only the
modules below are compiled into the binary. The list is current as of
**2026-11-16**.

| Module | Version | License | Why it is compiled | Reachable here |
|---|---|---|---|---|
| `github.com/markus-wa/demoinfocs-golang/v5` | v5.3.0 | MIT | CS2 demo parsing | Yes |
| `github.com/fsnotify/fsnotify` | v1.10.1 | BSD-3-Clause | watch-mode acceleration | Yes |
| `github.com/oklog/ulid/v2` | v2.1.1 | Apache-2.0 | equipment IDs in demoinfocs; grenade map key | Yes |
| `github.com/golang/geo` | v0.0.0-20230421003525-6adc56603217 | Apache-2.0 | vector types in demoinfocs' public API | Yes |
| `github.com/golang/snappy` | v1.0.0 | BSD-3-Clause | compressed demo string tables | Yes |
| `github.com/markus-wa/gobitread` | v0.2.5-0.20241202000432-3c3e0bc797c6 | MIT | demoinfocs bit reader | Yes |
| `github.com/markus-wa/godispatch` | v1.4.1 | MIT | demoinfocs event dispatch | Yes |
| `github.com/markus-wa/go-unassert` | v0.1.3 | MIT | demoinfocs internal assertions | Yes |
| `github.com/markus-wa/ice-cipher-go` | v0.0.0-20230901094113-348096939ba7 | MIT | encrypted net messages in demoinfocs | No — we never configure a decryption key |
| `github.com/markus-wa/quickhull-go/v2` | v2.2.0 | MIT | `Fires.ConvexHull3D` in demoinfocs | No — method is never called |
| `github.com/pkg/errors` | v0.9.1 | BSD-2-Clause | error wrapping in demoinfocs | Yes |
| `golang.org/x/exp` | v0.0.0-20230817173708-d852ddb80c63 | BSD-3-Clause | one `maps.Keys` call in demoinfocs | Yes |
| `golang.org/x/sys` | v0.50.0 | BSD-3-Clause | Unix backend of fsnotify | Yes |
| `google.golang.org/protobuf` | v1.36.11 | BSD-3-Clause | demo wire format | Yes |

A module being compiled does not mean its code is reachable. `govulncheck`
reports reachable vulnerabilities separately from module advisories.

### Why we do not remove more

- **demoinfocs is required.** There is no Valve-official Go demo parser. Writing
  one would mean reimplementing Source 2 demos, protobuf schemas, string tables,
  send tables, and encryption.
- **`ulid` cannot be dropped from the build.** demoinfocs returns `ulid.ULID`
  from `Equipment.UniqueID2()`. Removing our direct map-key use would not remove
  the module.
- **`geo`, `snappy`, `gobitread`, `godispatch`, `go-unassert`, and `protobuf`**
  are part of demoinfocs' parser core or public API.
- **`x/sys`** only exists because of fsnotify. We keep fsnotify for watch-mode
  latency and rely on the polling fallback only when it is unavailable.

### Dead code in demoinfocs

`quickhull-go` and `ice-cipher-go` are compiled but unreachable in this project:

- `Fires.ConvexHull3D` is a public demoinfocs helper we never call.
- ICE decryption only runs when `ParserConfig.NetMessageDecryptionKey` is set.
  We never set it, so encrypted messages produce a parser warning instead.

Removing them requires an upstream change or a fork. We deliberately do **not**
fork: a private fork would drift from upstream parser and security fixes.

## Upstream proposals

These changes are prepared for demoinfocs upstream. They are not applied
locally until they ship in a release.

### 1. Replace `golang.org/x/exp/maps`

`pkg/demoinfocs/sendtables/sendtablescs2/entity.go` uses
`maps.Keys` from `golang.org/x/exp`. Since demoinfocs targets Go 1.24+, the
standard library can provide the same result:

```go
props := slices.Sorted(maps.Keys(e.updateHandlers))
```

The `golang.org/x/exp` requirement can then be dropped.

### 2. Replace `github.com/pkg/errors`

Nine demoinfocs files use `pkg/errors` (6 `New`, 11 `Wrap`, 1 `Wrapf`, 3 `Is`).
These map directly to `fmt.Errorf` with `%w` and the standard `errors` package.
Replacing them removes a stale dependency. The largest observable difference is
the loss of `pkg/errors` stack traces, which demoinfocs does not expose.

### 3. Make `quickhull` optional

`Fires.ConvexHull3D` is the only use of `quickhull-go`. Upstream should either
move it behind a separate helper package or drop it from the core `common`
package.

### 4. Keep `ice-cipher-go`

The dependency is required for demos with encrypted net messages. It is not
dead from upstream's perspective, even if this project does not enable it.

## Non-module build inputs

- **Go toolchain:** `go.mod` declares `toolchain go1.27.4`. `GOTOOLCHAIN=local`
  in Docker prevents implicit toolchain downloads; the image itself is pinned by
  digest.
- **Build image:** `golang:1.27.4-alpine` pinned by digest in both the
  `Dockerfile` and the Makefile.
- **Runtime image:** `alpine:3.24.4` pinned by digest.
- **Dockerfile frontend:** `docker/dockerfile:1` pinned by digest.
- **GitHub Actions:** pinned to commit SHAs. Dependabot proposes updates for Go
  modules, Docker images, and Actions.
- **CI:** `.github/workflows/ci.yaml` runs tests, vet, module tidiness, module
  verification, and `govulncheck`.

## Adding a dependency

1. Prefer the standard library, then an official Go project module.
2. Explain why the dependency is required in the pull request.
3. Add it to this page and to `docs/THIRD_PARTY.md`.
4. Run `make tidy-check && make audit`.
5. Do not vendor a private fork of demoinfocs without a documented rebase and
   security-update policy.
