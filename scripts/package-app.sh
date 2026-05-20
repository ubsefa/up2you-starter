#!/usr/bin/env sh
set -eu

APP_DIR="${1:-examples/my-todo}"

if [ ! -f "$APP_DIR/app.yaml" ]; then
  echo "app.yaml not found in $APP_DIR" >&2
  exit 1
fi

APP_NAME="$(basename "$APP_DIR")"
ROOT_DIR="$(pwd)"
OUT_DIR="$ROOT_DIR/dist"
OUT_FILE="$OUT_DIR/$APP_NAME.zip"

mkdir -p "$OUT_DIR"
rm -f "$OUT_FILE"

(cd "$APP_DIR" && zip -qr "$OUT_FILE" .)

echo "$OUT_FILE"
