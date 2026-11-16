#!/bin/sh
# CS2 Demo Analyzer installer for Linux.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Kandru/cs2-demo-analyzer/main/install.sh | sudo sh
#   wget -qO- https://raw.githubusercontent.com/Kandru/cs2-demo-analyzer/main/install.sh | sudo sh
#
# Optional environment variables:
#   VERSION=<tag>       install a specific release instead of the latest
#   INSTALL_DIR=<dir>   binary directory (default: /opt/cs2-demo-analyzer)
#   BIN_DIR=<dir>       symlink directory (default: /usr/local/bin)
set -eu

REPO="Kandru/cs2-demo-analyzer"
BINARY="cs2-demo-analyzer"
INSTALL_DIR="${INSTALL_DIR:-/opt/cs2-demo-analyzer}"
BIN_DIR="${BIN_DIR:-/usr/local/bin}"
VERSION="${VERSION:-}"

say() { printf '%s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

[ "$(uname -s)" = "Linux" ] || die "install.sh supports Linux only"

if [ "$(id -u)" -ne 0 ]; then
	die "root is required for $INSTALL_DIR; run the installer with sudo"
fi

for tool in tar sha256sum awk sed grep head mktemp cp chmod mkdir mv ln; do
	command -v "$tool" >/dev/null 2>&1 || die "required tool not found: $tool"
done

if command -v curl >/dev/null 2>&1; then
	fetch() { curl -fsSL --retry 3 --connect-timeout 10 "$1"; }
	fetch_to() { curl -fsSL --retry 3 --connect-timeout 10 -o "$2" "$1"; }
elif command -v wget >/dev/null 2>&1; then
	fetch() { wget -qO- "$1"; }
	fetch_to() { wget -qO "$2" "$1"; }
else
	die "curl or wget is required"
fi

case "$(uname -m)" in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	i386 | i486 | i586 | i686 | x86) arch=386 ;;
	*) die "unsupported architecture: $(uname -m)" ;;
esac

if [ -z "$VERSION" ]; then
	say "resolving latest release..."
	VERSION="$(fetch "https://api.github.com/repos/$REPO/releases/latest" |
		sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' |
		head -n 1)"
	[ -n "$VERSION" ] || die "could not determine the latest release; set VERSION=<tag> to pin one"
fi
VERSION="${VERSION#v}"
printf '%s' "$VERSION" | grep -Eq '^[0-9]{2}\.[0-9]{2}\.[1-9][0-9]*$' ||
	die "invalid release version: $VERSION"

asset="$BINARY-$VERSION-linux-$arch.tar.gz"
base="https://github.com/$REPO/releases/download/$VERSION"
tmp="$(mktemp -d "${TMPDIR:-/tmp}/cs2-demo-analyzer.XXXXXX")" || die "could not create a temporary directory"
tmpbin=""
cleanup() {
	[ -z "$tmpbin" ] || rm -f "$tmpbin"
	rm -rf "$tmp"
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM

say "downloading $asset"
fetch_to "$base/$asset" "$tmp/$asset" || die "download failed: $base/$asset"
fetch_to "$base/SHA256SUMS" "$tmp/SHA256SUMS" || die "download failed: $base/SHA256SUMS"

expected="$(awk -v name="$asset" 'NF >= 2 { n = $2; sub(/^\*/, "", n); if (n == name) { print $1; exit } }' "$tmp/SHA256SUMS")"
[ -n "$expected" ] || die "SHA256SUMS does not list $asset"
actual="$(sha256sum "$tmp/$asset" | awk '{print $1}')"
[ "$expected" = "$actual" ] || die "checksum verification failed for $asset"

tar -xzf "$tmp/$asset" -C "$tmp" || die "could not extract $asset"
src="$tmp/${asset%.tar.gz}/$BINARY"
[ -f "$src" ] || die "archive does not contain $BINARY at the expected path"

mkdir -p "$INSTALL_DIR" "$BIN_DIR"
tmpbin="$INSTALL_DIR/.$BINARY.new.$$"
cp "$src" "$tmpbin"
chmod 0755 "$tmpbin"
mv -f "$tmpbin" "$INSTALL_DIR/$BINARY"
ln -sfn "$INSTALL_DIR/$BINARY" "$BIN_DIR/$BINARY"

say "installed $BINARY $VERSION to $INSTALL_DIR/$BINARY"
say "linked $BIN_DIR/$BINARY"
"$INSTALL_DIR/$BINARY" version || true
say ""
say "Run 'cs2-demo-analyzer help'. Update later with 'sudo cs2-demo-analyzer update'."
