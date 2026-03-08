#!/usr/bin/env bash
set -euo pipefail

REPO_URL="${REPO_URL:-https://github.com/zz3310969/max-skills.git}"
BRANCH="${BRANCH:-main}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
TMP_DIR="$(mktemp -d)"

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

if ! command -v git >/dev/null 2>&1; then
  echo "error: git is required" >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "error: go is required" >&2
  exit 1
fi

echo "[1/4] cloning $REPO_URL (branch: $BRANCH)"
git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$TMP_DIR/repo"

echo "[2/4] building rw"
cd "$TMP_DIR/repo/cli/rw"
go build -o rw .

echo "[3/4] installing to $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"
cp rw "$INSTALL_DIR/rw"
chmod +x "$INSTALL_DIR/rw"

echo "[4/4] done"
echo "Installed: $INSTALL_DIR/rw"
echo
echo "Next:"
echo "  rw setup --server-url <your-mcp-http-endpoint>"
echo "  # or"
echo "  rw setup --server-cmd 'python /path/to/mcp_server/server.py'"
echo "  rw doctor"
echo "  rw tools"
