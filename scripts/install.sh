#!/usr/bin/env bash
set -euo pipefail

REPO="${REPO:-vietrix/vcontext}"
VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "unsupported arch: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "unsupported os: $OS" >&2; exit 1 ;;
esac

ASSET="vcontext_${OS}_${ARCH}"
if [[ "$OS" == "windows" ]]; then
  ASSET="${ASSET}.exe"
fi

release_url="https://api.github.com/repos/${REPO}/releases/latest"
if [[ "$VERSION" != "latest" ]]; then
  release_url="https://api.github.com/repos/${REPO}/releases/tags/${VERSION}"
fi

export RELEASE_URL="$release_url"
export ASSET_NAME="$ASSET"

download_url="$(python - <<'PY'
import json, os, sys, urllib.request

repo_url = os.environ["RELEASE_URL"]
asset = os.environ["ASSET_NAME"]
with urllib.request.urlopen(repo_url) as resp:
    data = json.load(resp)
for item in data.get("assets", []):
    if item.get("name") == asset:
        print(item.get("browser_download_url", ""))
        sys.exit(0)
print("", end="")
sys.exit(1)
PY
)"

if [[ -z "$download_url" ]]; then
  echo "asset not found for $ASSET" >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR"
dest="$INSTALL_DIR/vcontext"
curl -fsSL "$download_url" -o "$dest"
chmod +x "$dest"

echo "installed $dest"
