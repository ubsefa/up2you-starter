#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
EXAMPLES_DIR="${1:-$ROOT_DIR/examples}"

env_value() {
  local key="$1"
  if [ -f "$ROOT_DIR/.env" ]; then
    grep -E "^${key}=" "$ROOT_DIR/.env" | tail -n1 | cut -d= -f2-
  fi
}

REGISTRY="${REGISTRY:-$(env_value REGISTRY)}"
REGISTRY="${REGISTRY:-ubsefa}"
VERSION="${VERSION:-$(env_value VERSION)}"
VERSION="${VERSION:-latest}"
IMAGE="$REGISTRY/core-engine:$VERSION"

pass=0
fail=0
failed=""

for dir in "$EXAMPLES_DIR"/*/; do
  [ -f "${dir}app.yaml" ] || continue
  app="$(basename "$dir")"
  abs="$(cd "$dir" && pwd)"
  if out="$(docker run --rm -v "$abs":/work:ro "$IMAGE" validate /work 2>&1)"; then
    printf '  PASS  %s\n' "$app"
    pass=$((pass + 1))
  else
    printf '  FAIL  %s\n' "$app"
    printf '%s\n' "$out" | sed 's/^/        /'
    fail=$((fail + 1))
    failed="$failed $app"
  fi
done

echo
printf 'validated %d examples with %s: %d passed, %d failed\n' "$((pass + fail))" "$IMAGE" "$pass" "$fail"
if [ "$fail" -ne 0 ]; then
  printf 'failed:%s\n' "$failed"
  exit 1
fi
