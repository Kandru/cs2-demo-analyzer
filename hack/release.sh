#!/bin/sh
# Package release archives for every supported target.
#
# Usage: hack/release.sh <output-root> <build command...>
#
# The build command runs from src/ with GOOS and GOARCH set and must write
# ./cmd/cs2-demo-analyzer to the requested -o path. Both `make release`
# (Docker) and `make release-local` (native Go) use this script so the
# archive layout stays identical.
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
out_root=$1
shift
[ "$#" -gt 0 ] || { echo "usage: $0 <output-root> <build command...>" >&2; exit 2; }

version=$(tr -d '[:space:]' < "$root/VERSION")
binary=cs2-demo-analyzer
targets="linux/amd64 linux/arm64 linux/386 windows/amd64 windows/arm64 darwin/amd64 darwin/arm64"

cd "$root/src"
for target in $targets; do
  os=${target%/*}
  arch=${target#*/}
  ext=
  [ "$os" = windows ] && ext=.exe
  name="$binary-$version-$os-$arch"
  printf 'building %s...\n' "$name"
  mkdir -p "$out_root/$name"
  GOOS=$os GOARCH=$arch "$@" -o "$out_root/$name/$binary$ext" ./cmd/$binary || exit 1
  cp "$root/README.md" "$root/VERSION" "$out_root/$name/" 2>/dev/null || true
  ( cd "$out_root" && tar -czf "$name.tar.gz" "$name" && rm -rf "$name" ) || exit 1
done

cd "$out_root"
sha256sum *.tar.gz > SHA256SUMS
printf '\nRelease %s artifacts:\n' "$version"
ls -lh *.tar.gz SHA256SUMS
