#!/bin/sh
# Trellis setup CLI installer.
#
#   curl -fsSL https://raw.githubusercontent.com/kodhama/trellis/main/install.sh | sh
#
# No package manager: this downloads a single static binary from GitHub Releases
# (decision-0023). Set TRELLIS_VERSION to pin a release; defaults to the latest.
# Set TRELLIS_INSTALL_DIR to choose where the binary lands (default: ~/.local/bin).
#
# Uninstall:  curl -fsSL .../install.sh | sh -s -- --uninstall
set -eu

REPO="kodhama/trellis"
BIN="trellis"
VERSION="${TRELLIS_VERSION:-latest}"
dir="${TRELLIS_INSTALL_DIR:-$HOME/.local/bin}"

# --uninstall: remove the binary and exit (mirror of `trellis uninstall`).
if [ "${1:-}" = "--uninstall" ]; then
  if [ -f "$dir/$BIN" ]; then
    rm -f "$dir/$BIN" && echo "trellis: removed $dir/$BIN"
  else
    echo "trellis: nothing to remove at $dir/$BIN"
  fi
  exit 0
fi

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *) echo "trellis: unsupported architecture: $arch" >&2; exit 1 ;;
esac
case "$os" in
  linux | darwin) ;;
  *) echo "trellis: unsupported OS: $os (build from ./cli with Go, or use git-copy)" >&2; exit 1 ;;
esac

asset="${BIN}_${os}_${arch}"
if [ "$VERSION" = "latest" ]; then
  url="https://github.com/${REPO}/releases/latest/download/${asset}"
else
  url="https://github.com/${REPO}/releases/download/${VERSION}/${asset}"
fi

mkdir -p "$dir"

echo "trellis: downloading ${asset} (${VERSION}) …"
if ! curl -fsSL "$url" -o "$dir/$BIN"; then
  echo "trellis: download failed from $url" >&2
  echo "  (Releases may not be published yet — until then, build from source:" >&2
  echo "     git clone https://github.com/${REPO} && cd trellis/cli && go build -o $dir/$BIN .)" >&2
  exit 1
fi
chmod +x "$dir/$BIN"

echo "trellis: installed to $dir/$BIN"
case ":$PATH:" in
  *":$dir:"*) : ;;
  *) echo "trellis: add $dir to your PATH, then run: trellis setup" ;;
esac
