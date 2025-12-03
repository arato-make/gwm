#!/usr/bin/env bash
set -euo pipefail

# Usage: scripts/make_release_tarball.sh v0.1.0 [output-dir]
# 環境変数 GOOS/GOARCH を指定するとクロスコンパイルできます。
# 例: GOOS=linux GOARCH=amd64 scripts/make_release_tarball.sh v0.1.0

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
VERSION="${1:-}"
OUT_DIR="${2:-$ROOT_DIR/dist}"

if [[ -z "$VERSION" ]]; then
  echo "usage: $0 <version> [output-dir]" >&2
  exit 1
fi

GOOS_CURRENT="${GOOS:-$(go env GOOS)}"
GOARCH_CURRENT="${GOARCH:-$(go env GOARCH)}"

BIN_NAME="gwm"
PKG_NAME="${BIN_NAME}_${VERSION}_${GOOS_CURRENT}_${GOARCH_CURRENT}"
STAGE_DIR="$OUT_DIR/$PKG_NAME"
TARBALL="$OUT_DIR/${PKG_NAME}.tar.gz"

mkdir -p "$STAGE_DIR"

# ビルド（静的リンク、再現性のため CGO 無効）
CGO_ENABLED=0 GOOS="$GOOS_CURRENT" GOARCH="$GOARCH_CURRENT" \
  go build -o "$STAGE_DIR/$BIN_NAME" "$ROOT_DIR/cmd/gwm"

# 同梱ファイル
cp "$ROOT_DIR/README.md" "$STAGE_DIR/"
if [[ -f "$ROOT_DIR/LICENSE" ]]; then
  cp "$ROOT_DIR/LICENSE" "$STAGE_DIR/"
fi

# 圧縮
mkdir -p "$OUT_DIR"
tar -C "$OUT_DIR" -czf "$TARBALL" "$PKG_NAME"

# 後片付け
rm -rf "$STAGE_DIR"

echo "created: $TARBALL"
