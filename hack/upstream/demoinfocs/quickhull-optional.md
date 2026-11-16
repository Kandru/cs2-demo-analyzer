# Proposal: make `quickhull-go` optional

## Problem

`pkg/demoinfocs/common/inferno.go` imports
`github.com/markus-wa/quickhull-go/v2` for a single public helper:

```go
func (f Fires) ConvexHull3D() quickhull.ConvexHull
```

Consumers that only parse events and player state still compile the convex-hull
implementation and its dependency into their binaries.

## Proposal

Pick one:

1. Move `ConvexHull3D` (and the private `convexHull` helper) to a separate
   helper package that consumers import explicitly, or
2. Remove `ConvexHull3D` in the next major version and document
   `ConvexHull2D` as the supported helper.

`common.Fires` already returns `[]Fire` through `List()`, so consumers that need
3D hulls can compute them from the points outside the parser.

## Compatibility

- Option 1 keeps the API available but changes its import path; it can ship in a
  minor release with a deprecation note on the old method.
- Option 2 is a breaking change and belongs in the next major release.

This project uses neither helper today, so either option removes the dependency
without affecting its reports.
