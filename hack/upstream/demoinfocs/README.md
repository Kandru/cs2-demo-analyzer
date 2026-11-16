# Proposed upstream demoinfocs patches

These patches target `github.com/markus-wa/demoinfocs-golang/v5` at **v5.3.0**.
They remove `github.com/pkg/errors` and `golang.org/x/exp` from the module
graph. We do **not** apply them locally or maintain a fork; the goal is to send
them upstream and pick them up in a normal release.

| Patch | Change |
|---|---|
| `0001-remove-pkg-errors.patch` | Replace `pkg/errors` with standard `errors` and `fmt.Errorf("%w")` |
| `0002-remove-x-exp.patch` | Replace `x/exp/maps` with standard `maps`/`slices` and `x/exp/constraints` with `cmp` |
| `0001-0002-combined.patch` | Both changes, for a single review branch |

`0001` and `0002` are independent against v5.3.0, but they edit adjacent
`go.mod` lines. Apply them with `git am`/`git apply` to separate branches, or
use the combined patch when testing both together.

## Apply

```sh
git clone https://github.com/markus-wa/demoinfocs-golang
cd demoinfocs-golang
git checkout v5.3.0

# One change per branch/PR:
git checkout -b remove-pkg-errors
git apply /path/to/0001-remove-pkg-errors.patch
go mod tidy
go build ./...

git checkout v5.3.0
git checkout -b remove-x-exp
git apply /path/to/0002-remove-x-exp.patch
go mod tidy
go build ./...
```

For a single branch:

```sh
git apply /path/to/0001-0002-combined.patch
go mod tidy
go build ./...
```

## Verification performed

Against v5.3.0 with Go 1.27.4:

- `go build ./...` passes for each patch and the combined patch.
- `go mod tidy` leaves no `pkg/errors` or `golang.org/x/exp` references.
- The full upstream test suite needs demo fixtures that are not in the module
  cache, so package tests were not used as the acceptance gate.

## Behavior notes

- `pkg/errors.Wrap`/`Wrapf` stack traces are lost. demoinfocs does not expose
  them, and errors still support `errors.Is`/`errors.As` through `%w`.
- `maps.Keys` is wrapped in `slices.Collect` because the standard library
  iterator form differs from the old `x/exp/maps` slice form.
- `cmp.Ordered` is a drop-in replacement for `constraints.Ordered`.

See `docs/dependencies.md` for the audit rationale and the optional
`quickhull-go` proposal.
