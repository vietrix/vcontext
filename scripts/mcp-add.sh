#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: $0 <codex|claude> [--db path]" >&2
  exit 1
fi

CLIENT="$1"
shift

DB_PATH=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --db)
      DB_PATH="${2:-}"
      shift 2
      ;;
    *)
      echo "unknown arg: $1" >&2
      exit 1
      ;;
  esac
done

VCMD="$(command -v vcontext || true)"
if [[ -z "$VCMD" ]]; then
  echo "vcontext not found in PATH. Install first (scripts/install.sh) or add it to PATH." >&2
  exit 1
fi

ARGS=()
if [[ -n "$DB_PATH" ]]; then
  ARGS+=("-db" "$DB_PATH")
fi

case "$CLIENT" in
  codex)
    codex mcp add vcontext -- "$VCMD" "${ARGS[@]}"
    ;;
  claude)
    claude mcp add --transport stdio vcontext -- "$VCMD" "${ARGS[@]}"
    ;;
  *)
    echo "unknown client: $CLIENT (expected codex or claude)" >&2
    exit 1
    ;;
esac
