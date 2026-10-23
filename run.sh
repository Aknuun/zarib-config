#!/usr/bin/env bash
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"

if [ -x "$HERE/zarib-config" ]; then
    exec "$HERE/zarib-config" "$@"
fi

if ! command -v go >/dev/null 2>&1; then
    echo "Go 1.22 or newer is required." >&2
    exit 1
fi

if [ -f "$HERE/go.mod" ]; then
    cd "$HERE"
    exec go run ./cmd/zarib-config "$@"
fi

exec go run github.com/Aknuun/zarib-config/cmd/zarib-config@latest "$@"
