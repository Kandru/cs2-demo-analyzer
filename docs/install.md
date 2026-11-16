# Install and update

## Linux install script

The installer downloads the latest GitHub release for the machine's
architecture, verifies `SHA256SUMS`, and installs the static binary under
`/opt/cs2-demo-analyzer`.

```sh
curl -fsSL https://raw.githubusercontent.com/Kandru/cs2-demo-analyzer/main/install.sh | sudo sh
```

`wget` works the same way:

```sh
wget -qO- https://raw.githubusercontent.com/Kandru/cs2-demo-analyzer/main/install.sh | sudo sh
```

The installer requires root because it writes to `/opt` and `/usr/local/bin`.
It supports Linux `amd64`, `arm64`, and `386`.

Installed layout:

```text
/opt/cs2-demo-analyzer/cs2-demo-analyzer   binary
/usr/local/bin/cs2-demo-analyzer          symlink
```

Check the installed version:

```sh
cs2-demo-analyzer version
```

Pin a release instead of installing the latest:

```sh
curl -fsSL https://raw.githubusercontent.com/Kandru/cs2-demo-analyzer/main/install.sh |
  sudo VERSION=26.11.1 sh
```

`INSTALL_DIR` and `BIN_DIR` override the two destination directories:

```sh
curl -fsSL https://raw.githubusercontent.com/Kandru/cs2-demo-analyzer/main/install.sh |
  sudo INSTALL_DIR=/srv/cs2-demo-analyzer BIN_DIR=/usr/bin sh
```

## Self-update

The binary can check GitHub releases and replace itself:

```sh
# report the current and latest versions without writing anything
cs2-demo-analyzer update --check

# update if a newer release exists
cs2-demo-analyzer update

# reinstall the latest release (also works for a dev build)
cs2-demo-analyzer update --force
```

A normal installation is root-owned, so updating it needs `sudo`:

```sh
sudo cs2-demo-analyzer update
```

Updates use the unauthenticated GitHub API. A request is made only when
`update` is run or when `--update` is passed to `analyze`/`watch`; ordinary
runs never contact the network.

The updater:

1. reads the latest release from
   `https://api.github.com/repos/Kandru/cs2-demo-analyzer/releases/latest`;
2. selects `cs2-demo-analyzer-<version>-<os>-<arch>.tar.gz`;
3. downloads `SHA256SUMS` and the archive;
4. verifies the archive checksum;
5. writes a temporary executable in the same directory and atomically renames
   it over the old binary.

Update failures report the underlying error; permission failures name the
target path and suggest running with `sudo` or reinstalling with `install.sh`.

**Trust model:** `SHA256SUMS` proves the archive matches the checksum published
in the same GitHub release. It does not prove who built it. A compromised
GitHub account or release can replace both the archive and its checksum, so
treat the GitHub release channel as the root of trust. Signed releases or
GitHub artifact attestations would remove that assumption; until then, verify
an archive independently if you need stronger provenance.

Windows self-update is not supported because a running `.exe` cannot be
replaced; use the release archive instead. macOS and Linux self-update are
supported.

## Manual upgrade and uninstall

Download an archive from the
[releases page](https://github.com/Kandru/cs2-demo-analyzer/releases), verify it
with `SHA256SUMS`, and replace `/opt/cs2-demo-analyzer/cs2-demo-analyzer`.
The `cs2-demo-analyzer update` command performs the same steps automatically.

To uninstall:

```sh
sudo rm -f /usr/local/bin/cs2-demo-analyzer
sudo rm -rf /opt/cs2-demo-analyzer
```

Reports, demo files, and `.cs2-demo-analyzer/manifest.json` caches are not
touched by the installer or updater.

## Docker

The Docker image is built from this repository and does not use `install.sh`.
Self-update is not meaningful inside the image; update by rebuilding it.
