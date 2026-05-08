#!/usr/bin/env sh
set -eu

REPO="${PIN_REPO:-muhammedikinci/pin}"
VERSION="${PIN_VERSION:-latest}"
BIN_DIR="${BIN_DIR:-/usr/local/bin}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
  linux) asset_os="Linux" ;;
  darwin) asset_os="Darwin" ;;
  *) echo "Unsupported OS: $os" >&2; exit 1 ;;
esac

case "$arch" in
  x86_64|amd64) asset_arch="x86_64" ;;
  arm64|aarch64) asset_arch="arm64" ;;
  *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
esac

asset="pin_${asset_os}_${asset_arch}.tar.gz"
if [ "$VERSION" = "latest" ]; then
  url="https://github.com/${REPO}/releases/latest/download/${asset}"
else
  url="https://github.com/${REPO}/releases/download/${VERSION}/${asset}"
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

echo "Downloading $url"
curl -fsSL "$url" -o "$tmpdir/$asset"
tar -xzf "$tmpdir/$asset" -C "$tmpdir"

if [ -w "$BIN_DIR" ]; then
  install -m 0755 "$tmpdir/pin" "$BIN_DIR/pin"
elif command -v sudo >/dev/null 2>&1; then
  sudo install -m 0755 "$tmpdir/pin" "$BIN_DIR/pin"
else
  echo "$BIN_DIR is not writable and sudo is unavailable" >&2
  exit 1
fi

echo "pin installed to $BIN_DIR/pin"
