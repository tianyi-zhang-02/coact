#!/bin/sh
# coact installer — downloads a prebuilt single binary for your platform.
# Usage: curl -fsSL https://raw.githubusercontent.com/tianyi-zhang-02/coact/main/install.sh | sh
set -eu

REPO="tianyi-zhang-02/coact"
BINARY="coact"
INSTALL_DIR="${COACT_INSTALL_DIR:-/usr/local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

case "$os" in
  linux) os=linux ;;
  darwin) os=darwin ;;
  *) echo "coact: unsupported OS '$os' (use go install on this platform)" >&2; exit 1 ;;
esac

case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "coact: unsupported arch '$arch'" >&2; exit 1 ;;
esac

version="${COACT_VERSION:-latest}"
if [ "$version" = "latest" ]; then
  version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | head -1 | cut -d'"' -f4)
fi
if [ -z "$version" ]; then
  echo "coact: could not resolve latest version" >&2; exit 1
fi

tarball="${BINARY}_${version#v}_${os}_${arch}.tar.gz"
url="https://github.com/${REPO}/releases/download/${version}/${tarball}"
checksums_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "coact: downloading ${version} for ${os}/${arch}"
curl -fsSL "$url" -o "$tmp/$tarball"
curl -fsSL "$checksums_url" -o "$tmp/checksums.txt"

expected=$(
  awk -v file="$tarball" '$2 == file || $2 == "./" file { print $1; exit }' "$tmp/checksums.txt"
)
if [ -z "$expected" ]; then
  echo "coact: checksum for $tarball not found" >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$tmp/$tarball" | awk '{print $1}')
else
  actual=$(shasum -a 256 "$tmp/$tarball" | awk '{print $1}')
fi
if [ "$actual" != "$expected" ]; then
  echo "coact: checksum mismatch for $tarball" >&2
  exit 1
fi

member=""
for candidate in "$BINARY" "./$BINARY"; do
  if tar -tzf "$tmp/$tarball" "$candidate" >/dev/null 2>&1; then
    member="$candidate"
    break
  fi
done
if [ -z "$member" ]; then
  echo "coact: archive does not contain $BINARY" >&2
  exit 1
fi
tar -xOf "$tmp/$tarball" "$member" > "$tmp/$BINARY"
chmod +x "$tmp/$BINARY"

if [ -w "$INSTALL_DIR" ]; then
  mv "$tmp/$BINARY" "$INSTALL_DIR/$BINARY"
else
  echo "coact: installing to $INSTALL_DIR (needs sudo)"
  sudo mv "$tmp/$BINARY" "$INSTALL_DIR/$BINARY"
fi
chmod +x "$INSTALL_DIR/$BINARY"

echo "coact: installed to $INSTALL_DIR/$BINARY"
"$INSTALL_DIR/$BINARY" --version || true
