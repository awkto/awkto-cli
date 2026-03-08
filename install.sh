#!/bin/bash
set -euo pipefail

REPO="awkto/awkto-cli"
INSTALL_DIR="/usr/local/bin"
BINARY="awkto"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  linux|darwin) ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

ASSET="awkto-${OS}-${ARCH}"

# Get latest version tag from GitHub API
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases?per_page=10" \
  | grep -o '"tag_name": *"v[^"]*"' | head -1 | grep -o 'v[^"]*')

if [ -z "$VERSION" ]; then
  echo "Failed to determine latest version"
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

echo "Downloading ${ASSET} ${VERSION}..."
TMP=$(mktemp)
if ! curl -fSL "$URL" -o "$TMP"; then
  echo "Failed to download from $URL"
  rm -f "$TMP"
  exit 1
fi

chmod +x "$TMP"

echo "Installing to ${INSTALL_DIR}/${BINARY}..."
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "${INSTALL_DIR}/${BINARY}"
else
  sudo mv "$TMP" "${INSTALL_DIR}/${BINARY}"
fi

echo "awkto installed successfully! Run 'awkto help' to get started."
